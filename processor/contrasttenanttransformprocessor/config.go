// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package contrasttenanttransformprocessor // Package contrasttenanttransformprocessor import "github.com/Contrast-Security-Inc/opentelemetry-collector-contrib/processor/contrasttenanttransformprocessor"

import (
	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for Contrast Tenant Transform processor.
type Config struct {
	// not configurable at this time
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
