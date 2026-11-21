package toolchain

import (
	"bytes"
	"testing"
)

func TestSourceEncodingRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		want SourceEncoding
	}{
		{"utf-8", []byte("Escribir \"año\"\n"), SourceEncoding{}},
		{"utf-8 with bom", []byte("\xEF\xBB\xBFEscribir \"año\"\n"), SourceEncoding{BOM: true}},
		{"windows-1252", []byte("Escribir \"a\xF1o \x80\"\n"), SourceEncoding{Windows1252: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, enc := DecodeSource(tc.data)
			if enc != tc.want {
				t.Fatalf("encoding = %+v, want %+v", enc, tc.want)
			}
			if got := EncodeSource(string(src), enc); !bytes.Equal(got, tc.data) {
				t.Fatalf("round trip = %q, want %q", got, tc.data)
			}
		})
	}
}
