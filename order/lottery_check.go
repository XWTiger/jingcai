package order

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CheckLottery(whenStart time.Time) error {
	log.Info("======= 足球比赛对账线程开始 ========")
	job := Job{
		Time: whenStart,
		CallBack: func(param interface{}) {
			time := time.Now()
			date := time.Format("2006-01-02 15:04:05")
			begin := strings.Split(date, " ")[0]
			time.AddDate(0, 0, 3)
			dateEnd := time.Format("2006-01-02 15:04:05")
			end := strings.Split(dateEnd, " ")[0]
			url := fmt.Sprintf("http://127.0.0.1:8090/lottery/sports/jc/result?matchBeginDate=%s&matchEndDate=%s&pageNo=1&pageSize=299", begin, end)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result LotteryResult
			err = json.Unmarshal(body, &result)
			if err != nil || result.Content == nil {
				log.Error("转换足彩结果为对象失败", err)
				return
			}
			tx := mysql.DB.Begin()
			//确认足彩结果结果
			for _, content := range result.Content {
				if content.WinFlag == "" {
					continue
				}
				bets, err := getBetByMatchId(strconv.Itoa(content.MatchId))
				if len(bets) <= 0 {
					continue
				}
				if err != nil {
					log.Error(err)
					continue
				}
				for _, bet := range bets {
					var countCorrect = 0
					var countCheck = 0
					for _, view := range bet.Group {

						if !view.Check && strings.Compare(strconv.Itoa(content.MatchId), view.MatchId) == 0 {
							//没有对比
							//SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
							switch view.Type {
							case "SFP":
								//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜
								switch view.ScoreVsScore {
								case "hada": //不让分
									//主负
									if len(content.GoalLine) > 0 {
										goal, err := strconv.Atoi(content.GoalLine)
										if err != nil {
											log.Error("转换让球失败")
											return
										}
										section := strings.Split(content.SectionsNo999, ":")
										masterScore, _ := strconv.Atoi(section[0])
										awayScore, _ := strconv.Atoi(section[1])
										if (masterScore + goal) < awayScore {
											view.Check = true
											view.Correct = true
										} else {
											view.Check = true
											view.Correct = false
										}
									} else {
										if strings.Compare(content.WinFlag, "A") == 0 {
											view.Check = true
											view.Correct = true
										} else {
											view.Check = true
											view.Correct = false
										}
									}
									break
								case "hadd": //让分
									if len(content.GoalLine) > 0 {
										goal, err := strconv.Atoi(content.GoalLine)
										if err != nil {
											log.Error("转换让球失败")
											return
										}
										section := strings.Split(content.SectionsNo999, ":")
										masterScore, _ := strconv.Atoi(section[0])
										awayScore, _ := strconv.Atoi(section[1])
										if (masterScore + goal) == awayScore {
											view.Check = true
											view.Correct = true
										} else {
											view.Check = true
											view.Correct = false
										}
									} else {
										if strings.Compare(content.WinFlag, "D") == 0 {
											view.Check = true
											view.Correct = true
										} else {
											view.Check = true
											view.Correct = false
										}
									}
									break
								case "hadh":
									if len(content.GoalLine) > 0 {
										goal, err := strconv.Atoi(content.GoalLine)
										if err != nil {
											log.Error("转换让球失败")
											return
										}
										section := strings.Split(content.SectionsNo999, ":")
										masterScore, _ := strconv.Atoi(section[0])
										awayScore, _ := strconv.Atoi(section[1])
										if (masterScore + goal) > awayScore {
											view.Check = true
											view.Correct = true
										} else {
											view.Check = true
											view.Correct = false
										}
									} else {
										if strings.Compare(content.WinFlag, "H") == 0 {
											view.Check = true
											view.Correct = true
										} else {
											view.Check = true
											view.Correct = false
										}
									}
								}
								break
							case "BF":
								//比分
								switch view.ScoreVsScore {
								case "s00s00":
									//比分 0:0
									if strings.Compare(content.SectionsNo999, "0:0") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s00s01":
									if strings.Compare(content.SectionsNo999, "0:1") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s00s02":
									if strings.Compare(content.SectionsNo999, "0:2") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s00s03":
									if strings.Compare(content.SectionsNo999, "0:3") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s00s04":
									if strings.Compare(content.SectionsNo999, "0:4") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s00s05":
									if strings.Compare(content.SectionsNo999, "0:5") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
								case "s01s00":
									if strings.Compare(content.SectionsNo999, "1:0") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s01s01":
									if strings.Compare(content.SectionsNo999, "1:1") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s01s02":
									if strings.Compare(content.SectionsNo999, "1:2") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s01s03":
									if strings.Compare(content.SectionsNo999, "1:3") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s01s04":
									if strings.Compare(content.SectionsNo999, "1:4") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s01s05":
									if strings.Compare(content.SectionsNo999, "1:5") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s02s00":
									if strings.Compare(content.SectionsNo999, "2:0") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s02s01":
									if strings.Compare(content.SectionsNo999, "2:1") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s02s02":
									if strings.Compare(content.SectionsNo999, "2:2") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s02s03":
									if strings.Compare(content.SectionsNo999, "2:3") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s02s04":
									if strings.Compare(content.SectionsNo999, "2:4") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s02s05":
									if strings.Compare(content.SectionsNo999, "2:5") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s03s00":
									if strings.Compare(content.SectionsNo999, "3:0") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s03s01":
									if strings.Compare(content.SectionsNo999, "3:1") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s03s02":
									if strings.Compare(content.SectionsNo999, "3:2") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s03s03":
									if strings.Compare(content.SectionsNo999, "3:3") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s04s00":
									if strings.Compare(content.SectionsNo999, "4:0") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s04s01":
									if strings.Compare(content.SectionsNo999, "4:1") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s04s02":
									if strings.Compare(content.SectionsNo999, "4:2") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s05s00":
									if strings.Compare(content.SectionsNo999, "5:0") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s05s01":
									if strings.Compare(content.SectionsNo999, "5:1") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s05s02":
									if strings.Compare(content.SectionsNo999, "5:2") == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s1sa":
									//负 其它比分 赔率
									section := strings.Split(content.SectionsNo999, ":")
									masterScore, _ := strconv.Atoi(section[0])
									awayScore, _ := strconv.Atoi(section[1])
									if masterScore > 2 && awayScore >= 5 {
										//3-5 2-6 ...
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s1sd":
									//平 其它比分 赔率
									section := strings.Split(content.SectionsNo999, ":")
									masterScore, _ := strconv.Atoi(section[0])
									awayScore, _ := strconv.Atoi(section[1])
									if masterScore > 3 && awayScore == masterScore {
										//3-5 2-6 ...
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s1sh":
									//胜 其它比分 赔率
									section := strings.Split(content.SectionsNo999, ":")
									masterScore, _ := strconv.Atoi(section[0])
									awayScore, _ := strconv.Atoi(section[1])
									if masterScore >= 5 && awayScore > 2 {
										//5-3 6-2 ...
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								default:
									log.Error("比分： 一种情况不符合")

								}
								break
							case "ZJQ":
								//总进球
								section := strings.Split(content.SectionsNo999, ":")
								masterScore, _ := strconv.Atoi(section[0])
								awayScore, _ := strconv.Atoi(section[1])
								switch view.ScoreVsScore {
								case "s0":
									if (masterScore + awayScore) == 0 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s1":
									if (masterScore + awayScore) == 1 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s2":
									if (masterScore + awayScore) == 2 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s3":
									if (masterScore + awayScore) == 3 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s4":
									if (masterScore + awayScore) == 4 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s5":
									if (masterScore + awayScore) == 5 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s6":
									if (masterScore + awayScore) == 6 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "s7":
									if (masterScore + awayScore) >= 7 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								default:
									log.Error("总进球： 一种情况不符合")

								}
								break
							case "BQSFP":
								//半场胜平负
								//全场
								section := strings.Split(content.SectionsNo999, ":")
								masterScore, _ := strconv.Atoi(section[0])
								awayScore, _ := strconv.Atoi(section[1])
								//半场
								section2 := strings.Split(content.SectionsNo1, ":")
								masterScore2, _ := strconv.Atoi(section2[0])
								awayScore2, _ := strconv.Atoi(section2[1])
								switch view.ScoreVsScore {
								case "aa":
									//负负
									if masterScore < awayScore && masterScore2 < awayScore2 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "ad":
									if masterScore == awayScore && masterScore2 < awayScore2 {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "ah":
									if masterScore2 < awayScore2 && masterScore > awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "da":
									//平负
									if masterScore2 == awayScore2 && masterScore < awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "dd":
									//平平
									if masterScore2 == awayScore2 && masterScore == awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "dh":
									if masterScore2 == awayScore2 && masterScore > awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "ha":
									//胜负
									if masterScore2 > awayScore2 && masterScore < awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "hd":
									//胜平
									if masterScore2 > awayScore2 && masterScore == awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								case "hh":
									if masterScore2 > awayScore2 && masterScore > awayScore {
										view.Check = true
										view.Correct = true
									} else {
										view.Check = true
										view.Correct = false
									}
									break
								default:
									log.Error("半全场胜平负： 一种情况不符合")

								}
								break
							default:
								log.Error("购买类型匹配不上")

							}
							if serr := tx.Save(&view).Error; serr != nil {
								log.Error("保存比赛")
								tx.Rollback()
								return
							}
						}
						if view.Check {
							countCheck++
							if view.Correct {
								countCorrect++
							}

						}

					}
					if countCheck == len(bet.Group) {
						bet.Check = true
						if countCorrect == len(bet.Group) && len(bet.Group) > 0 {
							bet.Win = true
						} else {
							bet.Win = false
						}
						if serr := tx.Save(&bet).Error; serr != nil {
							log.Error("保存")
							tx.Rollback()
							return
						}
					}

				}

			}
			tx.Commit()

			orderTx := mysql.DB.Begin()
			//订单更新
			orders := getNotFinishedOrders()
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
					order.Bonus = bonus * float32(order.Times)
					if err := orderTx.Save(&order).Error; err != nil {
						log.Error("定时任务，更新订单失败")
						log.Error(err)
						orderTx.Rollback()
						return
					}
					if order.AllWinId > 0 {
						log.Info("=====  确定合买订单  ======")
						all := GetAllWinByParentId(order.AllWinId)
						for _, win := range all {
							win.Bonus = order.Bonus / float32(win.BuyNumber)
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
		},
		Type: FOOTBALL,
	}
	return AddJob(job)
}

type LotteryResult struct {
	Content    []Content   `json:"content"`
	Resource   interface{} `json:"resource"`
	StatusCode int         `json:"status_code"`
	StatusMes  string      `json:"status_mes"`
}

type Content struct {
	A                 string `json:"a"`
	AllAwayTeam       string `json:"allAwayTeam"`
	AllHomeTeam       string `json:"allHomeTeam"`
	AwayTeam          string `json:"awayTeam"`
	AwayTeamId        int    `json:"awayTeamId"`
	BettingSingle     int    `json:"bettingSingle"`
	D                 string `json:"d"`
	GoalLine          string `json:"goalLine"`
	H                 string `json:"h"`
	HomeTeam          string `json:"homeTeam"`
	HomeTeamId        int    `json:"homeTeamId"`
	LeagueBackColor   string `json:"leagueBackColor"`
	LeagueId          int    `json:"leagueId"`
	LeagueName        string `json:"leagueName"`
	LeagueNameAbbr    string `json:"leagueNameAbbr"`
	MatchDate         string `json:"matchDate"`
	MatchId           int    `json:"matchId"`
	MatchNum          string `json:"matchNum"`
	MatchNumStr       string `json:"matchNumStr"`
	MatchResultStatus string `json:"matchResultStatus"`
	PoolStatus        string `json:"poolStatus"`
	//半场比分
	SectionsNo1 string `json:"sectionsNo1"`
	//全场比分
	SectionsNo999 string `json:"sectionsNo999"`
	WinFlag       string `json:"winFlag"`
}

// @Summary 手动触发对账接口
// @Description 手动触发对账接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param date  query string true "2023-01-01 21:27:00"
// @Router /api/super/check/lottery_check [post]
func AddCheckForManual(c *gin.Context) {
	date := c.Query("date")
	startTime, err := util.StrToTime(date)
	if err != nil {
		common.FailedReturn(c, "添加手动对账失败")
		return
	}
	CheckLottery(startTime)
	CheckBasketBallLottery(startTime)
}
