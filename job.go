package goexpire

import (
	"reflect"
	"time"

	"github.com/rs/xid"
)

type JobType int

const (
	Today JobType = iota
	Later
)

type Job struct {
	ID        string
	TotalTime int
	Type      JobType

	fn            interface{}
	args          []interface{}
	jobTimer      *JobTimer
	untilMidnight int64
}

type JobTimer struct {
	ID    string
	timer *time.Timer
}

func afterMidnight(second int) (bool, int64) {
	currentTime := time.Now()
	newTime := currentTime
	newTime.Add(time.Duration(second))

	if newTime.Day() > currentTime.Day() {
		midnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, time.UTC)
		return true, midnight.Unix() - currentTime.Unix()
	} else {
		return false, -1
	}
}

func (j *Job) Second(n int) *Job {
	if ok, diffTime := afterMidnight(n); j.Type == Today && ok {
		j.Type = Later
		j.untilMidnight = diffTime
		Expire.laterJob[j.ID] = j
	}

	j.TotalTime += n

	return j
}

func (j *Job) Minute(n int) *Job {
	second := n * 60

	if ok, diffTime := afterMidnight(n); j.Type == Today && ok {
		j.Type = Later
		j.untilMidnight = diffTime
		Expire.laterJob[j.ID] = j
	}

	j.TotalTime += second

	return j
}

func (j *Job) Hour(n int) *Job {
	second := n * 60 * 60

	if ok, diffTime := afterMidnight(n); j.Type == Today && ok {
		j.Type = Later
		j.untilMidnight = diffTime
		Expire.laterJob[j.ID] = j
	}

	j.TotalTime += second

	return j
}

func (j *Job) Day(n int) *Job {
	second := n * 60 * 60 * 24

	if ok, diffTime := afterMidnight(n); j.Type == Today && ok {
		j.Type = Later
		j.untilMidnight = diffTime
		Expire.laterJob[j.ID] = j
	}

	j.TotalTime += second

	return j
}

func (j *Job) Do(fn interface{}, args ...interface{}) {
	j.fn = fn
	j.args = args

	if j.Type == Today {
		j.jobTimer = new(JobTimer)
		j.jobTimer.ID = xid.New().String()
		j.jobTimer.timer = time.NewTimer(time.Duration(j.TotalTime) * time.Second)

		go func() {
			<-j.jobTimer.timer.C

			j.exec()

			Expire.setJobDone(j.ID)
		}()
	}
}

func (j *Job) exec() {
	reflectFn := reflect.ValueOf(j.fn)

	reflectArgs := make([]reflect.Value, len(j.args))
	for i, arg := range j.args {
		reflectArgs[i] = reflect.ValueOf(arg)
	}

	reflectFn.Call(reflectArgs)
}
