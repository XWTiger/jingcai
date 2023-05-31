package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"jingcai/cache"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/util"
	"jingcai/validatior"
	"sort"
	"strconv"
	"strings"
	"time"
)

var log = ilog.Logger

const (
	FOOTBALL    = "FOOTBALL"
	SUPER_LOTTO = "SUPER_LOTTO"
	P3          = "P3"
	P5          = "P5"
	BASKETBALL  = "BASKETBALL"
	SEVEN_STAR  = "SEVEN_STAR"
)

type Match struct {
	gorm.Model
	//比赛编号
	MatchNum string `validate:"required"`
	//比赛时间 2023-05-23 01:10:00
	MatchDate time.Time `validate:"required"`

	//比赛时间票
	MatchNumStr string
	//主队缩写
	HomeTeamCode string `validate:"required"`
	//客队缩写
	AwayTeamCode string `validate:"required"`

	//联赛id
	LeagueId string `validate:"required"`
	//联赛编号
	LeagueCode string `validate:"required"`
	//联赛名称
	LeagueName string `validate:"required"`
	//联赛全名
	LeagueAllName string `validate:"required"`

	//主队id
	HomeTeamId string `validate:"required"`
	//客队id
	AwayTeamId string `validate:"required"`

	//比赛id
	MatchId string `validate:"required"`

	//主队名称
	HomeTeamName string `validate:"required"`
	//主队全名
	HomeTeamAllName string `validate:"required"`

	//客队名称
	AwayTeamName string `validate:"required"`
	//客队全名
	AwayTeamAllName string `validate:"required"`
	//彩票组合
	Combines []LotteryDetail `gorm:"-:all" validate:"required"`
	OrderId  string          `validate:"required"`
}

type LotteryDetail struct {
	gorm.Model
	//类型 枚举：SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	Type string
	//赔率
	Odds float32

	PoolCode string
	PoolId   string

	//比分， 类型BF才有 s00s00 s05s02
	//半全场胜平负， 类型BQSFP  aa hh
	//总进球数， 类型ZJQ s0 - s7
	//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
	ScoreVsScore string
	//让球
	GoalLine string
	ParentId uint
}

type Order struct {
	UUID string `gorm:"primary_key"`
	//倍数
	Times int `validate:"required"`
	//过关
	Way string
	//足彩（FOOTBALL） 大乐透（SUPER_LOTTO）  排列三（P3） 篮球(BASKETBALL) 七星彩（SEVEN_STAR） 排列五（P5）
	LotteryType string

	//逻辑最小中奖
	LogicWinMin float32

	//逻辑最大中奖
	LogicWinMaX float32

	//比赛
	Matches []Match `gorm:"-:all"`

	//查询id
	LotteryUuid string `validate:"required" gorm:"-:all"`

	//数字内容
	Content string

	//保存类型 TEMP（临时保存） TOMASTER（提交到店）  合买(ALLWIN)
	SaveType string

	//是否让人跟单
	Share bool

	//合买id
	AllWinId uint

	//用户编号
	UserID uint
}
type Bet struct {
	gorm.Model
	OrderId  string
	Group    []FootView `gorm:"-:all"`
	Way      string
	MatchNum string
	MatchId  string
	//奖金
	Bonus float32
}
type FootView struct {
	gorm.Model
	//过关方式
	Way string
	//比赛时间
	Time string
	//比赛
	League string
	//实际购买种类（比分 胜负等）
	Mode string

	//奖金倍率
	Odd      float32
	BetId    uint
	MatchNum string
	MatchId  string
}

// @Summary 订单创建接口
// @Description 订单创建接口
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Order false "订单对象"
// @Router /order [post]
func OrderCreate(c *gin.Context) {
	var order Order
	c.BindJSON(&order)
	if order.Times <= 0 {
		common.FailedReturn(c, "倍数不能为空")
		return
	}
	if order.LotteryType == "" {
		common.FailedReturn(c, "购买类型不能为空")
		return
	}

	if order.SaveType == "" {
		common.FailedReturn(c, "订单类型不能为空")
		return
	}
	order.UUID = uuid.NewV4().String()
	var user = user.FetUserInfo(c)
	order.UserID = user.ID
	switch order.LotteryType {

	case FOOTBALL:
		validatior.Validator(c, order)
		football(c, &order)

		return
	case P3:
		return
	case P5:
		return
	case BASKETBALL:
		return
	case SUPER_LOTTO:
		return
	case SEVEN_STAR:
		return
	default:
		common.FailedReturn(c, "购买类型不正确")
		return
	}
}

func football(c *gin.Context, order *Order) {
	if len(order.Matches) <= 0 {
		common.FailedReturn(c, "比赛场数不能为空")
		return
	}
	mysql.DB.AutoMigrate(&Order{})
	mysql.DB.AutoMigrate(&Match{})
	mysql.DB.AutoMigrate(&LotteryDetail{})
	mysql.DB.AutoMigrate(&Bet{})
	mysql.DB.AutoMigrate(&FootView{})
	tx := mysql.DB.Begin()

	//保存所有组合
	mm, err := order.WayDetail()
	bonus := make([]float32, 0)
	if err != nil {
		log.Error("解析足彩组合失败", err)
		common.FailedReturn(c, "解析足彩组合失败")
		tx.Rollback()
		return
	}
	for s, v := range mm {
		fmt.Println(s, ":")
		for _, bet := range v.([]Bet) {
			if err := mysql.DB.Create(&bet).Error; err != nil {
				log.Error(err)
				common.FailedReturn(c, "保存组合失败")
				tx.Rollback()
				return
			}
			bonus = append(bonus, bet.Bonus)
			for _, view := range bet.Group {
				view.BetId = bet.ID
				if err := mysql.DB.Create(&view).Error; err != nil {
					log.Error(err)
					common.FailedReturn(c, "解析场次失败")
					tx.Rollback()
					return
				}
				fmt.Printf("时间：%s \n", view.Time)
				fmt.Printf("联赛：%s \n", view.League)
				fmt.Printf("%s@%f \n", view.Mode, view.Odd)
				fmt.Println("----------------------------------")
			}
			fmt.Println("奖金：", bet.Bonus)
			fmt.Println("=========================================")
		}
	}
	sort.Slice(bonus, func(i, j int) bool {
		return bonus[i] < bonus[j]
	})
	order.LogicWinMin = bonus[0]
	order.LogicWinMaX = bonus[len(bonus)-1]

	if err := tx.Create(order).Error; err != nil {
		log.Error("创建订单失败 ", err)
		common.FailedReturn(c, "创建订单失败， 请联系店主")
		tx.Rollback()
		return
	}
	//反填胜率
	data := cache.Get(order.LotteryUuid).(string)
	var game cache.FootBallGames
	jerr := json.Unmarshal([]byte(data), &game)
	if jerr != nil {
		log.Error(jerr)
		common.FailedReturn(c, "获取公布信息失败")
		tx.Rollback()
		return
	}

	for _, ele := range order.Matches {
		ele.OrderId = order.UUID

		if err := mysql.DB.Create(&ele).Error; err != nil {
			log.Error("save match failed", err)
			common.FailedReturn(c, "创建订单失败， 请联系店主")
			tx.Rollback()
			return
		}
		if len(ele.Combines) > 0 {
			for _, combine := range ele.Combines {
				odd := FindOdd(ele.MatchId, &combine, game)
				if odd == 0 {
					common.FailedReturn(c, "获取赔率失败")
					tx.Rollback()
					return
				}
				combine.Odds = float32(odd)
				combine.ParentId = ele.ID
				if err := mysql.DB.Create(&combine).Error; err != nil {
					log.Error("save lottery detail  failed", err)
					common.FailedReturn(c, "创建订单失败， 请联系店主")
					tx.Rollback()
					return
				}
			}
		}
	}

	cache.Remove(order.LotteryUuid)
	tx.Commit()
}

func FindOdd(matchId string, lotto *LotteryDetail, game cache.FootBallGames) float64 {
	mapper := game.MatchListToMap()
	match, ok := mapper[matchId]
	if !ok {
		return 0
	}
	//SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	switch lotto.Type {
	case "SFP":
		//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
		switch lotto.ScoreVsScore {
		case "hada":
			//主负
			odd, err := strconv.ParseFloat(match.Had.A, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				lotto.GoalLine = match.Had.GoalLine
				return odd
			}
		case "hadd":
			odd, err := strconv.ParseFloat(match.Had.D, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				lotto.GoalLine = match.Had.GoalLine
				return odd
			}
		case "hadh":
			odd, err := strconv.ParseFloat(match.Had.H, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				lotto.GoalLine = match.Had.GoalLine
				return odd
			}
		case "hhada":
			//客负
			odd, err := strconv.ParseFloat(match.Hhad.A, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				lotto.GoalLine = match.Hhad.GoalLine
				return odd
			}
		case "hhadd":
			odd, err := strconv.ParseFloat(match.Hhad.D, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				lotto.GoalLine = match.Hhad.GoalLine
				return odd
			}
		case "hhadh":
			//客胜
			odd, err := strconv.ParseFloat(match.Hhad.H, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				lotto.GoalLine = match.Hhad.GoalLine
				return odd
			}
		}
		break
	case "BF":
		//比分
		switch lotto.ScoreVsScore {
		case "s00s00":
			//比分 0:0
			odd, err := strconv.ParseFloat(match.Crs.S00S00, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s00s01":
			odd, err := strconv.ParseFloat(match.Crs.S00S01, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s00s02":
			odd, err := strconv.ParseFloat(match.Crs.S00S02, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s00s03":
			odd, err := strconv.ParseFloat(match.Crs.S00S03, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s00s04":
			odd, err := strconv.ParseFloat(match.Crs.S00S04, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s00s05":
			odd, err := strconv.ParseFloat(match.Crs.S00S05, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s01s00":
			odd, err := strconv.ParseFloat(match.Crs.S01S00, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s01s01":
			odd, err := strconv.ParseFloat(match.Crs.S01S01, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s01s02":
			odd, err := strconv.ParseFloat(match.Crs.S01S02, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s01s03":
			odd, err := strconv.ParseFloat(match.Crs.S01S03, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s01s04":
			odd, err := strconv.ParseFloat(match.Crs.S01S04, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s01s05":
			odd, err := strconv.ParseFloat(match.Crs.S01S05, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s1sa":
			//负 其它比分 赔率
			odd, err := strconv.ParseFloat(match.Crs.S1Sa, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s1sd":
			//平 其它比分 赔率
			odd, err := strconv.ParseFloat(match.Crs.S1Sd, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s1sh":
			//胜 其它比分 赔率
			odd, err := strconv.ParseFloat(match.Crs.S1Sh, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s02s00":
			odd, err := strconv.ParseFloat(match.Crs.S02S00, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s02s01":
			odd, err := strconv.ParseFloat(match.Crs.S02S01, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s02s02":
			odd, err := strconv.ParseFloat(match.Crs.S02S02, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s02s03":
			odd, err := strconv.ParseFloat(match.Crs.S02S03, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s02s04":
			odd, err := strconv.ParseFloat(match.Crs.S02S04, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s02s05":
			odd, err := strconv.ParseFloat(match.Crs.S02S05, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s03s00":
			odd, err := strconv.ParseFloat(match.Crs.S03S00, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s03s01":
			odd, err := strconv.ParseFloat(match.Crs.S03S01, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s03s02":
			odd, err := strconv.ParseFloat(match.Crs.S03S02, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s03s03":
			odd, err := strconv.ParseFloat(match.Crs.S03S03, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s04s00":
			odd, err := strconv.ParseFloat(match.Crs.S04S00, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s04s01":
			odd, err := strconv.ParseFloat(match.Crs.S04S01, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s04s02":
			odd, err := strconv.ParseFloat(match.Crs.S04S02, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s05s00":
			odd, err := strconv.ParseFloat(match.Crs.S05S00, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s05s01":
			odd, err := strconv.ParseFloat(match.Crs.S05S01, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s05s02":
			odd, err := strconv.ParseFloat(match.Crs.S05S02, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		default:
			return 0

		}
		break
	case "ZJQ":
		//总进球
		switch lotto.ScoreVsScore {
		case "s0":
			odd, err := strconv.ParseFloat(match.Ttg.S0, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s1":
			odd, err := strconv.ParseFloat(match.Ttg.S1, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s2":
			odd, err := strconv.ParseFloat(match.Ttg.S2, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s3":
			odd, err := strconv.ParseFloat(match.Ttg.S3, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s4":
			odd, err := strconv.ParseFloat(match.Ttg.S4, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s5":
			odd, err := strconv.ParseFloat(match.Ttg.S5, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s6":
			odd, err := strconv.ParseFloat(match.Ttg.S6, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "s7":
			odd, err := strconv.ParseFloat(match.Ttg.S7, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		default:
			return 0

		}
		break
	case "BQSFP":
		//半场胜平负
		switch lotto.ScoreVsScore {
		case "aa":
			//负负
			odd, err := strconv.ParseFloat(match.Hafu.Aa, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "ad":
			odd, err := strconv.ParseFloat(match.Hafu.Ad, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
			break
		case "ah":
			odd, err := strconv.ParseFloat(match.Hafu.Ah, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "da":
			//平负
			odd, err := strconv.ParseFloat(match.Hafu.Da, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "dd":
			//平平
			odd, err := strconv.ParseFloat(match.Hafu.Dd, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "dh":
			odd, err := strconv.ParseFloat(match.Hafu.Dh, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "ha":
			//胜负
			odd, err := strconv.ParseFloat(match.Hafu.Ha, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "hd":
			//胜平
			odd, err := strconv.ParseFloat(match.Hafu.Hd, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		case "hh":
			odd, err := strconv.ParseFloat(match.Hafu.Hh, 32)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0
			} else {
				return odd
			}
		default:
			return 0
		}
		break
	default:
		return 0

	}
	return 0
}

func (order *Order) WayDetail() (map[string]interface{}, error) {

	ways := strings.Split(order.Way, ",")
	oddCombines := make(map[string]interface{})
	data := cache.Get(order.LotteryUuid).(string)
	var game cache.FootBallGames
	jerr := json.Unmarshal([]byte(data), &game)
	if jerr != nil {
		log.Error(jerr)
		return oddCombines, errors.New("缓存解析失败")
	}
	poolMap := game.GetSinglePoolMap()
	for _, way := range ways {
		switch way {
		case "1x1":
			//单关 BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
			Single := make([]Bet, 0)
			for _, match := range order.Matches {
				for _, combine := range match.Combines {
					value, _ := strconv.Atoi(combine.PoolId)

					if poolMap[value].CbtSingle > 0 {
						Single = append(Single, Bet{
							OrderId: order.UUID,
							Group: []FootView{{
								Way:    "过关方式 1x1",
								Time:   match.MatchNumStr,
								League: fmt.Sprintf("主队:%s Vs 客队:%s"),
								Mode:   GetDesc(combine.Type, combine.ScoreVsScore),
								Odd:    combine.Odds,
							}},
							Way:      "过关方式 1x1",
							MatchId:  match.MatchId,
							MatchNum: match.MatchNum,
						})

					}
				}
			}
			oddCombines["1x1"] = Single
			break
		case "2x1":
			if len(order.Matches) < 2 {
				return nil, errors.New("场数必须不小于2")
			}
			res := util.Combine(len(order.Matches), 2)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 2x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
					}
					bets = append(bets, bet)
				}

			}
			break
		case "3x1":
			if len(order.Matches) < 3 {
				return nil, errors.New("场数必须不小于3")
			}
			res := util.Combine(len(order.Matches), 3)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					var bonus float32 = 1
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 3x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Bonus:    bonus,
					}
					bets = append(bets, bet)
				}

			}
			break
		case "4x1":
			if len(order.Matches) < 4 {
				return nil, errors.New("场数必须不小于4")
			}
			res := util.Combine(len(order.Matches), 4)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 4x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
					}
					bets = append(bets, bet)
				}

			}
			break
		case "5x1":
			if len(order.Matches) < 5 {
				return nil, errors.New("场数必须不小于3")
			}
			res := util.Combine(len(order.Matches), 5)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 5x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
					}
					bets = append(bets, bet)
				}

			}
			break
		case "6x1":
			if len(order.Matches) < 6 {
				return nil, errors.New("场数必须不小于3")
			}
			res := util.Combine(len(order.Matches), 6)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 6x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
					}
					bets = append(bets, bet)
				}

			}
			break
		case "7x1":
			if len(order.Matches) < 3 {
				return nil, errors.New("场数必须不小于3")
			}
			res := util.Combine(len(order.Matches), 7)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 7x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
					}
					bets = append(bets, bet)
				}

			}
			break
		case "8x1":
			if len(order.Matches) < 8 {
				return nil, errors.New("场数必须不小于3")
			}
			res := util.Combine(len(order.Matches), 8)
			bets := make([]Bet, 0)
			for _, re := range res {

				foots := make([]FootView, 0)
				betsTmp := make([][]FootView, 0)
				getBets(re, 0, &foots, &betsTmp, order.Matches)
				for _, views := range betsTmp {
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 8x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
					}
					bets = append(bets, bet)
				}

			}
			break

		}
	}

	return nil, nil
}

func getBets(list []int, index int, foots *[]FootView, bets *[][]FootView, matches []Match) {
	if index >= len(list) {
		temp := make([]FootView, len(*foots))
		copy(temp, *foots)
		*bets = append(*bets, temp)
		return
	}
	match := matches[list[index]]
	for _, combine := range match.Combines {
		*foots = append(*foots, FootView{
			Way:      "过关方式 1x1",
			Time:     match.MatchNumStr,
			League:   fmt.Sprintf("主队:%s Vs 客队:%s"),
			Mode:     GetDesc(combine.Type, combine.ScoreVsScore),
			Odd:      combine.Odds,
			MatchNum: match.MatchNum,
			MatchId:  match.MatchId,
		})
		getBets(list, index+1, foots, bets, matches)
		*foots = (*foots)[:0]
	}
}

// t 类型
func GetDesc(t string, scoreVsScore string) string {
	//SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	switch t {
	case "SFP":
		//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
		switch scoreVsScore {
		case "hada":
			//主负
			return "主负"
		case "hadd":
			return "主平"
		case "hadh":
			return "主胜"
		case "hhada":
			return "客负"
		case "hhadd":
			return "客平"
		case "hhadh":
			return "客胜"
		}
		break
	case "BF":
		//比分
		switch scoreVsScore {
		case "s00s00":
			//比分 0:0
			return "(0:0)"
		case "s00s01":
			return "(0:1)"
		case "s00s02":
			return "(0:2)"
		case "s00s03":
			return "(0:3)"
		case "s00s04":
			return "(0:4)"
		case "s00s05":
			return "(0:5)"
		case "s01s00":
			return "(1:0)"
		case "s01s01":
			return "(1:1)"
		case "s01s02":
			return "(1:2)"
		case "s01s03":
			return "(1:3)"
		case "s01s04":
			return "(1:4)"
		case "s01s05":
			return "(1:5)"
		case "s1sa":
			//负 其它比分 赔率
			return "(负其它)"
		case "s1sd":
			//平 其它比分 赔率
			return "(平其它)"
		case "s1sh":
			//胜 其它比分 赔率
			return "(胜其它)"
		case "s02s00":
			return "(2:0)"
		case "s02s01":
			return "(2:1)"
		case "s02s02":
			return "(2:2)"
		case "s02s03":
			return "(2:3)"
		case "s02s04":
			return "(2:4)"
		case "s02s05":
			return "(2:5)"
		case "s03s00":
			return "(3:0)"
		case "s03s01":
			return "(3:1)"
		case "s03s02":
			return "(3:2)"
		case "s03s03":
			return "(3:3)"
		case "s04s00":
			return "(4:0)"
		case "s04s01":
			return "(4:1)"
		case "s04s02":
			return "(4:2)"
		case "s05s00":
			return "(5:0)"
		case "s05s01":
			return "(5:1)"
		case "s05s02":
			return "(5:2)"
		default:
			return ""

		}
		break
	case "ZJQ":
		//总进球
		switch scoreVsScore {
		case "s0":
			return "(0)"
		case "s1":
			return "(1)"
		case "s2":
			return "(2)"
		case "s3":
			return "(3)"
		case "s4":
			return "(4)"
		case "s5":
			return "(5)"
		case "s6":
			return "(6)"
		case "s7":
			return "(7+)"
		default:
			return ""

		}
		break
	case "BQSFP":
		//半场胜平负
		switch scoreVsScore {
		case "aa":
			//负负
			return "(负负)"
		case "ad":
			return "(负平)"
		case "ah":
			return "(负胜)"
		case "da":
			//平负
			return "(平负)"
		case "dd":
			//平平
			return "(平平)"
		case "dh":
			return "(平胜)"
		case "ha":
			//胜负
			return "(胜负)"
		case "hd":
			//胜平
			return "(胜平)"
		case "hh":
			return "(胜胜)"
		default:
			return ""
		}
		break
	default:
		return ""

	}
	return ""
}
