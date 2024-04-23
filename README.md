# gin_tmp

// 多个中间件串联时

- a -> b -> c, 则执行完 a -> b -> c 的 c.Next()后，再执行 c -> b -> a

// 请求限流实现

-使用一个 sync.Map()来实现对一个 map，map[time_stamp]bool

- when received one request, set time_stamp=false, when the request done, set time_stamp=true
- set time_stamp 的精度来做为时间间隔，比如精确到秒，则每秒只能有一个请求，
