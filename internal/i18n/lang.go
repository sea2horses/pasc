package i18n

import "fmt"

type Lang uint8

const (
	ES Lang = iota
	EN
)

type Msg uint16

//go:generate stringer -type=Msg
const (
	IncorrectType Msg = iota
	VariableDoesntExist
)

var catalog = map[Lang]map[Msg]string{
	ES: {
		IncorrectType:       "tipo incorrecto: se esperaba %[1]s, se obtuvo %[2]s",
		VariableDoesntExist: "la variable '%[1]s' no existe",
	},
	EN: {
		IncorrectType:       "incorrect type: expected %[1]s, got %[2]s",
		VariableDoesntExist: "variable '%[1]s' doesn't exist",
	},
}

var current = ES

func Set(l Lang) { current = l }

func T(m Msg, args ...any) string {
	f, ok := catalog[current][m]
	// english fallback
	if !ok {
		f = catalog[EN][m]
	}
	return fmt.Sprintf(f, args...)
}
