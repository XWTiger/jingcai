package creeper

import (
	"bytes"
	"fmt"
	"github.com/gocolly/colly"
	ilog "jingcai/log"
	"jingcai/mysql"
	"time"
)

type Leisu struct {
}

var leiVisited = false
var leiLog = ilog.Logger
var leiBaseUrl = "https://www.leisu.com/guide/"

func (Lei Leisu) Creep() []Content {
	fmt.Println("===================== 雷速开始 ========================")
	contentList := make([]Content, 0)
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.leisu.com"),
		colly.MaxDepth(1),
	)

	c.OnHTML(".guide-match > .guide-match-date", func(e *colly.HTMLElement) {

		mTime := e.ChildText(".match-date")
		e.ForEach("div[class=guide-match-list]", func(i int, element *colly.HTMLElement) {

			timeTail := element.ChildText(".match-time-vip > .time")
			league := element.ChildText(".match-comp > .comp-name")
			homename := element.ChildText(".match-home > .team-name-ranking > .name")
			homeRanking := element.ChildText(".match-home > .team-name-ranking > .ranking")
			round := element.ChildText(".match-round")
			guestName := element.ChildText(".match-away > .team-name-ranking > .name")
			guestRanking := element.ChildText(".match-away > .team-name-ranking > .ranking")

			childUrl := element.ChildAttr(".match-live-news > a:nth-of-type(2)", "href")

			match := fmt.Sprintf("%s vs %s", homename, guestName)
			content := &Content{
				Type:    "雷速",
				Extra:   "",
				Url:     childUrl,
				Title:   match,
				Match:   match,
				Summery: fmt.Sprintf("%s(%s)  %s(%s)  %s", homename, homeRanking, guestName, guestRanking, round),
				league:  league,
				time:    fmt.Sprintf("%s %s", mTime, timeTail),
			}
			content.CreatedAt = time.Now()
			content.UpdatedAt = time.Now()
			fmt.Println("==============", league, "=================")
			time.Sleep(5000 * time.Microsecond)
			if !Lei.checkIfExist(childUrl) {
				Lei.childCreeper(childUrl, content)
				contentList = append(contentList, *content)

				//bytes, _ := json.Marshal(content)
				//log.Info(string(bytes))
				time.Sleep(2000 * time.Microsecond)
			} else {
				log.Info("=======>", childUrl, " 已经被爬过了！")
			}

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

	log.Info("url =============>", leiBaseUrl)
	c.Visit(leiBaseUrl)

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

	c.OnHTML(".layout-score-top > .content-max", func(e *colly.HTMLElement) {
		var conditionMaster = ""
		var conditionGuest = ""
		e.ForEach(".thead-bottom > .clearfix-row > .highcharts > .text", func(i int, element *colly.HTMLElement) {

			conditionMasterRate := element.ChildText(".txt")
			conditionName := element.ChildText(".clearfix-row")
			if i == 1 {
				conditionMaster = fmt.Sprintf("主：(%s %s)", conditionName, conditionMasterRate)
			} else {
				conditionGuest = fmt.Sprintf("客：(%s %s)", conditionName, conditionMasterRate)
			}
			fmt.Println(i, "")
		})
		var condition = make([]string, 0)
		content.Extra = fmt.Sprintf("%s   %s", conditionMaster, conditionGuest)
		//conTitle := e.ChildText(".thead-bottom > .clearfix-row > .center > .title")
		e.ForEach(".thead-bottom > .clearfix-row > .center > .bar-list > .children", func(i int, element *colly.HTMLElement) {
			var masterWinScore = ""
			var masterDrawScore = ""
			var masterFailedSore = ""
			var guestWinScore = ""
			var guestDrawScore = ""
			var guestFailedSore = ""
			element.ForEach(".f-s-16 > .float-left > span", func(i int, ele *colly.HTMLElement) {
				if i == 0 {
					masterWinScore = ele.Text
				}
				if i == 1 {
					masterDrawScore = ele.Text
				}
				if i == 2 {
					masterFailedSore = ele.Text
				}
			})
			var name = element.ChildText(".f-s-16 > .f-s-14")
			element.ForEach(".f-s-16 > .float-right > span", func(i int, ele *colly.HTMLElement) {
				if i == 0 {
					guestWinScore = ele.Text
				}
				if i == 1 {
					guestDrawScore = ele.Text
				}
				if i == 2 {
					guestFailedSore = ele.Text
				}
			})
			masterRate := element.ChildText(".bar > .num1")
			guesetRate := element.ChildText(".bar > .num2")
			condition1 := fmt.Sprintf("%s \n 主: (胜率: %s  胜%s, 平%s, 负%s ) \n 客: (胜率: %s 胜%s,平%s,负%s )", name, masterRate, masterWinScore, masterDrawScore, masterFailedSore, guesetRate, guestWinScore, guestDrawScore, guestFailedSore)
			condition = append(condition, condition1)
		})

		condition3 := e.ChildText(".thead-bottom > .clearfix-row > .clearfix-row  > .cols-table")
		condition = append(condition, condition3)
		content.Conditions = condition
	})

	c.OnHTML("#oddvue > .content-max > .clearfix-row > .information", func(e *colly.HTMLElement) {
		name := e.ChildText(".team-item:nth-child(1) > .team > .name")
		title := e.ChildText(".team-item:nth-child(1) > .good > .title")
		buf := new(bytes.Buffer)
		e.ForEach(".team-item:nth-child(1) > .good > .list > li", func(i int, ele *colly.HTMLElement) {
			buf.WriteString(fmt.Sprintf("%d. %s \n", i, ele.Text))
		})
		favora := buf.String() //有利情报
		harmfulTitle := e.ChildText(".team-item:nth-child(1) > .harmful > .title")
		buf2 := new(bytes.Buffer)
		e.ForEach(".team-item:nth-child(1) > .harmful > .list > li", func(i int, ele *colly.HTMLElement) {
			buf2.WriteString(fmt.Sprintf("%d. %s \n", i, ele.Text))
		})
		harmful := buf2.String()
		favoraContent := fmt.Sprintf("%s: \n %s: \n %s: \n %s: \n %s \n", name, title, favora, harmfulTitle, harmful)
		//fmt.Println(favoraContent)

		cname := e.ChildText(".common-item > .middle > .name")
		ctitle := e.ChildText(".common-item > .middle > .title")
		cbuf := new(bytes.Buffer)
		e.ForEach(".common-item > .good > .list > li", func(i int, ele *colly.HTMLElement) {
			cbuf.WriteString(fmt.Sprintf("%d. %s \n", i, ele.Text))
		})
		cfavora := cbuf.String() //有利情报

		cfavoraContent := fmt.Sprintf("%s: \n %s \n %s \n ", cname, ctitle, cfavora)
		//fmt.Println(cfavoraContent)

		gname := e.ChildText(".team-item:nth-last-child(1) > .team > .name")
		gtitle := e.ChildText(".team-item:nth-last-child(1) > .good > .title")
		gbuf := new(bytes.Buffer)
		e.ForEach(".team-item:nth-last-child(1) > .good > .list > li", func(i int, ele *colly.HTMLElement) {
			gbuf.WriteString(fmt.Sprintf("%d. %s \n", i, ele.Text))
		})
		gfavora := gbuf.String() //有利情报
		gharmfulTitle := e.ChildText(".team-item:nth-last-child(1) > .harmful > .title")
		gbuf2 := new(bytes.Buffer)
		e.ForEach(".team-item:nth-last-child(1) > .harmful > .list > li", func(i int, ele *colly.HTMLElement) {
			gbuf2.WriteString(fmt.Sprintf("%d. %s \n", i, ele.Text))
		})
		gharmful := gbuf2.String()
		gfavoraContent := fmt.Sprintf("%s: \n %s: \n %s: \n %s: \n %s \n", gname, gtitle, gfavora, gharmfulTitle, gharmful)
		//fmt.Println(gfavoraContent)
		content.Content = fmt.Sprintf("%s \n %s \n %s \n", favoraContent, cfavoraContent, gfavoraContent)

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
	time.Sleep(500 * time.Microsecond)
	c.Visit(url)
}
func (Lei Leisu) checkIfExist(url string) bool {
	var content Content
	if err := mysql.DB.Model(Content{}).Where("title=?", url).First(&content).Error; err != nil {
		return false
	}
	return true
}

func NewInstance() *Leisu {
	return &Leisu{}
}
