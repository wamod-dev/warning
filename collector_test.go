package warning_test

import (
	"errors"
	"testing"

	"go.wamod.dev/warning"
)

func TestCollector_Close(t *testing.T) {
	collector := warning.NewCollector()

	err := collector.Close()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	err = collector.Close()
	if !errors.Is(err, warning.ErrClosed) {
		t.Fatalf("expected %v, got %v", warning.ErrClosed, err)
	}
}

func TestCollector_WriteWarning(t *testing.T) {
	collector := warning.NewCollector()

	err := collector.WriteWarning(warning.New("test-1"))
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	err = collector.Close()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	err = collector.WriteWarning(warning.New("test-2"))
	if !errors.Is(err, warning.ErrClosed) {
		t.Fatalf("expected %v, got %v", warning.ErrClosed, err)
	}
}

func TestCollector_ReadWarning(t *testing.T) {
	collector := warning.NewCollector()

	err := collector.WriteWarning(warning.New("test-1"))
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	wrr, err := collector.ReadWarning()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	} else if wrr.Warn() != "test-1" {
		t.Fatalf("expected test-1, got %v", wrr.Warn())
	}

	err = collector.Close()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	wrr, err = collector.ReadWarning()
	if !errors.Is(err, warning.ErrClosed) {
		t.Fatalf("expected %v, got %v", warning.ErrClosed, err)
	} else if wrr != nil {
		t.Fatalf("expected nil, got %v", wrr)
	}
}
