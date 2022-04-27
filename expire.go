package goexpire

import (
	"errors"
	"time"

	"github.com/rs/xid"
)

type JobExpire struct {
	queue *JobQueue
}

func cronJobLater() {
	ticket := time.NewTicker(1 * time.Second)

	for range ticket.C {
		if time.Now().Second() == 59 &&
			time.Now().Minute() == 59 &&
			time.Now().Hour() == 23 {
			Expire.checkJobToday()
		}
	}
}

func NewExpire() *JobExpire {
	go cronJobLater()

	return &JobExpire{
		queue: Expire,
	}
}

func (j *JobExpire) Add() *Job {
	return &Job{
		ID:        xid.New().String(),
		TotalTime: 0,
		Type:      Today,
	}
}

// Terminal goroutine
func (j *JobExpire) Remove(id string) error {
	Expire.mu.Lock()
	defer Expire.mu.Unlock()

	if _, ok := Expire.currentJob[id]; ok {
		delete(Expire.currentJob, id)
		return nil
	}

	if _, ok := Expire.laterJob[id]; ok {
		delete(Expire.laterJob, id)
		return nil
	}

	return errors.New("Job not exit")
}

func (j *JobExpire) Stop(id string) error {
	Expire.mu.Lock()
	defer Expire.mu.Unlock()

	if _, ok := Expire.stopJob[id]; ok {
		return nil
	}

	if _, ok := Expire.currentJob[id]; ok {
		Expire.stopJob[id] = Expire.currentJob[id]
		delete(Expire.currentJob, id)
		return nil
	}

	if _, ok := Expire.laterJob[id]; ok {
		Expire.stopJob[id] = Expire.laterJob[id]
		delete(Expire.laterJob, id)
		return nil
	}

	return errors.New("Job not exit")
}

// Check time pass
func (j *JobExpire) Start(id string) error {
	Expire.mu.Lock()
	defer Expire.mu.Unlock()

	if _, ok := Expire.stopJob[id]; ok {
		delete(Expire.stopJob, id)
		return nil
	}

	return errors.New("Job not exit")
}
