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
	"jingcai/util"
	"net/http"
	"strconv"
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
			LotteryDrawNum       string `json:"lotteryDrawNum,omitempty"`
			LotteryDrawResult    string `json:"lotteryDrawResult,omitempty"`
			LotteryDrawTime      string `json:"lotteryDrawTime,omitempty"`
			LotteryGameName      string `json:"lotteryGameName,omitempty"`
			LotteryGameNum       string `json:"lotteryGameNum,omitempty"`
			PoolBalanceAfterdraw string `json:"poolBalanceAfterdraw,omitempty"`
			PrizeLevelList       []struct {
				AwardType        int    `json:"awardType,omitempty"`
				Group            string `json:"group,omitempty"`
				LotteryCondition string `json:"lotteryCondition,omitempty"`
				PrizeLevel       string `json:"prizeLevel,omitempty"`
				Sort             int    `json:"sort,omitempty"`
				StakeAmount      string `json:"stakeAmount,omitempty"`
				StakeCount       string `json:"stakeCount,omitempty"`
				TotalPrizeamount string `json:"totalPrizeamount,omitempty"`
			} `json:"prizeLevelList,omitempty"`
		} `json:"lastPoolDraw,omitempty"`
		List []struct {
			DrawFlowFund           string `json:"drawFlowFund,omitempty"`
			DrawFlowFundRj         string `json:"drawFlowFundRj,omitempty"`
			DrawPdfUrl             string `json:"drawPdfUrl,omitempty"`
			EstimateDrawTime       string `json:"estimateDrawTime,omitempty"`
			IsGetKjpdf             int    `json:"isGetKjpdf,omitempty"`
			IsGetXlpdf             int    `json:"isGetXlpdf,omitempty"`
			LotteryDrawNum         string `json:"lotteryDrawNum,omitempty"`
			LotteryDrawResult      string `json:"lotteryDrawResult,omitempty"`
			LotteryDrawStatus      int    `json:"lotteryDrawStatus,omitempty"`
			LotteryDrawStatusNo    string `json:"lotteryDrawStatusNo,omitempty"`
			LotteryDrawTime        string `json:"lotteryDrawTime,omitempty"`
			LotteryEquipmentCount  int    `json:"lotteryEquipmentCount,omitempty"`
			LotteryGameName        string `json:"lotteryGameName,omitempty"`
			LotteryGameNum         string `json:"lotteryGameNum,omitempty"`
			LotteryGamePronum      int    `json:"lotteryGamePronum,omitempty"`
			LotteryNotice          int    `json:"lotteryNotice,omitempty"`
			LotteryNoticeShowFlag  int    `json:"lotteryNoticeShowFlag,omitempty"`
			LotteryPaidBeginTime   string `json:"lotteryPaidBeginTime,omitempty"`
			LotteryPaidEndTime     string `json:"lotteryPaidEndTime,omitempty"`
			LotteryPromotionFlag   int    `json:"lotteryPromotionFlag,omitempty"`
			LotteryPromotionFlagRj int    `json:"lotteryPromotionFlagRj,omitempty"`
			LotterySaleBeginTime   string `json:"lotterySaleBeginTime,omitempty"`
			//LotterySaleEndTimeUnix  int           `json:"lotterySaleEndTimeUnix"`
			LotterySaleEndtime      string        `json:"lotterySaleEndtime,omitempty"`
			LotterySuspendedFlag    int           `json:"lotterySuspendedFlag,omitempty"`
			LotteryUnsortDrawresult string        `json:"lotteryUnsortDrawresult,omitempty"`
			MatchList               []interface{} `json:"matchList,omitempty"`
			PdfType                 int           `json:"pdfType,omitempty"`
			PoolBalanceAfterdraw    string        `json:"poolBalanceAfterdraw,omitempty"`
			PoolBalanceAfterdrawRj  string        `json:"poolBalanceAfterdrawRj,omitempty"`
			PrizeLevelList          []struct {
				AwardType        int    `json:"awardType,omitempty"`
				Group            string `json:"group,omitempty"`
				LotteryCondition string `json:"lotteryCondition,omitempty"`
				PrizeLevel       string `json:"prizeLevel,omitempty"`
				Sort             int    `json:"sort,omitempty"`
				StakeAmount      string `json:"stakeAmount,omitempty"`
				StakeCount       string `json:"stakeCount,omitempty"`
				TotalPrizeamount string `json:"totalPrizeamount,omitempty"`
			} `json:"prizeLevelList,omitempty"`
			PrizeLevelListRj  []interface{} `json:"prizeLevelListRj,omitempty"`
			RuleType          int           `json:"ruleType,omitempty"`
			SurplusAmount     string        `json:"surplusAmount,omitempty"`
			SurplusAmountRj   string        `json:"surplusAmountRj,omitempty"`
			TermList          []interface{} `json:"termList,omitempty"`
			TermResultList    []interface{} `json:"termResultList,omitempty"`
			TotalSaleAmount   string        `json:"totalSaleAmount,omitempty"`
			TotalSaleAmountRj string        `json:"totalSaleAmountRj,omitempty"`
			Verify            int           `json:"verify,omitempty"`
			VtoolsConfig      struct {
			} `json:"vtoolsConfig,omitempty"`
		} `json:"list,omitempty"`
		PageNo   int `json:"pageNo,omitempty"`
		PageSize int `json:"pageSize,omitempty"`
		Pages    int `json:"pages,omitempty"`
		Total    int `json:"total,omitempty"`
	} `json:"value,omitempty"`
}

// 排列5
type Plw struct {
	DataFrom     string `json:"dataFrom,omitempty"`
	EmptyFlag    bool   `json:"emptyFlag,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Success      bool   `json:"success,omitempty"`
	Value        struct {
		List []struct {
			DrawFlowFund           string `json:"drawFlowFund,omitempty"`
			DrawFlowFundRj         string `json:"drawFlowFundRj,omitempty"`
			DrawPdfUrl             string `json:"drawPdfUrl,omitempty"`
			DrawPdfUrlPls          string `json:"drawPdfUrlPls,omitempty"`
			EstimateDrawTime       string `json:"estimateDrawTime,omitempty"`
			IsGetKjpdf             int    `json:"isGetKjpdf,omitempty"`
			IsGetXlpdf             int    `json:"isGetXlpdf,omitempty"`
			LotteryDrawNum         string `json:"lotteryDrawNum,omitempty"`
			LotteryDrawResult      string `json:"lotteryDrawResult,omitempty"`
			LotteryDrawStatus      int    `json:"lotteryDrawStatus,omitempty"`
			LotteryDrawStatusNo    string `json:"lotteryDrawStatusNo,omitempty"`
			LotteryDrawTime        string `json:"lotteryDrawTime,omitempty"`
			LotteryEquipmentCount  int    `json:"lotteryEquipmentCount,omitempty"`
			LotteryGameName        string `json:"lotteryGameName,omitempty"`
			LotteryGameNum         string `json:"lotteryGameNum,omitempty"`
			LotteryGamePronum      int    `json:"lotteryGamePronum,omitempty"`
			LotteryNotice          int    `json:"lotteryNotice,omitempty"`
			LotteryNoticeShowFlag  int    `json:"lotteryNoticeShowFlag,omitempty"`
			LotteryPaidBeginTime   string `json:"lotteryPaidBeginTime,omitempty"`
			LotteryPaidEndTime     string `json:"lotteryPaidEndTime,omitempty"`
			LotteryPromotionFlag   int    `json:"lotteryPromotionFlag,omitempty"`
			LotteryPromotionFlagRj int    `json:"lotteryPromotionFlagRj,omitempty"`
			LotterySaleBeginTime   string `json:"lotterySaleBeginTime,omitempty"`
			//LotterySaleEndTimeUnix  map[interface{}]interface{} `json:"lotterySaleEndTimeUnix"`
			LotterySaleEndtime      string        `json:"lotterySaleEndtime,omitempty"`
			LotterySuspendedFlag    int           `json:"lotterySuspendedFlag,omitempty"`
			LotteryUnsortDrawresult string        `json:"lotteryUnsortDrawresult,omitempty"`
			MatchList               []interface{} `json:"matchList,omitempty"`
			PdfType                 int           `json:"pdfType,omitempty"`
			PoolBalanceAfterdraw    string        `json:"poolBalanceAfterdraw,omitempty"`
			PoolBalanceAfterdrawRj  string        `json:"poolBalanceAfterdrawRj,omitempty"`
			PrizeLevelList          []struct {
				AwardType        int    `json:"awardType,omitempty"`
				Group            string `json:"group,omitempty"`
				LotteryCondition string `json:"lotteryCondition,omitempty"`
				PrizeLevel       string `json:"prizeLevel,omitempty"`
				Sort             int    `json:"sort,omitempty"`
				StakeAmount      string `json:"stakeAmount,omitempty"`
				StakeCount       string `json:"stakeCount,omitempty"`
				TotalPrizeamount string `json:"totalPrizeamount,omitempty"`
			} `json:"prizeLevelList,omitempty"`
			PrizeLevelListRj  []interface{} `json:"prizeLevelListRj,omitempty"`
			RuleType          int           `json:"ruleType,omitempty"`
			SurplusAmount     string        `json:"surplusAmount,omitempty"`
			SurplusAmountRj   string        `json:"surplusAmountRj,omitempty"`
			TermList          []interface{} `json:"termList,omitempty"`
			TermResultList    []interface{} `json:"termResultList,omitempty"`
			TotalSaleAmount   string        `json:"totalSaleAmount,omitempty"`
			TotalSaleAmountRj string        `json:"totalSaleAmountRj,omitempty"`
			Verify            int           `json:"verify,omitempty"`
			VtoolsConfig      struct {
			} `json:"vtoolsConfig,omitempty"`
		} `json:"list,omitempty"`
		PageNo   int `json:"pageNo,omitempty"`
		PageSize int `json:"pageSize,omitempty"`
		Pages    int `json:"pages,omitempty"`
		Total    int `json:"total,omitempty"`
	} `json:"value,omitempty"`
}

// 七星彩
type SevenStar struct {
	DataFrom     string `json:"dataFrom,omitempty"`
	EmptyFlag    bool   `json:"emptyFlag,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Success      bool   `json:"success,omitempty"`
	Value        struct {
		LastPoolDraw struct {
			LotteryDrawNum       string `json:"lotteryDrawNum,omitempty"`
			LotteryDrawResult    string `json:"lotteryDrawResult,omitempty"`
			LotteryDrawTime      string `json:"lotteryDrawTime,omitempty"`
			LotteryGameName      string `json:"lotteryGameName,omitempty"`
			LotteryGameNum       string `json:"lotteryGameNum,omitempty"`
			PoolBalanceAfterdraw string `json:"poolBalanceAfterdraw,omitempty"`
			PrizeLevelList       []struct {
				AwardType        int    `json:"awardType,omitempty"`
				Group            string `json:"group,omitempty"`
				LotteryCondition string `json:"lotteryCondition,omitempty"`
				PrizeLevel       string `json:"prizeLevel,omitempty"`
				Sort             int    `json:"sort,omitempty"`
				StakeAmount      string `json:"stakeAmount,omitempty"`
				StakeCount       string `json:"stakeCount,omitempty"`
				TotalPrizeamount string `json:"totalPrizeamount,omitempty"`
			} `json:"prizeLevelList,omitempty"`
		} `json:"lastPoolDraw,omitempty"`
		List []struct {
			DrawFlowFund           string `json:"drawFlowFund,omitempty"`
			DrawFlowFundRj         string `json:"drawFlowFundRj,omitempty"`
			DrawPdfUrl             string `json:"drawPdfUrl,omitempty"`
			EstimateDrawTime       string `json:"estimateDrawTime,omitempty"`
			IsGetKjpdf             int    `json:"isGetKjpdf,omitempty"`
			IsGetXlpdf             int    `json:"isGetXlpdf,omitempty"`
			LotteryDrawNum         string `json:"lotteryDrawNum,omitempty"`
			LotteryDrawResult      string `json:"lotteryDrawResult,omitempty"`
			LotteryDrawStatus      int    `json:"lotteryDrawStatus,omitempty"`
			LotteryDrawStatusNo    string `json:"lotteryDrawStatusNo,omitempty"`
			LotteryDrawTime        string `json:"lotteryDrawTime,omitempty"`
			LotteryEquipmentCount  int    `json:"lotteryEquipmentCount,omitempty"`
			LotteryGameName        string `json:"lotteryGameName,omitempty"`
			LotteryGameNum         string `json:"lotteryGameNum,omitempty"`
			LotteryGamePronum      int    `json:"lotteryGamePronum,omitempty"`
			LotteryNotice          int    `json:"lotteryNotice,omitempty"`
			LotteryNoticeShowFlag  int    `json:"lotteryNoticeShowFlag,omitempty"`
			LotteryPaidBeginTime   string `json:"lotteryPaidBeginTime,omitempty"`
			LotteryPaidEndTime     string `json:"lotteryPaidEndTime,omitempty"`
			LotteryPromotionFlag   int    `json:"lotteryPromotionFlag,omitempty"`
			LotteryPromotionFlagRj int    `json:"lotteryPromotionFlagRj,omitempty"`
			LotterySaleBeginTime   string `json:"lotterySaleBeginTime,omitempty"`
			//LotterySaleEndTimeUnix  map[interface{}]interface{} `json:"lotterySaleEndTimeUnix"`
			LotterySaleEndtime      string        `json:"lotterySaleEndtime,omitempty"`
			LotterySuspendedFlag    int           `json:"lotterySuspendedFlag,omitempty"`
			LotteryUnsortDrawresult string        `json:"lotteryUnsortDrawresult,omitempty"`
			MatchList               []interface{} `json:"matchList,omitempty"`
			PdfType                 int           `json:"pdfType,omitempty"`
			PoolBalanceAfterdraw    string        `json:"poolBalanceAfterdraw,omitempty"`
			PoolBalanceAfterdrawRj  string        `json:"poolBalanceAfterdrawRj,omitempty"`
			PrizeLevelList          []struct {
				AwardType        int    `json:"awardType,omitempty"`
				Group            string `json:"group,omitempty"`
				LotteryCondition string `json:"lotteryCondition,omitempty"`
				PrizeLevel       string `json:"prizeLevel,omitempty"`
				Sort             int    `json:"sort,omitempty"`
				StakeAmount      string `json:"stakeAmount,omitempty"`
				StakeCount       string `json:"stakeCount,omitempty"`
				TotalPrizeamount string `json:"totalPrizeamount,omitempty"`
			} `json:"prizeLevelList,omitempty"`
			PrizeLevelListRj  []interface{} `json:"prizeLevelListRj,omitempty"`
			RuleType          int           `json:"ruleType,omitempty"`
			SurplusAmount     string        `json:"surplusAmount,omitempty"`
			SurplusAmountRj   string        `json:"surplusAmountRj,omitempty"`
			TermList          []interface{} `json:"termList,omitempty"`
			TermResultList    []interface{} `json:"termResultList,omitempty"`
			TotalSaleAmount   string        `json:"totalSaleAmount,omitempty"`
			TotalSaleAmountRj string        `json:"totalSaleAmountRj,omitempty"`
			Verify            int           `json:"verify,omitempty"`
			VtoolsConfig      struct {
			} `json:"vtoolsConfig,omitempty"`
		} `json:"list,omitempty"`
		PageNo   int `json:"pageNo,omitempty"`
		PageSize int `json:"pageSize,omitempty"`
		Pages    int `json:"pages,omitempty"`
		Total    int `json:"total,omitempty"`
	} `json:"value,omitempty"`
}

var LotteryStatistics = cache2go.Cache("lottery-statistics")

// @Summary 超级大乐透
// @Description 超级大乐透
// @Tags lotto 其它彩票
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /lottery-api/super-lottery [get]
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
	key := GetSuperLotteryIssueKey()
	if !LotteryStatistics.Exists(key) {
		LotteryStatistics.Add(key, 9*time.Hour, result.Value.LastPoolDraw.LotteryDrawNum)
	}
	common.SuccessReturn(c, result)
}

func GetSuperLotteryIssueKey() string {
	var key = fmt.Sprintf("%s,%s", "super-lottery:", util.GetTodayYYHHMMSS())
	return key
}

func GetSuperLotteryIssueId() (int, error) {
	key := GetSuperLotteryIssueKey()
	if LotteryStatistics.Exists(key) {
		item, err := LotteryStatistics.Value(key)
		if err != nil {
			log.Error(err)
			return 0, errors.New("获取缓存失败")
		}
		num := item.Data().(string)
		issueId, err := strconv.Atoi(num)
		return issueId, nil
	}

	resp, err := http.Get(SUPER_LOTTO_URL)
	if err != nil {
		fmt.Println(err)
		return 0, errors.New("请求在线失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result SuperLottery
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)
		return 0, errors.New("转换大乐透结果为对象失败")
	}

	LotteryStatistics.Add(key, 9*time.Hour, result.Value.LastPoolDraw.LotteryDrawNum)
	issueId, err := strconv.Atoi(result.Value.LastPoolDraw.LotteryDrawNum)
	return issueId, nil
}

// @Summary 排列五
// @Description 排列五
// @Tags lotto 其它彩票
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /lottery-api/plw [get]
func PlwFun(c *gin.Context) {

	var url = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListPlwV1.qry?gameNo=350133&provinceId=0&isVerify=1&termLimits=5"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("请求大乐透列表失败: ", err)
		return
	}

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
	key := GetPLWIssueKey()
	if !LotteryStatistics.Exists(key) {
		LotteryStatistics.Add(key, 9*time.Hour, result.Value.List[0].LotteryDrawNum)
	}
	common.SuccessReturn(c, result)

}

func GetPLWIssueKey() string {
	var key = fmt.Sprintf("%s,%s", "p3-5:", util.GetTodayYYHHMMSS())
	return key
}

func GetPLWIssueId() (int, error) {
	key := GetPLWIssueKey()
	if LotteryStatistics.Exists(key) {
		item, err := LotteryStatistics.Value(key)
		if err != nil {
			log.Error(err)
			return 0, errors.New("获取缓存失败")
		}
		num := item.Data().(string)
		issueId, err := strconv.Atoi(num)
		return issueId, nil
	}

	var url = "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListPlwV1.qry?gameNo=350133&provinceId=0&isVerify=1&termLimits=5"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return 0, errors.New("请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("请求大乐透列表失败: ", err)
		return 0, errors.New("请求大乐透列表失败")
	}

	var result Plw
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)
		return 0, errors.New("查询失败， 请稍后重试")
	}
	LotteryStatistics.Add(key, 9*time.Hour, result.Value.List[0].LotteryDrawNum)
	issueId, err := strconv.Atoi(result.Value.List[0].LotteryDrawNum)
	return issueId, nil
}

// @Summary 七星彩
// @Description 七星彩
// @Tags lotto 其它彩票
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /lottery-api/seven-star [get]
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
	key := GetSevenStarKey()
	if !LotteryStatistics.Exists(key) {
		LotteryStatistics.Add(key, 9*time.Hour, result.Value.List[0].LotteryDrawNum)
	}
	common.SuccessReturn(c, result)
}

func GetSevenStarKey() string {
	var key = fmt.Sprintf("%s,%s", "seven-start:", util.GetTodayYYHHMMSS())
	return key
}

func GetSevenStarIssueId() (int, error) {
	key := GetSevenStarKey()
	if LotteryStatistics.Exists(key) {
		item, err := LotteryStatistics.Value(key)
		if err != nil {
			log.Error(err)
			return 0, errors.New("获取缓存失败")
		}
		num := item.Data().(string)
		issueId, err := strconv.Atoi(num)
		return issueId, nil
	}
	resp, err := http.Get(SEVEN_START_URL)
	if err != nil {
		fmt.Println(err)
		return 0, errors.New("请求失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result SevenStar
	err = json.Unmarshal(body, &result)
	if err != nil || &result.Value == nil {
		log.Error("转换大乐透结果为对象失败", err)
		return 0, errors.New("对象转换失败")
	}
	LotteryStatistics.Add(key, 9*time.Hour, result.Value.List[0].LotteryDrawNum)
	issueId, err := strconv.Atoi(result.Value.List[0].LotteryDrawNum)
	return issueId, nil
}
