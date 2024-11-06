# warning
[![License](https://img.shields.io/badge/license-mit-green.svg)](https://go.wamod.dev/warning/blob/main/LICENSE)
[![Go](https://github.com/wamod-dev/warning/actions/workflows/go.yml/badge.svg)](https://github.com/wamod-dev/warning/actions/workflows/go.yml)
[![Codecov](https://codecov.io/gh/wamod-dev/warning/graph/badge.svg?token=QJRUAO7RZJ)](https://codecov.io/gh/wamod-dev/warning)
[![Go Report Card](https://goreportcard.com/badge/go.wamod.dev/warning)](https://goreportcard.com/report/go.wamod.dev/warning)
[![GoDoc](https://godoc.org/go.wamod.dev/warning?status.svg)](http://godoc.org/go.wamod.dev/warning)
[![Release](https://img.shields.io/github/release/wamod-dev/warning.svg)](https://go.wamod.dev/warning/releases/latest)

`warning` package provides mechanisms for capturing diagnostics using `context.Context`.
It allows to easily capture warning without modifying existing function signatures.
warning are captured using a `Collector`, which is attached to the `context.Context`.

## Overview

The package offers the following functionalities:

- **Capturing**: The package provides mechanisms to collect warning generated during the execution of a Go program. warning can be captured and stored for later analysis or reporting.
- **Filtering**: Developers can apply filters to selectively capture or ignore certain types of warning based on specified criteria. This allows for targeted handling of warning that meet specific conditions.
- **Transformation of warning**: warning can be transformed or modified in various ways to suit different requirements. This functionality enables developers to manipulate warning messages, format them differently, or perform other operations before storing or handling them.
- **Side-Effects**: The package supports the application of side-effects to warning, such as logging or triggering additional actions based on the occurrence of specific warning. This capability allows for flexible handling of warning beyond simple collection and storage.

## Installation

To use this package in your Go project, you can import it using:

```go
import "go.wamod.dev/warning"
```

## Usage

To start capturing warning create a new `Collector`:

```go
collector := warning.NewCollector()
defer collector.Close() // close before exiting
```

Attach this collector to your context:

```go
ctx := warning.Attach(context.Background(), collector)
```

Then, you can write warning to the context:

```go
warning.Warnf(ctx, "this is a warning")
```

Finally, capture all of warning back from the collector:

```go
wrrs, err := warning.ReadAll(collector)
```
### Helpers

#### Filter

This example demonstrates how to use the `Filter` function to filter warning:

```go
// filter warnings
ctx = warning.Filter(ctx, func(wrr warning.Warning) bool {
    return !strings.HasPrefix(wrr.Warn(), "ignore:")
})

// This warning will be captured
warning.Warnf(ctx, "capture: warning 1")

// This warning will be ignored
warning.Warnf(ctx, "ignore: warning 2") 
```

#### Map

Transform each written warning.

```go
ctx = warning.Map(ctx, func(wrr warning.Warning) warning.Warning {
    return warning.New(strings.ToUpper(wrr.Warn()))
})

// This warning will be transformed to "WARNING 1"
warning.Warnf(ctx, "warning 1")
```

#### Reduce

Combine all written warning into a single warning.

```go
// create a custom warning
type multiWarn struct {
	details []string
}

func (w *multiWarn) Warn() string {
	return strings.Join(w.details, ", ")
}

// reduce warnings
ctx, flush := warning.Reduce(ctx, func(acc *multiWarn*, wrr warning.Warning) *multiWarn {
    if acc == nil {
        acc = new(multiWarn)
    }
    acc.details = append(acc.details, wrr.Warn())
    return acc
})

// flush on exit
defer flush()

// Write warnings
warning.Warnf(ctx, "warning 1")
warning.Warnf(ctx, "warning 2")
warning.Warnf(ctx, "warning 3")

// The captured warning will be:
// &multiWarn{"warning 1", "warning 2", "warning 3"}
```

#### Tap

It does not modify the warning or the context but is useful for side effects like logging.

```go
ctx = warning.Tap(ctx, func(wrr warning.Warning) {
    slog.Warn(wrr.Warn())
})

// Now every new warning will be logged using `slog`
warning.Warnf(ctx, "this is a warning")
warning.Warnf(ctx, "this is another warning")
```

## Contributing

Thank you for your interest in contributing to the `warning` Go library! We welcome and appreciate any contributions, whether they be bug reports, feature requests, or code changes.

If you've found a bug, please create an issue in the GitHub repository describing the problem, including any relevant error messages and a minimal reproduction of the issue.

## License

`warning` is licensed under the [MIT License](LICENSE).