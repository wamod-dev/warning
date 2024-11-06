// Package warnings implements mechanisms for capturing diagnostics using [context.Context].
//
// It is designed to provide an easy way to capture warnings without modifying existing function signature.
//
// To start capturing warnings, you need to attach a [Collector] to the context using [Attach] function.
// This will create a new context with the collector attached and all the warnings written to
// the context will be captured.
//
//	collector := warning.NewCollector()
//	defer collector.Close()
//	ctx := warning.Attach(context.Background(), collector)
//
// Use [Warn] or [Warnf] functions to write warnings to the context. H
//
//	warning.Warnf(ctx, "this is a warning")
//
// To read all the warnings from the collector, use [ReadAll] function
//
//	wrrs, err := warning.ReadAll(collector)
//
// Or you can use [Scanner] function to read warnings one by one.
//
//	scanner := warning.NewScanner(collector)
//	for scanner.Scan() {
//		wrr := scanner.Warning()
//	}
//	if err := scanner.Err(); err != nil {
//		// handle error
//	}
//
// If you need a new context that does not collect warnings anymore, use [Detach] function.
//
//	ctx = warning.Detach(ctx)
//
// Use [Map], [Filter], [Reduce] or [Tap] helper functions to apply transformations,
// filters or side-effects to the warnings.
package warning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Warning is an interface representing a warning.
type Warning interface {
	Warn() string
}

type warningString struct {
	msg string
}

func (wrr *warningString) Warn() string {
	return wrr.msg
}

func (wrr *warningString) String() string {
	return wrr.msg
}

func (wrr *warningString) MarshalJSON() ([]byte, error) {
	return json.Marshal(wrr.msg)
}

// New creates a new warning from a given string.
func New(msg string) Warning {
	return &warningString{msg}
}

type writerKey struct{}

func setWriter(ctx context.Context, w Writer) context.Context {
	return context.WithValue(ctx, writerKey{}, w)
}

func getWriter(ctx context.Context) Writer {
	writer, ok := ctx.Value(writerKey{}).(Writer)
	if !ok {
		return nil
	}

	return writer
}

// Warn writes warning to the context. When multiple warnings are provided, they are written in order.
// If no writer is attached to the context, it does nothing and returns nil.
// If any of the warning fail to write, all the warnings are returned as one error.
func Warn(ctx context.Context, wrrs ...Warning) error {
	writer := getWriter(ctx)
	if writer == nil {
		return nil
	}

	var errs []error

	for _, wrr := range wrrs {
		if err := writer.WriteWarning(wrr); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Warnf is a helper function that formats the warning and writes it to the context.
// If the format string contains any [Warning] arguments, they are converted to strings before formatting.
func Warnf(ctx context.Context, format string, args ...any) error {
	for i, arg := range args {
		if wrr, ok := arg.(Warning); ok {
			args[i] = wrr.Warn()
		}
	}

	return Warn(ctx, New(fmt.Sprintf(format, args...)))
}

// Attach returns a new context that collects warnings using the provided writer.
// If a writer is already attached to the context, it creates a new writer that writes to both.
func Attach(ctx context.Context, writer Writer) context.Context {
	if found := getWriter(ctx); found != nil {
		writer = NewMultiWriter(found, writer)
	}

	return setWriter(ctx, writer)
}

// Detach returns a new context that does not propagate warnings up the chain.
func Detach(ctx context.Context) context.Context {
	writer := getWriter(ctx)
	if writer == nil {
		return ctx
	}

	return setWriter(ctx, nil)
}
