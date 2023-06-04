package order

import (
	"github.com/jinzhu/gorm"
	"jingcai/user"
)

// 合买
type AllWinVO struct {
	gorm.Model
	//份数
	Number int

	//关联Order
	Order Order

	//发起人
	Initiator user.UserDTO

	//发起成功/失败
	Status string

	//购买份数
	BuyNumber int
}
