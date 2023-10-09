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
	//如果是历史任务的话，传入job id 方便执行完修改状态
	JobId uint
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
	if job.JobId >= 0 {
		return
	}
	start, end := common.GetDateStartAndEnd(job.Time)
	var count int64
	mysql.DB.Model(&JobExecution{}).Where("time between ? and ?", start, end).Where("type = ?", job.Type).Count(&count)
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

func OrderCheckInit() {
	var jobs []JobExecution
	var counts []int
	mysql.DB.Model(JobExecution{}).Where(&JobExecution{Status: false}).Find(&jobs)
	for _, job := range jobs {
		switch job.Type {
		case FOOTBALL:
			if job.Time.Unix() > time.Now().Unix() {
				CheckLottery(job.Time)
			} else if counts[0] <= 0 {
				now := time.Now()
				CheckLottery(now.Add(time.Minute * 5))
				counts[0] = 1
			}
			break
		case P3:
			if counts[1] <= 0 {
				AddPlwCheck(3)
				counts[1] = 1
			}
			break
		case BASKETBALL:
			if job.Time.Unix() > time.Now().Unix() {
				CheckBasketBallLottery(job.Time)
			} else if counts[2] <= 0 {
				now := time.Now()
				CheckBasketBallLottery(now.Add(time.Minute * 5))
				counts[2] = 1
			}
			break
		case SEVEN_STAR:
			if counts[3] <= 0 {
				AddSevenStarCheck()
				counts[3] = 1
			}
			break
		case SUPER_LOTTO:
			if counts[4] <= 0 {
				AddSuperLottoCheck()
				counts[4] = 1
			}
			break
		case P5:
			if counts[5] <= 0 {
				AddPlwCheck(5)
				counts[5] = 1
			}
			break

		}
		mysql.DB.Model(JobExecution{}).Update("status", true).Where("id = ?", job.ID)
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
