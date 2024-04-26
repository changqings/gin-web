package middle

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var lastRequestTime int64
var mutex sync.Mutex

// should be as close as request begin
func Limiter(interval time.Duration) gin.HandlerFunc {
	mutex.Lock()
	defer mutex.Unlock()

	// !!notice, if now define here, it will only be exec once
	return func(ctx *gin.Context) {
		// time from 1970 second
		now := time.Now().UTC().Unix()
		if now-lastRequestTime < int64(interval.Seconds()) {
			ctx.AbortWithStatusJSON(429, gin.H{"error": "Request too frequently"})
			return
		}

		//
		lastRequestTime = now

		ctx.Next()
	}

}
