package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"io"
	"jingcai/cache"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/lottery"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/util"
	"jingcai/validatior"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var log = ilog.Logger

// 兑奖状态 NO_BONUS(未中奖) READY(已发放) NO_PAY(未发放)
// 大乐透：DIRECT 直选/随机;胆拖（DT，DTQQ 前区胆拖、 DTHQ 后区胆拖、 DTSQ 双区胆拖）;复式（FS， 前区复式 FSQQ、后区复式 FSHQ、双区复式 FSSQ
// 七星彩：复式(FSSTAR),DIRECT 直选/随机;
// ==========大乐透========
// 复式前区大于5个数字，后区 大于2个,只有QQ 为前区复式 ， 只有HQ 为后区复式  都有为双区复式
// 复式：
// QQ:01 02 04 05 11 12 35 33,HQ:06 07 12
// 胆拖：
// QQD:09,QQT:01 02 06 07 09,HQD:12 HQT:02 08
// 后区胆拖
// 03 05 08 09 11 13,HQD:12 HQT:02 08
// ==========七星彩=================
// 复式：
// I1:3 4,I2:0 8,I3:8 9,I4:0 8,I5: 1 7,I6:3,I7:7 8
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
	FSQQ        = "FSQQ"   //前区复式
	FSHQ        = "FSHQ"   //后区复式
	FSSQ        = "FSSQ"   //双区复式
	DTQQ        = "DTQQ"   //前区胆拖
	DTHQ        = "DTHQ"   //后区胆拖
	DTSQ        = "DTSQ"   //双区胆拖
	DIRECT      = "DIRECT" //直选\随机
	FSSTAR      = "FSSTAR" //按位复式
	QQ          = "QQ"     //前区
	HQ          = "HQ"     //后区
	QQD         = "QQD"    //前区胆
	QQT         = "QQT"    //前区拖
	HQD         = "HQD"    //后区胆
	HQT         = "HQT"    //后区拖
	I1          = "I1"     //七星彩 第一位
	I2          = "I2"     //七星彩 第二位
	I3          = "I3"     //七星彩 第三位
	I4          = "I4"     //七星彩 第四位
	I5          = "I5"     //七星彩 第五位
	I6          = "I6"     //七星彩 第六位
	I7          = "I7"     //七星彩 第七位
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
	Images []OrderImage
	//如果是合买
	AllWin       []AllWinVO
	AllWinOrders []Order
}
type Order struct {
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UUID      string         `gorm:"primary_key"`
	//倍数
	Times int `validate:"required"`
	//过关 足球，篮球 2x1,3x1;
	// 大乐透：DIRECT 直选/随机;胆拖（DT，DTQQ 前区胆拖、 DTHQ 后区胆拖、 DTSQ 双区胆拖）;复式（FS， 前区复式 FSQQ、后区复式 FSHQ、双区复式 FSSQ
	// 七星彩：复式(FSSTAR),DIRECT 直选/随机;
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
	//==========大乐透========
	//复式前区大于5个数字，后区 大于2个,只有QQ 为前区复式 ， 只有HQ 为后区复式  都有为双区复式
	//复式：
	// QQ:01 02 04 05 11 12 35 33,HQ:06 07 12
	//胆拖：
	// QQD:09,QQT:01 02 06 07 09,HQD:12 HQT:02 08
	//后区胆拖
	// 03 05 08 09 11 13,HQD:12 HQT:02 08
	//==========七星彩=================
	//复式：
	//I1:3 4,I2:0 8,I3:8 9,I4:0 8,I5: 1 7,I6:3,I7:7 8
	Content string
	//追加
	Append bool `gorm:"type:tinyint(1)"`

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
	IssueId string `validate:"required" message:"需要期号"`
	//SIGNAL（单注）   C6 （组合6） C3 （组合3）
	//SIGNAL    一注
	//C3      组合3
	//C6组合6
	//ZX_GSB 直选个十百
	//CALL 直选全组合
	//CALL_FS 直选 复式
	//C3_FS 组选三 复式
	//C3_DT 组选三 胆拖
	//C3_FS 组选六 复式
	//C3_DT 组选六 胆拖
	PL3Way string

	//是否已经出票？
	BetUpload bool

	Comment string `grom:"type:varchar(64)"`
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

	now := time.Now()
	//校验是否在售票时间内
	finishedTime := getFinishedTime(order)
	fmt.Println("dead time ==============>", util.GetTodayYYHHMMSSFrom(finishedTime))
	if now.UnixMicro() > finishedTime.UnixMicro() {
		common.FailedReturn(c, "现在已经不在营业时间")
		return
	}
	order.DeadTime = finishedTime
	order.Bonus = 0
	order.BetUpload = false
	var userInfo = user.FetUserInfo(c)
	if userInfo == (user.User{}) {
		common.FailedReturn(c, "用户未登录/不存在")
		return
	}
	order.UserID = userInfo.ID
	order.UUID = GetOrderId(&order)
	//校验字段
	verr := validatior.Validator(c, order)
	if verr != nil {
		log.Error(verr)
		return
	}
	switch order.LotteryType {

	case FOOTBALL:
		ferr := football(c, &order)
		if ferr != nil {
			return
		}
		break
	case BASKETBALL:
		berr := basketball(c, &order)
		if berr != nil {
			return
		}
		break
	case P3:
		err := CreatePLW(&order)
		if err != nil {
			log.Error(err)
			common.FailedReturn(c, err.Error())
			return
		}
		fmt.Println("=================排列3===================", order.UUID)
		fmt.Sprintf("逻辑总奖金: %s", order.LogicWinMaX)
		fmt.Println("期号: ", order.IssueId)
		fmt.Println("号码: ", order.Content)
		fmt.Println("倍数: ", order.Times)
		fmt.Println("实际付款: ", order.ShouldPay)
		fmt.Println("=========================================")
		common.SuccessReturn(c, order.UUID)
		break
	case P5:
		err := CreatePLW(&order)
		if err != nil {
			log.Error(err)
			common.FailedReturn(c, err.Error())
			return
		}

		fmt.Println("=================排列5===================", order.UUID)
		fmt.Sprintf("逻辑总奖金: %s", order.LogicWinMaX)
		fmt.Println("期号: ", order.IssueId)
		fmt.Println("号码: ", order.Content)
		fmt.Println("倍数: ", order.Times)
		fmt.Println("实际付款: ", order.ShouldPay)
		fmt.Println("=========================================")
		common.SuccessReturn(c, order.UUID)
		break
	case SUPER_LOTTO:
		//大乐透
		tx := mysql.DB.Begin()
		err := checkSuperLotto(&order)
		if err != nil {
			log.Error(err)
			common.FailedReturn(c, err.Error())
			return
		}

		if order.AllWinId == 0 {
			billErr := user.CheckScoreOrDoBill(order.UserID, order.ShouldPay, true, tx)
			if billErr != nil {
				log.Error("扣款失败， 无法提交订单")
				common.FailedReturn(c, billErr.Error())
				tx.Rollback()
				return
			}
			order.PayStatus = true
		}
		if err := tx.Model(&Order{}).Create(&order).Error; err != nil {
			log.Error("创建订单失败 ", err)
			common.FailedReturn(c, "创建订单失败， 请联系店主")
			tx.Rollback()
			return
		}
		fmt.Println("=================大乐透===================", order.UUID)
		fmt.Sprintf("逻辑总奖金: %s", order.LogicWinMaX)
		fmt.Println("期号: ", order.IssueId)
		fmt.Println("号码: ", order.Content)
		fmt.Println("倍数: ", order.Times)
		fmt.Println("实际付款: ", order.ShouldPay)
		fmt.Println("=========================================")
		common.SuccessReturn(c, order.UUID)
		tx.Commit()
		break
	case SEVEN_STAR:
		err := checkSevenStar(&order)
		if err != nil {
			log.Error(err)
			common.FailedReturn(c, err.Error())
			return
		}
		tx := mysql.DB.Begin()
		if order.AllWinId == 0 {
			billErr := user.CheckScoreOrDoBill(order.UserID, order.ShouldPay, true, tx)
			if billErr != nil {
				log.Error("扣款失败， 无法提交订单")
				common.FailedReturn(c, billErr.Error())
				tx.Rollback()
				return
			}
			order.PayStatus = true
		}
		if err := tx.Model(&Order{}).Create(&order).Error; err != nil {
			log.Error("创建订单失败 ", err)
			common.FailedReturn(c, "创建订单失败， 请联系店主")
			tx.Rollback()
			return
		}
		fmt.Println("=================七星彩===================", order.UUID)
		fmt.Sprintf("逻辑总奖金: %s", order.LogicWinMaX)
		fmt.Println("期号: ", order.IssueId)
		fmt.Println("号码: ", order.Content)
		fmt.Println("倍数: ", order.Times)
		fmt.Println("实际付款: ", order.ShouldPay)
		fmt.Println("=========================================")
		common.SuccessReturn(c, order.UUID)
		tx.Commit()
		break
	default:
		common.FailedReturn(c, "购买类型不正确")
		return
	}
	//TODO 扣款逻辑/扣积分逻辑
	//积分逻辑 在上面已经完成积分扣除， 这里只创建流水
	err := user.BillForScore(order.UUID, userInfo.ID, order.ShouldPay, user.SUBTRACT)
	if err != nil {
		log.Error(err)
		common.FailedReturn(c, err.Error())
		return
	}
}

func checkSuperLotto(ord *Order) error {
	if len(ord.Content) <= 0 {
		return errors.New("选号不能为空")
	}
	//判断是否多注
	if strings.Contains(ord.Content, ",") {
		buffer := strings.Split(ord.Content, ",")
		for _, s := range buffer {
			temp := strings.Split(s, " ")
			if len(temp) != 7 {
				return errors.New("选号存在问题")
			}
		}
	} else {
		buffer := strings.Split(ord.Content, " ")
		if len(buffer) != 7 {
			return errors.New("选号存在问题")
		}
	}

	if len(ord.IssueId) <= 0 {
		return errors.New("订单期号不能为空")
	}
	//校验期号
	drawNum, cerr := lottery.GetSuperLotteryIssueId()
	fmt.Println("大乐透最新期号=================>", drawNum)
	issueId, err := strconv.Atoi(ord.IssueId)
	if err != nil || cerr != nil {
		log.Error(err)
		return errors.New("校验期号失败")
	}
	if issueId-drawNum != 1 {
		return errors.New("购买期号不正确")
	}

	nums, err := getArr(ord.Content, ord.LotteryType, ord.Way)
	if err != nil {
		return err
	}

	if ord.Append {
		ord.ShouldPay += float32(len(nums) * 3 * ord.Times)
	} else {
		ord.ShouldPay = float32(len(nums) * 2 * ord.Times)
	}
	if nil == nums || len(nums) <= 0 {
		return errors.New("参数异常")
	}
	for _, num := range nums {
		numbers := strings.Split(num, " ")
		for i, number := range numbers {
			numb, err := strconv.Atoi(number)
			if err != nil {
				log.Error(err)
				return errors.New("选号存在问题")
			}
			if i <= 5 {
				if numb < 1 || numb > 35 {
					return errors.New("大乐透前五位只能在01—35之间")
				}
			}

			if i >= 6 {
				if numb < 1 || numb > 12 {
					return errors.New("大乐透后2位只能在01—12之间")
				}
			}
		}
	}

	if !lottery.LotteryStatistics.Exists("super_lotto_check") {
		lottery.LotteryStatistics.Add("super_lotto_check", 8*time.Hour, 1)
		AddSuperLottoCheck(nil)
	}

	return nil
}

func checkSevenStar(ord *Order) error {
	if len(ord.Content) <= 0 {
		return errors.New("选号不能为空")
	}
	buffer := strings.Split(ord.Content, " ")
	if len(buffer) != 7 {
		return errors.New("选号存在问题")
	}

	if len(ord.IssueId) <= 0 {
		return errors.New("订单期号不能为空")
	}

	//校验期号
	drawNum, cerr := lottery.GetSevenStarIssueId()
	fmt.Println("七星彩最新期号=================>", drawNum)
	issueId, err := strconv.Atoi(ord.IssueId)
	if err != nil || cerr != nil {
		log.Error(err)
		return errors.New("校验期号失败")
	}
	if issueId-drawNum != 1 {
		return errors.New("购买期号不正确")
	}

	nums, err := getArr(ord.Content, ord.LotteryType, ord.Way)
	if err != nil {
		return nil
	}
	ord.ShouldPay = float32(len(nums) * 2 * ord.Times)
	if nil == nums || len(nums) <= 0 {
		return errors.New("参数异常")
	}
	for _, num := range nums {
		numbers := strings.Split(num, " ")
		for i, number := range numbers {
			numb, err := strconv.Atoi(number)
			if err != nil {
				log.Error(err)
				return errors.New("选号存在问题")
			}
			if i <= 6 {
				if numb < 0 || numb > 9 {
					return errors.New("七星彩前六位只能在000000-999999之间")
				}
			}

			if i > 6 {
				if numb < 0 || numb > 14 {
					return errors.New("七星彩后1位只能在0—14之间")
				}
			}
		}
	}

	if !lottery.LotteryStatistics.Exists("seven_star_check") {
		lottery.LotteryStatistics.Add("seven_star_check", 8*time.Hour, 1)
		AddSevenStarCheck(nil)
	}
	return nil
}

// @Summary 用户自己的订单查询接口
// @Description 用户自己的订单查询接口
// @Tags order 订单
// @Accept json
// @Produce json
// @Success 200 {object} OrderVO
// @failure 500 {object} common.BaseResponse
// @param saveType  query string false "保存类型 TEMP（临时保存） TOMASTER（提交到店）  ALLWIN（合买）"
// @param lotteryType  query string false "足彩（FOOTBALL） 大乐透（SUPER_LOTTO）  排列三（P3） 篮球(BASKETBALL) 七星彩（SEVEN_STAR） 排列五（P5）"
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @param win  query bool false "是否只查赢的"
// @Router /api/order [get]
func OrderList(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	if userInfo == (user.User{}) {
		log.Error("用户不存在")
	}
	saveType := c.Query("saveType")
	lotteryType := c.Query("lotteryType")
	flag := c.Query("win")
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	var param = Order{
		UserID: userInfo.ID,
	}
	if saveType != "" {
		param.SaveType = saveType
	}
	if lotteryType != "" {
		param.LotteryType = lotteryType
	}
	if flag != "" {
		param.Win = true
	}
	var list = make([]Order, 0)
	var resultList = make([]OrderVO, 0)
	var count int64
	mysql.DB.Debug().Model(&param).Where(&param).Order("created_at desc").Count(&count).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	for index, order := range list {
		//如果是足球和篮球 就把比赛回填回来
		if strings.Compare(order.LotteryType, FOOTBALL) == 0 || strings.Compare(order.LotteryType, BASKETBALL) == 0 {
			var mathParam = Match{
				OrderId: list[index].UUID,
			}
			var matchList = make([]Match, 0)
			mysql.DB.Model(&mathParam).Where("order_id=?", order.UUID).Find(&matchList)
			list[index].Matches = matchList
			for idx, match := range matchList {
				var detailParam = LotteryDetail{
					ParentId: match.ID,
				}
				var detailList = make([]LotteryDetail, 0)
				mysql.DB.Model(&detailParam).Where("parent_id=?", match.ID).Find(&detailList)
				matchList[idx].Combines = detailList
			}
		}
		var uuid string
		//如果是参加别人的合买 就把票查回来
		if strings.Compare(order.SaveType, ALL_WIN) == 0 {
			var initAllWin AllWin
			mysql.DB.Model(&AllWin{
				Model: gorm.Model{
					ID: list[index].AllWinId,
				},
			}).First(&initAllWin)
			if initAllWin.ParentId == 0 {
				uuid = initAllWin.OrderId
			} else {
				uuid = initAllWin.ParentOrderId
			}

		}
		images := getImageByOrderId(uuid)
		resultList = append(resultList, OrderVO{
			Order:  &list[index],
			Images: images,
		})
	}

	common.SuccessReturn(c, common.PageCL{
		Total:    int(count),
		Content:  resultList,
		PageSize: pageSize,
		PageNo:   page,
	})
}

// @Summary 订单分享接口
// @Description 订单分享接口
// @Tags order 订单
// @Accept json
// @Produce json
// @Success 200 {object} OrderVO
// @failure 500 {object} common.BaseResponse
// @param lotteryType  query string false "足彩（FOOTBALL） 大乐透（SUPER_LOTTO）  排列三（P3） 篮球(BASKETBALL) 七星彩（SEVEN_STAR） 排列五（P5）"
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @Router /api/order/shared [get]
func SharedOrderList(c *gin.Context) {
	saveType := c.Query("saveType")
	lotteryType := c.Query("lotteryType")
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	var param Order

	param = Order{
		Share:            true,
		AllMatchFinished: false,
	}
	if lotteryType != "" {
		param.LotteryType = lotteryType
	}

	if saveType != "" {
		param.SaveType = saveType
	}
	var list = make([]Order, 0)
	var count int64

	mysql.DB.Debug().Model(Order{}).Where(&param).Where("TIMESTAMPDIFF(SECOND, now(), dead_time) > 0 ").Order("created_at desc").Count(&count).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	for i := 0; i < len(list); i++ {
		order := list[i]
		if strings.Compare(order.LotteryType, FOOTBALL) == 0 || strings.Compare(order.LotteryType, BASKETBALL) == 0 {
			var mathParam = Match{
				OrderId: list[i].UUID,
			}
			var matchList = make([]Match, 0)
			mysql.DB.Model(&mathParam).Where(&mathParam).Find(&matchList)
			list[i].Matches = matchList
			for idx, match := range matchList {
				var detailParam = LotteryDetail{
					ParentId: match.ID,
				}
				var detailList = make([]LotteryDetail, 0)
				mysql.DB.Model(&detailParam).Where(&detailParam).Find(&detailList)
				matchList[idx].Combines = detailList
			}
		}
	}
	common.SuccessReturn(c, common.PageCL{page, pageSize, int(count), list})
}
func FindOrderVOById(uuid string, searchMatch bool) OrderVO {
	var param = Order{
		UUID: uuid,
	}
	var order Order
	mysql.DB.Model(&param).Where(&param).First(&order)
	var mathParam = Match{
		OrderId: order.UUID,
	}
	if searchMatch && (strings.Compare(FOOTBALL, order.LotteryType) == 0 || strings.Compare(BASKETBALL, order.LotteryType) == 0) {
		var matchList = make([]Match, 0)
		mysql.DB.Model(&mathParam).Where(&mathParam).Find(&matchList)
		order.Matches = matchList
		for _, match := range matchList {
			var detailParam = LotteryDetail{
				ParentId: match.ID,
			}
			var detailList = make([]LotteryDetail, 0)
			mysql.DB.Model(&detailParam).Where(&detailParam).Find(&detailList)
			match.Combines = detailList
		}
	}
	images := getImageByOrderId(uuid)

	vo := OrderVO{
		Order:  &order,
		Images: images,
	}
	if order.AllWinId > 0 {
		vo.AllWinOrders = GetAllOrderByAllWinId(order.AllWinId)
	}
	return vo

}
func GetAllOrderByAllWinId(allWinId uint) []Order {
	var orders []Order
	mysql.DB.Model(Order{AllWinId: allWinId}).Find(&orders)
	return orders
}

func FindById(uuid string, searchMatch bool) Order {
	var param = Order{
		UUID: uuid,
	}
	var order Order
	mysql.DB.Model(&param).Where(&param).First(&order)
	var mathParam = Match{
		OrderId: order.UUID,
	}
	if searchMatch {
		var matchList = make([]Match, 0)
		mysql.DB.Model(&mathParam).Where(&mathParam).Find(&matchList)
		order.Matches = matchList
		for _, match := range matchList {
			var detailParam = LotteryDetail{
				ParentId: match.ID,
			}
			var detailList = make([]LotteryDetail, 0)
			mysql.DB.Model(&detailParam).Where(&detailParam).Find(&detailList)
			match.Combines = detailList
		}
	}
	return order
}

func getNotFinishedOrders() []Order {
	var param = Order{
		AllMatchFinished: false,
	}
	var list = make([]Order, 0)
	mysql.DB.Debug().Model(&param).Where("all_match_finished=?", false).Order("created_at desc").Find(&list)
	return list
}

func football(c *gin.Context, order *Order) error {
	if len(order.Matches) <= 0 {
		common.FailedReturn(c, "比赛场数不能为空")
		return errors.New("比赛场数不能为空")
	}
	tx := mysql.DB.Begin()
	defer func() {
		tx.Rollback()
	}()
	//回填比赛信息 以及反填胜率
	officalMatch, err := cache.GetOnTimeFootballMatch(order.LotteryUuid)
	if officalMatch == nil {
		common.FailedReturn(c, "查公布信息异常， 请联系管理员！")
		return errors.New("查公布信息异常， 请联系管理员！")
	}
	fillStatus := fillMatches(*officalMatch, order, c, tx)
	if fillStatus == nil {
		return errors.New("查公布信息异常， 请联系管理员！")
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
				fmt.Printf("联赛：%s \n", view.League)
				fmt.Printf("%s@%f \n", view.Mode, view.Odd)
				fmt.Println("----------------------------------")
			}
			fmt.Println("倍数：", order.Times)
			fmt.Println("单个组合奖金：", bet.Bonus)

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
	var logicCount = bonusCout * float32(order.Times)
	if order.LogicWinMaX != logicCount {
		log.Warn("逻辑奖金和后台算出对不上", order.LogicWinMaX, logicCount)
	}
	order.LogicWinMaX = logicCount
	order.ShouldPay = float32(2 * len(bonus) * order.Times)
	order.CreatedAt = time.Now()
	fmt.Println("实际付款：", order.ShouldPay)
	if order.AllWinId == 0 {
		billErr := user.CheckScoreOrDoBill(order.UserID, order.ShouldPay, true, tx)
		if err != nil {
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

	fmt.Println("逻辑总奖金: ", order.LogicWinMaX)
	fmt.Println("=========================================")

	CheckLottery(util.AddTwoHToTime(order.Matches[len(order.Matches)-1].TimeDate))
	//cache.Remove(order.LotteryUuid)
	tx.Commit()
	common.SuccessReturn(c, order.UUID)
	return nil
}
func fillMatches(games cache.FootBallGames, order *Order, c *gin.Context, tx *gorm.DB) *Order {
	if len(order.Matches) <= 0 {
		return nil
	}
	var mapper = games.MatchListToMap()
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
			order.Matches[index].AwayTeamCode = matchMapper.AwayTeamCode
			order.Matches[index].AwayTeamId = strconv.Itoa(matchMapper.AwayTeamId)
			order.Matches[index].AwayTeamName = matchMapper.AwayTeamAbbName
			order.Matches[index].HomeTeamId = strconv.Itoa(matchMapper.HomeTeamId)
			order.Matches[index].HomeTeamAllName = matchMapper.HomeTeamAllName
			order.Matches[index].HomeTeamCode = matchMapper.HomeTeamCode
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
					odd, err := FindOdd(order.Matches[index].MatchId, &order.Matches[index].Combines[in], mapper)
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

func (order *Order) WayDetail() (map[string]interface{}, error) {

	ways := strings.Split(order.Way, ",")
	oddCombines := make(map[string]interface{})
	data := cache.Get(order.LotteryUuid).(string)
	var poolMap map[int]cache.Pool
	if strings.Compare(order.LotteryType, "BASKETBALL") == 0 {
		var game cache.BasketBallGames
		jerr := json.Unmarshal([]byte(data), &game)
		if jerr != nil {
			log.Error(jerr)
			return oddCombines, errors.New("缓存解析失败")
		}
		poolMap = game.GetSinglePoolMap()

	} else {
		var game cache.LotteryResult
		jerr := json.Unmarshal([]byte(data), &game)
		if jerr != nil {
			log.Error(jerr)
			return oddCombines, errors.New("缓存解析失败")
		}
		poolMap = game.Content.GetSinglePoolMap()
	}

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
								Way:          "过关方式 1x1",
								Time:         match.MatchNumStr,
								League:       fmt.Sprintf("主队:%s Vs 客队:%s", match.HomeTeamName, match.AwayTeamName),
								Mode:         GetDesc(combine.Type, combine.ScoreVsScore),
								Odd:          combine.Odds,
								Type:         combine.Type,
								ScoreVsScore: combine.ScoreVsScore,
								GoalLine:     combine.GoalLine,
								Check:        false,
								Correct:      false,
							}},
							Way:      "过关方式 1x1",
							MatchId:  match.MatchId,
							MatchNum: match.MatchNum,
							Check:    false,
							Win:      false,
						})

					}
				}
			}
			oddCombines["1x1"] = Single
			return oddCombines, nil
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
					var bonus float32 = 2
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 2x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Check:    false,
						Win:      false,
						Bonus:    bonus,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["2x1"] = bets
			return oddCombines, nil
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
					var bonus float32 = 2
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
						Check:    false,
						Win:      false,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["3x1"] = bets
			return oddCombines, nil
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
					var bonus float32 = 2
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 4x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Bonus:    bonus,
						Check:    false,
						Win:      false,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["4x1"] = bets
			return oddCombines, nil
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
					var bonus float32 = 2
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 5x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Bonus:    bonus,
						Check:    false,
						Win:      false,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["5x1"] = bets
			return oddCombines, nil
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
					var bonus float32 = 2
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 6x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Bonus:    bonus,
						Check:    false,
						Win:      false,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["6x1"] = bets
			return oddCombines, nil
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
					var bonus float32 = 2
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 7x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Bonus:    bonus,
						Check:    false,
						Win:      false,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["7x1"] = bets
			return oddCombines, nil
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
					var bonus float32 = 2
					for _, view := range views {
						bonus = bonus * view.Odd
					}
					bet := Bet{
						OrderId:  order.UUID,
						Way:      "过关方式 8x1",
						Group:    views,
						MatchId:  views[0].MatchId,
						MatchNum: views[0].MatchNum,
						Bonus:    bonus,
						Check:    false,
						Win:      false,
					}
					bets = append(bets, bet)
				}

			}
			oddCombines["8x1"] = bets
			return oddCombines, nil
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
func FollowOrder(c *gin.Context) {
	var follow FollowDto
	c.BindJSON(&follow)

	if len(follow.OrderId) <= 0 {
		common.FailedReturn(c, "订单id不能为空")
	}
	order := FindById(follow.OrderId, true)
	order.UUID = ""
	if len(order.Matches) > 0 {
		for _, match := range order.Matches {
			match.OrderId = ""
			match.ID = 0
			if len(match.Combines) > 0 {
				for _, combine := range match.Combines {
					combine.ID = 0
				}
			}
		}
	}

	order.Times = follow.Times

	if strings.Compare(order.LotteryType, FOOTBALL) == 0 || strings.Compare(order.LotteryType, BASKETBALL) == 0 {
		var uid = uuid.NewV4().String()
		cache.GetOnTimeFootballMatch(uid)
		order.LotteryUuid = uid
	} else {
		order.LotteryUuid = "8888888"
	}
	order.Share = false
	orderCreateFunc(c, &order)
}

func GetOrderByLotteryType(tp string) []Order {
	var orders []Order
	if err := mysql.DB.Model(Order{}).Where(&Order{LotteryType: tp, AllMatchFinished: false}).Find(&orders).Error; err != nil {
		log.Error(err)
		return nil
	}
	return orders
}

func CreatePLW(ord *Order) error {

	if len(ord.Content) <= 0 {
		return errors.New("选号不能为空")
	}

	if len(ord.IssueId) <= 0 {
		return errors.New("订单期号不能为空")
	}

	var tp = 0
	if strings.Compare(ord.LotteryType, "P3") == 0 {
		tp = 3
	}

	if strings.Compare(ord.LotteryType, "P5") == 0 {
		tp = 5
	}
	//校验期号
	drawNum, cerr := lottery.GetPLWIssueId()
	fmt.Println("排列最新期号=================>", drawNum)
	issueId, err := strconv.Atoi(ord.IssueId)
	if err != nil || cerr != nil {
		log.Error(err)
		return errors.New("校验期号失败")
	}
	if issueId-drawNum != 1 {
		return errors.New("购买期号不正确")
	}

	tx := mysql.DB.Begin()
	if strings.Contains(ord.Content, ",") {
		arr := strings.Split(ord.Content, ",")

		ord.ShouldPay = float32(len(arr) * 2 * ord.Times)
		if ord.LotteryType == P3 {
			for _, s := range arr {
				numArr := strings.Split(s, " ")
				if ord.PL3Way == PL_SIGNAL || ord.PL3Way == PL_C3 || ord.PL3Way == PL_C6 {
					if len(numArr) != tp {
						return errors.New("选号存在问题")
					}
				}
				for _, s2 := range numArr {
					num, err := strconv.Atoi(s2)
					if err != nil {
						log.Error(err)
						return errors.New("号码存在异常")
					}
					if !(0 <= num && num <= 9) {
						return errors.New("号码存在异常,数字不在0-9 之间")
					}
				}
			}
		}
	} else {
		numArr := strings.Split(ord.Content, " ")
		if ord.LotteryType == P5 && ord.PL3Way == PL_SIGNAL {
			if len(numArr) != tp {
				return errors.New("选号存在问题")
			}
		}

		for _, s2 := range numArr {
			num, err := strconv.Atoi(s2)
			if err != nil {
				log.Error(err)
				return errors.New("号码存在异常")
			}
			if !(0 <= num && num <= 9) {
				return errors.New("号码存在异常,数字不在0-9 之间")
			}
		}
		ord.ShouldPay = float32(1 * 2 * ord.Times)

	}
	if ord.AllWinId == 0 {
		billErr := user.CheckScoreOrDoBill(ord.UserID, ord.ShouldPay, true, tx)
		if billErr != nil {
			log.Error("扣款失败， 无法提交订单")
			tx.Rollback()
			return billErr
		}
		ord.PayStatus = true
	}
	if err := tx.Create(ord).Error; err != nil {
		log.Error(err)
		tx.Rollback()
		return errors.New("保存订单失败")
	}
	if !lottery.LotteryStatistics.Exists("plw_check") {
		lottery.LotteryStatistics.Add("plw_check", 8*time.Hour, 1)
		AddPlwCheck(tp, nil)
	}
	tx.Commit()
	return nil
}

func AddPlwCheck(p int, when *time.Time) {
	log.Info("============ 排列对账任务开启==============")
	var job Job
	switch p {
	case 3:
		job = Job{
			Time:  util.GetPLWFinishedTime().Add(320 * time.Second),
			Param: nil,
			CallBack: func(param interface{}) {
				log.Info("========排列3对账任务执行=============")
				resp, err := http.Get(lottery.PLW_URL)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				var result lottery.Plw
				err = json.Unmarshal(body, &result)
				if err != nil || &result.Value == nil {
					log.Error("转换排列5结果为对象失败", err)

					return
				}
				orders := GetOrderByLotteryType("P3")
				//用于映射是否有这个期号
				var mapper = make(map[string]int)
				for i, s := range result.Value.List {
					_, ok := mapper[s.LotteryDrawNum]
					if !ok {
						mapper[s.LotteryDrawNum] = i
					}
				}
				tx := mysql.DB.Begin()
				if len(orders) > 0 {
					for _, o := range orders {
						value, ok := mapper[o.IssueId]
						if ok {

							content, types := GetPlAllNums(&o)
							releaseNum := result.Value.List[value].LotteryDrawResult[0:3]
							releaseArr := strings.Split(releaseNum, " ")
							for _, s := range content {
								arr := strings.Split(s, " ")
								if strings.Compare(s, releaseNum) == 0 {
									if strings.Compare(types, PL_SIGNAL) == 0 {
										o.Bonus = o.Bonus + 1040
									}
									if strings.Compare(types, PL_C3) == 0 {
										numArr := util.CovertStrArrToInt(arr)
										sendWorld := util.GetCombine3(numArr)
										release := util.CovertStrArrToInt(releaseArr)
										win := 0
										for _, ints := range sendWorld {
											count := 0
											for i, value := range ints {
												if release[i] == value {
													count++
													continue
												} else {
													break
												}
											}
											if count == 3 {
												win = 3
												break
											}
										}
										if win == 3 {
											o.Bonus = o.Bonus + 346
										}
										o.Way = fmt.Sprintf("%s 组合3：%s 中奖", o.Way, s)
									}
									if strings.Compare(types, PL_C6) == 0 {
										numArr := util.CovertStrArrToInt(arr)
										sendWorld := util.Permute(numArr)
										release := util.CovertStrArrToInt(releaseArr)
										win := 0
										for _, ints := range sendWorld {
											count := 0
											for i, value := range ints {
												if release[i] == value {
													count++
													continue
												} else {
													break
												}
											}
											if count == 3 {
												win = 3
												break
											}
										}
										if win == 3 {
											o.Bonus = o.Bonus + 173
										}
										o.Way = fmt.Sprintf("%s 组合6：%s 中奖", o.Way, s)
									}
									o.Win = true
								}
							}
							//TODO 组合3和组合6计算
							o.AllMatchFinished = true
						} else {
							continue
						}
						if o.Win == true {
							o.BonusStatus = NO_PAY
							o.Bonus = o.Bonus * float32(o.Times)
						} else {
							o.BonusStatus = NO_BONUS
						}
						tx.Save(o)
					}
				}
				tx.Commit()
				if param != nil {
					id := fmt.Sprintf("%d", param)
					err := mysql.DB.Model(JobExecution{}).Where("id = ?", id).Update("status", true).Error
					if err != nil {
						log.Error("更新job状态失败!")
					}
				}
			},
			Type: P3,
		}

		break
	case 5:
		job = Job{
			Time:  util.GetPLWFinishedTime().Add(320 * time.Second),
			Param: nil,
			CallBack: func(param interface{}) {
				log.Info("========排列5对账任务执行=============")
				resp, err := http.Get(lottery.PLW_URL)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				var result lottery.Plw
				err = json.Unmarshal(body, &result)
				if err != nil || &result.Value == nil {
					log.Error("转换排列5结果为对象失败", err)

					return
				}
				orders := GetOrderByLotteryType("P5")
				//用于映射是否有这个期号
				var mapper map[string]int = make(map[string]int, 0)
				for i, s := range result.Value.List {
					_, ok := mapper[s.LotteryDrawNum]
					if !ok {
						mapper[s.LotteryDrawNum] = i
					}
				}
				tx := mysql.DB.Begin()
				if len(orders) > 0 {
					for _, o := range orders {
						index, ok := mapper[o.IssueId]
						if ok {
							content, err := getArr(o.Content, o.LotteryType, o.PL3Way)
							if err != nil {
								log.Error(err)
								continue
							}
							releaseNum := result.Value.List[index].LotteryDrawResult
							for _, s := range content {
								if strings.Compare(s, releaseNum) == 0 {
									o.Bonus = o.Bonus + 100000
									o.Win = true
								}
							}
							o.AllMatchFinished = true
							if o.Win == true {
								o.BonusStatus = NO_PAY
								o.Bonus = o.Bonus * float32(o.Times)
							} else {
								o.BonusStatus = NO_BONUS
							}

						}
						tx.Save(o)
					}
				}
				tx.Commit()
				if param != nil {
					id := fmt.Sprintf("%d", param)
					err := mysql.DB.Model(JobExecution{}).Where("id = ?", id).Update("status", true).Error
					if err != nil {
						log.Error("更新job状态失败!")
					}
				}
			},
			Type: P5,
		}
		break
	}
	if when != nil {
		job.Time = *when
	}
	AddJob(job)

}

// 通过 level:“一等奖” 获取奖金(append: 是否追加) 一等奖(追加) 二等奖(追加)
func getStakeAmountByPrizeLevel(prizes []lottery.PrizeLevel, level string) int {
	for _, s := range prizes {
		if s.PrizeLevel == level {
			newStr := strings.ReplaceAll(s.StakeAmount, ",", "")
			v, _ := strconv.Atoi(newStr)
			return v
		}
	}
	return 0
}

func AddSuperLottoCheck(when *time.Time) {
	log.Info("============= 大乐透对账任务开启==============")
	var job Job

	job = Job{
		Time:  util.GetPLWFinishedTime(),
		Param: nil,
		CallBack: func(param interface{}) {
			log.Info("========大乐透对账任务执行=============")
			/*week := time.Now().Weekday()
			if !(week == 1 || week == 3 || week == 6) {
				log.Info("========== 不是 1 3 6 不检测大乐透============")
				return
			}*/
			resp, err := http.Get(lottery.SUPER_LOTTO_URL)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result lottery.SuperLottery
			err = json.Unmarshal(body, &result)
			if err != nil || &result.Value == nil {
				log.Error("转换大乐透结果为对象失败", err)

				return
			}
			//用于映射是否有这个期号
			var mapper = make(map[string]int)
			for i, s := range result.Value.List {
				_, ok := mapper[s.LotteryDrawNum]
				if !ok {
					mapper[s.LotteryDrawNum] = i
				}
			}
			orders := GetOrderByLotteryType("SUPER_LOTTO")
			tx := mysql.DB.Begin()
			if len(orders) > 0 {
				for _, o := range orders {
					index, ok := mapper[o.IssueId]

					if ok {
						content, err := getArr(o.Content, o.LotteryType, o.Way)
						if err != nil {
							log.Error(err)
							continue
						}
						releaseNum := result.Value.List[index].LotteryDrawResult

						for _, s := range content {
							yes, count := NoOrderHeaderCompare(5, 5, s, releaseNum)
							indexR := []int{5, 6}
							state, countR := NoOrderCompareTailDirectNum(indexR, 1, s, releaseNum)
							if count == 5 && countR == 2 {
								//一等奖
								bonus := getStakeAmountByPrizeLevel(result.Value.List[index].PrizeLevelList, "一等奖")
								o.Bonus = o.Bonus + float32(bonus)
								if o.Append {
									//	一等奖(追加) 二等奖(追加)
									bonusAppend := getStakeAmountByPrizeLevel(result.Value.List[index].PrizeLevelList, "一等奖(追加)")
									o.Bonus = o.Bonus + float32(bonusAppend)
								}
								o.Win = true
								o.Comment = "一等奖"
								continue
							}

							if yes && state {
								//前5相同 后面两个任意一个相同
								bonus := getStakeAmountByPrizeLevel(result.Value.List[index].PrizeLevelList, "二等奖")
								o.Bonus = o.Bonus + float32(bonus)
								if o.Append {
									//	一等奖(追加) 二等奖(追加)
									bonusAppend := getStakeAmountByPrizeLevel(result.Value.List[index].PrizeLevelList, "二等奖(追加)")
									o.Bonus = o.Bonus + float32(bonusAppend)
								}
								o.Win = true
								o.Comment = "二等奖"
								continue
							}
							if yes {
								//五个前区号码相同
								o.Bonus = o.Bonus + 10000
								o.Win = true
								o.Comment = "三等奖"
								continue
							}

							//任意四个前区号码及两个后区号码相同
							if count == 4 && countR == 2 {
								o.Bonus = o.Bonus + 3000
								o.Win = true
								o.Comment = "四等奖"
								continue
							}

							//任意前区4个号码以及后区一个
							if count == 4 && countR == 1 {
								o.Bonus = o.Bonus + 300
								o.Win = true
								o.Comment = "五等奖"
								continue
							}

							if 3 == count && countR == 2 {
								o.Bonus = o.Bonus + 200
								o.Win = true
								o.Comment = "六等奖"
								continue
							}

							if 4 == count {
								o.Bonus = o.Bonus + 100
								o.Win = true
								o.Comment = "七等奖"
								continue
							}

							if 3 == count && countR == 1 {
								o.Bonus = o.Bonus + 15
								o.Win = true
								o.Comment = "八等奖"
								continue
							}

							if 3 == count || (1 == count && 2 == countR) || (2 == count && 1 == countR) || countR == 2 {
								o.Bonus = o.Bonus + 5
								o.Win = true
								o.Comment = "九等奖"
								continue
							}
							o.Win = false
						}
					} else {
						continue
					}
					o.AllMatchFinished = true
					if o.Win == true {
						o.BonusStatus = NO_PAY
						o.Bonus = o.Bonus * float32(o.Times)
					} else {
						o.BonusStatus = NO_BONUS
					}

					tx.Save(&o)
				}
			}
			tx.Commit()
			if param != nil {
				id := fmt.Sprintf("%d", param)
				err := mysql.DB.Model(JobExecution{}).Where("id = ?", id).Update("status", true).Error
				if err != nil {
					log.Error("更新job状态失败!")
				}
			}
		},
		Type: SUPER_LOTTO,
	}
	if when != nil {
		job.Time = *when
	}
	AddJob(job)

}

func AddSevenStarCheck(when *time.Time) {

	log.Info("============= 七星彩对账任务开启==============")
	var job Job

	job = Job{
		Time:  util.GetPLWFinishedTime(),
		Param: nil,
		CallBack: func(param interface{}) {
			log.Info("========七星彩对账任务执行=============")
			/*week := time.Now().Weekday()
			if !(week == 0 || week == 2 || week == 5) {
				log.Info("========== 不是 0 2 5 不检测七星彩============")
				return
			}*/
			resp, err := http.Get(lottery.SEVEN_START_URL)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result lottery.SevenStar
			err = json.Unmarshal(body, &result)
			if err != nil || &result.Value == nil {
				log.Error("转换大乐透结果为对象失败", err)

				return
			}
			orders := GetOrderByLotteryType("SEVEN_STAR")
			//用于映射是否有这个期号
			var mapper = make(map[string]int)
			for i, s := range result.Value.List {
				_, ok := mapper[s.LotteryDrawNum]
				if !ok {
					mapper[s.LotteryDrawNum] = i
				}
			}
			tx := mysql.DB.Begin()
			if len(orders) > 0 {
				for _, o := range orders {
					index, ok := mapper[o.IssueId]
					if ok {
						content, err := getArr(o.Content, o.LotteryType, o.Way)
						if err != nil {
							log.Error(err)
							continue
						}
						releaseNum := result.Value.List[index].LotteryDrawResult
						for _, s := range content {
							if strings.Compare(s, releaseNum) == 0 {
								//一等奖
								o.Bonus = o.Bonus + 5000000
								o.Win = true
								o.Way = "一等奖"
								continue
							}
							yes, count := randomNumBeforeDirect(6, 6, s, releaseNum)
							if yes {
								//前6相同
								o.Bonus = o.Bonus + 2000000
								o.Win = true
								o.Way = fmt.Sprintf("%s + %s", o.Way, "二等奖")
								continue
							}

							//投注号码前6位中的任意5个数字与开奖号码对应位置数字相同且最后一个数字与开奖号码对应位置数字相同，即中奖

							index2 := []int{6}
							yr, _ := CompareDirectNum(index2, 1, s, releaseNum)
							if count == 5 && yr {
								o.Bonus = o.Bonus + 3000
								o.Win = true
								o.Way = fmt.Sprintf("%s + %s", o.Way, "三等奖")
								continue
							}
							_, countR := randomNumBeforeDirect(7, 5, s, releaseNum)
							if countR == 5 {
								o.Bonus = o.Bonus + 500
								o.Win = true
								o.Way = fmt.Sprintf("%s + %s", o.Way, "四等奖")
								continue
							}

							if 4 == countR {
								o.Bonus = o.Bonus + 30
								o.Win = true
								o.Way = fmt.Sprintf("%s + %s", o.Way, "五等奖")
								continue
							}

							if 3 == countR || (1 == countR && yr) || yr {
								o.Bonus = o.Bonus + 5
								o.Win = true
								o.Way = fmt.Sprintf("%s + %s", o.Way, "六等奖")
								continue
							}
							o.Win = false
						}
					} else {
						continue
					}
					o.AllMatchFinished = true
					if o.Win == true {
						o.BonusStatus = NO_PAY
						o.Bonus = o.Bonus * float32(o.Times)
					} else {
						o.BonusStatus = NO_BONUS
					}
					tx.Save(o)
				}
			}
			tx.Commit()
			if param != nil {
				id := fmt.Sprintf("%d", param)
				err := mysql.DB.Model(JobExecution{}).Where("id = ?", id).Update("status", true).Error
				if err != nil {
					log.Error("更新job状态失败!")
				}
			}
		},
		Type: SEVEN_STAR,
	}
	if when != nil {
		job.Time = *when
	}
	AddJob(job)

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

// 前区对比号码（顺序不限），length 对比前面多少个数  是否满足num个
func NoOrderHeaderCompare(length int, num int, userNum string, releaseNum string) (bool, int) {

	var count = 0
	numBuffer := strings.Split(userNum, " ")
	releaseBuffer := strings.Split(releaseNum, " ")
	for i := 0; i < length; i++ {
		for j := 0; j < 5; j++ {
			if util.PaddingZeroCompare(numBuffer[i], releaseBuffer[j]) {
				count += 1
			}
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

// 后区对比号码（没有顺序）， length 个数  是否满足num个
func NoOrderCompareTailDirectNum(index []int, number int, userNum string, releaseNum string) (bool, int) {
	var count = 0
	numBuffer := strings.Split(userNum, " ")
	releaseBuffer := strings.Split(releaseNum, " ")
	for i := 0; i < len(index); i++ {
		for j := 0; j < 5; j++ {
			if util.PaddingZeroCompare(numBuffer[index[i]], releaseBuffer[j]) {
				count += 1
			}
		}

	}
	if count >= number {
		return true, count
	}
	return false, count
}

// 按位组合
func GetIndexCmn(ii int, arr [][]string, trace *[]string, res *[][]string) {
	if len(*trace) == 7 {
		temp := make([]string, len(*trace))
		copy(temp, *trace)
		*res = append(*res, temp)
		return
	}
	for i := ii; i < len(arr); i++ {
		for j := 0; j < len(arr[i]); j++ {
			*trace = append(*trace, arr[i][j])
			GetIndexCmn(i+1, arr, trace, res)
			// 撤销选择
			*trace = (*trace)[:len(*trace)-1]
		}
	}
}

// 大乐透：DIRECT 直选/随机;胆拖（DT，DTQQ 前区胆拖、 DTHQ 后区胆拖、 DTSQ 双区胆拖）;复式（FS， 前区复式 FSQQ、后区复式 FSHQ、双区复式 FSSQ
// 七星彩：复式(FSSTAR),DIRECT 直选/随机;
// ==========大乐透========
// 复式前区大于5个数字，后区 大于2个,只有QQ 为前区复式 ， 只有HQ 为后区复式  都有为双区复式
// 复式：
// QQ:01 02 04 05 11 12 35 33,HQ:06 07 12
// 胆拖：
// QQD:09,QQT:01 02 06 07 09,HQD:12 HQT:02 08
// 后区胆拖
// 03 05 08 09 11 13,HQD:12,HQT:02 08
// ==========七星彩=================
// 复式：
// I1:3 4,I2:0 8,I3:8 9,I4:0 8,I5: 1 7,I6:3,I7:7 8
func getArr(content string, ty string, way string) ([]string, error) {

	switch ty {
	case P3:
	case P5:
		break
	case SEVEN_STAR: // I1:3 4,I2:0 8,I3:8 9,I4:0 8,I5: 1 7,I6:3,I7:7 8
		switch way {
		case FSSTAR:
			arr := strings.Split(content, ",")
			var tmp = [][]string{}
			for _, s := range arr {
				tmp = append(tmp, strings.Split(s[3:], " "))
			}
			var combines [][]string
			var strace []string
			GetIndexCmn(0, tmp, &strace, &combines)
			var realNums []string
			for _, combine := range combines {
				realNums = append(realNums, strings.Join(combine, " "))
			}
			return realNums, nil
		case DIRECT:
		default:
			if strings.Contains(content, ",") {
				return strings.Split(content, ","), nil
			} else {
				var strs []string
				strs = append(strs, content)
				return strs, nil
			}
			break

		}

		break
	case SUPER_LOTTO:
		switch way {
		case DTQQ: //前区胆拖 QQD:09,QQT:01 02 06 07 09,01 02
			var qqd, qqt, hq []string
			if !(strings.Contains(content, QQD) && strings.Contains(content, QQT)) {
				return nil, errors.New("前区胆拖数据错误")
			}
			arr := strings.Split(content, ",")
			if len(arr) != 3 {
				return nil, errors.New("前区胆拖数据错误")
			}
			qqdStr := strings.ReplaceAll(arr[0], "QQD:", "")
			qqd = strings.Split(qqdStr, " ")
			qqtStr := strings.ReplaceAll(arr[1], "QQT:", "")
			qqt = strings.Split(qqtStr, " ")
			hq = strings.Split(arr[2], " ")
			if len(qqd) <= 0 || len(qqt) < 0 || len(qqd)+len(qqt) <= 5 || len(hq) != 2 {
				return nil, errors.New("前区胆拖数据错误,胆拖位数不对！")
			}

			k := 5 - len(qqd)
			numArr := util.CombineArray(qqt, k)
			var realNums []string
			for _, v := range numArr {
				numStr := qqdStr + strings.Join(v, " ") + " " + strings.Join(hq, " ")
				realNums = append(realNums, numStr)
			}
			return realNums, nil
		case DTHQ: //后区胆拖 01 02 03 04 05,HQD:01,HQT:12 11 10
			var qq, hqt []string
			arr := strings.Split(content, ",")
			qq = strings.Split(arr[0], " ")
			hqdStr := strings.ReplaceAll(arr[1], "HQD:", "")
			hqtStr := strings.ReplaceAll(arr[2], "HQT:", "")
			hqt = strings.Split(hqtStr, " ")
			if len(qq) != 5 || len(hqdStr) != 2 || len(hqt) <= 1 {
				return nil, errors.New("后区胆拖数据错误")
			}
			var realNums []string
			for _, s := range hqt {
				realStr := arr[0] + " " + hqdStr + " " + s
				realNums = append(realNums, realStr)
			}
			return realNums, nil
		case DTSQ: //QQD:09,QQT:01 02 06 07 09,HQD:01,HQT:12 11 10
			var qqd, qqt, hqt []string
			arr := strings.Split(content, ",")
			qqdStr := strings.ReplaceAll(arr[0], "QQD:", "")
			qqd = strings.Split(qqdStr, " ")
			qqtStr := strings.ReplaceAll(arr[1], "QQT:", "")
			qqt = strings.Split(qqtStr, " ")
			hqdStr := strings.ReplaceAll(arr[2], "HQD:", "")
			hqtStr := strings.ReplaceAll(arr[2], "HQT:", "")
			hqt = strings.Split(hqtStr, " ")
			if len(qqd)+len(qqt) <= 5 || len(hqt) <= 1 || len(hqdStr) != 2 {
				return nil, errors.New("双区胆拖数据错误")
			}
			var realNums []string
			kqq := 5 - len(qqd)
			qqArr := util.CombineArray(qqt, kqq)
			var buffer []string
			for _, i := range qqArr {
				str := qqdStr + " " + strings.Join(i, " ")
				buffer = append(buffer, str)
			}
			for _, b := range buffer {
				for _, s := range hqt {
					realNum := b + " " + hqdStr + " " + s
					realNums = append(realNums, realNum)
				}
			}
			return realNums, nil
		case FSQQ: //QQ:01 02 04 05 11 12 35 33,01 12
			var qq, tail []string
			arr := strings.Split(content, ",")
			qqStr := strings.ReplaceAll(arr[0], "QQ:", "")
			qq = strings.Split(qqStr, " ")
			tail = strings.Split(arr[1], " ")
			if len(qq) <= 5 {
				return nil, errors.New("复式前区，位数不够！")
			}
			if len(tail) != 2 {
				return nil, errors.New("复式前区，后区位数不够！")
			}
			res := util.CombineArray(qq, 5)
			var realNums []string
			for _, re := range res {
				realStr := strings.Join(re, " ") + " " + strings.Join(tail, " ")
				realNums = append(realNums, realStr)
			}
			return realNums, nil

		case FSHQ: //01 02 03 05 09,HQ:06 07 12
			var qq, tail []string
			arr := strings.Split(content, ",")
			tailStr := strings.ReplaceAll(arr[1], "HQ:", "")
			tail = strings.Split(tailStr, " ")
			qq = strings.Split(arr[0], " ")
			if len(qq) != 5 {
				return nil, errors.New("复式后区，前区位数不够2位！")
			}
			if len(tail) <= 2 {
				return nil, errors.New("复式前区，后区位数不够3位！")
			}
			res := util.CombineArray(tail, 2)
			var realNums []string
			for _, re := range res {
				realStr := arr[0] + " " + strings.Join(re, " ")
				realNums = append(realNums, realStr)
			}
			return realNums, nil

		case FSSQ: // QQ:01 02 04 05 11 12 35 33,HQ:06 07 12
			var qq, tail []string
			arr := strings.Split(content, ",")
			qqStr := strings.ReplaceAll(arr[0], "QQ:", "")
			qq = strings.Split(qqStr, " ")
			tailStr := strings.ReplaceAll(arr[1], "HQ:", "")
			tail = strings.Split(tailStr, " ")
			if len(qq) <= 5 {
				return nil, errors.New("复式双区，位数不够！")
			}
			if len(tail) <= 2 {
				return nil, errors.New("复式双区，后区位数不够！")
			}
			resHeader := util.CombineArray(qq, 5)
			resTail := util.CombineArray(tail, 2)
			var realNums []string
			for _, re := range resHeader {
				for _, t := range resTail {
					realStr := strings.Join(re, " ") + " " + strings.Join(t, " ")
					realNums = append(realNums, realStr)
				}

			}
			return realNums, nil
		case DIRECT:
		default:
			if strings.Contains(content, ",") {
				return strings.Split(content, ","), nil
			} else {
				var strs []string
				strs = append(strs, content)
				return strs, nil
			}
			break

		}

		break
	}
	return make([]string, 0), nil
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
