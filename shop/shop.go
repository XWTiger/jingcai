package shop

import "gorm.io/gorm"

type Shop struct {
	gorm.Model
	//门店地址
	Addr string

	UserId uint

	//证件地址
	Certificate string

	//经度
	Longitude float32

	//纬度
	Latitude float32

	//门店名称
	Name string

	//审核状态
	Status bool
}

//TODO 分享二维码 和门店流水
