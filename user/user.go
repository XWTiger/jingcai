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
	"gorm.io/gorm/clause"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
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
const (
	SCORE    = "SCORE"
	RMB      = "RMB"
	ADD      = "ADD" //增加
	SUBTRACT = "SUBTRACT"
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
	Role string //"enum: Admin,User"
	//微信号
	Wechat string
	//支付宝号
	Ali string
	//余额
	Score float32
	//头像地址
	HeaderImageUrl string

	//来自推荐的码/店铺（管理员的用户id）
	From uint
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

	From uint
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

	if mysql.DB.Model(param).First(&userPO).Error != nil {
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
	common.SuccessReturn(c, userDTO)
}

// @Summary 更新用户信息
// @Description 更新用户信息
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

func CreateUser(user UserVO) error {
	var userPo = User{
		Phone:  user.Phone,
		Secret: user.Secret,
		Name:   user.Name,
		Salt:   uuid.NewV4().String()[0:16],
		Role:   "User",
		Score:  0.00,
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
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body UserVO true "用户对象"
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
	var intId int
	id := c.Query("sharedId")
	if strings.Compare(id, "") == 0 {
		intId = 1
	} else {
		ind, err := strconv.Atoi(id)
		if err != nil {
			log.Error(err)
			intId = 1
		} else {
			intId = ind
		}
	}
	user.From = uint(intId)
	if CreateUser(user) != nil {
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
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body TokenVO true "token对象"
// @Router /api/user/logout [post]
func Logout(c *gin.Context) {
	var tokenVO TokenVO
	c.BindJSON(&tokenVO)
	if tokenVO.Token != "" {
		userCahe.Delete(tokenVO.Token)
	}
	common.SuccessReturn(c, "注销成功")
}

func getUserInfo(c *gin.Context) User {
	user, _ := c.Get("userInfo")
	return user.(User)
}

func FetUserInfo(c *gin.Context) User {
	user, _ := c.Get("userInfo")
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

func CheckScoreOrDoBill(userId uint, score float32, doBill bool) error {
	var user User
	tx := mysql.DB.Begin()
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{ID: userId}}).Clauses(clause.Locking{Strength: "UPDATE"}).First(&user).Error; err != nil {
		return errors.New("用户查询失败")
	}
	if score > user.Score {
		return errors.New("积分不足，无法进行后续操作")
	}
	if doBill {
		user.Score = user.Score - score

		tx.Model(&user).Update("score", user.Score)

	}
	tx.Commit()
	return nil
}

func ReturnScore(userId uint, score float32) error {
	var user User
	tx := mysql.DB.Begin()
	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{ID: userId}}).Clauses(clause.Locking{Strength: "UPDATE"}).First(&user).Error; err != nil {
		return errors.New("用户查询失败")
	}
	if score > user.Score {
		return errors.New("积分不足，无法进行后续操作")
	}
	user.Score = user.Score + score
	tx.Model(&user).Update("score", user.Score)

	tx.Commit()
	return nil
}

type Bill struct {
	gorm.Model

	//SCORE、RMB
	Type string
	//订单id
	OrderId string

	//数量
	Num float32

	//用户id
	UserId uint

	//ADD SUBTRACT
	Option string

	ShopId uint
}

func BillForScore(OrderId string, userId uint, score float32) error {
	var lock sync.Mutex
	//扣积分逻辑
	lock.Lock()
	tx := mysql.DB.Begin()
	var bill = Bill{
		Num:     score,
		UserId:  userId,
		OrderId: OrderId,
		Type:    SCORE,
		Option:  SUBTRACT,
	}
	var user User
	tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: userId,
	}}).First(&user)
	if user.Score < score {

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
	}

	if billerr := tx.Model(Bill{}).Save(bill).Error; billerr != nil {
		tx.Rollback()
		lock.Unlock()
		return errors.New("创建账单失败")
	}
	tx.Commit()
	lock.Unlock()
	return nil
}

type Score struct {
	//分数
	Num float32

	//用户id
	UserId uint
}

// @Summary 积分充值
// @Description 积分充值
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Score true "投诉对象"
// @Router /api/super/add-score [post]
func AddScore(c *gin.Context) {
	var score Score
	err := c.BindJSON(&score)
	if err != nil {
		common.FailedReturn(c, "参数获取失败")
		return
	}
	var lock sync.Mutex
	if score.Num <= 0 {
		common.FailedReturn(c, "积分需要大于0")
		return
	}

	lock.Lock()
	tx := mysql.DB.Begin()
	var user User
	tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: score.UserId,
	}}).First(&user)

	if err := tx.Model(User{}).Where(&User{Model: gorm.Model{
		ID: score.UserId,
	}}).Update("score", user.Score+score.Num).Error; err != nil {
		tx.Rollback()
		lock.Unlock()
		common.FailedReturn(c, "更新订单失败")
		return
	}
	tx.Commit()
	lock.Unlock()
	common.FailedReturn(c, "上分成功")
}
