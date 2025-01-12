package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var metricLinesRead = promauto.NewCounter(prometheus.CounterOpts{
	Name: "jsumo_lines_read_total",
	Help: "The total number of lines read from journalctl",
})

var metricBytesSentToReceiver = promauto.NewCounter(prometheus.CounterOpts{
	Name: "jsumo_bytes_sent_total",
	Help: "The total number of bytes sent to the receiver",
})

var metricStatusCodesFromReceiver = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "jsumo_receiver_status_codes_total",
	Help: "The total number of requests to the receiver",
}, []string{"status_code"})

var metricErrorsWhenSendingToReceiver = promauto.NewCounter(prometheus.CounterOpts{
	Name: "jsumo_errors_sending_to_receiver_total",
	Help: "The total number of errors when sending logs to the receiver",
})
