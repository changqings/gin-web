## gin_web

### 多个中间件串联时

- a -> b -> c, 则执行完 a -> b -> c 的 c.Next()后，再执行 c -> b -> a

### 请求限流实现

- 使用 time duration 实现，只做时间间隔的限流，如果要使用桶，请自行百度
- middle.Limiter(1\*time.Second),这里设置每个请求的时间间隔，小于此间隔，则会被禁止

// prometheus metrics, value from tenceont monitor

- 请求 tencent monitor 的接口，获取数据并生成 prometheus 指标，可以基于此指标进行一些活动，比如`keda`
- prometheus metrics 使用了 time.Ticker 在后台定期更新，与用户访问就触发访问 tencent api 解耦
