package score

import (
	"errors"
	"gorm.io/gorm"
	"jingcai/mysql"
)

// 赠送积分
type FreeScore struct {
	gorm.Model
	UserId  uint
	Score   float32 `gorm:"type: decimal(20,6)"`
	Comment string  `grom:"type: varchar(255)"`
	//来自活动名称
	ActiveName string `grom:"type: varchar(255)"`
	//活动码
	ActiveCode string `grom:"type: varchar(255)"`
}

func QueryByUserId(userId uint) (*FreeScore, error) {
	var fee FreeScore
	if err := mysql.DB.Where(&FreeScore{Model: gorm.Model{
		ID: userId,
	}}).First(&fee).Error; err != nil {
		return nil, err
	}
	return &fee, nil
}

func (fs *FreeScore) Subtract(score float32) error {

	if fs.Score < score {
		return errors.New("积分不够")
	}
	fs.Score -= score
	err := mysql.DB.Save(fs).Error
	if err != nil {
		return err
	}
	return nil
}

func (fs *FreeScore) Add(score float32) error {

	fs.Score += score
	err := mysql.DB.Save(fs).Error
	if err != nil {
		return err
	}
	return nil
}
