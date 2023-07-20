package lottery

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"io"
	"jingcai/common"
	alog "jingcai/log"
	"jingcai/mysql"
	"jingcai/order"
	"jingcai/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var log = alog.Logger

const SUPER_LOTTERY_POOL = "super_lottery_pool"
const PLW_POOL = "plw_pool"
const BASKET_BALL_COUNT = "basket_ball_count"
const FOOT_BALL_COUNT = "foot_ball_count"
const SEVEN_STAR_POOL = "seven_star_pool"
const SEVEN_START_URL = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListV1.qry?gameNo=04&provinceId=0&isVerify=1&termLimits=13"
const PLW_URL = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListPlwV1.qry?gameNo=350133&provinceId=0&isVerify=1&termLimits=5"

// 大乐透
type SuperLottery struct {
	DataFrom     string `json:"dataFrom"`
	EmptyFlag    bool   `json:"emptyFlag"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
	Value        struct {
		LastPoolDraw struct {
			LotteryDrawNum       string `json:"lotteryDrawNum"`
			LotteryDrawResult    string `json:"lotteryDrawResult"`
			LotteryDrawTime      string `json:"lotteryDrawTime"`
			LotteryGameName      string `json:"lotteryGameName"`
			LotteryGameNum       string `json:"lotteryGameNum"`
			PoolBalanceAfterdraw string `json:"poolBalanceAfterdraw"`
			PrizeLevelList       []struct {
				AwardType        int    `json:"awardType"`
				Group            string `json:"group"`
				LotteryCondition string `json:"lotteryCondition"`
				PrizeLevel       string `json:"prizeLevel"`
				Sort             int    `json:"sort"`
				StakeAmount      string `json:"stakeAmount"`
				StakeCount       string `json:"stakeCount"`
				TotalPrizeamount string `json:"totalPrizeamount"`
			} `json:"prizeLevelList"`
		} `json:"lastPoolDraw"`
		List []struct {
			DrawFlowFund            string        `json:"drawFlowFund"`
			DrawFlowFundRj          string        `json:"drawFlowFundRj"`
			DrawPdfUrl              string        `json:"drawPdfUrl"`
			EstimateDrawTime        string        `json:"estimateDrawTime"`
			IsGetKjpdf              int           `json:"isGetKjpdf"`
			IsGetXlpdf              int           `json:"isGetXlpdf"`
			LotteryDrawNum          string        `json:"lotteryDrawNum"`
			LotteryDrawResult       string        `json:"lotteryDrawResult"`
			LotteryDrawStatus       int           `json:"lotteryDrawStatus"`
			LotteryDrawStatusNo     string        `json:"lotteryDrawStatusNo"`
			LotteryDrawTime         string        `json:"lotteryDrawTime"`
			LotteryEquipmentCount   int           `json:"lotteryEquipmentCount"`
			LotteryGameName         string        `json:"lotteryGameName"`
			LotteryGameNum          string        `json:"lotteryGameNum"`
			LotteryGamePronum       int           `json:"lotteryGamePronum"`
			LotteryNotice           int           `json:"lotteryNotice"`
			LotteryNoticeShowFlag   int           `json:"lotteryNoticeShowFlag"`
			LotteryPaidBeginTime    string        `json:"lotteryPaidBeginTime"`
			LotteryPaidEndTime      string        `json:"lotteryPaidEndTime"`
			LotteryPromotionFlag    int           `json:"lotteryPromotionFlag"`
			LotteryPromotionFlagRj  int           `json:"lotteryPromotionFlagRj"`
			LotterySaleBeginTime    string        `json:"lotterySaleBeginTime"`
			LotterySaleEndTimeUnix  int           `json:"lotterySaleEndTimeUnix"`
			LotterySaleEndtime      string        `json:"lotterySaleEndtime"`
			LotterySuspendedFlag    int           `json:"lotterySuspendedFlag"`
			LotteryUnsortDrawresult string        `json:"lotteryUnsortDrawresult"`
			MatchList               []interface{} `json:"matchList"`
			PdfType                 int           `json:"pdfType"`
			PoolBalanceAfterdraw    string        `json:"poolBalanceAfterdraw"`
			PoolBalanceAfterdrawRj  string        `json:"poolBalanceAfterdrawRj"`
			PrizeLevelList          []struct {
				AwardType        int    `json:"awardType"`
				Group            string `json:"group"`
				LotteryCondition string `json:"lotteryCondition"`
				PrizeLevel       string `json:"prizeLevel"`
				Sort             int    `json:"sort"`
				StakeAmount      string `json:"stakeAmount"`
				StakeCount       string `json:"stakeCount"`
				TotalPrizeamount string `json:"totalPrizeamount"`
			} `json:"prizeLevelList"`
			PrizeLevelListRj  []interface{} `json:"prizeLevelListRj"`
			RuleType          int           `json:"ruleType"`
			SurplusAmount     string        `json:"surplusAmount"`
			SurplusAmountRj   string        `json:"surplusAmountRj"`
			TermList          []interface{} `json:"termList"`
			TermResultList    []interface{} `json:"termResultList"`
			TotalSaleAmount   string        `json:"totalSaleAmount"`
			TotalSaleAmountRj string        `json:"totalSaleAmountRj"`
			Verify            int           `json:"verify"`
			VtoolsConfig      struct {
			} `json:"vtoolsConfig"`
		} `json:"list"`
		PageNo   int `json:"pageNo"`
		PageSize int `json:"pageSize"`
		Pages    int `json:"pages"`
		Total    int `json:"total"`
	} `json:"value"`
}

// 排列5
type Plw struct {
	DataFrom     string `json:"dataFrom"`
	EmptyFlag    bool   `json:"emptyFlag"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
	Value        struct {
		List []struct {
			DrawFlowFund            string        `json:"drawFlowFund"`
			DrawFlowFundRj          string        `json:"drawFlowFundRj"`
			DrawPdfUrl              string        `json:"drawPdfUrl"`
			DrawPdfUrlPls           string        `json:"drawPdfUrlPls"`
			EstimateDrawTime        string        `json:"estimateDrawTime"`
			IsGetKjpdf              int           `json:"isGetKjpdf"`
			IsGetXlpdf              int           `json:"isGetXlpdf"`
			LotteryDrawNum          string        `json:"lotteryDrawNum"`
			LotteryDrawResult       string        `json:"lotteryDrawResult"`
			LotteryDrawStatus       int           `json:"lotteryDrawStatus"`
			LotteryDrawStatusNo     string        `json:"lotteryDrawStatusNo"`
			LotteryDrawTime         string        `json:"lotteryDrawTime"`
			LotteryEquipmentCount   int           `json:"lotteryEquipmentCount"`
			LotteryGameName         string        `json:"lotteryGameName"`
			LotteryGameNum          string        `json:"lotteryGameNum"`
			LotteryGamePronum       int           `json:"lotteryGamePronum"`
			LotteryNotice           int           `json:"lotteryNotice"`
			LotteryNoticeShowFlag   int           `json:"lotteryNoticeShowFlag"`
			LotteryPaidBeginTime    string        `json:"lotteryPaidBeginTime"`
			LotteryPaidEndTime      string        `json:"lotteryPaidEndTime"`
			LotteryPromotionFlag    int           `json:"lotteryPromotionFlag"`
			LotteryPromotionFlagRj  int           `json:"lotteryPromotionFlagRj"`
			LotterySaleBeginTime    string        `json:"lotterySaleBeginTime"`
			LotterySaleEndTimeUnix  int           `json:"lotterySaleEndTimeUnix"`
			LotterySaleEndtime      string        `json:"lotterySaleEndtime"`
			LotterySuspendedFlag    int           `json:"lotterySuspendedFlag"`
			LotteryUnsortDrawresult string        `json:"lotteryUnsortDrawresult"`
			MatchList               []interface{} `json:"matchList"`
			PdfType                 int           `json:"pdfType"`
			PoolBalanceAfterdraw    string        `json:"poolBalanceAfterdraw"`
			PoolBalanceAfterdrawRj  string        `json:"poolBalanceAfterdrawRj"`
			PrizeLevelList          []struct {
				AwardType        int    `json:"awardType"`
				Group            string `json:"group"`
				LotteryCondition string `json:"lotteryCondition"`
				PrizeLevel       string `json:"prizeLevel"`
				Sort             int    `json:"sort"`
				StakeAmount      string `json:"stakeAmount"`
				StakeCount       string `json:"stakeCount"`
				TotalPrizeamount string `json:"totalPrizeamount"`
			} `json:"prizeLevelList"`
			PrizeLevelListRj  []interface{} `json:"prizeLevelListRj"`
			RuleType          int           `json:"ruleType"`
			SurplusAmount     string        `json:"surplusAmount"`
			SurplusAmountRj   string        `json:"surplusAmountRj"`
			TermList          []interface{} `json:"termList"`
			TermResultList    []interface{} `json:"termResultList"`
			TotalSaleAmount   string        `json:"totalSaleAmount"`
			TotalSaleAmountRj string        `json:"totalSaleAmountRj"`
			Verify            int           `json:"verify"`
			VtoolsConfig      struct {
			} `json:"vtoolsConfig"`
		} `json:"list"`
		PageNo   int `json:"pageNo"`
		PageSize int `json:"pageSize"`
		Pages    int `json:"pages"`
		Total    int `json:"total"`
	} `json:"value"`
}

// 七星彩
type SevenStar struct {
	DataFrom     string `json:"dataFrom"`
	EmptyFlag    bool   `json:"emptyFlag"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
	Value        struct {
		LastPoolDraw struct {
			LotteryDrawNum       string `json:"lotteryDrawNum"`
			LotteryDrawResult    string `json:"lotteryDrawResult"`
			LotteryDrawTime      string `json:"lotteryDrawTime"`
			LotteryGameName      string `json:"lotteryGameName"`
			LotteryGameNum       string `json:"lotteryGameNum"`
			PoolBalanceAfterdraw string `json:"poolBalanceAfterdraw"`
			PrizeLevelList       []struct {
				AwardType        int    `json:"awardType"`
				Group            string `json:"group"`
				LotteryCondition string `json:"lotteryCondition"`
				PrizeLevel       string `json:"prizeLevel"`
				Sort             int    `json:"sort"`
				StakeAmount      string `json:"stakeAmount"`
				StakeCount       string `json:"stakeCount"`
				TotalPrizeamount string `json:"totalPrizeamount"`
			} `json:"prizeLevelList"`
		} `json:"lastPoolDraw"`
		List []struct {
			DrawFlowFund            string        `json:"drawFlowFund"`
			DrawFlowFundRj          string        `json:"drawFlowFundRj"`
			DrawPdfUrl              string        `json:"drawPdfUrl"`
			EstimateDrawTime        string        `json:"estimateDrawTime"`
			IsGetKjpdf              int           `json:"isGetKjpdf"`
			IsGetXlpdf              int           `json:"isGetXlpdf"`
			LotteryDrawNum          string        `json:"lotteryDrawNum"`
			LotteryDrawResult       string        `json:"lotteryDrawResult"`
			LotteryDrawStatus       int           `json:"lotteryDrawStatus"`
			LotteryDrawStatusNo     string        `json:"lotteryDrawStatusNo"`
			LotteryDrawTime         string        `json:"lotteryDrawTime"`
			LotteryEquipmentCount   int           `json:"lotteryEquipmentCount"`
			LotteryGameName         string        `json:"lotteryGameName"`
			LotteryGameNum          string        `json:"lotteryGameNum"`
			LotteryGamePronum       int           `json:"lotteryGamePronum"`
			LotteryNotice           int           `json:"lotteryNotice"`
			LotteryNoticeShowFlag   int           `json:"lotteryNoticeShowFlag"`
			LotteryPaidBeginTime    string        `json:"lotteryPaidBeginTime"`
			LotteryPaidEndTime      string        `json:"lotteryPaidEndTime"`
			LotteryPromotionFlag    int           `json:"lotteryPromotionFlag"`
			LotteryPromotionFlagRj  int           `json:"lotteryPromotionFlagRj"`
			LotterySaleBeginTime    string        `json:"lotterySaleBeginTime"`
			LotterySaleEndTimeUnix  int           `json:"lotterySaleEndTimeUnix"`
			LotterySaleEndtime      string        `json:"lotterySaleEndtime"`
			LotterySuspendedFlag    int           `json:"lotterySuspendedFlag"`
			LotteryUnsortDrawresult string        `json:"lotteryUnsortDrawresult"`
			MatchList               []interface{} `json:"matchList"`
			PdfType                 int           `json:"pdfType"`
			PoolBalanceAfterdraw    string        `json:"poolBalanceAfterdraw"`
			PoolBalanceAfterdrawRj  string        `json:"poolBalanceAfterdrawRj"`
			PrizeLevelList          []struct {
				AwardType        int    `json:"awardType"`
				Group            string `json:"group"`
				LotteryCondition string `json:"lotteryCondition"`
				PrizeLevel       string `json:"prizeLevel"`
				Sort             int    `json:"sort"`
				StakeAmount      string `json:"stakeAmount"`
				StakeCount       string `json:"stakeCount"`
				TotalPrizeamount string `json:"totalPrizeamount"`
			} `json:"prizeLevelList"`
			PrizeLevelListRj  []interface{} `json:"prizeLevelListRj"`
			RuleType          int           `json:"ruleType"`
			SurplusAmount     string        `json:"surplusAmount"`
			SurplusAmountRj   string        `json:"surplusAmountRj"`
			TermList          []interface{} `json:"termList"`
			TermResultList    []interface{} `json:"termResultList"`
			TotalSaleAmount   string        `json:"totalSaleAmount"`
			TotalSaleAmountRj string        `json:"totalSaleAmountRj"`
			Verify            int           `json:"verify"`
			VtoolsConfig      struct {
			} `json:"vtoolsConfig"`
		} `json:"list"`
		PageNo   int `json:"pageNo"`
		PageSize int `json:"pageSize"`
		Pages    int `json:"pages"`
		Total    int `json:"total"`
	} `json:"value"`
}

var LotteryStatistics = cache2go.Cache("lottery-statistics")

// @Summary 超级大乐透
// @Description 超级大乐透
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /lottery/super-lottery [get]
func SuperLotteryFun(c *gin.Context) {
	var url = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListV1.qry?gameNo=04&provinceId=0&isVerify=1&termLimits=13"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result SuperLottery
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)
		common.FailedReturn(c, "查询失败")
		return
	}
	//把一些奖池缓存起来
	if !LotteryStatistics.Exists(SUPER_LOTTERY_POOL) {
		LotteryStatistics.Add(SUPER_LOTTERY_POOL, 6*time.Hour, result.Value.LastPoolDraw.PoolBalanceAfterdraw)
	}
	common.SuccessReturn(c, result)
}

// @Summary 排列五
// @Description 排列五
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /lottery/plw [get]
func PlwFun(c *gin.Context) {
	var url = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListPlwV1.qry?gameNo=350133&provinceId=0&isVerify=1&termLimits=5"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result Plw
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)
		common.FailedReturn(c, "查询失败")
		return
	}
	//把一些奖池缓存起来
	if !LotteryStatistics.Exists(PLW_POOL) {
		LotteryStatistics.Add(PLW_POOL, 6*time.Hour, result.Value.List[0].PoolBalanceAfterdraw)
	}
	common.SuccessReturn(c, result)

}

// @Summary 七星彩
// @Description 七星彩
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /lottery/seven-star [get]
func SevenStarFun(c *gin.Context) {

	resp, err := http.Get(SEVEN_START_URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result SevenStar
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)
		common.FailedReturn(c, "查询失败")
		return
	}
	//把一些奖池缓存起来
	if !LotteryStatistics.Exists(SEVEN_STAR_POOL) {
		LotteryStatistics.Add(SEVEN_STAR_POOL, 6*time.Hour, result.Value.LastPoolDraw.PoolBalanceAfterdraw)
	}
	common.SuccessReturn(c, result)
}

func CreatePLW(ord *order.Order) error {

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

	if strings.Contains(ord.Content, ",") {
		arr := strings.Split(ord.Content, ",")
		if len(arr) == tp {
			for _, s := range arr {
				numArr := strings.Split(s, " ")
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
	} else if len(ord.Content) == tp {
		numArr := strings.Split(ord.Content, " ")
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

	} else {
		return errors.New("参数异常")
	}
	if err := mysql.DB.Create(ord).Error; err != nil {
		log.Error(err)
		return errors.New("保存订单失败")
	}
	if !LotteryStatistics.Exists("plw_check") {
		LotteryStatistics.Add("plw_check", 8*time.Hour, 1)
		AddPlwCheck(tp)
	}
	return nil
}

func AddPlwCheck(p int) {

	resp, err := http.Get(PLW_URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result Plw
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)

		return
	}
	var job order.Job
	switch p {
	case 3:
		job = order.Job{
			Time:  util.GetPLWFinishedTime(),
			Param: nil,
			CallBack: func(param interface{}) {
				orders := order.GetOrderByLotteryType("P3")
				tx := mysql.DB.Begin()
				if len(orders) > 0 {
					for _, o := range orders {
						if strings.Compare(result.Value.List[0].LotteryDrawNum, o.IssueId) == 0 {
							content := getArr(o.Content)
							releaseNum := result.Value.List[0].LotteryDrawResult[0:4]
							for _, s := range content {
								if strings.Compare(s, releaseNum) == 0 {
									o.Bonus = o.Bonus + 1
								}
							}
						}
						tx.Save(o)
					}
				}
				tx.Commit()

			},
		}

		break
	case 5:
		job = order.Job{
			Time:  util.GetPLWFinishedTime(),
			Param: nil,
			CallBack: func(param interface{}) {
				orders := order.GetOrderByLotteryType("P5")
				tx := mysql.DB.Begin()
				if len(orders) > 0 {
					for _, o := range orders {
						if strings.Compare(result.Value.List[0].LotteryDrawNum, o.IssueId) == 0 {
							content := getArr(o.Content)
							releaseNum := result.Value.List[0].LotteryDrawResult
							for _, s := range content {
								if strings.Compare(s, releaseNum) == 0 {
									o.Bonus = o.Bonus + 1
								}
							}
						}
						tx.Save(o)
					}
				}
				tx.Commit()

			},
		}

		break
	}
	order.AddJob(job)

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
