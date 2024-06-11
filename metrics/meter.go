package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	otelmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	ResourceKindAttrKey attribute.Key = "resource_kind"
	TypeURLAttrKey      attribute.Key = "type_url"
)

// GetGlobalMeter ... get the global meter from otel library
func GetGlobalMeter() metric.Meter {
	return otel.Meter("xds-mgnt-srv")
}

// SetPromGlobalMeter ... set Prometheus Exporter as a global meter in otel library
func SetPromGlobalMeter() error {
	promExporter, err := otelprom.New()
	if err != nil {
		return err
	}
	promProvider := otelmetric.NewMeterProvider(otelmetric.WithReader(promExporter))
	otel.SetMeterProvider(promProvider)
	return nil
}
