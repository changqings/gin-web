package router

import (
	"github.com/changqings/gin-web/handle"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	tencent_api_secret_id  = ""
	tencent_api_secert_key = ""
	clb_id                 = "lb-xxx"
	clb_port               = "443"
	clb_protocol           = "tcp"
	clb_metrics_name       = "ClientAccIntraffic"
	tencent_resources_ns   = "QCE/LB_PUBLIC" // or QCE/LB_PRIVATE
)

func TxMetrics(app *gin.Engine) error {
	metricsGroup := app.Group("/metrics")
	// metrics usage

	cm, err := handle.NewClbMetrics(
		tencent_api_secret_id,
		tencent_api_secert_key,
		clb_id, clb_port, clb_protocol,
		tencent_resources_ns,
		clb_metrics_name,
	)
	if err != nil {
		return err
	}

	// set prometehsu gauge
	cm.PrometheusMetrics = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "clb",
			Name:      "client_acc_intraffic",
			Help:      "A gauage of clb in duation of 60s.",
		})

	// registry prometheus metrics
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	prometheus.MustRegister(cm.PrometheusMetrics)

	// watching, every 60s update value
	go cm.WatchMetricsValue()
	metricsGroup.GET("/tx_clb", handle.AdaptHttpHandler(promhttp.Handler()))
	return nil

}
