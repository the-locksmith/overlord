package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"overlord/models"
	"overlord/scanners/port-scanner"

	. "libs/color"

	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	//"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/proxy"
	"gopkg.in/mgo.v2/bson"
)

///// App ///////////////////////////////////////////////
type Application struct {
	Name      string
	Version   models.Version
	Database  *storm.DB
	Collector colly.Collector
	Config    models.Config
}

func (self Application) SetCollectorLimits(parallelism int, delay time.Duration) {
	self.Collector.Limit(&colly.LimitRule{
		Parallelism: parallelism,
		RandomDelay: delay * time.Second,
	})
}

func (self Application) SetCollectorProxies(socksProxies string) (err error) {
	proxies, err := proxy.RoundRobinProxySwitcher(socksProxies)
	if err == nil {
		self.Collector.SetProxyFunc(proxies)
	}
	return err
}

///// Main ////////////////////////////////////////////
func (self Application) PrintBanner() {
	fmt.Println(Bold(Magenta(fmt.Sprintf("%v: Network Detector (v%v)", self.Name, self.Version.String()))))
	fmt.Println(Gray("=================================="))
}

func LoadJSONConfig(path string) (config app.Config) {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil
	}
	json.Unmarshal(fileData, &config)
	return config
}

func main() {
	config := LoadJSONConfig("./config.json")
	if _, err := os.Stat("~/.local/share/overlord"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("~/.local/share/overlord", 700)
		}
	}
	db, _ := storm.Open("servers.db")
	app := Application{
		Name:     "Overlord",
		Version:  models.Version{0, 1, 0},
		Database: db,
		Config:   models.Config{},
		Collector: *colly.NewCollector(
			//colly.Debugger(&debug.LogDebugger{}),
			colly.IgnoreRobotsTxt(),
			colly.MaxDepth(15),
			colly.CacheDir("./cache/"),
			//colly.DisallowedDomains("facebook.com", "twitter.com"),
			colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0"),
			//colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
			colly.Async(false),
			//colly.URLFilters(
			//	regexp.MustCompile("(.+|^facebook|^twitter)$"),
			//),
		),
	}
	app.PrintBanner()
	defer app.Database.Close()

	app.InitWebUI()

	app.SetCollectorLimits(2, 4)
	///////////////////////////////////////////////////////////////
	// On HTML Load ///////////////////////////////////////////////
	app.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})
	///////////////////////////////////////////////////////////////
	// On Request /////////////////////////////////////////////////
	fmt.Println("Starting Server Scanning...")
	app.Collector.OnRequest(func(r *colly.Request) {
		if len(r.URL.String()) > 0 {
			// Parse URL
			u, _ := url.Parse(r.URL.String())
			// Parse Addresses
			addresses, _ := net.LookupIP(u.Host)
			var stringAddresses []string
			for _, address := range addresses {
				stringAddresses = append(stringAddresses, address.String())
			}
			// Parse Headers
			headersString := fmt.Sprintf("%s", r.Headers)

			var s models.Server
			err := db.One("Host", u.Host, &s)
			if err != nil {
				for _, ip := range addresses {
					ps := portscanner.NewPortScanner(ip.String(), 2*time.Second, 5)
					fmt.Printf("scanning port %d-%d...\n", 1, 65536)
					openTCPPorts := ps.GetOpenedPort(1, 65536)

					var services []models.Service
					for i := 0; i < len(openTCPPorts); i++ {
						port := openTCPPorts[i]
						fmt.Print(" ", port, " [open]")
						fmt.Println("  -->  ", ps.DescribePort(port))
						services = append(services, models.Service{
							Description: ps.DescribePort(port),
							Port:        port,
						})
					}
					fmt.Println("Listing open ports found with port-scanner lib:", openTCPPorts)
					fmt.Println("Listing open services found with port-scanner lib:", services)

					err := app.Database.Save(&models.Server{
						ID:           bson.NewObjectId(),
						IPAddress:    ip,
						Host:         u.Host,
						OpenTCPPorts: openTCPPorts,
						Services:     services,
						CreatedAt:    time.Now(),
					})
					if err != nil {
						fmt.Println("Error: ", err)
					} else {
						fmt.Println("Server successfully saved.")
					}
				}
				err := app.Database.Save(&models.Page{
					ID:        bson.NewObjectId(),
					Path:      u.Path,
					CreatedAt: time.Now(),
					URL:       r.URL.String(),
					Scheme:    u.Scheme,
					Body:      "",
					RawQuery:  u.RawQuery,
					Headers:   headersString[5:(len(headersString) - 1)],
				})
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println("Page successfully saved.")
				}
				err = app.Database.Save(&models.Domain{
					ID:          bson.NewObjectId(),
					CreatedAt:   time.Now(),
					Host:        u.Host,
					IPAddresses: addresses,
					Headers:     headersString[5:(len(headersString) - 1)],
				})
				if err != nil {
					fmt.Println("Error: ", err)
				} else {
					fmt.Println("Domain successfully saved.")
				}
			}
		}
	})
	//app.Collector.OnResponse(func(r *colly.Response) {
	//	fmt.Println(r.Ctx.Get("url"))
	//})
	//// Set error handler
	//app.Collector.OnError(func(r *colly.Response, err error) {
	//	fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	//})
	// Seed
	app.Collector.Visit(app.Config.SeedURL)
}

func (app *Application) InitWebUI() {
	// TODO: Provide a web API that serves results of crawling
	router := gin.Default()
	//gin.SetMode(gin.ReleaseMode)
	v1 := router.Group("/api/v1")
	v1.GET("/pages", app.Pages)
	v1.GET("/domains", app.Domains)
	v1.GET("/servers", app.Servers)

	go router.Run(":8080")
}

func (app *Application) Servers(c *gin.Context) {
	var servers []models.Server
	err := app.Database.All(&servers)
	if err != nil {
		c.JSON(200, gin.H{"error": err})
	} else {
		c.IndentedJSON(200, servers)
	}
}

func (app *Application) Domains(c *gin.Context) {
	var domains []models.Domain
	err := app.Database.All(&domains)
	if err != nil {
		c.JSON(200, gin.H{"error": err})
	} else {
		c.IndentedJSON(200, domains)
	}
}

func (app *Application) Pages(c *gin.Context) {
	var pages []models.Page
	err := app.Database.All(&pages)
	if err != nil {
		c.JSON(200, gin.H{"error": err})
	} else {
		c.IndentedJSON(200, pages)
	}
}
