package creeper

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	ilog "jingcai/log"
	"time"
)

type Leisu struct {
}

var leiVisited = false
var leiLog = ilog.Logger
var leiBaseUrl = "https://www.leisu.com/guide/"

func (Lei Leisu) Creep() []Content {
	contentList := make([]Content, 0)
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.leisu.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML(".guide-match > .guide-match-date", func(e *colly.HTMLElement) {

		mTime := e.ChildText(".match-date")
		e.ForEach("div[guide-match-list]", func(i int, element *colly.HTMLElement) {

			timeTail := element.ChildText(".match-time-vip > .time")
			league := element.ChildText(".match-comp > .comp-name")
			homename := element.ChildText(".match-home > .team-name-ranking > .name")
			homeRanking := element.ChildText(".match-home > .team-name-ranking > .ranking")
			round := element.ChildText(".match-round")
			guestName := element.ChildText(".match-away > .team-name-ranking > .name")
			guestRanking := element.ChildText(".match-away > .team-name-ranking > .ranking")

			childUrl := element.ChildAttr(".match-live-news > a", "href")

			if !visited && Lei.checkIfExist(childUrl) {
				visited = true
			}
			content := &Content{
				Type:    "球探",
				Extra:   "",
				Url:     childUrl,
				Title:   fmt.Sprintf("%s vs %s", homename, guestName),
				Summery: fmt.Sprintf("%s(%s)  %s(%s)  %s", homename, homeRanking, guestName, guestRanking, round),
				league:  league,
				time:    fmt.Sprintf("%s %s", mTime, timeTail),
			}
			Lei.childCreeper(childUrl, content)
			contentList = append(contentList, *content)
			time.Sleep(500 * time.Microsecond)

			bytes, _ := json.Marshal(content)
			log.Info(string(bytes))
		})

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

func (Lei *Leisu) Key() string {
	//TODO implement me
	return "www.leisu.com"
}

func (tan Leisu) childCreeper(url string, content *Content) {
	c := colly.NewCollector(
		colly.AllowedDomains("www.leisu.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML(".guide-match-list", func(e *colly.HTMLElement) {
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
func (Lei Leisu) checkIfExist(url string) bool {
	return false
}

func NewInstance() *Leisu {
	return &Leisu{}
}
