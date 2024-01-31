package order

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/util"
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
	if job.JobId > 0 {
		return
	}
	start, end := common.GetDateStartAndEnd(job.Time)
	var count int64
	mysql.DB.Model(&JobExecution{}).Where("time between ? and ?", start, end).Where("type = ? and status = ?", job.Type, 0).Count(&count)
	if count <= 1 {
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
	var counts []int = make([]int, 7)
	mysql.DB.Debug().Model(JobExecution{}).Where("status=?", false).Find(&jobs)
	now := time.Now()
	nowAdd := now.Add(time.Minute * 5)
	for _, job := range jobs {
		switch job.Type {
		case FOOTBALL:
			if job.Time.Unix() > time.Now().Unix() {
				CheckLottery(job.Time)
			} else if counts[0] <= 0 && !job.Status {

				CheckLottery(now.Add(time.Minute * 5))
				counts[0] = 1
			}
			break
		case P3:
			if job.Time.Unix() > time.Now().Unix() {
				AddPlwCheck(3, &job.Time)
			} else if counts[1] <= 0 && !job.Status {
				AddPlwCheck(3, &nowAdd)
				counts[1] = 1
			}
			break
		case BASKETBALL:
			if job.Time.Unix() > time.Now().Unix() {
				CheckBasketBallLottery(job.Time)
			} else if counts[2] <= 0 && !job.Status {
				now := time.Now()
				CheckBasketBallLottery(now.Add(time.Minute * 5))
				counts[2] = 1
			}
			break
		case SEVEN_STAR:
			if job.Time.Unix() > time.Now().Unix() {
				AddSevenStarCheck(&job.Time)
			} else if counts[3] <= 0 && !job.Status {
				AddSevenStarCheck(&nowAdd)
				counts[3] = 1
			}
			break
		case SUPER_LOTTO:
			if job.Time.Unix() > time.Now().Unix() {
				AddSuperLottoCheck(&job.Time)
			} else if counts[4] <= 0 && !job.Status {
				AddSuperLottoCheck(&nowAdd)
				counts[4] = 1
			}
			break
		case P5:
			if job.Time.Unix() > time.Now().Unix() {
				AddPlwCheck(5, &job.Time)
			} else if counts[5] <= 0 && !job.Status {
				AddPlwCheck(5, &nowAdd)
				counts[5] = 1
			}
			break
		case ALL_WIN:
			if job.Time.Unix() > time.Now().Unix() {
				AllWinCheck(job.Time)
			} else if counts[6] <= 0 && !job.Status {
				AllWinCheck(nowAdd)
				counts[6] = 1
			}

		}
		mysql.DB.Model(JobExecution{}).Where("id = ?", job.ID).Update("status", true)
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

// 时间（2023223141）+ userId（000001） + 订单类型（0-5）+ 是否分享（00/01）
func GetOrderId(order *Order) string {
	now := time.Now()
	y, m, d := now.Date()
	strDate := fmt.Sprintf("%d%s%s%s%s%d", y, common.GetNum(int(m)), common.GetNum(d), common.GetNum(now.Hour()), common.GetNum(now.Minute()), now.Second())
	usrId := util.GetPaddingId(order.UserID)
	var typ string
	switch order.LotteryType {
	case FOOTBALL:
		//01

		typ = "01"
		break
	case BASKETBALL:
		//02
		typ = "02"
		break
	case P3:
		//03
		typ = "03"
		break
	case P5:
		//04
		typ = "04"
		break
	case SUPER_LOTTO:
		//05
		typ = "05"
		break
	case SEVEN_STAR:
		//06
		typ = "06"
		break
	}
	var share string
	if order.Share {
		share = "01"
	} else {
		share = "00"
	}

	return fmt.Sprintf("%s%s%s%s", strDate, usrId, typ, share)

}
