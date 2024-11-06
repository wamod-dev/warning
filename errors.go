package warning

import "fmt"

// ErrClosed is returned when the warning stream is closed.
var ErrClosed = fmt.Errorf("warning stream is closed")
