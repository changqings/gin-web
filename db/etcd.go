package db

import (
	"context"
	"log/slog"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	etcd_host_port  = "http://139.199.218.214:2379"
	etcd_user       = "root"
	etcd_pass       = "lzwk_ops_0517"
	election_prefix = "gin-web-master"
	election_val    = "got"
)

type Etcd struct {
	Context context.Context
	Client  *clientv3.Client
}

func (e *Etcd) Campaign(f func() error) error {
	s, err := concurrency.NewSession(e.Client, concurrency.WithTTL(10))
	if err != nil {
		slog.Error("get concurrency session", "msg", err)
		return err
	}
	defer s.Close()

	el := concurrency.NewElection(s, election_prefix)
	slog.Info("waitting to be master...")
	if err := el.Campaign(e.Context, election_val); err != nil {
		slog.Error("el campaign", "msg", err)
		return err
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

	return f()

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
