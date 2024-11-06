package warning

import (
	"context"
)

// Map returns a new context that transforms each written warning using the provided function.
func Map(ctx context.Context, mapFunc func(wrr Warning) Warning) context.Context {
	writer := getWriter(ctx)
	if writer == nil {
		return ctx
	}

	return setWriter(ctx, &mapWriter{writer, mapFunc})
}

type mapWriter struct {
	next    Writer
	mapFunc func(Warning) Warning
}

func (writer *mapWriter) WriteWarning(wrr Warning) error {
	return writer.next.WriteWarning(writer.mapFunc(wrr))
}

// Filter returns a new context that filters written warnings using the provided function.
func Filter(ctx context.Context, filterFunc func(wrr Warning) bool) context.Context {
	writer := getWriter(ctx)
	if writer == nil {
		return ctx
	}

	return setWriter(ctx, &filterWriter{writer, filterFunc})
}

type filterWriter struct {
	next       Writer
	filterFunc func(Warning) bool
}

func (writer *filterWriter) WriteWarning(wrr Warning) error {
	if writer.filterFunc(wrr) {
		return writer.next.WriteWarning(wrr)
	}

	return nil
}

// Reduce returns a new context that reduces written warnings using the provided function.
// It also returns a flush() function that once called, writes the reduced warning to the underlying writer.
// If no warnings are written, it does nothing.
func Reduce[T Warning](ctx context.Context, reduceFunc func(acc T, wrr Warning) T) (_ context.Context, flush func()) {
	writer := getWriter(ctx)
	if writer == nil {
		return ctx, func() {}
	}

	collector := NewCollector()

	return setWriter(ctx, collector), func() {
		defer collector.Close()

		wrrs, err := ReadAll(collector)
		if err != nil || len(wrrs) == 0 {
			return
		}

		acc := *new(T)

		for _, wrr := range wrrs {
			acc = reduceFunc(acc, wrr)
		}

		_ = writer.WriteWarning(acc)
	}
}

// Tap returns a new context that taps written warnings using the provided function.
// It does not modify the warnings or the context but is useful for side effects like logging.
func Tap(ctx context.Context, tapFunc func(wrr Warning)) context.Context {
	writer := getWriter(ctx)
	if writer == nil {
		return ctx
	}

	return setWriter(ctx, &tapWriter{writer, tapFunc})
}

type tapWriter struct {
	next    Writer
	tapFunc func(Warning)
}

func (writer *tapWriter) WriteWarning(wrr Warning) error {
	writer.tapFunc(wrr)

	return writer.next.WriteWarning(wrr)
}
