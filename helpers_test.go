package warning_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.wamod.dev/warning"
)

// ExampleTap demonstrates how to use the Tap function to apply a side-effect to each warning.
func ExampleTap() {
	// create a new collector
	collector := warning.NewCollector()
	defer collector.Close() // make sure to close the collector when done

	// attach the collector to a context
	ctx := warning.Attach(context.Background(), collector)

	// use Tap to apply a side-effect to each warning
	ctx = warning.Tap(ctx, func(wrr warning.Warning) {
		fmt.Println("side-effect:", wrr.Warn())
	})

	// use Warn or Warnf to write warning to the context
	warning.Warnf(ctx, "this is a warning 1")
	warning.Warnf(ctx, "this is a warning 2")

	// read all warning from the collector
	wrrs, err := warning.ReadAll(collector)
	if err != nil {
		panic(err)
	}

	for _, wrr := range wrrs {
		fmt.Println(wrr.Warn())
	}

	// Output:
	// side-effect: this is a warning 1
	// side-effect: this is a warning 2
	// this is a warning 1
	// this is a warning 2
}

type multiWarn struct {
	details []string
}

func (w *multiWarn) Warn() string {
	return strings.Join(w.details, ", ")
}

// ExampleReduce demonstrates how to use the Reduce function to reduce warnings into a single value.
func ExampleReduce() {
	// create a new collector
	collector := warning.NewCollector()
	defer collector.Close() // make sure to close the collector when done

	// attach the collector to a context
	ctx := warning.Attach(context.Background(), collector)

	// use Reduce to reduce warning into a single value
	ctx, flush := warning.Reduce(ctx, func(acc *multiWarn, wrr warning.Warning) *multiWarn {
		if acc == nil { // initialize the accumulator
			acc = new(multiWarn)
		}

		acc.details = append(acc.details, wrr.Warn())

		return acc
	})

	// use Warn or Warnf to write warning to the context
	warning.Warnf(ctx, "this is a warning 1")
	warning.Warnf(ctx, "this is a warning 2")
	warning.Warnf(ctx, "this is a warning 3")

	// flush the reduced warning
	flush()

	// read all warning from the collector
	wrrs, err := warning.ReadAll(collector)
	if err != nil {
		panic(err)
	}

	for i, wrr := range wrrs {
		fmt.Printf("[%d]: %s\n", i, wrr.Warn())
	}

	// Output:
	// [0]: this is a warning 1, this is a warning 2, this is a warning 3
}

// ExampleFilter demonstrates how to use the Filter function to filter warnings.
func ExampleFilter() {
	// create a new collector
	collector := warning.NewCollector()
	defer collector.Close() // make sure to close the collector when done

	// attach the collector to a context
	ctx := warning.Attach(context.Background(), collector)

	// use Filter to filter warning
	ctx = warning.Filter(ctx, func(wrr warning.Warning) bool {
		return !strings.HasPrefix(wrr.Warn(), "ignore")
	})

	// use Warn or Warnf to write warning to the context
	warning.Warnf(ctx, "this is a warning")
	warning.Warnf(ctx, "ignore this warning")
	warning.Warnf(ctx, "this is another warning")

	// read all warning from the collector
	wrrs, err := warning.ReadAll(collector)
	if err != nil {
		panic(err)
	}

	for i, wrr := range wrrs {
		fmt.Printf("[%d]: %s\n", i, wrr.Warn())
	}

	// Output:
	// [0]: this is a warning
	// [1]: this is another warning
}

// ExampleMap demonstrates how to use the Map function to map warnings.
func ExampleMap() {
	// create a new collector
	collector := warning.NewCollector()
	defer collector.Close() // make sure to close the collector when done

	// attach the collector to a context
	ctx := warning.Attach(context.Background(), collector)

	// use Map to map warning
	ctx = warning.Map(ctx, func(wrr warning.Warning) warning.Warning {
		return warning.New(strings.ToUpper(wrr.Warn()))
	})

	// use Warn or Warnf to write warning to the context
	warning.Warnf(ctx, "this is a warning")
	warning.Warnf(ctx, "this is another warning")

	// read all warning from the collector
	wrrs, err := warning.ReadAll(collector)
	if err != nil {
		panic(err)
	}

	for i, wrr := range wrrs {
		fmt.Printf("[%d]: %s\n", i, wrr.Warn())
	}

	// Output:
	// [0]: THIS IS A WARNING
	// [1]: THIS IS ANOTHER WARNING
}

func TestMap(t *testing.T) {
	writer := &mockWriter{}

	ctx := warning.Attach(context.Background(), writer)
	ctx = warning.Map(ctx, func(wrr warning.Warning) warning.Warning {
		return warning.New(strings.ToUpper(wrr.Warn()))
	})

	warning.Warn(ctx, warning.New("test"))

	if len(writer.buf) != 1 {
		t.Fatalf("expected 1 warning, got %v", len(writer.buf))
	}

	if writer.buf[0].Warn() != "TEST" {
		t.Errorf("expected TEST, got %v", writer.buf[0].Warn())
	}
}

func TestMapNoWriter(t *testing.T) {
	ctx := warning.Map(context.Background(), func(wrr warning.Warning) warning.Warning {
		return warning.New(strings.ToUpper(wrr.Warn()))
	})

	err := warning.Warn(ctx, warning.New("test"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestFilter(t *testing.T) {
	writer := &mockWriter{}

	ctx := warning.Attach(context.Background(), writer)
	ctx = warning.Filter(ctx, func(wrr warning.Warning) bool {
		return wrr.Warn() != "ignore"
	})

	warning.Warn(ctx, warning.New("this"))
	warning.Warn(ctx, warning.New("ignore"))
	warning.Warn(ctx, warning.New("that"))

	if len(writer.buf) != 2 {
		t.Fatalf("expected 2 warning, got %v", len(writer.buf))
	}

	if writer.buf[0].Warn() != "this" {
		t.Errorf("expected this, got %v", writer.buf[0].Warn())
	}

	if writer.buf[1].Warn() != "that" {
		t.Errorf("expected that, got %v", writer.buf[1].Warn())
	}
}

func TestFilterNoWriter(t *testing.T) {
	ctx := warning.Filter(context.Background(), func(wrr warning.Warning) bool {
		return wrr.Warn() != "ignore"
	})

	err := warning.Warn(ctx, warning.New("ignore"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestReduce(t *testing.T) {
	writer := &mockWriter{}

	ctx := warning.Attach(context.Background(), writer)
	ctx, flush := warning.Reduce(ctx, func(acc *multiWarn, wrr warning.Warning) *multiWarn {
		if acc == nil {
			acc = new(multiWarn)
		}

		acc.details = append(acc.details, wrr.Warn())

		return acc
	})

	warning.Warn(ctx, warning.New("this"))
	warning.Warn(ctx, warning.New("that"))

	flush()

	if len(writer.buf) != 1 {
		t.Fatalf("expected 1 warning, got %v", len(writer.buf))
	}

	if writer.buf[0].Warn() != "this, that" {
		t.Fatalf("expected this, that, got %v", writer.buf[0].Warn())
	}
}

func TestReduceNoWriter(t *testing.T) {
	ctx, flush := warning.Reduce(context.Background(), func(acc *multiWarn, wrr warning.Warning) *multiWarn {
		if acc == nil {
			acc = new(multiWarn)
		}

		acc.details = append(acc.details, wrr.Warn())

		return acc
	})

	if ctx != context.Background() {
		t.Errorf("expected same context, got %v", ctx)
	}

	warning.Warn(ctx, warning.New("this"))
	warning.Warn(ctx, warning.New("that"))

	flush()
}

func TestReduceNoWarning(t *testing.T) {
	writer := &mockWriter{}

	ctx := warning.Attach(context.Background(), writer)

	_, flush := warning.Reduce(ctx, func(acc *multiWarn, wrr warning.Warning) *multiWarn {
		if acc == nil {
			acc = new(multiWarn)
		}

		acc.details = append(acc.details, wrr.Warn())

		return acc
	})

	flush()

	if len(writer.buf) > 0 {
		t.Fatalf("expected zero warnings, got %v", len(writer.buf))
	}
}

func TestTap(t *testing.T) {
	writer := &mockWriter{}
	touched := false

	ctx := warning.Attach(context.Background(), writer)
	ctx = warning.Tap(ctx, func(_ warning.Warning) {
		touched = true
	})

	warning.Warn(ctx, warning.New("test"))

	if !touched {
		t.Errorf("expected touched, got not touched")
	}

	if len(writer.buf) != 1 {
		t.Errorf("expected 1 warning, got %v", len(writer.buf))
	}
}

func TestTapNoWriter(t *testing.T) {
	touched := false

	ctx := warning.Tap(context.Background(), func(_ warning.Warning) {
		touched = true
	})

	if ctx != context.Background() {
		t.Errorf("expected same context, got %v", ctx)
	}

	warning.Warn(ctx, warning.New("test"))

	if touched {
		t.Fatalf("expected not touched, got touched")
	}
}
