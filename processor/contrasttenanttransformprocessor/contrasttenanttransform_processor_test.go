// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package contrasttenanttransformprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"
)

func TestTenetHdrProcessor(t *testing.T) {
	tests := []struct {
		name             string
		metadata         client.Metadata
		isError          bool
		sourceAttributes map[string]string
		wantAttributes   map[string]string
	}{
		{
			name:             "no_header__no_org__dropped",
			metadata:         client.NewMetadata(map[string][]string{}),
			isError:          true,
			sourceAttributes: nil,
			wantAttributes:   map[string]string{},
		},
		{
			name:             "no_header__org_attr__dropped",
			metadata:         client.NewMetadata(map[string][]string{}),
			isError:          true,
			sourceAttributes: map[string]string{TenetAttrKey: "someone"},
			wantAttributes:   map[string]string{},
		},
		{
			name:     "org_header__org_attr__dropped",
			metadata: client.NewMetadata(map[string][]string{TenetHeaderKey: {"someone"}}),
			isError:  true,
			sourceAttributes: map[string]string{
				TenetAttrKey: "someone",
			},
			wantAttributes: map[string]string{},
		},
		{
			name:             "org_header_empty__no_org_attr__dropped",
			metadata:         client.NewMetadata(map[string][]string{TenetHeaderKey: {""}}),
			isError:          true,
			sourceAttributes: map[string]string{},
			wantAttributes:   map[string]string{},
		},
		{
			name:             "org_header__no_org_attr__org_attr_injected",
			metadata:         client.NewMetadata(map[string][]string{TenetHeaderKey: {"someone"}}),
			isError:          false,
			sourceAttributes: map[string]string{},
			wantAttributes: map[string]string{
				TenetAttrKey: "someone",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory()
			ctx := client.NewContext(context.Background(), client.Info{Metadata: tt.metadata})

			// Test trace consumer
			ttn := new(consumertest.TracesSink)
			rtp, err := factory.CreateTracesProcessor(context.Background(), processortest.NewNopCreateSettings(), &Config{}, ttn)
			require.NoError(t, err)
			// advertise to system that this processor mutates data.  It affects the ownership handoff of the data.
			assert.True(t, rtp.Capabilities().MutatesData)

			sourceTraceData := generateTraceData(tt.sourceAttributes)
			wantTraceData := generateTraceData(tt.wantAttributes)
			err = rtp.ConsumeTraces(ctx, sourceTraceData)
			if tt.isError {
				require.Error(t, err)
				traces := ttn.AllTraces()
				require.Len(t, traces, 0)
			} else {
				require.NoError(t, err)
				traces := ttn.AllTraces()
				require.Len(t, traces, 1)
				assert.NoError(t, ptracetest.CompareTraces(wantTraceData, traces[0]))
			}

			// Test metrics consumer
			tmn := new(consumertest.MetricsSink)
			rmp, err := factory.CreateMetricsProcessor(context.Background(), processortest.NewNopCreateSettings(), &Config{}, tmn)
			require.NoError(t, err)
			assert.True(t, rmp.Capabilities().MutatesData)

			sourceMetricData := generateMetricData(tt.sourceAttributes)
			wantMetricData := generateMetricData(tt.wantAttributes)
			err = rmp.ConsumeMetrics(ctx, sourceMetricData)
			if tt.isError {
				require.Error(t, err)
				metrics := tmn.AllMetrics()
				require.Len(t, metrics, 0)
			} else {
				require.NoError(t, err)
				metrics := tmn.AllMetrics()
				require.Len(t, metrics, 1)
				assert.NoError(t, pmetrictest.CompareMetrics(wantMetricData, metrics[0]))
			}

			// Test logs consumer
			tln := new(consumertest.LogsSink)
			rlp, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopCreateSettings(), &Config{}, tln)
			require.NoError(t, err)
			assert.True(t, rlp.Capabilities().MutatesData)

			sourceLogData := generateLogData(tt.sourceAttributes)
			wantLogData := generateLogData(tt.wantAttributes)
			err = rlp.ConsumeLogs(ctx, sourceLogData)
			if tt.isError {
				require.Error(t, err)
				logs := tln.AllLogs()
				require.Len(t, logs, 0)
			} else {
				require.NoError(t, err)
				logs := tln.AllLogs()
				require.Len(t, logs, 1)
				assert.NoError(t, plogtest.CompareLogs(wantLogData, logs[0]))
			}
		})
	}
}

func generateTraceData(attributes map[string]string) ptrace.Traces {
	td := testdata.GenerateTracesOneSpanNoResource()
	if attributes == nil {
		return td
	}
	resource := td.ResourceSpans().At(0).Resource()
	for k, v := range attributes {
		resource.Attributes().PutStr(k, v)
	}
	return td
}

func generateMetricData(attributes map[string]string) pmetric.Metrics {
	md := testdata.GenerateMetricsOneMetricNoResource()
	if attributes == nil {
		return md
	}
	resource := md.ResourceMetrics().At(0).Resource()
	for k, v := range attributes {
		resource.Attributes().PutStr(k, v)
	}
	return md
}

func generateLogData(attributes map[string]string) plog.Logs {
	ld := testdata.GenerateLogsOneLogRecordNoResource()
	if attributes == nil {
		return ld
	}
	resource := ld.ResourceLogs().At(0).Resource()
	for k, v := range attributes {
		resource.Attributes().PutStr(k, v)
	}
	return ld
}
