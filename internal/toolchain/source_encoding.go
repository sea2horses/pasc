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

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// SourceEncoding records how a source file was stored, so tools that rewrite
// it (such as the formatter) can write it back the same way.
type SourceEncoding struct {
	BOM         bool /* The file started with a UTF-8 byte order mark */
	Windows1252 bool /* The file was not valid UTF-8 and was read as Windows-1252 */
}

func decodeSource(data []byte) []rune {
	src, _ := DecodeSource(data)
	return src
}

// DecodeSource prefers UTF-8 and falls back to Windows-1252 for legacy PSeInt
// files. These encodings have no in-band marker in the common case, so invalid
// UTF-8 is the signal used to select the legacy decoding.
func DecodeSource(data []byte) ([]rune, SourceEncoding) {
	var enc SourceEncoding
	if trimmed, ok := bytes.CutPrefix(data, utf8BOM); ok {
		data, enc.BOM = trimmed, true
	}
	if utf8.Valid(data) {
		return []rune(string(data)), enc
	}

	enc.Windows1252 = true
	runes := make([]rune, len(data))
	for i, b := range data {
		switch {
		case b < 0x80, b >= 0xA0:
			runes[i] = rune(b)
		default:
			runes[i] = windows1252High[int(b)-0x80]
		}
	}
	return runes, enc
}

// EncodeSource is the inverse of DecodeSource. Runes that Windows-1252 cannot
// represent are written as '?'; text decoded from such a file only gains
// ASCII when it is reformatted, so this does not happen in practice.
func EncodeSource(text string, enc SourceEncoding) []byte {
	var out []byte
	if enc.BOM {
		out = append(out, utf8BOM...)
	}
	if !enc.Windows1252 {
		return append(out, text...)
	}
	for _, r := range text {
		out = append(out, encodeWindows1252(r))
	}
	return out
}

func encodeWindows1252(r rune) byte {
	if r < 0x80 || (r >= 0xA0 && r <= 0xFF) {
		return byte(r)
	}
	for i, high := range windows1252High {
		if high == r {
			return byte(0x80 + i)
		}
	}
	return '?'
}
