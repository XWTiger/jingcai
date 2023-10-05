package permission

import (
	"gorm.io/gorm"
)

// 权限对象
type permissions struct {
	gorm.Model
	//接口路径
	UrlPath string

	//描述
	Describe string

	//权限编码
	Code string
}
