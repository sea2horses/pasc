//go:build ignore

package main

func WaitKey() {
	writer.WriteString("Presione Enter para continuar...")
	writer.Flush()
	if _, err := reader.ReadString('\n'); err != nil {
		RuntimeError(ErrCouldNotReadKey)
	}
	writer.WriteRune('\n')
	writer.Flush()
}
