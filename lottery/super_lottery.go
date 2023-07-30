package lottery

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"io"
	"jingcai/common"
	alog "jingcai/log"
	"net/http"
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
const SUPER_LOTTO_URL = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListV1.qry?gameNo=04&provinceId=0&isVerify=1&termLimits=13"

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
	resp, err := http.Get(SUPER_LOTTO_URL)
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
