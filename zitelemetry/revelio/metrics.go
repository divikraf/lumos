package revelio

import (
	"errors"
	"fmt"
	"regexp"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const packageName = "revelio"

func errStrFormatter(str string) string {
	return fmt.Sprintf("%s: %s", packageName, str)
}

type scopeHolder struct {
	scope Scope
}

// globalDefaultScope, no-op meter will be set on app start.
var globalDefaultScope = initDefaultScope()

func initDefaultScope() *atomic.Value {
	meter := otel.GetMeterProvider().Meter("")
	v := &atomic.Value{}
	v.Store(scopeHolder{
		scope: NewFromMeter(meter),
	})
	return v
}

// GetDefault returns the global default Scope.
func GetDefault() Scope {
	return globalDefaultScope.Load().(scopeHolder).scope
}

// SetDefault replaces the global default Scope.
func SetDefault(s Scope) {
	if s == nil {
		panic(packageName + ": SetDefault: cannot assign nil Meter for global default meter")
	}
	globalDefaultScope.Store(scopeHolder{scope: s})
}

// NewFromMeter wraps OpenTelemetry's [go.opentelemetry.io/otel/metric.Meter]
// into our own Scope.
func NewFromMeter(meter metric.Meter) Scope {
	return &scope{
		meter: meter,
	}
}

const scopeNameRegexStr = `^([a-z]{1}[a-z0-9-]{1,}[a-z0-9]{1})$`

var scopeNameRegex = regexp.MustCompile(scopeNameRegexStr)

func validateScopeName(scopeName string) error {
	if scopeName == "" {
		return errors.New(errStrFormatter("scopeName must not be empty"))
	}

	if !scopeNameRegex.MatchString(scopeName) {
		return errors.New(errStrFormatter("scopeName must conform to the regex " + scopeNameRegexStr))
	}

	return nil
}

// New returns a new Scope with the provided name and configuration.
//
// The name needs to be unique so it does not collide with other names used by
// an application, nor other applications.
//
// Returns error if name is empty or doesn't conform to the naming spec.
func New(name string, opts ...metric.MeterOption) (Scope, error) {
	if err := validateScopeName(name); err != nil {
		return nil, errors.New(errStrFormatter("New: name must not be empty"))
	}
	met := otel.GetMeterProvider().Meter(name, opts...)
	return NewFromMeter(met), nil
}

// MustNew is a syntactic sugar for [New].
// This function will trigger panic when err is occurred.
func MustNew(name string, opts ...metric.MeterOption) Scope {
	scope, err := New(name, opts...)
	if err != nil {
		panic(err)
	}
	return scope
}

// Duration is an instrument to record duration thingy, such as process latencies.
// It's basically a Float64Histogram instrument identified by
// name, unit of `ms` and configured with additional options.
func Duration(name string, description string, options ...DurationOption) (DurationRecorder, error) {
	return GetDefault().Duration(name, description, options...)
}

// MustDuration is a syntactic sugar for [Duration].
// This function will trigger panic when err is occurred.
func MustDuration(name string, description string, options ...DurationOption) DurationRecorder {
	instr, err := Duration(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64Counter returns a new Int64Counter instrument identified by name
// and configured with options. The instrument is used to synchronously
// record increasing int64 measurements during a computational operation.
func Int64Counter(name string, description string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return GetDefault().Int64Counter(name, description, options...)
}

// MustInt64Counter is a syntactic sugar for [Int64Counter].
// This function will trigger panic when err is occurred.
func MustInt64Counter(name string, description string, options ...metric.Int64CounterOption) metric.Int64Counter {
	instr, err := Int64Counter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64UpDownCounter returns a new Int64UpDownCounter instrument
// identified by name and configured with options. The instrument is used
// to synchronously record int64 measurements during a computational
// operation.
func Int64UpDownCounter(name string, description string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return GetDefault().Int64UpDownCounter(name, description, options...)
}

// MustInt64UpDownCounter is a syntactic sugar for [Int64UpDownCounter].
// This function will trigger panic when err is occurred.
func MustInt64UpDownCounter(name string, description string, options ...metric.Int64UpDownCounterOption) metric.Int64UpDownCounter {
	instr, err := Int64UpDownCounter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64Histogram returns a new Int64Histogram instrument identified by
// name and configured with options. The instrument is used to
// synchronously record the distribution of int64 measurements during a
// computational operation.
func Int64Histogram(name string, description string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return GetDefault().Int64Histogram(name, description, options...)
}

// MustInt64Histogram is a syntactic sugar for [Int64Histogram].
// This function will trigger panic when err is occurred.
func MustInt64Histogram(name string, description string, options ...metric.Int64HistogramOption) metric.Int64Histogram {
	instr, err := GetDefault().Int64Histogram(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64Gauge returns a new Int64Gauge instrument identified by name and
// configured with options. The instrument is used to synchronously record
// instantaneous int64 measurements during a computational operation.
func Int64Gauge(name string, description string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	return GetDefault().Int64Gauge(name, description, options...)
}

// MustInt64Gauge is a syntactic sugar for [Int64Gauge].
// This function will trigger panic when err is occurred.
func MustInt64Gauge(name string, description string, options ...metric.Int64GaugeOption) metric.Int64Gauge {
	instr, err := Int64Gauge(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64ObservableCounter returns a new Int64ObservableCounter identified
// by name and configured with options. The instrument is used to
// asynchronously record increasing int64 measurements once per a
// measurement collection cycle.
//
// Measurements for the returned instrument are made via a callback. Use
// the WithInt64Callback option to register the callback here, or use the
// RegisterCallback method of this Meter to register one later. See the
// Measurements section of the package documentation for more information.
func Int64ObservableCounter(name string, description string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	return GetDefault().Int64ObservableCounter(name, description, options...)
}

// MustInt64ObservableCounter is a syntactic sugar for [Int64ObservableCounter].
// This function will trigger panic when err is occurred.
func MustInt64ObservableCounter(name string, description string, options ...metric.Int64ObservableCounterOption) metric.Int64ObservableCounter {
	instr, err := Int64ObservableCounter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64ObservableUpDownCounter returns a new Int64ObservableUpDownCounter
// instrument identified by name and configured with options. The
// instrument is used to asynchronously record int64 measurements once per
// a measurement collection cycle.
//
// Measurements for the returned instrument are made via a callback. Use
// the WithInt64Callback option to register the callback here, or use the
// RegisterCallback method of this Meter to register one later. See the
// Measurements section of the package documentation for more information.
func Int64ObservableUpDownCounter(name string, description string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	return GetDefault().Int64ObservableUpDownCounter(name, description, options...)
}

// MustInt64ObservableUpDownCounter is a syntactic sugar for [Int64ObservableUpDownCounter].
// This function will trigger panic when err is occurred.
func MustInt64ObservableUpDownCounter(name string, description string, options ...metric.Int64ObservableUpDownCounterOption) metric.Int64ObservableUpDownCounter {
	instr, err := Int64ObservableUpDownCounter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Int64ObservableGauge returns a new Int64ObservableGauge instrument
// identified by name and configured with options. The instrument is used
// to asynchronously record instantaneous int64 measurements once per a
// measurement collection cycle.
//
// Measurements for the returned instrument are made via a callback. Use
// the WithInt64Callback option to register the callback here, or use the
// RegisterCallback method of this Meter to register one later. See the
// Measurements section of the package documentation for more information.
func Int64ObservableGauge(name string, description string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	return GetDefault().Int64ObservableGauge(name, description, options...)
}

// MustInt64ObservableGauge is a syntactic sugar for [Int64ObservableGauge].
// This function will trigger panic when err is occurred.
func MustInt64ObservableGauge(name string, description string, options ...metric.Int64ObservableGaugeOption) metric.Int64ObservableGauge {
	instr, err := Int64ObservableGauge(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64Counter returns a new Float64Counter instrument identified by
// name and configured with options. The instrument is used to
// synchronously record increasing float64 measurements during a
// computational operation.
func Float64Counter(name string, description string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return GetDefault().Float64Counter(name, description, options...)
}

// MustFloat64Counter is a syntactic sugar for [Float64Counter].
// This function will trigger panic when err is occurred.
func MustFloat64Counter(name string, description string, options ...metric.Float64CounterOption) metric.Float64Counter {
	instr, err := Float64Counter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64UpDownCounter returns a new Float64UpDownCounter instrument
// identified by name and configured with options. The instrument is used
// to synchronously record float64 measurements during a computational
// operation.
func Float64UpDownCounter(name string, description string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	return GetDefault().Float64UpDownCounter(name, description, options...)
}

// MustFloat64UpDownCounter is a syntactic sugar for [Float64UpDownCounter].
// This function will trigger panic when err is occurred.
func MustFloat64UpDownCounter(name string, description string, options ...metric.Float64UpDownCounterOption) metric.Float64UpDownCounter {
	instr, err := Float64UpDownCounter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64Histogram returns a new Float64Histogram instrument identified by
// name and configured with options. The instrument is used to
// synchronously record the distribution of float64 measurements during a
// computational operation.
func Float64Histogram(name string, description string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return GetDefault().Float64Histogram(name, description, options...)
}

// MustFloat64Histogram is a syntactic sugar for [Float64Histogram].
// This function will trigger panic when err is occurred.
func MustFloat64Histogram(name string, description string, options ...metric.Float64HistogramOption) metric.Float64Histogram {
	instr, err := Float64Histogram(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64Gauge returns a new Float64Gauge instrument identified by name and
// configured with options. The instrument is used to synchronously record
// instantaneous float64 measurements during a computational operation.
func Float64Gauge(name string, description string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	return GetDefault().Float64Gauge(name, description, options...)
}

// MustFloat64Gauge is a syntactic sugar for [Float64Gauge].
// This function will trigger panic when err is occurred.
func MustFloat64Gauge(name string, description string, options ...metric.Float64GaugeOption) metric.Float64Gauge {
	instr, err := Float64Gauge(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64ObservableCounter returns a new Float64ObservableCounter
// instrument identified by name and configured with options. The
// instrument is used to asynchronously record increasing float64
// measurements once per a measurement collection cycle.
//
// Measurements for the returned instrument are made via a callback. Use
// the WithFloat64Callback option to register the callback here, or use the
// RegisterCallback method of this Meter to register one later. See the
// Measurements section of the package documentation for more information.
func Float64ObservableCounter(name string, description string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	return GetDefault().Float64ObservableCounter(name, description, options...)
}

// MustFloat64ObservableCounter is a syntactic sugar for [Float64ObservableCounter].
// This function will trigger panic when err is occurred.
func MustFloat64ObservableCounter(name string, description string, options ...metric.Float64ObservableCounterOption) metric.Float64ObservableCounter {
	instr, err := Float64ObservableCounter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64ObservableUpDownCounter returns a new
// Float64ObservableUpDownCounter instrument identified by name and
// configured with options. The instrument is used to asynchronously record
// float64 measurements once per a measurement collection cycle.
//
// Measurements for the returned instrument are made via a callback. Use
// the WithFloat64Callback option to register the callback here, or use the
// RegisterCallback method of this Meter to register one later. See the
// Measurements section of the package documentation for more information.
func Float64ObservableUpDownCounter(name string, description string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	return GetDefault().Float64ObservableUpDownCounter(name, description, options...)
}

// MustFloat64ObservableUpDownCounter is a syntactic sugar for [Float64ObservableUpDownCounter].
// This function will trigger panic when err is occurred.
func MustFloat64ObservableUpDownCounter(name string, description string, options ...metric.Float64ObservableUpDownCounterOption) metric.Float64ObservableUpDownCounter {
	instr, err := Float64ObservableUpDownCounter(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// Float64ObservableGauge returns a new Float64ObservableGauge instrument
// identified by name and configured with options. The instrument is used
// to asynchronously record instantaneous float64 measurements once per a
// measurement collection cycle.
//
// Measurements for the returned instrument are made via a callback. Use
// the WithFloat64Callback option to register the callback here, or use the
// RegisterCallback method of this Meter to register one later. See the
// Measurements section of the package documentation for more information.
func Float64ObservableGauge(name string, description string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	return GetDefault().Float64ObservableGauge(name, description, options...)
}

// MustFloat64ObservableGauge is a syntactic sugar for [Float64ObservableGauge].
// This function will trigger panic when err is occurred.
func MustFloat64ObservableGauge(name string, description string, options ...metric.Float64ObservableGaugeOption) metric.Float64ObservableGauge {
	instr, err := Float64ObservableGauge(name, description, options...)
	if err != nil {
		panic(err)
	}
	return instr
}

// RegisterCallback registers f to be called during the collection of a
// measurement cycle.
//
// If Unregister of the returned Registration is called, f needs to be
// unregistered and not called during collection.
//
// The instruments f is registered with are the only instruments that f may
// observe values for.
//
// If no instruments are passed, f should not be registered nor called
// during collection.
//
// The function f needs to be concurrent safe.
func RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	return GetDefault().RegisterCallback(f, instruments...)
}
