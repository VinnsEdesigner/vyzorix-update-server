package outcome

import (
	"fmt"
	"strings"
)

//nolint:govet // fieldalignment: field order chosen for readability, not packing
type Evidence struct {
	Detail    string
	Truncated bool
	Bytes     int
	ExitCode  *int
	Signal    string
	Facts     map[string]string
}

const maxEvidenceBytes = 8 * 1024

func Capture(detail string) Evidence {
	if len(detail) <= maxEvidenceBytes {
		return Evidence{Detail: detail, Bytes: len(detail)}
	}
	return Evidence{
		Detail:    detail[:maxEvidenceBytes],
		Truncated: true,
		Bytes:     len(detail),
	}
}

func (e Evidence) WithExitCode(code int) Evidence {
	e.ExitCode = &code
	return e
}

func (e Evidence) WithSignal(signal string) Evidence {
	e.Signal = signal
	return e
}

func (e Evidence) WithFact(key, value string) Evidence {
	if e.Facts == nil {
		e.Facts = map[string]string{}
	}
	e.Facts[key] = value
	return e
}

func (e Evidence) IsEmpty() bool {
	return e.Detail == "" && e.Bytes == 0 && e.ExitCode == nil && e.Signal == "" && len(e.Facts) == 0
}

func (e Evidence) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "evidence(%dB", e.Bytes)
	if e.Truncated {
		b.WriteString(", truncated")
	}
	if e.ExitCode != nil {
		fmt.Fprintf(&b, ", exit=%d", *e.ExitCode)
	}
	if e.Signal != "" {
		fmt.Fprintf(&b, ", signal=%s", e.Signal)
	}
	if len(e.Facts) > 0 {
		fmt.Fprintf(&b, ", facts=%d", len(e.Facts))
	}
	b.WriteString(")")
	return b.String()
}
