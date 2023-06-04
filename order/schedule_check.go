package order

import (
	"errors"
	"time"
)

type JobFun interface {
	Execute() error
}

type Job struct {
	Time     time.Time
	Param    interface{}
	CallBack func(param interface{})
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
