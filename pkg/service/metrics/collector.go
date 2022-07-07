package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics label.
	MetricsLabelTenant = "tenant_id"

	// metrics name.
	MetricsNameDeviceNumTotal      = "device_num_total"
	MetricsNameDeviceTemplateTotal = "device_template_total"
	MetricsNameDeviceOnlineTotal   = "device_online_total"
)

var CollectorDeviceNumRequest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameDeviceNumTotal,
		Help: "device num total",
	},
	[]string{MetricsLabelTenant},
)
var CollectorDeviceTemplateRequest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameDeviceTemplateTotal,
		Help: "device template total",
	},
	[]string{MetricsLabelTenant},
)
var CollectorDeviceOnlineRequest = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsNameDeviceOnlineTotal,
		Help: "device online total",
	},
	[]string{MetricsLabelTenant},
)
