package loadtest

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type LoadTestCpu struct {
	DurationSeconds int
	workers         int
}

func NewLoadTestCpu() *LoadTestCpu {
	return &LoadTestCpu{
		DurationSeconds: 60,
		workers:         10,
	}
}

func (l *LoadTestCpu) Run() gin.HandlerFunc {
	return func(c *gin.Context) {
		var duration int

		durationStr := c.Query("duration")
		if len(durationStr) > 0 {
			var err error
			duration, err = strconv.Atoi(durationStr)
			if err != nil {
				slog.Error("parse duration error, use default duration", "seconds", l.DurationSeconds, "err", err)
				duration = l.DurationSeconds
			}
		} else {
			duration = l.DurationSeconds
		}

		go burnCpu(duration, l.workers)

		c.JSON(http.StatusOK, gin.H{
			"msg":              "start cpu load test",
			"start time":       time.Now().Format(time.RFC3339),
			"duration seconds": duration,
		})

	}
}

func burnCpu(duration, workers int) {
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					for x := 0; x < 1e6; x++ {
						x = x * x
					}
				}
			}
		}()
	}

	go func() {
		<-time.After(time.Duration(duration) * time.Second)
		close(stopCh)
	}()

	wg.Wait()
	slog.Info("cpu load test finished", "duration seconds", duration)
}
