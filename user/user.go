package user

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	gorsa "github.com/Lyafei/go-rsa"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"github.com/pascaldekloe/jwt"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
	"jingcai/score"
	"jingcai/validatior"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = ilog.Logger
var userCahe = cache2go.Cache("user")

const TOKEN_TIME_OUT = 4 * time.Hour
const ADMIN = "Admin"
const USER = "User"
const LOCK_TIMES = 5

// ADD(增加) SUBTRACT(扣除)
type Option string

// 清账方式  WECHAT(微信) CARD(银行卡) ALI(支付宝) SCORE(积分清账)
const (
	SCORE                 = "SCORE"
	RMB                   = "RMB"
	FREE_SCORE            = "FREE_SCORE"
	ADD                   = "ADD" //增加
	SUBTRACT              = "SUBTRACT"
	WECHAT                = "WECHAT" //微信
	CARD                  = "CARD"   //银行卡
	ALI                   = "ALI"    //支付宝
	SALE                  = "Sale"   //销售角色
	SUPER_ADMIN           = "SuperAdmin"
	BILL_COMMENT_CASHED   = "BILL_COMMENT_CASHED"   //兑奖
	BILL_COMMENT_ADD      = "BILL_COMMENT_ADD"      //充值
	BILL_COMMENT_CLEAR    = "BILL_COMMENT_CLEAR"    //清账
	BILL_COMMENT_ACTIVITY = "BILL_COMMENT_ACTIVITY" //活动赠送
	BILL_COMMENT_BUY      = "BILL_COMMENT_BUY"      //购彩

)

type User struct {
	gorm.Model
	Phone string
	//昵称
	Name string
	//密码
	Secret string
	//盐
	Salt string
	Role string //"enum: Admin,User,Sale"
	//微信号
	Wechat string
	//支付宝号
	Ali string
	//余额
	Score float32 `gorm:"type: decimal(20,6)"`
	//头像地址
	HeaderImageUrl string

	//来自店铺码（管理员的用户id）
	From uint

	//来自某个用户推广
	FromUser uint
}

func (o Option) String() string {
	switch o {
	case ADD:
		return ADD
	case SUBTRACT:
		return SUBTRACT
	default:
		return ""
	}
}

// 用户对象
type UserVO struct {
	Phone string `minLength:"11" maxLength:"11"`
	//昵称
	Name string `minLength:"4" maxLength:"16"`
	//密码
	Secret string `minLength:"6" maxLength:"16"`
	//头像
	Avatar string
}

// token 对象
type TokenVO struct {
	Token string
}

type UserDTO struct {
	Phone string
	//昵称
	Name string

	Role string //"enum: Admin,User"
	//微信号
	Wechat string
	//支付宝号
	Ali string

	//余额
	Score float32

	//头像地址
	HeaderImageUrl string

	//赠送积分
	FreeScore float32
}

func (u User) GetDTO() UserDTO {
	return UserDTO{
		Phone:          u.Phone,
		Name:           u.Name,
		Role:           u.Role,
		Wechat:         u.Wechat,
		Ali:            u.Ali,
		Score:          u.Score,
		HeaderImageUrl: u.HeaderImageUrl,
	}
}

// @Summary 查询用户信息
// @Description 查询用户信息
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/user/info [get]
func GetUserInfo(c *gin.Context) {
	var user = getUserInfo(c)
	var userPO User
	var param = User{}
	param.ID = user.ID

	if mysql.DB.Model(param).Where(&param).First(&userPO).Error != nil {
		common.FailedReturn(c, "查询信息失败")
	}
	var userDTO = UserDTO{
		Phone:          userPO.Phone,
		Name:           userPO.Name,
		Role:           userPO.Role,
		Wechat:         userPO.Wechat,
		Ali:            userPO.Ali,
		Score:          userPO.Score,
		HeaderImageUrl: userPO.HeaderImageUrl,
	}
	free, err := score.QueryByUserId(user.ID)
	if err == nil {
		userDTO.FreeScore = free.Score
	}
	common.SuccessReturn(c, userDTO)
}

// @Summary 更新用户信息
// @Description 更新用户信息
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body UserDTO true "用户对象, socre 可以不传"
// @Router /api/user/info [post]
func UpdateUser(c *gin.Context) {
	var user = getUserInfo(c)
	var userDTO UserDTO
	var param = User{
		Model: gorm.Model{
			ID: user.ID,
		},
	}

	var realUsr User
	c.BindJSON(&userDTO)
	var update = User{
		Model: gorm.Model{
			ID: user.ID,
		},
		Phone:          userDTO.Phone,
		Name:           userDTO.Name,
		Role:           userDTO.Role,
		Wechat:         userDTO.Wechat,
		Ali:            userDTO.Ali,
		HeaderImageUrl: userDTO.HeaderImageUrl,
		Secret:         realUsr.Secret,
	}
	if mysql.DB.Model(param).First(&realUsr).Error != nil {
		common.FailedReturn(c, "查不到当前用户")
	}
	if err := mysql.DB.Model(param).Where(param).Updates(&update).Error; err != nil {
		log.Error("update user failed, id: ", user.ID, " err: ", err)
		common.FailedReturn(c, "更新用户失败")
	}
	common.SuccessReturn(c, update)
	return
}

func CreateUser(user UserVO, shopId uint, sharedUserId uint) error {
	var userPo = User{
		Phone:    user.Phone,
		Secret:   user.Secret,
		Name:     user.Name,
		Salt:     uuid.NewV4().String()[0:16],
		Role:     "User",
		Score:    0.00,
		From:     shopId,
		FromUser: sharedUserId,
	}

	pwd, err := common.EnPwdCode([]byte(user.Secret), []byte(userPo.Salt))
	if err != nil {

		log.Error("加密密码失败", err)
		return err
	}
	userPo.Secret = pwd
	mysql.DB.AutoMigrate(&userPo)
	return mysql.DB.Create(&userPo).Error
}

func (u User) ChangePass(user UserVO) error {

	return nil
}

// @Summary 创建用户
// @Description 创建用户
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body UserVO true "用户对象"
// @param sharedShopId query string   false  "店铺id"
// @param sharedUserId query string  false  "分享者id"
// @Router /api/user [post]
func UserCreateHandler(c *gin.Context) {
	var user UserVO
	c.Header("Content-Type", "application/json; charset=utf-8")
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &common.BaseResponse{
			Code:    0,
			Message: "参数获取失败",
		})
	}
	//校验手机号是否已经存在
	//校验昵称是否存在
	var nameExist int64
	mysql.DB.AutoMigrate(&User{})
	mysql.DB.Model(&User{}).Where("name = ?", user.Name).Count(&nameExist)
	if nameExist > 0 {
		common.FailedReturn(c, "昵称已经存在")
		return
	}
	var phoneExist int64
	mysql.DB.Model(&User{}).Where("phone = ?", user.Phone).Count(&phoneExist)
	if phoneExist > 0 {
		common.FailedReturn(c, "手机号已经存在")
		return
	}
	var shopId int
	id := c.Query("sharedShopId")
	shardUserId := c.Query("sharedUserId")
	if id == "" {
		shopId = 1
	} else {
		ind, err := strconv.Atoi(id)
		if err != nil {
			log.Error(err)
			shopId = 1
		} else {
			shopId = ind
		}
	}
	var sharedId uint
	if shardUserId == "" {
		sharedId = 1
	} else {
		ind, err := strconv.Atoi(shardUserId)
		if err != nil {
			log.Error(err)
			sharedId = 1
		} else {
			sharedId = uint(ind)
		}
	}

	if CreateUser(user, uint(shopId), sharedId) != nil {
		c.JSON(http.StatusInternalServerError, &common.BaseResponse{
			Code:    0,
			Message: "创建用户失败",
		})
	} else {
		c.JSON(http.StatusOK, common.Success(""))
	}
}

// @Summary 登录接口
// @Description 公钥放在头里 salt， 密码：需要和公钥rsa 加密 账号为手机号
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body UserVO true "用户对象"
// @Router /api/user/login [post]
func Login(c *gin.Context) {
	var userVo UserVO
	salt := c.GetHeader("salt")

	if salt == "" {
		log.Error("====== salt is empty ========")
		common.FailedReturn(c, "认证失败")
	}
	//decodeSalt, _ := base64.StdEncoding.DecodeString(salt)
	res, err := common.CacheJingCai.Value(salt)
	if err != nil {
		common.FailedReturn(c, "公钥已经超时")
	} else {
		privateKey := res.Data().(string)
		c.BindJSON(&userVo)
		pwd, err := gorsa.PriKeyDecrypt(userVo.Secret, privateKey)
		var user User
		if mysql.DB.Model(&User{Phone: userVo.Phone}).Where(&User{Phone: userVo.Phone}).Find(&user).Error != nil {
			common.FailedReturn(c, "账户错误")
			return
		}
		if checkLock(user.Phone) {
			common.FailedReturn(c, "用户已经锁定")
		}
		real, errDePwd := common.DePwdCode(user.Secret, []byte(user.Salt))
		if errDePwd != nil || err != nil {
			log.Error("decode user password failed")
			if checkLock(user.Phone) {
				common.FailedReturn(c, "用户已经锁定")
			} else {
				common.FailedReturn(c, "解析密码失败, 错误5次将会锁30分钟")
			}
			return
		}
		if strings.Compare(pwd, string(real)) != 0 {
			if checkLock(user.Phone) {
				common.FailedReturn(c, "用户已经锁定")
			} else {
				common.FailedReturn(c, "账户或者密码错误, 错误5次将会锁30分钟")
			}
			return
		}
		//生成token
		var claims jwt.Claims
		claims.Subject = "alice"
		claims.Issued = jwt.NewNumericTime(time.Now().Round(time.Second))
		claims.Set = map[string]interface{}{"name": user.Name, "role": user.Role}
		// issue a JWT
		block, _ := pem.Decode([]byte(privateKey))
		privateKeyObj, covererr := x509.ParsePKCS1PrivateKey(block.Bytes)

		if covererr != nil {
			log.Error(covererr)
			common.FailedReturn(c, "解密失败")
			return
		}
		token, err := claims.RSASign(jwt.RS256, privateKeyObj)
		userCahe.Add(base64.StdEncoding.EncodeToString(token), TOKEN_TIME_OUT, user)
		common.SuccessReturn(c, token)
	}
}

func checkLock(account string) bool {
	var count int = 0
	if userCahe.Exists(account) {
		item, _ := userCahe.Value(account)
		var data = item.Data()
		if data.(int) > LOCK_TIMES {
			return true
		}
		count = count + data.(int)
		userCahe.Add(account, 30*time.Minute, count)
	} else {
		userCahe.Add(account, 30*time.Minute, count)
	}
	return false
}

// 投诉对象
type Complain struct {
	gorm.Model
	//投诉类型
	Type string
	//投诉详情
	Content string
	//图片
	Image string
	//联系电话
	Phone string
	//建议
	Proposal string
	UserId   uint
	UserName string
	//修改备注
	Comment string
}

// @Summary 注销登录
// @Description 注销当前用户
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/user/logout [post]
func Logout(c *gin.Context) {

	var token string
	token = c.Query("token") // 访问令牌
	if token == "" {
		token = c.GetHeader("token")
	}

	if token != "" {
		userCahe.Delete(token)
	}
	common.SuccessReturn(c, "注销成功")
}

func getUserInfo(c *gin.Context) User {
	user, _ := c.Get("userInfo")
	if user == nil {
		log.Error("获取不到用户信息！")
		common.FailedReturn(c, "获取不到用户信息")
		c.Abort()
	}
	return user.(User)
}

func FetUserInfo(c *gin.Context) User {
	user, exist := c.Get("userInfo")
	if !exist {
		log.Warn("用户信息不存在！")
		return User{}
	}
	return user.(User)
}

func FindUserById(id uint) User {
	var user User
	mysql.DB.Model(&User{
		Model: gorm.Model{
			ID: id,
		},
	}).Where(&User{
		Model: gorm.Model{
			ID: id,
		},
	}).First(&user)
	return user
}
func FindUserVOById(id uint) UserVO {
	var user User
	mysql.DB.Model(&User{
		Model: gorm.Model{
			ID: id,
		},
	}).Where(&User{
		Model: gorm.Model{
			ID: id,
		},
	}).First(&user)

	return UserVO{
		Phone:  user.Phone,
		Name:   user.Name,
		Avatar: user.HeaderImageUrl,
	}
}

func FindUsserMapById(id []uint) map[uint]UserVO {
	var user []User
	mysql.DB.Model(&User{
		Model: gorm.Model{},
	}).Where("id in (?)", id).Find(&user)
	var mapp = make(map[uint]UserVO, 0)
	if len(user) > 0 {
		for _, u := range user {
			mapp[u.ID] = UserVO{
				Phone:  fmt.Sprintf("%s****%s", u.Phone[0:3], u.Phone[8:11]),
				Name:   u.Name,
				Avatar: u.HeaderImageUrl,
			}
		}
		return mapp
	}
	return mapp
}

// @Summary 投诉
// @Description 投诉
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Complain true "投诉对象"
// @Router /api/user/complain [post]
func UserComplain(c *gin.Context) {
	var user = getUserInfo(c)
	var complain Complain
	c.BindJSON(&complain)
	complain.UserId = user.ID
	complain.UserName = user.Name
	mysql.DB.AutoMigrate(&complain)
	mysql.DB.Create(&complain)
	common.SuccessReturn(c, "提交成功")
}

func CheckScoreOrDoBill(userId uint, orderId string, scoreNum float32, doBill bool, tx *gorm.DB) error {
	var userInfo User
	var lock sync.Mutex

	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{ID: userId}}).First(&userInfo).Error; err != nil {
		return errors.New("用户查询失败")
	}
	freeScore, err := score.QueryByUserId(userId)
	if err != nil {
		log.Error(err)
	}
	if scoreNum > userInfo.Score+freeScore.Score {
		return errors.New("积分不足，无法进行后续操作")
	}
	if !doBill {
		return nil
	}
	lock.Lock()
	var userScore = float32(0)
	//先扣赠送积分
	if freeScore != nil && freeScore.Score > 0 {
		var free = float32(0)
		if scoreNum >= freeScore.Score {
			userScore = scoreNum - freeScore.Score
			free = freeScore.Score
		} else {
			free = scoreNum
		}
		freeScore.Subtract(free)
		var bill = Bill{
			Num:     free,
			UserId:  userId,
			OrderId: orderId,
			Type:    FREE_SCORE,
			Option:  SUBTRACT,
			Comment: BILL_COMMENT_BUY, //购彩
		}
		err := BillForScore(bill, tx)
		if err != nil {
			lock.Unlock()
			tx.Rollback()
			return errors.New("记录账单失败")
		}
	} else {
		userScore = scoreNum
	}
	if userScore > 0 {
		userInfo.Score = userInfo.Score - scoreNum
		tx.Model(&userInfo).Update("score", userInfo.Score)
		var bill = Bill{
			Num:     userScore,
			UserId:  userId,
			OrderId: orderId,
			Type:    FREE_SCORE,
			Option:  SUBTRACT,
			Comment: BILL_COMMENT_BUY,
		}
		err := BillForScore(bill, tx)
		if err != nil {
			lock.Unlock()
			tx.Rollback()
			return errors.New("记录账单失败")
		}

	}
	lock.Unlock()
	return nil
}

func ReturnScore(userId uint, score float32) error {
	var user User
	var mu sync.Mutex
	mu.Lock()
	tx := mysql.DB.Begin()
	if err := tx.Model(&User{}).Where(&User{Model: gorm.Model{ID: userId}}).First(&user).Error; err != nil {
		mu.Unlock()
		return errors.New("用户查询失败")
	}

	user.Score = user.Score + score
	tx.Model(&User{}).Where(&user).Update("score", user.Score)

	mu.Unlock()
	tx.Commit()
	log.Info("退还成功， 金额：", score)
	return nil
}

type Bill struct {
	gorm.Model

	//SCORE、RMB、FREE_SCORE（赠送的积分）
	Type string `validate:"required"`
	//订单id id如果是空说明是后台加账
	OrderId string `validate:"required"`

	//数量
	Num float32 `validate:"required"`

	//用户id
	UserId uint `validate:"required"`

	//ADD SUBTRACT
	Option string `validate:"required"`

	ShopId uint

	//原因
	//"BILL_COMMENT_CASHED":   "兑奖",
	//	"BILL_COMMENT_ADD":      "充值",
	//	"BILL_COMMENT_CLEAR":    "清账",
	//	"BILL_COMMENT_ACTIVITY": "活动赠送",
	//	"BILL_COMMENT_BUY":      "购票",
	Comment string `grom:"type: varchar(255)"`
	//活动码
	ActiveCode string `grom:"type: varchar(255)"`
}

type BillVO struct {
	gorm.Model

	//SCORE、RMB、FREE_SCORE（赠送的积分）
	Type string `validate:"required"`
	//订单id id如果是空说明是后台加账
	OrderId string `validate:"required"`

	//数量
	Num float32 `validate:"required"`

	//用户id
	UserInfo UserVO

	//ADD SUBTRACT
	Option string `validate:"required"`

	ShopId uint

	//原因
	Comment string `grom:"type: varchar(255)"`
	//活动码
	ActiveCode string `grom:"type: varchar(255)"`
}

func (bill Bill) GetVO() *BillVO {
	vo := &BillVO{
		Model:      bill.Model,
		Type:       bill.Type,
		OrderId:    bill.OrderId,
		Num:        bill.Num,
		Option:     bill.Option,
		ShopId:     bill.ShopId,
		Comment:    bill.Comment,
		ActiveCode: bill.ActiveCode,
	}
	if vo.Comment == "购彩" {
		vo.Comment = BILL_COMMENT_BUY
	}
	userVo := FindUserVOById(bill.UserId)
	if userVo != (UserVO{}) {
		vo.UserInfo = userVo
	}
	return vo
}

/*
*
option  ADD(增加) SUBTRACT(扣除)
ty 账单类型 SCORE(用户积分)、RMB(人民币)、FREE_SCORE（赠送的积分）
reason 原因
*/
func BillForScore(bill Bill, tx *gorm.DB) error {
	//扣积分逻辑

	err := validatior.Validator(nil, bill)
	if err != nil {
		return err
	}
	var user User
	tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: bill.UserId,
	}}).First(&user)
	bill.ShopId = user.From
	/*if user.Score < score {

		tx.Rollback()
		lock.Unlock()
		return errors.New("余额不足")
	}
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: userId,
	}}).Update("score", user.Score-score).Error; err != nil {
		tx.Rollback()
		lock.Unlock()
		return errors.New("更新订单失败！")
	}*/

	if billerr := tx.Model(Bill{}).Save(&bill).Error; billerr != nil {
		tx.Rollback()
		return errors.New("创建账单失败")
	}
	return nil
}

type Score struct {
	//分数
	Num float32

	//用户id
	UserId uint
}

// @Summary 积分加账
// @Description 积分加账
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Score true "积分对象"
// @Router /api/super/add-score [post]
func AddScore(c *gin.Context) {
	var score Score
	err := c.BindJSON(&score)
	if err != nil {
		common.FailedReturn(c, "参数获取失败")
		return
	}
	var userSelf = getUserInfo(c)
	var lock sync.Mutex
	if score.Num <= 0 {
		common.FailedReturn(c, "积分需要大于0")
		return
	}

	lock.Lock()
	tx := mysql.DB.Begin()
	var user User
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: score.UserId,
	}, FromUser: userSelf.ID,
	}).First(&user).Error; err != nil {
		lock.Unlock()
		tx.Rollback()
		common.FailedReturn(c, "该用户不是您的用户!")
		return
	}

	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: score.UserId,
	}}).Update("score", user.Score+score.Num).Error; err != nil {
		tx.Rollback()
		lock.Unlock()
		common.FailedReturn(c, "更新订单失败")
		return
	}
	var bill = Bill{
		Num:     score.Num,
		UserId:  score.UserId,
		Type:    SCORE,
		Option:  ADD,
		Comment: BILL_COMMENT_ADD,
	}
	bill.ShopId = user.From
	if billerr := tx.Model(Bill{}).Save(&bill).Error; billerr != nil {
		common.FailedReturn(c, "创建账单失败")
		tx.Rollback()
		lock.Unlock()
		return

	}
	tx.Commit()
	lock.Unlock()
	common.SuccessReturn(c, "上分成功")
}

// @Summary 清账接口
// @Description 清账接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Score true "积分对象"
// @Router /api/super/substract-score [post]
func BillClear(c *gin.Context) {
	var score Score
	var lock sync.Mutex
	err := c.BindJSON(&score)
	if err != nil {
		common.FailedReturn(c, "参数获取失败")
		return
	}

	tx := mysql.DB.Begin()

	var user User
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: score.UserId,
	}}).First(&user).Error; err != nil {
		common.FailedReturn(c, "用户不存在")
		lock.Unlock()
		return
	}
	lock.Lock()

	if user.Score < score.Num {
		tx.Rollback()
		common.FailedReturn(c, "分数不足，请确定清账数额")
		lock.Unlock()
		return
	}

	var bill = Bill{
		Num:     score.Num,
		UserId:  score.UserId,
		Type:    SCORE,
		Option:  SUBTRACT,
		Comment: BILL_COMMENT_CLEAR,
	}

	bill.ShopId = user.From
	if billerr := tx.Model(Bill{}).Save(&bill).Error; billerr != nil {
		common.FailedReturn(c, "创建账单失败")
		tx.Rollback()
		lock.Unlock()
		return

	}

	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: score.UserId,
	}}).Update("score", user.Score-score.Num).Error; err != nil {
		tx.Rollback()
		common.FailedReturn(c, "更新积分失败")
		lock.Unlock()
		return
	}
	var notify ScoreUserNotify
	tx.Model(ScoreUserNotify{}).Where(&ScoreUserNotify{Initiator: score.UserId}).First(&notify)
	if notify != (ScoreUserNotify{}) {
		if err := tx.Model(ScoreUserNotify{}).Update("num", score.Num).Update("status", true).Where(&notify).Error; err != nil {
			tx.Rollback()
			common.FailedReturn(c, "更新清账通知失败")
			lock.Unlock()
		}
	}
	lock.Unlock()
	tx.Commit()
	common.SuccessReturn(c, "清账成功！")

}

type ScoreUserNotify struct {
	//发起人
	Initiator uint
	//给哪个店主
	ToAdminId uint
	//是否完成清账
	Status bool
	//清账方式  WECHAT(微信) CARD(银行卡) ALI(支付宝)
	Way string
	//清账分数（以实际清账为主）
	NUM float32

	gorm.Model
}

// @Summary 清账发起接口
// @Description 清账发起， 提示：清账积分以协商为主, 默认清除剩余所有！
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/user/bill/notify [post]
func BillClearNotify(c *gin.Context) {
	var user = getUserInfo(c)
	var userInfo User
	mysql.DB.Model(User{}).Where(&User{
		Model: gorm.Model{
			ID: user.ID,
		},
	}).First(&userInfo)
	if userInfo.From <= 0 {
		common.FailedReturn(c, "您没有对应的店铺， 请联系管理员！")
		return
	}
	if userInfo.Score <= 0 {
		common.FailedReturn(c, "该账户没有积分不用清账！")
		return
	}
	var notify = ScoreUserNotify{
		Initiator: user.ID,
		ToAdminId: userInfo.From,
		Way:       SCORE,
	}

	if err := mysql.DB.Save(&notify).Error; err != nil {
		common.FailedReturn(c, "提交清账失败！")
		return
	}

	common.SuccessReturn(c, "提交清账成功！")
}

// @Summary 查询自己发起清账接口
// @Description 查询自己发起清账接口
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @Router /api/user/bill/notify [get]
func BillClearNotifyList(c *gin.Context) {
	var user = getUserInfo(c)
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	var notifies []ScoreUserNotify
	mysql.DB.Model(ScoreUserNotify{}).Where(&ScoreUserNotify{Initiator: user.ID, Status: false}).Offset((page - 1) * pageSize).Limit(pageSize).Find(&notifies)
	common.SuccessReturn(c, notifies)
}

// @Summary 查询需要清账通知
// @Description 查询自己发起清账通知
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @Router /api/super/bill/notify [get]
func BillClearShopNotifyList(c *gin.Context) {
	var user = getUserInfo(c)
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	var notifies []ScoreUserNotify
	mysql.DB.Model(ScoreUserNotify{}).Where(&ScoreUserNotify{ToAdminId: user.ID, Status: false}).Offset((page - 1) * pageSize).Limit(pageSize).Find(&notifies)
	common.SuccessReturn(c, notifies)
}

// @Summary 查询店主信息
// @Description 查询店主信息
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/user/owner [get]
func GetShopOwnerInfo(c *gin.Context) {
	var user = getUserInfo(c)
	var owner User
	mysql.DB.Model(User{}).Where(&User{
		Model: gorm.Model{ID: user.From},
	}).Find(&owner)
	owner.Score = 0
	owner.Salt = ""
	common.SuccessReturn(c, owner)

}

// 修改密码对象
type UserChangePasswordDTO struct {
	//新密码
	NewPassWord string

	//老密码
	OldPassWord string

	//手机验证码（忘记密码，手机验证才用）
	Code string

	//手机号（忘记密码，手机验证才用）
	Phone string
}

func GetUserInfoByPhoneNum(num string) (User, error) {
	var user User
	mysql.DB.Where(&User{Phone: num}).First(&user)
	if user == (User{}) {
		return user, errors.New("用户不存在")
	}
	return user, nil
}

// @Summary 手机重置密码
// @Description 修改密码-手机校验
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body UserDTO true "用户对象, socre 可以不传"
// @Router /api/user/passwordByPhoneCode [post]
func ChangePasswordByPhoneCodeHandler(c *gin.Context) {
	var upd UserChangePasswordDTO
	c.BindJSON(&upd)
	//TODO 校验手机code 拿到手机号
	//通过手机号拿到用户信息

	user, err := GetUserInfoByPhoneNum("")
	if err != nil {
		common.FailedReturn(c, err.Error())
		return
	}
	salt := uuid.NewV4().String()[0:16]
	pwd, err := common.EnPwdCode([]byte(user.Secret), []byte(salt))
	if err != nil {
		log.Error("加密密码失败", err)
		common.FailedReturn(c, "加密密码失败")
		return
	}

	mysql.DB.Model(user).Where(user).Update("secret", pwd).Update("Salt", salt)
}

// @Summary 修改密码
// @Description 修改密码
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body UserChangePasswordDTO true "需改密码对象"
// @Router /api/user/password [post]
func ChangePasswordHandler(c *gin.Context) {
	var upd UserChangePasswordDTO
	c.BindJSON(&upd)
	var user = getUserInfo(c)
	salt := uuid.NewV4().String()[0:16]
	pwd, err := common.EnPwdCode([]byte(user.Secret), []byte(salt))
	if err != nil {

		log.Error("加密密码失败", err)
		common.FailedReturn(c, "加密密码失败")
		return
	}

	mysql.DB.Model(user).Where(user).Update("secret", pwd).Update("Salt", salt)
}

// 充值
func AddScoreInner(score float32, userId uint, ownerId uint, way string, tx *gorm.DB) error {
	var lock sync.Mutex

	lock.Lock()

	var user User
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: userId,
	}, From: ownerId,
	}).First(&user).Error; err != nil {
		lock.Unlock()
		tx.Rollback()
		return errors.New("该用户不是您的用户!")
	}

	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: userId,
	}}).Update("score", user.Score+score).Error; err != nil {
		tx.Rollback()
		lock.Unlock()
		return errors.New("更新订单失败")
	}
	var bill = Bill{
		Num:     score,
		UserId:  userId,
		Type:    way,
		Option:  ADD,
		Comment: BILL_COMMENT_ADD,
	}
	bill.ShopId = user.From
	if billerr := tx.Model(Bill{}).Save(&bill).Error; billerr != nil {
		tx.Rollback()
		lock.Unlock()
		return errors.New("创建账单失败")

	}
	lock.Unlock()
	return nil
}

// 机器兑奖
func AddScoreInnerByMachine(score float32, userId uint, way string, tx *gorm.DB) error {
	var lock sync.Mutex

	lock.Lock()

	var user User
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: userId,
	},
	}).First(&user).Error; err != nil {
		lock.Unlock()
		tx.Rollback()
		return errors.New("该用户不是您的用户!")
	}

	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: userId,
	}}).Update("score", user.Score+score).Error; err != nil {
		tx.Rollback()
		lock.Unlock()
		return errors.New("更新订单失败")
	}
	var bill = Bill{
		Num:     score,
		UserId:  userId,
		Type:    way,
		Option:  ADD,
		Comment: BILL_COMMENT_CASHED,
	}
	bill.ShopId = user.From
	if billerr := tx.Model(Bill{}).Save(&bill).Error; billerr != nil {
		tx.Rollback()
		lock.Unlock()
		return errors.New("创建账单失败")

	}
	lock.Unlock()
	return nil
}
