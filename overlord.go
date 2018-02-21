package main

import (
	//"bytes"
	"fmt"
	//"io/ioutil"
	"net"
	"net/url"
	//"regexp"
	"time"
	//"strings"
	//"net/http"
	//"net/url"

	//"libs/networking/ping"

	"github.com/1lann/cete"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/proxy"
)

///// App ///////////////////////////////////////////////
type App struct {
	Name string
	Version
	Collector *colly.Collector
	Database
}

func (self App) SetCollectorLimits(parallelism int, delay time.Duration) {
	self.Collector.Limit(&colly.LimitRule{
		Parallelism: parallelism,
		RandomDelay: delay * time.Second,
	})
}

///// Version //////////////////////////////////////////
type Version struct {
	Major int
	Minor int
	Patch int
}

func (self Version) String() string {
	return fmt.Sprintf("%v.%v.%v", self.Major, self.Minor, self.Patch)
}

///// Database /////////////////////////////////////////
type Database struct {
	Data cete.DB
	// Cached
	Servers     *cete.Range
	ServerCount int64
}

func InitDB(path string, table string, indices []string) (db cete.DB) {
	// Load Database
	ptr, _ := cete.Open(path)
	db = *(ptr)
	// Initialize Servers DB
	db.NewTable(table)
	// Initialize DB Indices
	for _, index := range indices {
		db.Table(table).NewIndex(index)
	}
	return db
}

func (self App) SetCollectorProxies(socksProxies string) (err error) {
	proxies, err := proxy.RoundRobinProxySwitcher(socksProxies)
	if err == nil {
		self.Collector.SetProxyFunc(proxies)
	}
	return err
}

func (self Database) CacheServers() *cete.Range {
	// Cache Server Data
	self.Servers = self.Data.Table("servers").All()
	self.ServerCount, _ = self.Servers.Count()
	return self.Servers
}

func (self Database) PrintServers() {
	fmt.Println("  Server Count: ", self.ServerCount)
	fmt.Println("=================================================")
	self.Servers.Do(func(k string, c uint64, d cete.Document) error {
		var s Server
		d.Decode(&s)
		fmt.Println("-------------")
		fmt.Println("Key:         ", k)
		fmt.Println("Counter:     ", c)
		fmt.Println("-------------")
		fmt.Println("URL:         ", s.URL)
		fmt.Println("User-Agent:  ", s.UserAgent)
		fmt.Println("Headers:     ", s.Headers)
		fmt.Println("-------------")
		return nil
	}, 1)
}

func (self Database) InsertServer(s Server) error {
	return self.Data.Table("servers").Set(s.Host, Server{})
}

///// Server ///////////////////////////////////////////
type Server struct {
	CrawledAt   time.Time
	URL         string
	Scheme      string
	Host        string
	Path        string
	RawQuery    string
	Headers     string
	Body        string
	UserAgent   string
	IPAddress   string
	IPAddresses []string
	// Use other tools to get these
	//OpenPorts      []int
	//WebApplication string
	//Banner         string
}

///// Main ////////////////////////////////////////////
func main() {
	app := App{
		Name: "Overlord",
		Database: Database{
			Data: InitDB("./servers_db/", "servers", []string{"Host", "IPAddress"}),
		},
		Version: Version{0, 1, 0},
		Collector: colly.NewCollector(
			colly.Debugger(&debug.LogDebugger{}),
			colly.IgnoreRobotsTxt(),
			colly.MaxDepth(15),
			colly.CacheDir("./_overload_cache/"),
			colly.DisallowedDomains("facebook.com", "twitter.com"),
			colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0"),
			//colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
			//colly.Async(true),
			//colly.URLFilters(
			//	regexp.MustCompile("(.+|^facebook|^twitter)$"),
			//),
		),
	}
	defer app.Database.Data.Close()
	app.SetCollectorLimits(2, 4)
	///////////////////////////////////////////////////////////////
	// On HTML Load ///////////////////////////////////////////////
	app.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println(link)
		e.Request.Visit(link)
	})
	///////////////////////////////////////////////////////////////
	// On Request /////////////////////////////////////////////////
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
			// Add Server
			app.Database.Data.Table("servers").Set(u.Host, Server{
				Host:        u.Host,
				URL:         r.URL.String(),
				Scheme:      u.Scheme,
				Path:        u.Path,
				RawQuery:    u.RawQuery,
				IPAddress:   addresses[0].String(),
				IPAddresses: stringAddresses,
				Headers:     headersString[5:(len(headersString) - 1)],
				UserAgent:   r.Headers.Get("User-Agent"),
				CrawledAt:   time.Now(),
			})
			fmt.Println("=================================================")
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
	app.Collector.Visit("https://coinmarketcap.com/currencies/volume/24-hour/")
}
