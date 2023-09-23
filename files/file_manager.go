package files

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/muesli/cache2go"
	uuid "github.com/satori/go.uuid"
	"jingcai/common"
	"jingcai/config"
	ilog "jingcai/log"
	"jingcai/mysql"
	"net/url"
	"os"
	"runtime"
	"strings"
)

var log = ilog.Logger

const PATH_LINUX = "/opt/jingcai/"
const PATH_WINDOW = "D:\\opt\\jingcai\\"

var ImageUrl string
var userCache = cache2go.Cache("user")

type FileStore struct {
	gorm.Model
	FilePath string
	From     string //BBS,USER,HEADER
}

func Init(c *config.Config) {
	ImageUrl = c.HttpConf.ImagePrefix
}

// @Summary 上传文件或者图片
// @Description 上传文件或者图片
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param type query string false "BBS,USER,HEADER 或则null"
// @Param files formData file true "文件"
// @Router /api/upload [post]
func Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		log.Error("get file failed", err)
		common.FailedReturn(c, "获取文件失败")
		return
	}
	// 获取所有图片
	files := form.File["files"]
	typeFile := c.Query("type")
	var filestores = make([]FileStore, 0)
	// 遍历所有图片
	for _, file := range files {
		// 逐个存
		dir, _ := os.Getwd()
		fmt.Println(dir)
		var path string
		if strings.Compare(runtime.GOOS, "windows") == 0 {
			path = PATH_WINDOW
		} else {
			path = PATH_LINUX
		}
		os.Mkdir(path, 0755)
		id := uuid.NewV4().String()
		savePath := fmt.Sprintf("%s%s_%s", path, id, file.Filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			log.Error("upload err %s", err.Error())
			common.FailedReturn(c, "上传图片失败")
			return
		}
		filestores = append(filestores, FileStore{
			From:     typeFile,
			FilePath: fmt.Sprintf("%s%s_%s", ImageUrl, id, file.Filename),
		})
	}
	log.Info("=== upload ok %d files ===", len(files))
	mysql.DB.AutoMigrate(&FileStore{})
	if err := mysql.DB.Create(&filestores).Error; err != nil {
		log.Error("mysql create failed", err)
		common.FailedReturn(c, "保存图片失败")
		return
	}
	common.SuccessReturn(c, filestores)
}

// @Summary 下载文件/图片
// @Description 下载文件/图片
// @Accept json
// @Produce json
// @Success 200 {object}  common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param file query string true "文件名称"
// @Router /api/download/:name [get]
func DownLoad(c *gin.Context) {
	file := c.Param("name")

	if file == "" {
		common.FailedReturn(c, "填正确的图片地址")
	}
	var path string
	if strings.Compare(runtime.GOOS, "windows") == 0 {
		path = PATH_WINDOW
	} else {
		path = PATH_LINUX
	}

	var fileStore FileStore
	if err := mysql.DB.Model(FileStore{}).Where(FileStore{FilePath: fmt.Sprintf("%s%s", "/api/download/", file)}).First(&fileStore).Error; err != nil {
		log.Error(err)
		common.FailedReturn(c, "图片不存在")
		return
	}

	if strings.Compare(fileStore.From, "BBS") == 0 {
		//论坛得图片不作限制
	} else {
		var token string
		token = c.Query("token") // 访问令牌
		if token == "" {
			token = c.GetHeader("token")
			if token == "" {
				fmt.Println("== get token failed ===")
				common.FailedAuthReturn(c, "访问未授权")
				c.Abort()
				return
			}
		}
		_, err := userCache.Value(token)
		if err != nil {
			c.Abort()
			common.FailedAuthReturn(c, "token已过期")
			return
		}
	}

	strFile := fmt.Sprintf("attachment; filename*=utf-8' '%s", url.QueryEscape(file))
	c.Writer.Header().Add("Content-Type", "application/octet-stream;charset=utf-8")
	c.Writer.Header().Add("Content-Disposition", strFile)
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(fmt.Sprintf("%s%s", path, file))

	return
}
