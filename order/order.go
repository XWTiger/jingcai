package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jingcai/cache"
	"jingcai/common"
	ilog "jingcai/log"

	"jingcai/mysql"
	"jingcai/user"
	"jingcai/util"
	"sort"
	"strconv"
	"strings"
	"time"
)

var log = ilog.Logger

// 兑奖状态 NO_BONUS(未中奖) READY(已发放) NO_PAY(未发放)
const (
	FOOTBALL    = "FOOTBALL"
	SUPER_LOTTO = "SUPER_LOTTO"
	P3          = "P3"
	P5          = "P5"
	BASKETBALL  = "BASKETBALL"
	SEVEN_STAR  = "SEVEN_STAR"
	TOP         = 4 //前4
	S_TOP       = 1 //连胜 1
	ALL_WIN     = "ALLWIN"
	NO_BONUS    = "NO_BONUS"
	READY       = "READY"
	NO_PAY      = "NO_PAY"
	PL_SIGNAL   = "SIGNAL"
	PL_C3       = "C3"      //组合3
	PL_C6       = "C6"      //组合6
	ZX_GSB      = "ZX_GSB"  // 直选个十百
	ALL_C       = "CALL"    // 直选全组合
	ALL_FS      = "CALL_FS" // 直选 复式
	C3_FS       = "C3_FS"   //组选三 复式
	C3_DT       = "C3_DT"   //组选三 胆拖
	C6_FS       = "C3_FS"   //组选六 复式
	C6_DT       = "C3_DT"   //组选六 胆拖
	TOMASTER    = "TOMASTER"
	SCORE       = "SCORE"
	RMB         = "RMB"
	TEMP        = "TEMP"
)

type Match struct {
	gorm.Model
	//比赛编号
	MatchNum string
	//比赛时间 2023-05-23 01:10:00
	TimeDate time.Time

	//比赛时间 2023-05-23
	MatchDate string

	//01:10:00
	MatchTime string

	//比赛时间票
	MatchNumStr string
	//主队缩写
	HomeTeamCode string
	//客队缩写
	AwayTeamCode string

	//联赛id
	LeagueId string
	//联赛编号
	LeagueCode string
	//联赛名称
	LeagueName string
	//联赛全名
	LeagueAllName string

	//主队id
	HomeTeamId string
	//客队id
	AwayTeamId string

	//比赛id
	MatchId string `validate:"required"`

	//主队名称
	HomeTeamName string
	//主队全名
	HomeTeamAllName string

	//客队名称
	AwayTeamName string
	//客队全名
	AwayTeamAllName string
	//彩票组合
	Combines []LotteryDetail `gorm:"-:all" validate:"required"`
	OrderId  string
}

type LotteryDetail struct {
	gorm.Model
	//足球类型 枚举：SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	//篮球类型 枚举：HDC （胜负）、 HILO（大小分）、 MNL（让分胜负）、 WNM（胜分差）
	Type string `validate:"required"`
	//赔率
	Odds float32

	PoolCode string `validate:"required"`
	PoolId   string `json:"poolId" validate:"required"`

	//=================足球=========================
	//比分， 类型BF才有 s00s00 s05s02
	//半全场胜平负， 类型BQSFP  aa hh
	//总进球数， 类型ZJQ s0 - s7
	//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
	//=================篮球=========================
	//让分胜负， 类型HDC a 负，  h 胜
	//大小分，类型HILO l 小， h 大
	//胜负，类型MNL a 主负， h 主胜
	//胜分差，类型WNM l1 客胜1-5分  l2 6-10分 ... l6 26+分， w1 主胜1-5分 ... w6 26+分
	ScoreVsScore string `validate:"required"`
	//让球 胜平负才有，篮球就是让分
	GoalLine string
	ParentId uint
}
type OrderVO struct {
	//订单
	Order *Order
	//图片

}
type Order struct {
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UUID      string         `gorm:"primary_key"`
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

	//比赛 按时间正序
	Matches []Match `gorm:"-:all"`

	//查询id
	LotteryUuid string `validate:"required" gorm:"-:all"`

	//数字内容 空格分隔 多组用英文逗号,
	Content string

	//保存类型 TEMP（临时保存） TOMASTER（提交到店）  合买(ALLWIN 舍弃)
	SaveType string `validate:"required"`

	//是否让人跟单
	Share bool

	//合买id
	AllWinId uint

	//用户编号
	UserID uint

	//奖金
	Bonus float32 `max:"0"`

	//兑奖状态 NO_BONUS(未中奖) READY(已发放) NO_PAY(未发放)
	BonusStatus string

	//付款金额
	ShouldPay float32 `max:"0"`

	//比赛完成且完成对比， true 全完成
	AllMatchFinished bool
	//售票结束时间
	DeadTime time.Time
	//是否中奖
	Win bool

	//付款是否成功？
	PayStatus bool

	//支付方式 ALI  WECHAT SCORE（积分）
	PayWay string

	//如果是大乐透 七星彩 排列3 5 需要填期号
	IssueId string `validate:"required" message："需要期号"`
	//SIGNAL（单注）   C6 （组合6） C3 （组合3）
	PL3Way string

	//是否已经出票？
	BetUpload bool
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

	//校验是否校验
	Check bool

	//是否中奖
	Win    bool
	UserId uint
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

	//足球类型 枚举：SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	//篮球类型 枚举：HDC （胜负）、 HILO（大小分）、 MNL（让分胜负）、 WNM（胜分差）
	Type string
	///=================足球=========================
	//比分， 类型BF才有 s00s00 s05s02
	//半全场胜平负， 类型BQSFP  aa hh
	//总进球数， 类型ZJQ s0 - s7
	//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
	//=================篮球=========================
	//让分胜负， 类型HDC a 负，  h 胜
	//大小分，类型HILO l 小， h 大
	//胜负，类型MNL a 主负， h 主胜
	//胜分差，类型WNM l1 客胜1-5分  l2 6-10分 ... l6 26+分， w1 主胜1-5分 ... w6 26+分
	ScoreVsScore string
	//让球 胜平负才有
	GoalLine string

	//是否已经对比 true 已对比
	Check bool
	//该场比赛是否买正确
	Correct bool
}

// @Summary 订单创建接口
// @Description 订单创建接口， matchs 比赛按时间从先到后排序, 提示：所有赔率以店主出票为准！
// @Tags order 订单
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Order false "订单对象"
// @Router /api/order [post]
func OrderCreate(c *gin.Context) {
	orderCreateFunc(c, nil)
}
func orderCreateFunc(c *gin.Context, orderFrom *Order) {
	var order Order
	if orderFrom != nil {
		log.Info("======== 跟单订单 ========")
		order = *orderFrom
	} else {
		log.Info("======== 发起订单 =========")
		c.BindJSON(&order)
	}
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
	if order.ShouldPay <= 0 {
		common.FailedReturn(c, "付款小于0")
		return
	}

	//TODO 扣款逻辑/扣积分逻辑
	//积分逻辑 在上面已经完成积分扣除， 这里只创建流水

}

// 回填比赛信息和赔率
func fillBasketBallMatches(games cache.BasketBallGames, order *Order, c *gin.Context, tx *gorm.DB) *Order {
	if len(order.Matches) <= 0 {
		return nil
	}
	var mapper = games.GetBasketMapper()
	for index, match := range order.Matches {
		matchMapper, ok := mapper[match.MatchId]
		if ok {
			date, error := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", matchMapper.MatchDate, matchMapper.MatchTime), time.Local)
			if error == nil {
				order.Matches[index].TimeDate = date
			} else {
				fmt.Println("====== 比赛日期转换失败， 要影响订单统计 order id：=======", order.UUID)
				log.Error(error)
				common.FailedReturn(c, "时间转换失败， 请联系店主")
				tx.Rollback()
				return nil
			}
			// 校验比赛是否已经有已经开始的了 或者超过时间
			now := time.Now()
			ftime := common.GetMatchFinishedTime(order.Matches[index].TimeDate)
			if now.UnixMicro() > ftime.UnixMicro() {
				log.Error("比赛已经开始或者已经停售了", "截止时间：", ftime.Format("2006-01-02 15:04:05"))
				common.FailedReturn(c, "比赛已经开始或者已经停售了")
				tx.Rollback()
				return nil
			}
			order.Matches[index].MatchDate = matchMapper.MatchDate
			order.Matches[index].AwayTeamAllName = matchMapper.AwayTeamAllName
			order.Matches[index].AwayTeamId = strconv.Itoa(matchMapper.AwayTeamId)
			order.Matches[index].AwayTeamName = matchMapper.AwayTeamAbbName
			order.Matches[index].HomeTeamId = strconv.Itoa(matchMapper.HomeTeamId)
			order.Matches[index].HomeTeamAllName = matchMapper.HomeTeamAllName

			order.Matches[index].HomeTeamName = matchMapper.HomeTeamAbbName
			order.Matches[index].LeagueAllName = matchMapper.LeagueAllName
			order.Matches[index].LeagueCode = matchMapper.LeagueCode
			order.Matches[index].LeagueId = strconv.Itoa(matchMapper.LeagueId)
			order.Matches[index].MatchDate = matchMapper.MatchDate
			order.Matches[index].MatchTime = matchMapper.MatchTime
			order.Matches[index].MatchNumStr = matchMapper.MatchNumStr
			order.Matches[index].MatchNum = strconv.Itoa(matchMapper.MatchNum)

			order.Matches[index].OrderId = order.UUID
			order.Matches[index].TimeDate = date
			if err := tx.Create(&order.Matches[index]).Error; err != nil {
				log.Error("save match failed", err)
				common.FailedReturn(c, "创建订单失败， 请联系店主")
				tx.Rollback()
				return nil
			}
			if len(match.Combines) > 0 {
				for in, _ := range order.Matches[index].Combines {
					odd, err := FindBasketBallOdd(order.Matches[index].MatchId, &order.Matches[index].Combines[in], mapper)
					if odd == 0 || err != nil {
						common.FailedReturn(c, "获取赔率失败")
						tx.Rollback()
						return nil
					}
					order.Matches[index].Combines[in].Odds = float32(odd)
					order.Matches[index].Combines[in].ParentId = order.Matches[index].ID
					if err := tx.Create(&order.Matches[index].Combines[in]).Error; err != nil {
						log.Error("save lottery detail  failed", err)
						common.FailedReturn(c, "创建订单失败， 请联系店主")
						tx.Rollback()
						return nil
					}
				}
			}
		} else {
			return nil
		}
	}
	return order
}

func FindOdd(matchId string, lotto *LotteryDetail, mapper map[string]cache.Match) (float64, error) {
	match, ok := mapper[matchId]
	if !ok {
		return 0, errors.New("mapper 解析失败")
	}
	if match.SellStatus != 1 || strings.Compare(match.MatchStatus, "Selling") != 0 {
		return 0, errors.New("该比赛已经停售")
	}
	//SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	switch lotto.Type {
	case "SFP":
		//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
		lotto.GoalLine = match.Had.GoalLine
		switch lotto.ScoreVsScore {
		case "hada":
			//主负
			odd, err := strconv.ParseFloat(match.Had.A, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				lotto.GoalLine = match.Had.GoalLine
				return odd, nil
			}
		case "hadd":
			//主平
			odd, err := strconv.ParseFloat(match.Had.D, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				lotto.GoalLine = match.Had.GoalLine
				return odd, nil
			}
		case "hadh":
			//主胜
			odd, err := strconv.ParseFloat(match.Had.H, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				lotto.GoalLine = match.Had.GoalLine
				return odd, nil
			}
		case "hhada":
			//客负
			lotto.GoalLine = match.Hhad.GoalLine
			odd, err := strconv.ParseFloat(match.Hhad.A, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				lotto.GoalLine = match.Hhad.GoalLine
				return odd, nil
			}
		case "hhadd":
			lotto.GoalLine = match.Hhad.GoalLine
			odd, err := strconv.ParseFloat(match.Hhad.D, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				lotto.GoalLine = match.Hhad.GoalLine
				return odd, nil
			}
		case "hhadh":
			//客胜
			lotto.GoalLine = match.Hhad.GoalLine
			odd, err := strconv.ParseFloat(match.Hhad.H, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				lotto.GoalLine = match.Hhad.GoalLine
				return odd, nil
			}
		}
		break
	case "BF":
		//比分
		switch lotto.ScoreVsScore {
		case "s00s00":
			//比分 0:0
			odd, err := strconv.ParseFloat(match.Crs.S00S00, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s00s01":
			odd, err := strconv.ParseFloat(match.Crs.S00S01, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s00s02":
			odd, err := strconv.ParseFloat(match.Crs.S00S02, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s00s03":
			odd, err := strconv.ParseFloat(match.Crs.S00S03, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s00s04":
			odd, err := strconv.ParseFloat(match.Crs.S00S04, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s00s05":
			odd, err := strconv.ParseFloat(match.Crs.S00S05, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s01s00":
			odd, err := strconv.ParseFloat(match.Crs.S01S00, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s01s01":
			odd, err := strconv.ParseFloat(match.Crs.S01S01, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s01s02":
			odd, err := strconv.ParseFloat(match.Crs.S01S02, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s01s03":
			odd, err := strconv.ParseFloat(match.Crs.S01S03, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s01s04":
			odd, err := strconv.ParseFloat(match.Crs.S01S04, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s01s05":
			odd, err := strconv.ParseFloat(match.Crs.S01S05, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s1sa":
			//负 其它比分 赔率
			odd, err := strconv.ParseFloat(match.Crs.S1Sa, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s1sd":
			//平 其它比分 赔率
			odd, err := strconv.ParseFloat(match.Crs.S1Sd, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s1sh":
			//胜 其它比分 赔率
			odd, err := strconv.ParseFloat(match.Crs.S1Sh, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s02s00":
			odd, err := strconv.ParseFloat(match.Crs.S02S00, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s02s01":
			odd, err := strconv.ParseFloat(match.Crs.S02S01, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s02s02":
			odd, err := strconv.ParseFloat(match.Crs.S02S02, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s02s03":
			odd, err := strconv.ParseFloat(match.Crs.S02S03, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s02s04":
			odd, err := strconv.ParseFloat(match.Crs.S02S04, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s02s05":
			odd, err := strconv.ParseFloat(match.Crs.S02S05, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s03s00":
			odd, err := strconv.ParseFloat(match.Crs.S03S00, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s03s01":
			odd, err := strconv.ParseFloat(match.Crs.S03S01, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s03s02":
			odd, err := strconv.ParseFloat(match.Crs.S03S02, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s03s03":
			odd, err := strconv.ParseFloat(match.Crs.S03S03, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s04s00":
			odd, err := strconv.ParseFloat(match.Crs.S04S00, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s04s01":
			odd, err := strconv.ParseFloat(match.Crs.S04S01, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s04s02":
			odd, err := strconv.ParseFloat(match.Crs.S04S02, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s05s00":
			odd, err := strconv.ParseFloat(match.Crs.S05S00, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s05s01":
			odd, err := strconv.ParseFloat(match.Crs.S05S01, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s05s02":
			odd, err := strconv.ParseFloat(match.Crs.S05S02, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		default:
			return 0, errors.New("类型不存在")

		}
		break
	case "ZJQ":
		//总进球
		lotto.GoalLine = match.Ttg.GoalLine
		switch lotto.ScoreVsScore {

		case "s0":
			odd, err := strconv.ParseFloat(match.Ttg.S0, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s1":
			odd, err := strconv.ParseFloat(match.Ttg.S1, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s2":
			odd, err := strconv.ParseFloat(match.Ttg.S2, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s3":
			odd, err := strconv.ParseFloat(match.Ttg.S3, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s4":
			odd, err := strconv.ParseFloat(match.Ttg.S4, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s5":
			odd, err := strconv.ParseFloat(match.Ttg.S5, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s6":
			odd, err := strconv.ParseFloat(match.Ttg.S6, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "s7":
			odd, err := strconv.ParseFloat(match.Ttg.S7, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		default:
			return 0, errors.New("类型不存在")

		}
		break
	case "BQSFP":
		//半场胜平负
		switch lotto.ScoreVsScore {
		case "aa":
			//负负
			odd, err := strconv.ParseFloat(match.Hafu.Aa, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "ad":
			odd, err := strconv.ParseFloat(match.Hafu.Ad, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
			break
		case "ah":
			odd, err := strconv.ParseFloat(match.Hafu.Ah, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "da":
			//平负
			odd, err := strconv.ParseFloat(match.Hafu.Da, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "dd":
			//平平
			odd, err := strconv.ParseFloat(match.Hafu.Dd, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "dh":
			odd, err := strconv.ParseFloat(match.Hafu.Dh, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "ha":
			//胜负
			odd, err := strconv.ParseFloat(match.Hafu.Ha, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "hd":
			//胜平
			odd, err := strconv.ParseFloat(match.Hafu.Hd, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "hh":
			odd, err := strconv.ParseFloat(match.Hafu.Hh, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		default:
			return 0, errors.New("类型不存在")
		}
		break
	default:
		return 0, errors.New("类型不存在")

	}
	return 0, errors.New("类型不存在")
}

func FindBasketBallOdd(matchId string, lotto *LotteryDetail, mapper map[string]cache.BasketMatch) (float64, error) {
	//篮球类型 枚举：HDC （胜负）、 HILO（大小分）、 MNL（让分胜负）、 WNM（胜分差）
	match := mapper[matchId]
	switch lotto.Type {
	case "HDC":
		//胜负
		if strings.Compare(lotto.ScoreVsScore, "a") == 0 {
			odd, err := strconv.ParseFloat(match.Hdc.A, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		}
		if strings.Compare(lotto.ScoreVsScore, "h") == 0 {
			odd, err := strconv.ParseFloat(match.Hdc.H, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		}
		break
	case "HILO":
		//大小分
		lotto.GoalLine = match.Hilo.GoalLine
		switch lotto.ScoreVsScore {
		case "l":
			odd, err := strconv.ParseFloat(match.Hilo.L, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "h":
			odd, err := strconv.ParseFloat(match.Hilo.H, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		}
		break
	case "MNL":
		//让分胜负
		lotto.GoalLine = match.Mnl.GoalLine
		if strings.Compare(lotto.ScoreVsScore, "a") == 0 {
			odd, err := strconv.ParseFloat(match.Mnl.A, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		}
		if strings.Compare(lotto.ScoreVsScore, "h") == 0 {
			odd, err := strconv.ParseFloat(match.Hdc.H, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		}
		break
	case "WNM":
		//胜分差
		switch lotto.ScoreVsScore {
		case "l1":
			odd, err := strconv.ParseFloat(match.Wnm.L1, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}

		case "l2":
			odd, err := strconv.ParseFloat(match.Wnm.L2, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "l3":
			odd, err := strconv.ParseFloat(match.Wnm.L3, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "l4":
			odd, err := strconv.ParseFloat(match.Wnm.L4, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "l5":
			odd, err := strconv.ParseFloat(match.Wnm.L5, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "l6":
			odd, err := strconv.ParseFloat(match.Wnm.L6, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "w1":
			odd, err := strconv.ParseFloat(match.Wnm.W1, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "w2":
			odd, err := strconv.ParseFloat(match.Wnm.W2, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "w3":
			odd, err := strconv.ParseFloat(match.Wnm.W3, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "w4":
			odd, err := strconv.ParseFloat(match.Wnm.W4, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "w5":
			odd, err := strconv.ParseFloat(match.Wnm.W5, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		case "w6":
			odd, err := strconv.ParseFloat(match.Wnm.W6, 8)
			if err != nil {
				log.Error("存在赔率无法转换")
				return 0.0, errors.New("存在赔率无法转换")
			} else {
				return odd, nil
			}
		}
		break
	default:
		return 0, errors.New("类型不存在")

	}
	return 0, errors.New("类型不存在")
}

func getBets(list []int, index int, foots *[]FootView, bets *[][]FootView, matches []Match) {
	if index >= len(list) {
		temp := make([]FootView, len(*foots))
		copy(temp, *foots)
		*bets = append(*bets, temp)
		return
	}
	match := matches[list[index]-1]
	for _, combine := range match.Combines {
		*foots = append(*foots, FootView{
			Way:          "过关方式 1x1",
			Time:         match.MatchNumStr,
			League:       fmt.Sprintf("主队:%s Vs 客队:%s", match.HomeTeamName, match.AwayTeamName),
			Mode:         GetDesc(combine.Type, combine.ScoreVsScore),
			Odd:          combine.Odds,
			MatchNum:     match.MatchNum,
			MatchId:      match.MatchId,
			Type:         combine.Type,
			ScoreVsScore: combine.ScoreVsScore,
			GoalLine:     combine.GoalLine,
			Check:        false,
			Correct:      false,
		})
		getBets(list, index+1, foots, bets, matches)
		*foots = (*foots)[:0]
	}
}

type TigerDragon struct {
	//中奖top
	Tops       []Top
	SerialWins []WinSer
}
type Top struct {
	//奖金 top4
	Bonus    float32
	UserInfo user.UserDTO
}

type WinSer struct {
	//连胜次数 top1
	Times    int
	UserInfo user.UserDTO
}

// @Summary 龙虎榜
// @Description 龙虎榜
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param start query   uint true "时间戳 unix time"
// @param end query   uint true "时间戳 unix time"
// @Router /api/tiger-dragon-list [get]
func TigerDragonList(c *gin.Context) {
	//默认前4名中奖 以及 连红
	//TODO 校验开通

	startValue, _ := strconv.Atoi(c.Query("start"))
	endValue, _ := strconv.Atoi(c.Query("end"))
	startTime := time.Unix(int64(startValue), 0)
	endTime := time.Unix(int64(endValue), 0)
	startDate := startTime.Format("2006-01-02 15:04:05")
	endDate := endTime.Format("2006-01-02 15:04:05")
	var orders = make([]Order, 0)
	users := mysql.DB.Model(&Order{}).Select("user_id").Group("user_id")
	mysql.DB.Model(&Order{}).Order("created_at desc").Where("user_id in (?) and created_at BETWEEN ? AND ?", users, startDate, endDate).Find(&orders)
	var win = make([]Top, 0)
	var seriel = make([]WinSer, 0)
	var buff = make(map[uint]*[]Order)
	for _, order := range orders {
		list, ok := buff[order.UserID]
		if ok {
			*list = append(*list, order)
		} else {
			orList := make([]Order, 0)
			buff[order.UserID] = &orList
		}

	}
	for k, v := range buff {
		var param = user.User{Model: gorm.Model{
			ID: k,
		}}
		var user user.User
		mysql.DB.Model(&param).First(&user)
		var top = Top{
			Bonus:    0,
			UserInfo: user.GetDTO(),
		}

		var ser = WinSer{
			Times:    0,
			UserInfo: user.GetDTO(),
		}
		var countBonus float32 = 0
		var coutSer = 0
		var flag = false
		for _, order := range *v {
			if order.AllMatchFinished && order.Win {
				countBonus += order.Bonus
				if !flag {
					coutSer++
				}

			}
			if order.AllMatchFinished && !order.Win {
				flag = true
			}

		}
		top.Bonus = countBonus
		ser.Times = coutSer
		win = append(win, top)
		seriel = append(seriel, ser)
		sort.Slice(win, func(i, j int) bool {
			return win[i].Bonus < win[j].Bonus
		})
		sort.Slice(seriel, func(i, j int) bool {
			return seriel[i].Times < seriel[j].Times
		})

		common.SuccessReturn(c, &TigerDragon{
			Tops:       win[:TOP],
			SerialWins: seriel[:S_TOP],
		})
	}

}

// t 类型
func GetDesc(t string, scoreVsScore string) string {
	//足球类型 枚举：SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	//篮球类型 枚举：HDC （胜负）、 HILO（大小分）、 MNL（让分胜负）、 WNM（胜分差）

	switch t {
	case "HDC":
		switch scoreVsScore {
		case "a":
			return "让分主胜"
		case "h":
			return "让分客胜"
		}

		break
	case "HILO":
		switch scoreVsScore {
		case "l":
			return "小"
		case "h":
			return "大"
		}
		break
	case "MNL":
		switch scoreVsScore {
		case "a":
			return "主负"
		case "h":
			return "主胜"
		}

		break
	case "WNM":
		switch scoreVsScore {
		case "l1":
			return "(客胜 1-5 分差)"
		case "l2":
			return "(客胜 6-10 分差)"
		case "l3":
			return "(客胜 10-15 分差)"
		case "l4":
			return "(客胜 6-20 分差)"
		case "l5":
			return "(客胜 21-25 分差)"
		case "l6":
			return "(客胜 26+ 分差)"
		case "w1":
			return "(主胜 1-5 分差)"
		case "w2":
			return "(主胜 6-10 分差)"
		case "w3":
			return "(主胜 11-15 分差)"
		case "w4":
			return "(主胜 16-20 分差)"
		case "w5":
			return "(主胜 21-25 分差)"
		case "w6":
			return "(主胜 26+ 分差)"
		}
		break

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

type FollowDto struct {
	OrderId string
	Times   int
}

// @Summary 跟单订单
// @Description 跟单订单
// @Tags order 订单
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param follow body FollowDto true "跟单对象"
// @Router /api/order/follow [post]

func GetOrderByLotteryType(tp string) []Order {
	var orders []Order
	if err := mysql.DB.Model(Order{}).Where(&Order{LotteryType: tp, AllMatchFinished: false}).Find(&orders).Error; err != nil {
		log.Error(err)
		return nil
	}
	return orders
}

// 对比号码， length 个数  是否满足num个
func randomNumBeforeDirect(length int, num int, userNum string, releaseNum string) (bool, int) {
	//前5任意数量的数值相同
	var count = 0
	numBuffer := strings.Split(userNum, " ")
	releaseBuffer := strings.Split(releaseNum, " ")

	for i := 0; i < length; i++ {
		if util.PaddingZeroCompare(numBuffer[i], releaseBuffer[i]) {
			count += 1
		}
	}
	if count >= num {
		return true, count
	}
	return false, count
}

// 对比号码， length 个数  是否满足num个
func CompareDirectNum(index []int, number int, userNum string, releaseNum string) (bool, int) {
	//前5任意数量的数值相同
	var count = 0
	numBuffer := strings.Split(userNum, " ")
	releaseBuffer := strings.Split(releaseNum, " ")

	for i := 0; i < len(index); i++ {
		if util.PaddingZeroCompare(numBuffer[index[i]], releaseBuffer[index[i]]) {
			count += 1
		}
	}
	if count >= number {
		return true, count
	}
	return false, count
}

func getArr(content string) []string {

	if strings.Contains(content, ",") {
		return strings.Split(content, ",")
	} else {
		var strs []string
		strs = append(strs, content)
		return strs
	}
}

func GetPlAllNums(order *Order) ([]string, string) {

	if strings.Compare(order.PL3Way, ZX_GSB) == 0 {
		//个十百 组合
		arr := strings.Split(order.Content, ",")
		combineArr := make([][]string, 0)
		for i, s := range arr {
			combineArr[i] = strings.Split(s, " ")
		}
		var childs = make([]string, 0)
		var sb = make([]byte, 0)
		util.GetZxGsb(0, combineArr, &sb, &childs)

		return childs, PL_SIGNAL
	}
	if strings.Compare(order.PL3Way, PL_C3) == 0 {
		if strings.Contains(order.Content, ",") {
			return strings.Split(order.Content, ","), PL_C3
		} else {
			var strs []string
			strs = append(strs, order.Content)
			return strs, PL_C3
		}
	}

	if strings.Compare(order.PL3Way, PL_C6) == 0 {
		if strings.Contains(order.Content, ",") {
			return strings.Split(order.Content, ","), PL_C6
		} else {
			var strs []string
			strs = append(strs, order.Content)
			return strs, PL_C6
		}
	}

	return nil, ""
}
