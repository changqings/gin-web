package db

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	etcd_host_port  = "http://139.199.218.214:2379"
	etcd_user       = "root"
	etcd_pass       = "lzwk_ops_0517"
	election_prefix = "gin-web-master"
	election_val    = "master_set"

	lock_prefix = "gin-web-lock"
)

type Etcd struct {
	Context context.Context
	Client  *clientv3.Client
}

// master election with auto lease
// ttl = 10s, and auto lease with 5s
func (e *Etcd) Campaign(f func()) {
	s, err := concurrency.NewSession(e.Client, concurrency.WithTTL(10))
	if err != nil {
		slog.Error("get concurrency session", "msg", err)
		return
	}
	defer s.Close()

	el := concurrency.NewElection(s, election_prefix)
	slog.Info("waitting to be master...")
	if err := el.Campaign(e.Context, election_val); err != nil {
		slog.Error("el campaign", "msg", err)
		return
	}

	slog.Info("campaign", "msg", "run as master")
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				_, err := e.Client.KeepAliveOnce(e.Context, s.Lease())
				if err != nil {
					slog.Error("keep alive once", "msg", err)
					return
				}
			case <-s.Done():
				slog.Info("Session done, exit keep-alive go func()")
				return
			}

		}
	}()

	f()

}

func getEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			etcd_host_port,
		},
		Username: etcd_user,
		Password: etcd_pass,
	})
	if err != nil {
		slog.Error("get etcd client error", "msg", err)
		return nil
	}

	return cli
}

func NewEtcd() *Etcd {
	return &Etcd{
		Context: context.Background(),
		Client:  getEtcdClient(),
	}
}

func (e *Etcd) NewSession(ttl int) *concurrency.Session {
	s, err := concurrency.NewSession(e.Client, concurrency.WithTTL(ttl))
	if err != nil {
		slog.Error("NewSession", "msg", err)
		return nil
	}
	return s
}

// for test etcd.Campaign(backgroundTask)
func BackgroundTask() error {
	stopCh, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	t := time.NewTicker(time.Second * 10)

	for {
		select {
		case <-stopCh.Done():
			slog.Info("get exit signal, good bye.")
			os.Exit(0)
		case <-t.C:
			slog.Info("background task running...")
		}
	}
}

// for lock
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

func LockTask02(e *Etcd) {
	s := e.NewSession(60)
	if s == nil {
		return
	}
	mu := concurrency.NewMutex(s, lock_prefix)
	defer mu.Unlock(e.Context)

	err := mu.TryLock(e.Context)
	if err != nil {
		slog.Error("lock task 02", "msg", err)
		return
	}
	slog.Info("task 02 run")
	time.Sleep(time.Second * 2)

}
