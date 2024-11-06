package warning_test

import (
	"errors"
	"fmt"
	"testing"

	"go.wamod.dev/warning"
)

type mockWriter struct {
	buf    []warning.Warning
	result error
}

func (w *mockWriter) WriteWarning(wrr warning.Warning) error {
	if w.result != nil {
		return w.result
	}

	w.buf = append(w.buf, wrr)

	return w.result
}

func TestNewMultiWriter(t *testing.T) {
	writers := []*mockWriter{
		{result: nil},
		{result: nil},
	}

	topWriter := warning.NewMultiWriter(writers[0], writers[1])
	wantWrr := warning.New("test")

	err := topWriter.WriteWarning(wantWrr)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	for _, writer := range writers {
		if len(writer.buf) != 1 || writer.buf[0] != wantWrr {
			t.Fatalf("expected %v, got %v", wantWrr, writer.buf)
		}
	}
}

func TestNewMultiWriterErrors(t *testing.T) {
	writers := []*mockWriter{
		{result: nil},
		{result: fmt.Errorf("test-error-1")}, //nolint:err113
		{result: fmt.Errorf("test-error-2")}, //nolint:err113
	}

	topWriter := warning.NewMultiWriter(writers[0], writers[1], writers[2])
	wantWrr := warning.New("test")

	err := topWriter.WriteWarning(wantWrr)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, writers[1].result) {
		t.Fatalf("expected %v, got %v", writers[1].result, err)
	}

	if !errors.Is(err, writers[2].result) {
		t.Fatalf("expected %v, got %v", writers[2].result, err)
	}
}
