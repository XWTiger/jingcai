package creeper

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	ilog "jingcai/log"
	"time"
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
		tan.childCreeper(realUrl, content)
		contentList = append(contentList, *content)
		time.Sleep(500 * time.Microsecond)

		bytes, _ := json.Marshal(content)
		log.Info(string(bytes))
		//c.Visit(e.Request.AbsoluteURL(url))
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

	c.OnHTML(".qiuba_Info", func(e *colly.HTMLElement) {
		time := e.ChildText(".info-box > .relatmatch > .time >.blue")
		theme := e.ChildText(".info-box > .relatmatch > .time")
		homename := e.ChildText(".info-box > .match > .homename")
		guestName := e.ChildText(".info-box > .match > .guestname")
		winer := e.ChildText(".info-box > .match   .on")
		var condition = make([]string, 0)
		condition1 := e.ChildText(".info-box > .match  .Nbg > span:first-child")
		condition2 := e.ChildText(".info-box > .match  .Nbg > span:last-child")
		condition = append(condition, condition1, condition2)
		contxt := e.ChildText("#openContentData")
		content.Content = contxt
		content.Match = fmt.Sprintf("%s vs %s", homename, guestName)
		content.time = time
		content.league = theme
		content.Predict = winer
		content.Conditions = condition
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
