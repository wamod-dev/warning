package warning_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"go.wamod.dev/warning"
)

func TestScanner(t *testing.T) {
	want := []warning.Warning{
		warning.New("test-1"),
		warning.New("test-2"),
	}

	reader := &mockReader{
		[]mockReaderResult{
			{want[0], nil},
			{want[1], nil},
			{nil, io.EOF},
		},
	}

	scanner := warning.NewScanner(reader)

	for i, wrr := range want {
		if !scanner.Scan() {
			t.Fatalf("expected to scan warning %v", i)
		}

		if got := scanner.Warning(); got != wrr {
			t.Errorf("expected %v, got %v", wrr, got)
		}
	}

	if scanner.Scan() {
		t.Fatalf("expected to not scan any more warning")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScanner_EOF(t *testing.T) {
	reader := &mockReader{
		[]mockReaderResult{
			{warning.New("test-1"), nil},
			{warning.New("test-2"), nil},
			{nil, io.EOF},
		},
	}

	want := [3]bool{
		true,
		true,
		false,
	}

	scanner := warning.NewScanner(reader)

	got := [3]bool{
		scanner.Scan(),
		scanner.Scan(),
		scanner.Scan(),
	}

	if want != got {
		t.Fatalf("expected scans = %v, got %v", want, got)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScanner_UnexpectedError(t *testing.T) {
	wantErr := fmt.Errorf("test-error") //nolint:err113

	reader := &mockReader{
		[]mockReaderResult{
			{warning.New("test-1"), nil},
			{warning.New("test-2"), nil},
			{nil, wantErr},
			{warning.New("test-4"), nil},
		},
	}

	want := [3]bool{
		true,
		true,
		false,
	}

	scanner := warning.NewScanner(reader)

	got := [3]bool{
		scanner.Scan(),
		scanner.Scan(),
		scanner.Scan(),
	}

	if want != got {
		t.Fatalf("expected scans = %v, got %v", want, got)
	}

	if err := scanner.Err(); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}

	if scanner.Scan() {
		t.Fatalf("expected to not scan any more warning")
	}

	if err := scanner.Err(); !errors.Is(err, wantErr) {
		t.Fatalf("expected same error %v, got %v", wantErr, err)
	}
}

func TestScanner_Empty(t *testing.T) {
	reader := &mockReader{
		[]mockReaderResult{
			{nil, io.EOF},
		},
	}

	scanner := warning.NewScanner(reader)

	if scanner.Scan() {
		t.Fatalf("expected to not scan any warning")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
