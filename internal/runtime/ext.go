//go:build ignore

package main

import (
	"os"

	"golang.org/x/term"
)

func WaitKey() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		if _, _, err := reader.ReadRune(); err != nil {
			RuntimeError(ErrCouldNotReadKey)
		}
		return
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		RuntimeError(ErrTermError)
	}
	defer func() {
		if err := term.Restore(fd, state); err != nil {
			RuntimeError(ErrRestoreTermError)
		}
	}()

	if _, _, err := reader.ReadRune(); err != nil {
		RuntimeError(ErrCouldNotReadKey)
	}
}
