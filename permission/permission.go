package permission

import (
	"gorm.io/gorm"
)

// 权限对象
type Permissions struct {
	gorm.Model
	//接口路径
	UrlPath string

	//描述
	Describe string

	//权限编码
	Code string

	//图片地址
	Icon string

	//MENU 菜单 / BUTTON 按钮 / INTERFACE 接口
	Type string

	//父级id
	ParentId uint

	//排序字段
	Sort int

	//名称
	Name string
}
