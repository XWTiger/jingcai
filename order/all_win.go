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

	//发起人
	Initiator user.UserDTO

	//发起成功/失败
	Status string
}
