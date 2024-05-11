package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	REDIS_HOST_PORT   = "192.168.1.201:6379"
	HighAvailable_key = "ha_key_should_be_uniq"
	ShouldRunAsMaster bool
)

type RedisLock struct {
	Key    string
	Value  int64
	Client *redis.Client
}

func init() {

	r := &RedisLock{}
	r.New(HighAvailable_key, 0)

	ticker := time.NewTicker(time.Minute)

	if r.SetLock() {
		//  run as master
		ShouldRunAsMaster = true
		go func() {

			// 定期更新key, 否则会自动过期，slave会watch这个key
			// slave如果发现这个key不存在了，则会获取成为master,并继续更新这个key
			for range ticker.C {
				if !r.Updatelock(time.Now().Unix()) {
					panic("更新redis key失败, 请检查")
				}
			}

		}()
	} else {
		ShouldRunAsMaster = false
		// watch
		go func() {
			for range ticker.C {
				if r.SetLock() {
					ShouldRunAsMaster = true
				}
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

func (r *RedisLock) Updatelock(v int64) bool {
	_, err := r.Client.SetXX(context.Background(), r.Key, v, time.Second*75).Result()
	return err == nil

}

func (r *RedisLock) SetLock() bool {

	_, err := r.Client.SetNX(context.Background(), r.Key, r.Value, time.Second*75).Result()
	return err != redis.Nil
}