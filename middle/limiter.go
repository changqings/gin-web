package middle

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var lastRequestTime int
var mutex sync.Mutex

// should be as close as request begin
func Limiter(interval time.Duration) gin.HandlerFunc {

	// !!notice, if ts define here, it will only be exec only once
	return func(ctx *gin.Context) {
		mutex.Lock()
		// time from 1970 second
		now := time.Now().UTC().Second()
		if now-int(lastRequestTime) < int(interval.Seconds()) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Request too frequently"})
			return
		}

		//
		lastRequestTime = now
		mutex.Unlock()

		ctx.Next()
	}

}
