package handle_func

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

type ClbMetrics struct {
	ID                string
	Port              string
	Protocol          string // TCP/tcp/HTTP/http
	MetricsName       string
	MontiorNs         string
	PrometheusMetrics prometheus.Gauge
	Client            *monitor.Client
	MonitorGetRequest *monitor.GetMonitorDataRequest
}

func NewClbMetrics(id, port, protocol, cloudMonitorNs, cloudMonitorMetricsName string) *ClbMetrics {

	return &ClbMetrics{
		ID:          id,
		Port:        port,
		Protocol:    protocol,
		MetricsName: cloudMonitorMetricsName,
		MontiorNs:   cloudMonitorNs,
	}

}

func (cm *ClbMetrics) SetMonitorClientAndRequest(id, key string) error {
	credential := common.NewCredential(
		id,  // secretId
		key, // secredKey
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"
	client, err := monitor.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return err
	}
	cm.Client = client

	// set request
	request := monitor.NewGetMonitorDataRequest()

	request.Namespace = common.StringPtr(cm.MontiorNs)
	request.MetricName = common.StringPtr(cm.MetricsName)

	//
	request.Instances = []*monitor.Instance{
		{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("loadBalancerId"),
					Value: common.StringPtr(cm.ID),
				},
				{
					Name:  common.StringPtr("loadBalancerPort"),
					Value: common.StringPtr(cm.Port),
				},
				{
					Name:  common.StringPtr("protocol"),
					Value: common.StringPtr(cm.Protocol),
				},
			},
		},
	}

	// time start and period, this two value will lead to the values count to get
	request.Period = common.Uint64Ptr(60)
	request.StartTime = common.StringPtr(time.Now().Add(-60 * time.Second).Format(time.RFC3339))

	// request.SpecifyStatistics = common.Int64Ptr(7)

	cm.MonitorGetRequest = request

	return nil
}

func (cm *ClbMetrics) GetMetricsVaule() (float64, error) {

	res, err := cm.Client.GetMonitorData(cm.MonitorGetRequest)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return -1, err
	}
	if err != nil {
		return -1, err
	}

	// fmt.Printf("%s\n", response.ToJsonString())
	if len(res.Response.DataPoints) > 0 && len(res.Response.DataPoints[0].Values) > 0 {
		return *res.Response.DataPoints[0].Values[0], nil

	}
	return -1, err
}

func (cm *ClbMetrics) SetMetricsValue() {

	value, err := cm.GetMetricsVaule()
	if err != nil {
		slog.Error("get metrics value", "msg", err, "value", -1)
		cm.PrometheusMetrics.Set(-1)
		return
	}
	cm.PrometheusMetrics.Set(value)
}

func (cm *ClbMetrics) WatchMetricsValue() {
	cm.SetMetricsValue()

	tk := time.NewTicker(time.Second * time.Duration(60))
	defer tk.Stop()

	for range tk.C {
		cm.SetMetricsValue()
	}
}

func AdaptHttpHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
