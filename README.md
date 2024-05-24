## gin_web, 使用 gin 实现一些接口调用样例

### 中间件串联

- a -> b -> c, 则执行完 a -> b -> c 的 c.Next()后，再执行 c -> b -> a

### 请求限流实现

- 使用 time duration 实现，只做时间间隔的限流，如果要使用桶，请使用官方限制中间件
- middle.Limiter(1\*time.Second),这里设置每个请求的时间间隔，小于此间隔，则会被禁止

### prometheus 指标实现 k8s pod 伸缩

- 请求 tencent monitor 的接口，获取数据并生成 prometheus 指标，可以基于此指标进行一些活动，比如`keda`
- prometheus metrics 使用了 time.Ticker 在后台定期更新，与用户访问就触发访问 tencent api 解耦

## pgsql 使用

- 使用 init()函数中的 autoMigrate()方法，创建表，并初始化一个\*gorm.DB 连接，可全局使用
- 使用 struct tag 创建及限制表字段，并继承了 gorm.Model，可以设置的标签值可在官网查询，可以设置外键，主键，非空等
- 为结构体创建方法，使用指针接收者方法，传递了\*gorm.DB，方便处理 db 请求，通过返回 gin.handlerFunc,以注册到 gin 接口方法

```bash
mkdir -p /data/pg
docker run -d --name pg \
  --restart=always \
  -e POSTGRES_PASSWORD="xxx" \
  -p 5432:5432 \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -v /data/pg:/var/lib/postgresql/data \
  postgres:16.2-bookworm
```

## ci_cd 简易实现

- 只考虑了在 linux 下运行，代码调试系统为 ubuntu24.04
- 解耦开发与 cicd 的强关联，可使用 cicd 仓库单独由运维管理
- 流程，clone --> build --> push --> deploy
- use exec.Command()来封装了`git clone`, `mkdir -p`,`docker build`,
- and `docker tag`, `docker push`, `kubectl apply -f`
- 附:devops_cicd 仓库目录结构

```bash
$ tree
.
├── config
│   ├── build_config.yaml
│   └── deploy_config.yaml
├── jobs
│   └── backend
│       └── go-micro
│           ├── build
│           │   └── Dockerfile
│           └── deploy
│               └── dev
│                   └── deployment.yaml
└── README.md

8 directories, 5 files
```

## etcd 选主

```go
	// master election
	etcd := db.NewEtcd()
	etcd.Campaign(main_func)

```

## etcd 的分布式锁

```go
func LockTask01(e *Etcd) {

	s := e.NewSession(4)
	if s == nil {
		return
	}
	mu := concurrency.NewMutex(s, lock_prefix)
	defer mu.Unlock(e.Context)

	err := mu.TryLock(e.Context)
	if err != nil {
		slog.Error("lock task 01", "msg", err)
		return
	}
	defer mu.Unlock(e.Context)

	slog.Info("task 01 run")
	time.Sleep(time.Second * 2)
}
```

## 代码结构

```yaml
- cicd/
- db/
- handle/
- middle/
- router/
./go.mod
./main.go
```
