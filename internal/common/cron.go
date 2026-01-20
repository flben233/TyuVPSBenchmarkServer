package common

import "time"

type CronJob struct {
	Interval time.Duration
	Job      func()
}

var jobs = make([]CronJob, 0)

func RegisterCronJob(interval time.Duration, job func()) {
	jobs = append(jobs, CronJob{
		Interval: interval,
		Job:      job,
	})
}

func RunCronJobs() {
	for _, cronJob := range jobs {
		go func(cj CronJob) {
			ticker := time.NewTicker(cj.Interval)
			defer ticker.Stop()
			for range ticker.C {
				cj.Job()
			}
		}(cronJob)
	}
}
