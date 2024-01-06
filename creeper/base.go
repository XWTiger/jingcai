package creeper

import (
	"gorm.io/gorm"
	"strings"
)

type Content struct {
	gorm.Model
	//网站名称
	Type     string `json:"type" example:"雷速" `
	Content  string
	ImageUrl []string `gorm:"-:all"`
	Url      string
	Summery  string
	//额外的一些信息
	Extra string
	Title string
	//比赛
	Match string
	//预测谁赢
	Predict string
	//条件 让球 1.25
	Conditions []string `gorm:"-:all"`
	time       string
	// 联赛
	league string

	UserID uint
}

/*
条件表
*/
type Condition struct {
	gorm.Model
	ParentId  uint
	Condition string
}

type Creeper interface {
	Creep() []Content
	Key() string
}

// 过滤特有字段（分收费68球币,已有3人查看）
func (c *Content) FilterStrInSummery() {

	start := strings.Index(c.Summery, "收费")

	end := strings.LastIndex(c.Summery, "查看")

	if start > 0 && end > 0 {
		c.Summery = c.Summery[start:end]
	}
}
