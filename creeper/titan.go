package creeper

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	ilog "jingcai/log"
)

type Titan struct {
}

var visited = false
var log = ilog.Logger
var baseUrl = "http://ba2.titan007.com/"

func (tan Titan) Creep() []Content {
	contentList := make([]Content, 0)
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("ba2.titan007.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML(".theme-content-box > .theme-content", func(e *colly.HTMLElement) {

		predict := e.ChildText(".time > .dpt")
		log.Info("predict: ", predict)
		title := e.ChildText(".title > a")
		if title == "" || predict == "" {
			return
		}
		url := e.ChildAttr(".title  > a", "href")
		realUrl := fmt.Sprintf("%s%s", baseUrl, url)
		info := e.ChildText(".info > .shareinfo")
		if info == "" {
			return
		}
		if !visited && tan.checkIfExist(realUrl) {
			visited = true
		}
		content := &Content{
			Type:    "球探",
			Extra:   predict,
			Url:     realUrl,
			Title:   title,
			Summery: info,
		}
		contentList = append(contentList, *content)
		bytes, _ := json.Marshal(content)
		log.Info(string(bytes))

		//c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Response %s: %d bytes\n", r.Request.URL, len(r.Body))
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error %s: %v\n", r.Request.URL, err)
	})
	c.Visit(baseUrl)

	return contentList
}

func (tan Titan) checkIfExist(url string) bool {
	return false
}

func (tan Titan) childCreeper(url string, content *Content) {
	c := colly.NewCollector(
		colly.AllowedDomains("ba2.titan007.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML(".theme-content", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		//c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Response %s: %d bytes\n", r.Request.URL, len(r.Body))
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error %s: %v\n", r.Request.URL, err)
	})
	c.Visit(url)
}

func (tan Titan) Key() string {
	return "ba2.titan007.com"
}

func NewTianInstance() *Titan {
	return &Titan{}
}
