package order

import (
	"errors"
	"gorm.io/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"time"
)

type JobFun interface {
	Execute() error
}

type Job struct {
	Time     time.Time
	Param    interface{}
	CallBack func(param interface{})
	//FOOTBALL    = "FOOTBALL"
	//	SUPER_LOTTO = "SUPER_LOTTO"
	//	P3          = "P3"
	//	P5          = "P5"
	//	BASKETBALL  = "BASKETBALL"
	//	SEVEN_STAR  = "SEVEN_STAR"
	Type string
}

type JobExecution struct {
	gorm.Model
	//false 未执行， true 执行
	Status bool
	Time   time.Time

	//FOOTBALL    = "FOOTBALL"
	//	SUPER_LOTTO = "SUPER_LOTTO"
	//	P3          = "P3"
	//	P5          = "P5"
	//	BASKETBALL  = "BASKETBALL"
	//	SEVEN_STAR  = "SEVEN_STAR"
	Type string
}

// 添加job 记录，一天内记录只有存一次
func RecordJob(job *Job) {
	start, end := common.GetDateStartAndEnd(job.Time)
	var count int64
	mysql.DB.Model(&JobExecution{}).Where("created_at between ? and ?", start, end).Where("type = ?", job.Type).Count(&count)
	if count <= 0 {
		var je = JobExecution{
			Type:   job.Type,
			Time:   job.Time,
			Status: false,
		}
		mysql.DB.Save(&je)
		job.Param = je.ID
	}
}
func doSchedule(job Job) {
	dtime := job.Time.Unix()
	now := time.Now().Unix()
	if dtime > now {
		after := dtime - now
		timer1 := time.NewTimer(time.Duration(after) * time.Second)
		select {
		case <-timer1.C:
			job.Execute()
		}
	}
}

func AddJob(job Job) error {
	RecordJob(&job)
	if job.Time == (time.Time{}) || job.CallBack == nil {
		return errors.New("time  callback required")
	}
	go doSchedule(job)
	return nil
}

func (check Job) Execute() error {
	check.CallBack(check.Param)
	return nil
}
