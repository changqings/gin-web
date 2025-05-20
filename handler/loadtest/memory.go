package loadtest

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	defultDuration = 60
	defaultSize    = 512
)

type LoadtestMemory struct {
	Duration int
	SizeMB   int
}

func NewLoadtestMemory() *LoadtestMemory {
	return &LoadtestMemory{
		Duration: defultDuration,
		SizeMB:   defaultSize,
	}
}

func (l *LoadtestMemory) Run() gin.HandlerFunc {
	return func(c *gin.Context) {
		var duration, sizeMB int

		durationStr := c.Query("duration")
		if len(durationStr) > 0 {
			var err error
			duration, err = strconv.Atoi(durationStr)
			if err != nil {
				slog.Error("parse duration error, use default duration", "seconds", l.Duration, "err", err)
				duration = l.Duration
			}
		} else {
			duration = l.Duration
		}
		sizeStr := c.Query("size")
		if len(sizeStr) > 0 {
			var err error
			sizeMB, err = strconv.Atoi(sizeStr)
			if err != nil {
				slog.Error("parse size error, use default size", "size", l.SizeMB, "err", err)
				sizeMB = l.SizeMB
			}

		} else {
			sizeMB = l.SizeMB
		}

		go burnMemory(duration, sizeMB)
		c.JSON(http.StatusOK, gin.H{
			"msg":              "start memory load test",
			"start time":       time.Now().Format(time.RFC3339),
			"duration seconds": duration,
			"size mb":          sizeMB,
		})

	}
}

func burnMemory(duration, sizeMB int) {
	byteSizeTotal := sizeMB * 1024 * 1024

	tmp := make([]byte, byteSizeTotal)
	for i := range byteSizeTotal {
		tmp[i] = 1
	}
	<-time.After(time.Duration(duration) * time.Second)
	slog.Info("memory load test finished", "duration seconds", duration, "size MB", sizeMB)
}
