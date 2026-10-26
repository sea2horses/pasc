package toolchain

import (
	"bytes"
	"unicode/utf8"
)

var windows1252High = [32]rune{
	0x20AC, 0x0081, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
	0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x008D, 0x017D, 0x008F,
	0x0090, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x009D, 0x017E, 0x0178,
}

// decodeSource prefers UTF-8 and falls back to Windows-1252 for legacy PSeInt
// files. These encodings have no in-band marker in the common case, so invalid
// UTF-8 is the signal used to select the legacy decoding.
func decodeSource(data []byte) []rune {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(data) {
		return []rune(string(data))
	}

	runes := make([]rune, len(data))
	for i, b := range data {
		switch {
		case b < 0x80, b >= 0xA0:
			runes[i] = rune(b)
		default:
			runes[i] = windows1252High[int(b)-0x80]
		}
	}
	return runes
}
