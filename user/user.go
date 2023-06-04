package user

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	gorsa "github.com/Lyafei/go-rsa"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"github.com/pascaldekloe/jwt"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
	"net/http"
	"strings"
	"time"
)

var log = ilog.Logger
var userCahe = cache2go.Cache("user")

const TOKEN_TIME_OUT = 4 * time.Hour

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
}

// 用户对象
type UserVO struct {
	Phone string `minLength:"11" maxLength:"11"`
	//昵称
	Name string `minLength:"4" maxLength:"16"`
	//密码
	Secret string `minLength:"6" maxLength:"16"`
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
// @Router /user/info [get]
func GetUserInfo(c *gin.Context) {
	var user = getUserInfo(c)
	var userPO User
	var param = User{}
	param.ID = user.ID

	if mysql.DB.Model(param).First(&userPO).Error != nil {
		common.FailedReturn(c, "查询信息失败")
	}
	var userDTO = UserDTO{
		Phone:  userPO.Phone,
		Name:   userPO.Name,
		Role:   userPO.Role,
		Wechat: userPO.Wechat,
		Ali:    userPO.Ali,
		Score:  userPO.Score,
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
// @Router /user/info [post]
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
		Phone:  userDTO.Phone,
		Name:   userDTO.Name,
		Role:   userDTO.Role,
		Wechat: userDTO.Wechat,
		Ali:    userDTO.Ali,
	}
	if mysql.DB.Model(param).First(&realUsr).Error != nil {
		common.FailedReturn(c, "查不到当前用户")
	}
	if err := mysql.DB.Model(param).Updates(&update).Error; err != nil {
		log.Error("update user failed, id: ", user.ID, " err: ", err)
		common.FailedReturn(c, "更新用户失败")
	}
}

func CreateUser(user UserVO) error {
	var userPo User = User{
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
// @Router /user [post]
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
// @Router /user/login [post]
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
		if mysql.DB.Model(&User{Phone: userVo.Phone}).Find(&user).Error != nil {
			common.FailedReturn(c, "账户错误")
			return
		}
		real, errDePwd := common.DePwdCode(user.Secret, []byte(user.Salt))
		if errDePwd != nil || err != nil {
			log.Error("decode user password failed")
			common.FailedReturn(c, "解析密码失败")
			return
		}
		if strings.Compare(pwd, string(real)) != 0 {
			common.FailedReturn(c, "账户或者密码错误")
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
// @Router /user/logout [post]
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

// @Summary 投诉
// @Description 投诉
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Complain true "投诉对象"
// @Router /user/complain [post]
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
