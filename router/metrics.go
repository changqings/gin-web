package router

import (
	"log/slog"

	"github.com/changqings/gin-web/handle"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	tencent_api_id     = ""
	tencent_api_secert = ""
)

func TxMetrics(app *gin.Engine) {
	metricsGroup := app.Group("/metrics")
	// metrics usage
	cm := &handle.ClbMetrics{
		ID:          "lb-xxx",
		Port:        "443",
		Protocol:    "tcp",
		MetricsName: "ClientAccIntraffic",
	}

	///
	// set prometehsu gauge
	cm.PrometheusMetrics = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "clb",
			Name:      "client_acc_intraffic",
			Help:      "A gauage of clb in duation of 60s.",
		})

	// set tecent monitor client and request
	err := cm.SetMonitorClientAndRequest(tencent_api_id, tencent_api_secert)
	if err != nil {
		slog.Error("get monitor client", "msg", err)
		return
	}

	// registry prometheus metrics
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	prometheus.MustRegister(cm.PrometheusMetrics)

	// watching, every 60s update value
	go cm.WatchMetricsValue()
	metricsGroup.GET("/tx_clb", handle.AdaptHttpHandler(promhttp.Handler()))

}
