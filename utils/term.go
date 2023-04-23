// inspired by a similar library from age, but they don't export functions so we
// can't just import and re-use them
package utils

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func withTerminal(f func(in, out *os.File) error) error {
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer tty.Close()
		return f(tty, tty)
	} else if term.IsTerminal(int(os.Stdin.Fd())) {
		return f(os.Stdin, os.Stdin)
	} else {
		return fmt.Errorf("standard input is not a terminal, and /dev/tty is not available: %v", err)
	}
}

func printfToTerminal(format string, v ...interface{}) error {
	return withTerminal(func(_, out *os.File) error {
		_, err := fmt.Fprintf(out, format+"\n", v...)
		return err
	})
}

func readSecret(prompt string) (s []byte, err error) {
	err = withTerminal(func(in, out *os.File) error {
		fmt.Fprintf(out, "%s ", prompt)
		defer clearLine(out)
		s, err = term.ReadPassword(int(in.Fd()))
		return err
	})
	return
}

func clearLine(out io.Writer) {
	const (
		CUI = "\033["   // Control Sequence Introducer
		CPL = CUI + "F" // Cursor Previous Line
		EL  = CUI + "K" // Erase in Line
	)
	fmt.Fprintf(out, "\r\n"+CPL+EL)
}
