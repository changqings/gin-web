package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	REDIS_HOST_PORT = "192.168.1.201:6379"

	HighAvailableKey  = "ha_key_should_be_uniq"
	ShouldRunAsMaster bool
)

type RedisLock struct {
	Key    string
	Value  int64
	Client *redis.Client
}

// 设置过期时间为25s,而定时更新时间为20s
// 保证在程序正常运行的情况下，key一直不会过期
func init() {

	r := &RedisLock{}
	r.New(HighAvailableKey, 0)

	ticker := time.NewTicker(time.Second * 20)
	updateFunc := func() {
		// 定期更新key, 否则会自动过期
		// slave会watch这个key
		// slave如果发现这个key不存在
		// 则会获取成为master,并继续更新这个key
		if !r.Updatelock(time.Now().Unix()) {
			slog.Error("db.redis.go updateFunc", "msg", "更新redis key失败, 会导致多副本同时运行，请检查")
		}
	}

	// 如果设置成功，则后台运行更新函数
	if r.SetLock() {
		//  run as master
		ShouldRunAsMaster = true
		go func() {
			for range ticker.C {
				updateFunc()
			}
		}()
		// 如果没有设置成功，则后台尝试更新，如果更新成功
		// 则说明原master已不存在
	} else {
		ShouldRunAsMaster = false
		go func() {
			for range ticker.C {
				// 定期尝试SetLock(),如果成功，则退出监听
				if r.SetLock() {
					ShouldRunAsMaster = true
					// slog.Info("redis debug", "shouldRunAsMaster", ShouldRunAsMaster)
					break
				}
			}
			// 上面的for range在SetLock()后，此for range接管
			// 使用updateFunc()继续定期更新
			for range ticker.C {
				updateFunc()
			}

		}()
	}

}

func (r *RedisLock) New(k string, db_num int) {

	r.Key = k
	r.Value = time.Now().Unix()
	r.Client = redis.NewClient(&redis.Options{
		Addr: REDIS_HOST_PORT,
		DB:   db_num,
	})
}

// 更新使用XX, 即存在key才会更新，否则不更新，只有master才会执行这个函数
// 否则，不应该执行这个函数
func (r *RedisLock) Updatelock(v int64) bool {
	b, _ := r.Client.SetXX(context.Background(), r.Key, v, time.Second*25).Result()
	return b

}

// 设置使用NX，如果原来不存在，即认为没有master在运行，即获取master权限
func (r *RedisLock) SetLock() bool {

	b, _ := r.Client.SetNX(context.Background(), r.Key, r.Value, time.Second*25).Result()

	return b
}
