package netstat

/*
Copyright 2016 Staples, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"syscall"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/shirou/gopsutil/net"
)

const (
	vendor        = "staples"
	fs            = "netstat"
	pluginName    = "netstat"
	pluginVersion = 1
	pluginType    = plugin.CollectorPluginType
)

// NetstatCollector type
type NetstatCollector struct {
}

// New returns a new netstat plugin object
func New() *NetstatCollector {
	netstat := &NetstatCollector{}
	return netstat
}

//Meta returns meta data for plugin
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		pluginName,
		pluginVersion,
		pluginType,
		[]string{},
		[]string{plugin.SnapGOBContentType},
		plugin.ConcurrencyCount(1))
}

// GetConfigPolicy returns plugin configuration
func (netstat *NetstatCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}

// GetMetricTypes returns MetricType slice collected by plugin
func (netstat *NetstatCollector) GetMetricTypes(cfg plugin.ConfigType) (metrics []plugin.MetricType, err error) {
	fields, err := getStats()
	if err != nil {
		return nil, fmt.Errorf("Error collecting metrics: %v", err)
	}

	for name := range fields {
		ns := core.NewNamespace(vendor, "procfs", fs, name)
		metric := plugin.MetricType{
			Namespace_: ns,
			Data_:      nil,
			Timestamp_: time.Now(),
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// CollectMetrics gathers netstat metrics
func (netstat *NetstatCollector) CollectMetrics(metricTypes []plugin.MetricType) (metrics []plugin.MetricType, err error) {
	fields, err := getStats()
	if err != nil {
		return nil, fmt.Errorf("Error collecting metrics: %v", err)
	}

	for _, metricType := range metricTypes {
		ns := metricType.Namespace()

		val, err := getMapValueByNamespace(fields, ns[3:].Strings())
		if err != nil {
			return nil, fmt.Errorf("Error collecting metrics: %v", err)
		}

		metric := plugin.MetricType{
			Namespace_: ns,
			Data_:      val,
			Timestamp_: time.Now(),
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

//getMapValueByNamespace gets value from map by namespace given in array of strings
func getMapValueByNamespace(m map[string]interface{}, ns []string) (val interface{}, err error) {
	if len(ns) == 0 {
		return val, fmt.Errorf("Namespace length equal to zero")
	}

	current := ns[0]
	var ok bool
	if len(ns) == 1 {
		val, ok = m[current]
		if ok {
			return val, err
		}
		return val, fmt.Errorf("Key does not exist in map {key %s}", current)
	}

	if v, ok := m[current].(map[string]interface{}); ok {
		val, err = getMapValueByNamespace(v, ns[1:])
		return val, err
	}
	return val, err
}

func getStats() (map[string]interface{}, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	counts["UDP"] = 0
	for _, conn := range conns {
		if conn.Type == syscall.SOCK_DGRAM {
			counts["UDP"]++
			continue
		}
		c, ok := counts[conn.Status]
		if !ok {
			counts[conn.Status] = 0
		}
		counts[conn.Status] = c + 1
	}

	fields := map[string]interface{}{
		"tcp_established": counts["ESTABLISHED"],
		"tcp_syn_sent":    counts["SYN_SENT"],
		"tcp_syn_recv":    counts["SYN_RECV"],
		"tcp_fin_wait1":   counts["FIN_WAIT1"],
		"tcp_fin_wait2":   counts["FIN_WAIT2"],
		"tcp_time_wait":   counts["TIME_WAIT"],
		"tcp_close":       counts["CLOSE"],
		"tcp_close_wait":  counts["CLOSE_WAIT"],
		"tcp_last_ack":    counts["LAST_ACK"],
		"tcp_listen":      counts["LISTEN"],
		"tcp_closing":     counts["CLOSING"],
		"tcp_none":        counts["NONE"],
		"udp_socket":      counts["UDP"],
	}

	return fields, nil
}
