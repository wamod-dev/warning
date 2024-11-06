package warning_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"go.wamod.dev/warning"
)

type mockReader struct {
	results []mockReaderResult
}

func (r *mockReader) ReadWarning() (warning.Warning, error) {
	if len(r.results) == 0 {
		return nil, io.EOF
	}

	var result mockReaderResult

	result, r.results = r.results[0], r.results[1:]

	return result.wrr, result.err
}

type mockReaderResult struct {
	wrr warning.Warning
	err error
}

func TestReadAll(t *testing.T) {
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

	wrrs, err := warning.ReadAll(reader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if l := len(wrrs); l != 2 {
		t.Fatalf("expected 2 warning, got %v", l)
	}

	for i, wrr := range want {
		if got := wrrs[i]; got != wrr {
			t.Errorf("expected %v, got %v", wrr, got)
		}
	}
}

func TestReadAll_UnexpectedError(t *testing.T) {
	wantErr := fmt.Errorf("test-error") //nolint:err113

	reader := &mockReader{
		[]mockReaderResult{
			{warning.New("test-1"), nil},
			{warning.New("test-2"), nil},
			{nil, wantErr},
		},
	}

	wrrs, err := warning.ReadAll(reader)
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error %v, got %v", wantErr, err)
	}

	if len(wrrs) > 0 {
		t.Errorf("expected no warning, got %v", wrrs)
	}
}

func TestReadAll_Empty(t *testing.T) {
	reader := &mockReader{
		[]mockReaderResult{
			{nil, io.EOF},
		},
	}

	wrrs, err := warning.ReadAll(reader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if l := len(wrrs); l > 0 {
		t.Fatalf("expected no warning, got %v", l)
	}
}
