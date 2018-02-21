package main

import (
	"fmt"
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
)

type Server struct {
	URL       string
	Headers   string
	UserAgent string
	// Use other tools to get these
	IpAddress      string
	OpenPorts      []int
	WebApplication string
	Banner         string
	CrawledAt      time.Time
}

func panicNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// Initialize Servers DB
	db, _ := cete.Open("./server_data")
	defer db.Close()
	db.NewTable("servers")
	db.Table("servers").NewIndex("URL")

	// Instantiate default collector
	c := colly.NewCollector(
		colly.IgnoreRobotsTxt(),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0"),
		colly.Debugger(&debug.LogDebugger{}),
		colly.MaxDepth(15),
		colly.Async(true),
		//colly.URLFilters(
		//	regexp.MustCompile("(.+|^facebook|^twitter)$"),
		//),
	)
	//rp, err := proxy.RoundRobinProxySwitcher("socks5://127.0.0.1:1337", "socks5://127.0.0.1:1338")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//c.SetProxyFunc(rp)

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println(link)
		e.Request.Visit(link)
	})

	c.OnRequest(func(r *colly.Request) {
		headersString := fmt.Sprintf("%s", r.Headers)
		db.Table("servers").Set(r.URL.String(), Server{
			URL:       r.URL.String(),
			Headers:   headersString[5:(len(headersString) - 1)],
			UserAgent: r.Headers.Get("User-Agent"),
			CrawledAt: time.Now(),
		})
		fmt.Println("Headers: ", r.Headers)
		fmt.Println("Visiting", r.URL)
		fmt.Println("Body: ", r.Body)
		//fmt.Println("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
		fmt.Println("=================================================")
		servers := db.Table("servers").All()
		serverCount, _ := servers.Count()
		fmt.Println("  Server Count: ", serverCount)
		fmt.Println("=================================================")
		db.Table("servers").All().Do(func(key string, counter uint64, d cete.Document) error {
			var s Server
			d.Decode(&s)
			//fmt.Println("d.Key():     ", d.Key())
			//fmt.Println("d.Counter(): ", d.Counter())
			fmt.Println("URL:         ", s.URL)
			fmt.Println("User-Agent:  ", s.UserAgent)
			fmt.Println("Headers:     ", s.Headers)

			u, err := url.Parse(s.URL)
			if err != nil {
				fmt.Println("Error: ", err)
			}
			addresses, err := net.LookupIP(u.Host)
			if err != nil {
				fmt.Println("Error: ", err)
			}
			fmt.Println("Addresses found at URL: ", addresses)
			//fmt.

			return nil
		}, 1)

		fmt.Println("=================================================")
	})

	//c.OnResponse(func(r *colly.Response) {
	//	fmt.Println(r.Ctx.Get("url"))
	//})
	//// Set error handler
	//c.OnError(func(r *colly.Response, err error) {
	//	fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	//})
	// Start scraping market index of coinmarketcap to find Markets
	//
	// Print out saved servers
	//////////////////////////

	// Seed
	c.Visit("https://coinmarketcap.com/currencies/volume/24-hour/")

}
