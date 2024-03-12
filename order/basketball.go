package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"io"
	"jingcai/cache"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/util"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BaseketBallResult struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Value        struct {
		LastUpdateTime string `json:"lastUpdateTime"`
		LeagueList     []struct {
			LeagueAbbName string `json:"leagueAbbName"`
			LeagueId      int    `json:"leagueId"`
			LeagueAllName string `json:"leagueAllName"`
		} `json:"leagueList"`
		MatchResult []struct {
			AllAwayTeam string `json:"allAwayTeam"`
			AllHomeTeam string `json:"allHomeTeam"`
			AwayTeam    string `json:"awayTeam"`
			AwayTeamId  int    `json:"awayTeamId"`
			FinalScore  string `json:"finalScore"`
			Hdc         struct {
				Combination     string `json:"combination"`
				CombinationDesc string `json:"combinationDesc"`
				GoalLine        string `json:"goalLine"`
				ResultStatus    string `json:"resultStatus"`
				Single          int    `json:"single"`
				WinOdds         string `json:"winOdds"`
			} `json:"hdc"`
			Hilo struct {
				Combination     string `json:"combination"`
				CombinationDesc string `json:"combinationDesc"`
				GoalLine        string `json:"goalLine"`
				ResultStatus    string `json:"resultStatus"`
				Single          int    `json:"single"`
				WinOdds         string `json:"winOdds"`
			} `json:"hilo"`
			HomeTeam        string `json:"homeTeam"`
			HomeTeamId      int    `json:"homeTeamId"`
			LeagueBackColor string `json:"leagueBackColor"`
			LeagueId        int    `json:"leagueId"`
			LeagueName      string `json:"leagueName"`
			LeagueNameAbbr  string `json:"leagueNameAbbr"`
			MatchDate       string `json:"matchDate"`
			MatchId         int    `json:"matchId"`
			MatchNum        string `json:"matchNum"`
			MatchNumStr     string `json:"matchNumStr"`
			MatchTime       string `json:"matchTime"`
			Mnl             struct {
				Combination     string `json:"combination"`
				CombinationDesc string `json:"combinationDesc"`
				GoalLine        string `json:"goalLine"`
				ResultStatus    string `json:"resultStatus"`
				Single          int    `json:"single"`
				WinOdds         string `json:"winOdds"`
			} `json:"mnl"`
			Status int `json:"status"`
			Wnm    struct {
				Combination     string `json:"combination"`
				CombinationDesc string `json:"combinationDesc"`
				GoalLine        string `json:"goalLine"`
				ResultStatus    string `json:"resultStatus"`
				Single          int    `json:"single"`
				WinOdds         string `json:"winOdds"`
			} `json:"wnm"`
		} `json:"matchResult"`
		PageNo      int `json:"pageNo"`
		PageSize    int `json:"pageSize"`
		Pages       int `json:"pages"`
		ResultCount int `json:"resultCount"`
		Total       int `json:"total"`
	} `json:"value"`
	EmptyFlag bool   `json:"emptyFlag"`
	DataFrom  string `json:"dataFrom"`
	Success   bool   `json:"success"`
}

func basketball(c *gin.Context, order *Order) error {
	if len(order.Matches) <= 0 {
		common.FailedReturn(c, "比赛场数不能为空")
		return errors.New("比赛场数不能为空")
	}
	tx := mysql.DB.Begin()

	//回填比赛信息 以及反填胜率
	officalMatch, err := cache.GetOnTimeBasketBallMatch(order.LotteryUuid)
	if officalMatch == nil {
		common.FailedReturn(c, "查公布信息异常， 请联系管理员！")
		return errors.New("查公布信息异常， 请联系管理员！")
	}
	fillStatus := fillBasketBallMatches(*officalMatch, order, c, tx)
	if fillStatus == nil {
		common.FailedReturn(c, "回填订单信息失败， 请联系管理员！")
		return errors.New("回填订单信息失败， 请联系管理员！")
	}

	//保存所有组合
	mm, err := order.WayDetail()
	bonus := make([]float32, 0)
	if err != nil {
		log.Error("解析足彩组合失败", err)
		common.FailedReturn(c, "解析足彩组合失败")
		tx.Rollback()
		return errors.New("解析足彩组合失败")
	}
	fmt.Println("======", order.UUID, "======")
	for s, v := range mm {
		fmt.Println(s, ":")
		for _, bet := range v.([]Bet) {
			bet.UserId = order.UserID
			if err := tx.Create(&bet).Error; err != nil {
				log.Error(err)
				common.FailedReturn(c, "保存组合失败")
				tx.Rollback()
				return errors.New("保存组合失败")
			}
			bonus = append(bonus, bet.Bonus)
			for _, view := range bet.Group {
				view.BetId = bet.ID
				if err := tx.Create(&view).Error; err != nil {
					log.Error(err)
					common.FailedReturn(c, "解析场次失败")
					tx.Rollback()
					return errors.New("解析场次失败")
				}
				fmt.Printf("时间：%s \n", view.Time)
				fmt.Printf("比赛：%s \n", exchangeHomeAwayTeam(view.League))
				fmt.Printf("%s@%f \n", view.Mode, view.Odd)
				fmt.Println("----------------------------------")
			}
			fmt.Println("倍数：", order.Times)
			fmt.Println("奖金：", bet.Bonus)
			fmt.Println("=========================================")
		}
	}
	sort.Slice(bonus, func(i, j int) bool {
		return bonus[i] < bonus[j]
	})
	order.Bonus = 0
	order.LogicWinMin = bonus[0] * float32(order.Times)
	var bonusCout float32 = 0
	for _, f := range bonus {
		bonusCout += f
	}
	order.LogicWinMaX = bonusCout * float32(order.Times)
	order.ShouldPay = float32(2 * len(bonus) * order.Times)
	order.CreatedAt = time.Now()
	fmt.Println("实际付款：", order.ShouldPay)
	if order.AllWinId == 0 {
		billErr := user.CheckScoreOrDoBill(order.UserID, order.UUID, order.ShouldPay, true, tx)
		if billErr != nil {
			log.Error("扣款失败， 无法提交订单")
			common.FailedReturn(c, billErr.Error())
			return errors.New("扣款失败， 无法提交订单")
		}
		order.PayStatus = true
	}
	if err := tx.Create(order).Error; err != nil {
		log.Error("创建订单失败 ", err)
		common.FailedReturn(c, "创建订单失败， 请联系店主")
		tx.Rollback()
		return errors.New("创建订单失败， 请联系店主")
	}

	CheckBasketBallLottery(util.AddTwoHToTime(order.Matches[len(order.Matches)-1].TimeDate))

	tx.Commit()

	common.SuccessReturn(c, order.UUID)

	return nil
}

func exchangeHomeAwayTeam(str string) string {
	temp := strings.Split(str, "Vs")
	return fmt.Sprintf("客队:%s Vs 主队:%s", temp[1], temp[0])
}

func CheckBasketBallLottery(checkTime time.Time) {
	log.Info("============= 篮球对账任务开启==============")
	job := Job{
		Time:  checkTime,
		Param: nil,
		CallBack: func(param interface{}) {
			var url = "https://webapi.sporttery.cn/gateway/jc/basketball/getMatchResultV2.qry?matchPage=1&matchBeginDate=%s&matchEndDate=%s&leagueId=&pageSize=299&pageNo=1&isFix=0&pcOrWap=1"
			time := time.Now()
			date := time.Format("2006-01-02 15:04:05")
			begin := strings.Split(date, " ")[0]
			realUrl := fmt.Sprintf(url, begin, begin)
			fmt.Println("basketball check ===========> %s", realUrl)
			resp, err := http.Get(realUrl)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result BaseketBallResult
			err = json.Unmarshal(body, &result)
			if err != nil || result.Value.MatchResult == nil {
				log.Error("转换篮球结果为对象失败", err)
				return
			}
			tx := mysql.DB.Begin()
			var orders = make([]Order, 0)
			if err := tx.Model(Order{}).Where(&Order{LotteryType: "BASKETBALL"}).Find(&orders).Error; err != nil {
				log.Error("查询订单失败", err)
				return
			}
			for _, s := range result.Value.MatchResult {
				bets, err := getBetByMatchId(strconv.Itoa(s.MatchId))
				if err != nil {
					log.Error(err)
					return
				}
				for _, bet := range bets {
					for _, view := range bet.Group {
						if !view.Check && strings.Compare(strconv.Itoa(s.MatchId), view.MatchId) == 0 {
							switch view.Type {
							case "HDC": //让分胜负
								if strings.Compare(view.ScoreVsScore, strings.ToLower(s.Hdc.Combination)) == 0 {
									view.Check = true
									view.Correct = true
								} else {
									view.Check = true
									view.Correct = false
								}
								break
							case "HILO": // 大小分
								if strings.Compare(view.ScoreVsScore, strings.ToLower(s.Hdc.Combination)) == 0 {
									view.Check = true
									view.Correct = true
								} else {
									view.Check = true
									view.Correct = false
								}
								break
							case "MNL": //胜负
								if strings.Compare(view.ScoreVsScore, strings.ToLower(s.Hdc.Combination)) == 0 {
									view.Check = true
									view.Correct = true
								} else {
									view.Check = true
									view.Correct = false
								}
								break
							case "WNM": //胜分差
								num, err := strconv.Atoi(s.Wnm.Combination)
								if err != nil {
									log.Error("数字转换你失败")
									return
								}
								var tag string
								if num < 0 {
									//客胜
									tag = fmt.Sprintf("%s%d", "l", int(math.Abs(float64(num))))
								} else {
									//主胜
									tag = fmt.Sprintf("%s%d", "w", num)
								}

								fmt.Println("tag =========>", tag)
								switch tag {
								case "l1":
									view.Check = true
									view.Correct = true
									break
								case "l2":
									view.Check = true
									view.Correct = true
									break
								case "l3":
									view.Check = true
									view.Correct = true
									break
								case "l4":
									view.Check = true
									view.Correct = true
									break
								case "l5":
									view.Check = true
									view.Correct = true
									break
								case "l6":
									view.Check = true
									view.Correct = true
									break
								case "w1":
									view.Check = true
									view.Correct = true
									break
								case "w2":
									view.Check = true
									view.Correct = true
									break
								case "w3":
									view.Check = true
									view.Correct = true
									break
								case "w4":
									view.Check = true
									view.Correct = true
									break
								case "w5":
									view.Check = true
									view.Correct = true
									break
								case "w6":
									view.Check = true
									view.Correct = true
									break
								default:
									view.Check = true
									view.Correct = false
								}

								break
							}

						}
						if serr := tx.Save(&view).Error; serr != nil {
							log.Error("保存篮球比赛")
							tx.Rollback()
							return
						}
					}

				}

			}
			tx.Commit()

			orderTx := mysql.DB.Begin()
			//订单更新
			orders = getNotFinishedOrders()
			for _, order := range orders {
				bets := getBetByOrderId(order.UUID)
				var countBet = 0
				for _, bet := range bets {
					if bet.Check {
						countBet++
					}
				}
				if countBet == len(bets) {
					//所有比赛都完成 并且中奖已经对账
					order.AllMatchFinished = true
					var bonus float32 = 0.0
					for _, bet := range bets {
						if bet.Check && bet.Win {
							order.Win = true
							bonus = bonus + bet.Bonus
						}
					}
					value, _ := decimal.NewFromFloat32(bonus).Mul(decimal.NewFromInt(int64(order.Times))).Float64()
					order.Bonus = float32(value)
					if err := orderTx.Save(&order).Error; err != nil {
						log.Error("定时任务，更新订单")
						orderTx.Rollback()
						return
					}
					if order.AllWinId > 0 {
						log.Info("=====  确定合买订单  ======")
						all := GetAllWinByParentId(order.AllWinId)
						for _, win := range all {
							value2, _ := decimal.NewFromFloat32(order.Bonus).Div(decimal.NewFromInt(int64(win.BuyNumber))).Float64()
							win.Bonus = float32(value2)
							win.Timeout = true
							orderTx.Save(&win)
						}

					}

				}
			}
			orderTx.Commit()
			if param != nil {
				id := fmt.Sprintf("%d", param)
				err := mysql.DB.Model(JobExecution{}).Where("id = ?", id).Update("status", true).Error
				if err != nil {
					log.Error("更新job状态失败!")
				}
			}
			log.Info("==================== 完成篮球中奖校验 =======================")
		},
		Type: BASKETBALL,
	}
	AddJob(job)
}
