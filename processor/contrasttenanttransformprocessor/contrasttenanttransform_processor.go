// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package contrasttenanttransformprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/contrasttenanttransformprocessor"

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TenetHeaderKey all communication to the collector comes from the api-gateway, which handles authentication and authorization.
// This value would be better fetched from a JWT, but the contrast-api-gateway isn't giving those out yet.
const TenetHeaderKey = "X-ORGANIZATION-ID"
const TenetAttrKey = "organization"

type tenantProcessor struct {
	logger *zap.Logger
}

func (rp *tenantProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	ci := client.FromContext(ctx)
	orgVal := strings.Join(ci.Metadata.Get(TenetHeaderKey), ";")
	if len(orgVal) == 0 {
		rp.logger.Error("tenant header not present on data",
			zap.String("metadata.key", TenetHeaderKey))
		return ptrace.NewTraces(), fmt.Errorf("tenant header was not present on request")
	}

	rts := td.ResourceSpans()
	for i := 0; i < rts.Len(); i++ {
		if val, present := rts.At(i).Resource().Attributes().Get(TenetAttrKey); present {
			rp.logger.Warn("invalid resource in span already includes the tenant information, dropping",
				zap.String(TenetAttrKey, orgVal))
			return ptrace.NewTraces(), fmt.Errorf("invalid tenant attribute '%s=%v' present in data", TenetAttrKey, val)
		}
		rts.At(i).Resource().Attributes().PutStr(TenetAttrKey, orgVal)
	}

	return td, nil
}

func (rp *tenantProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	ci := client.FromContext(ctx)
	orgVal := strings.Join(ci.Metadata.Get(TenetHeaderKey), ";")
	if len(orgVal) == 0 {
		rp.logger.Error(
			"Tenant metadata not present on request",
			zap.String("metadata.key", TenetHeaderKey))
		return pmetric.NewMetrics(), fmt.Errorf("tenant header was not present on request")
	}

	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		if val, present := rms.At(i).Resource().Attributes().Get(TenetAttrKey); present {
			rp.logger.Warn("invalid resource in metric already includes the tenant information, dropping",
				zap.String(TenetAttrKey, orgVal))
			return pmetric.NewMetrics(), fmt.Errorf("invalid tenant attribute '%s=%v' present in data", TenetAttrKey, val)
		}
		rms.At(i).Resource().Attributes().PutStr(TenetAttrKey, orgVal)
	}
	return md, nil
}

func (rp *tenantProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	ci := client.FromContext(ctx)
	orgVal := strings.Join(ci.Metadata.Get(TenetHeaderKey), ";")
	if len(orgVal) == 0 {
		rp.logger.Error(
			"Tenant metadata not present on request",
			zap.String("metadata.key", TenetHeaderKey))
		return plog.NewLogs(), fmt.Errorf("tenant header was not present on request")
	}
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		// If the metric already has "resource.attribute[organization]" value.  Something is wrong. Data producers don't
		// know this value to supply it.
		if val, present := rls.At(i).Resource().Attributes().Get(TenetAttrKey); present {
			rp.logger.Warn("invalid resource in log already includes the tenant information, dropping",
				zap.String(TenetAttrKey, orgVal))
			return plog.NewLogs(), fmt.Errorf("invalid tenant attribute '%s=%v' present in data", TenetAttrKey, val)
		}
		rls.At(i).Resource().Attributes().PutStr(TenetAttrKey, orgVal)
	}

	return ld, nil
}
