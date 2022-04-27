package goexpire

import "sync"

type JobQueue struct {
	mu         *sync.Mutex
	currentJob map[string]*Job
	laterJob   map[string]*Job
	stopJob    map[string]*Job
	doneJob    map[string]bool
}

var Expire = &JobQueue{
	&sync.Mutex{},
	map[string]*Job{},
	map[string]*Job{},
	map[string]*Job{},
	map[string]bool{},
}

func (j *JobQueue) setJobDone(id string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.currentJob[id]; ok {
		delete(j.currentJob, id)
		j.doneJob[id] = true
	}
}

func (j *JobQueue) setJobToday(id string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.laterJob[id]; ok {
		j.currentJob[id] = j.laterJob[id]
		delete(j.laterJob, id)
	}
}

func (j *JobQueue) checkJobToday() {
	j.mu.Lock()
	defer j.mu.Unlock()

	for id, job := range j.laterJob {
		second := job.TotalTime - int(job.untilMidnight)
		if ok, _ := afterMidnight(second); ok {
			job.untilMidnight = 60 * 60 * 24
			job.TotalTime = second
		} else {
			j.setJobToday(id)
		}
	}
}
