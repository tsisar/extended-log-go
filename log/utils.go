package log

import (
	"fmt"
	"io"
	"os"
)

func fprintf(w io.Writer, format string, a ...any) {
	n, err := fmt.Fprintf(w, format, a)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Fprintf: %v\n", err)
	}
	fmt.Printf("%d bytes written.\n", n)
}
