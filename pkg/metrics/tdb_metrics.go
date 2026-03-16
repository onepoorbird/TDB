// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

const (
	// TDB metric names
	TDBMetricAgentTotal        = "agent_total"
	TDBMetricSessionTotal      = "session_total"
	TDBMetricMemoryTotal       = "memory_total"
	TDBMetricEventTotal        = "event_total"
	TDBMetricRequestLatency    = "request_latency"
	TDBMetricRequestTotal      = "request_total"
	TDBMetricActiveConnections = "active_connections"

	// TDB metric labels
	TDBLabelMethod    = "method"
	TDBLabelStatus    = "status"
	TDBLabelAgentType = "agent_type"
	TDBLabelEventType = "event_type"

	// TDB metric values
	TDBStatusSuccess = "success"
	TDBStatusError   = "error"
)

var (
	// TDBAgentTotal tracks the total number of agents
	TDBAgentTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricAgentTotal,
			Help:      "Total number of agents managed by TDB",
		}, []string{TDBLabelAgentType},
	)

	// TDBSessionTotal tracks the total number of sessions
	TDBSessionTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricSessionTotal,
			Help:      "Total number of sessions managed by TDB",
		}, []string{},
	)

	// TDBMemoryTotal tracks the total number of memories
	TDBMemoryTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricMemoryTotal,
			Help:      "Total number of memories stored in TDB",
		}, []string{},
	)

	// TDBEventTotal tracks the total number of events
	TDBEventTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricEventTotal,
			Help:      "Total number of events processed by TDB",
		}, []string{TDBLabelEventType},
	)

	// TDBRequestLatency tracks the latency of gRPC requests
	TDBRequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricRequestLatency,
			Help:      "Latency of TDB gRPC requests in milliseconds",
			Buckets:   buckets,
		}, []string{TDBLabelMethod},
	)

	// TDBRequestTotal tracks the total number of requests
	TDBRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricRequestTotal,
			Help:      "Total number of TDB gRPC requests",
		}, []string{TDBLabelMethod, TDBLabelStatus},
	)

	// TDBActiveConnections tracks the number of active connections
	TDBActiveConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.TDBRole,
			Name:      TDBMetricActiveConnections,
			Help:      "Number of active gRPC connections to TDB",
		}, []string{},
	)
)

// RegisterTDB registers all TDB metrics with the provided registry
func RegisterTDB(registry *prometheus.Registry) {
	registry.MustRegister(TDBAgentTotal)
	registry.MustRegister(TDBSessionTotal)
	registry.MustRegister(TDBMemoryTotal)
	registry.MustRegister(TDBEventTotal)
	registry.MustRegister(TDBRequestLatency)
	registry.MustRegister(TDBRequestTotal)
	registry.MustRegister(TDBActiveConnections)
}
