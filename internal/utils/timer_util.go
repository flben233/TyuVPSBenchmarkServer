package utils

import (
	"time"
)

func SetInterval(fn func(), interval int) chan bool {
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				fn()
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}
	}()
	return stop
}
