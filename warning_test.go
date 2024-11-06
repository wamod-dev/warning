package warning_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"go.wamod.dev/warning"
)

// Example demonstrates how to use this package to read and write warnings.
func Example() {
	// create a new collector
	collector := warning.NewCollector()
	defer collector.Close() // make sure to close the collector when done

	// attach the collector to a context
	ctx := warning.Attach(context.Background(), collector)

	// use Warn or Warnf to write warning to the context
	warning.Warnf(ctx, "this is a warning 1")
	warning.Warnf(ctx, "this is a warning 2")

	// use Scanner to read warning one by one
	scanner := warning.NewScanner(collector)
	for scanner.Scan() {
		wrr := scanner.Warning()
		fmt.Println(wrr.Warn())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// Output:
	// this is a warning 1
	// this is a warning 2
}

func TestNew(t *testing.T) {
	want := "test-warning"
	wrr := warning.New(want)

	if wrr.Warn() != want {
		t.Errorf("expected %s, got %v", want, wrr.Warn())
	}

	if got, ok := wrr.(fmt.Stringer); !ok {
		t.Errorf("expected warning to implement fmt.Stringer")
	} else if got := got.String(); got != want {
		t.Errorf("expected %s, got %v", want, got)
	}

	if got, ok := wrr.(json.Marshaler); !ok {
		t.Errorf("expected warning to implement json.Marshaler")
	} else {
		got, err := got.MarshalJSON()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !bytes.Equal(got, []byte(`"`+want+`"`)) {
			t.Errorf("expected %v, got %v", want, got)
		}
	}
}

func TestWarn(t *testing.T) {
	want := warning.New("test-warning")
	writer := &mockWriter{}

	ctx := warning.Attach(context.Background(), writer)

	err := warning.Warn(ctx, want)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(writer.buf) != 1 || writer.buf[0] != want {
		t.Fatalf("expected %v, got %v", want, writer.buf)
	}
}

func TestWarnNoWriter(t *testing.T) {
	want := warning.New("test-warning")

	err := warning.Warn(context.Background(), want)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWarnError(t *testing.T) {
	wantErr := fmt.Errorf("test-error") //nolint:err113
	wantWrr := warning.New("test-warning")

	writer := &mockWriter{result: wantErr}
	ctx := warning.Attach(context.Background(), writer)

	if err := warning.Warn(ctx, wantWrr); !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}

	if len(writer.buf) > 0 {
		t.Errorf("expected zero warnings, got %v", writer.buf)
	}
}

func TestWarnf(t *testing.T) {
	writer := &mockWriter{}
	want := "test-warning: sub-warning"
	ctx := warning.Attach(context.Background(), writer)

	err := warning.Warnf(ctx, "test-warning: %s", warning.New("sub-warning"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(writer.buf) != 1 || writer.buf[0].Warn() != want {
		t.Fatalf("expected test-1, got %v", writer.buf)
	}
}

func TestAttach(t *testing.T) {
	writers := [2]*mockWriter{
		new(mockWriter),
		new(mockWriter),
	}

	ctx := warning.Attach(context.Background(), writers[0])
	warning.Warn(ctx, warning.New("test-1"))

	ctx = warning.Attach(ctx, writers[1])
	warning.Warn(ctx, warning.New("test-2"))

	if len(writers[0].buf) != 2 {
		t.Fatalf("expected 2 warning, got %v", writers[0].buf)
	}

	if len(writers[1].buf) != 1 {
		t.Fatalf("expected 1 warning, got %v", writers[1].buf)
	}
}

func TestDetach(t *testing.T) {
	writer := &mockWriter{}

	ctx := warning.Attach(context.Background(), writer)
	ctx = warning.Detach(ctx)

	wantWrr := warning.New("test-warning")

	err := warning.Warn(ctx, wantWrr)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(writer.buf) > 0 {
		t.Fatalf("expected no warning, got %v", writer.buf)
	}
}

func TestDetachNoWriter(t *testing.T) {
	ctx := warning.Detach(context.Background())

	err := warning.Warn(ctx, warning.New("test-warning"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
