package highlight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pascc/internal/toolchain"
)

// classes returns the non-plain segments as "class:text" pairs.
func classes(src string) []string {
	var out []string
	for _, s := range Segments([]rune(src)) {
		if s.Class != Plain {
			out = append(out, htmlClasses[s.Class]+":"+s.Text)
		}
	}
	return out
}

func TestSegmentsClassify(t *testing.T) {
	src := `Funcion r <- Doble(x) // comentario
	r = x * 2.5
Fin Funcion
Algoritmo demo
	Definir n Como entero
	n <- Doble(Raiz(PI)) MOD 3
	Si n >= 1 Y NO Falso Entonces Escribir 'hola', v[1]; FinSi
FinAlgoritmo`
	want := []string{
		"kw:Funcion", "id:r", "op:<-", "fn:Doble", "pu:(", "id:x", "pu:)", "co:// comentario",
		"id:r", "op:=", "id:x", "op:*", "nu:2.5",
		"kw:Fin", "kw:Funcion",
		"kw:Algoritmo", "id:demo",
		"kw:Definir", "id:n", "kw:Como", "ty:entero",
		"id:n", "op:<-", "fn:Doble", "pu:(", "bi:Raiz", "pu:(", "bi:PI", "pu:))", "op:MOD", "nu:3",
		"kw:Si", "id:n", "op:>=", "nu:1", "op:Y", "op:NO", "bo:Falso", "kw:Entonces", "kw:Escribir",
		"st:'hola'", "pu:,", "id:v", "pu:[", "nu:1", "pu:];", "kw:FinSi",
		"kw:FinAlgoritmo",
	}
	got := classes(src)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("unexpected segments\n got: %q\nwant: %q", got, want)
	}
}

func TestSegmentsMarkUnreadableText(t *testing.T) {
	got := classes("x = 1 $ 2\nEscribir \"sin cerrar\nFin")
	want := []string{"id:x", "op:=", "nu:1", "er:$", "nu:2", "kw:Escribir", `er:"sin`, `er:cerrar`, "kw:Fin"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("unexpected segments\n got: %q\nwant: %q", got, want)
	}
}

// Highlighting must never add, drop or reorder text, whatever the input.
func TestSegmentsPreserveSource(t *testing.T) {
	inputs := []string{
		"",
		"\r\n\tAlgoritmo a\r\n\tEscribir \"ñandú\" /* bloque\r\n de dos lineas */ 1\r\nFinAlgoritmo",
		"/* sin cerrar",
		"'otra comilla\"\n@#~ ¿?\n\"",
	}
	files, _ := filepath.Glob("../../demos/*.psc")
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		src, _ := toolchain.DecodeSource(data)
		inputs = append(inputs, string(src))
	}
	for _, src := range inputs {
		if got := Text(Segments([]rune(src))); got != src {
			t.Errorf("source not preserved\n got: %q\nwant: %q", got, src)
		}
	}
}

func TestANSI(t *testing.T) {
	got := ANSI(Segments([]rune("Si x /* a\nb */")))
	want := "\x1b[1;35mSi\x1b[0m x \x1b[3;90m/* a\x1b[0m\n\x1b[3;90mb */\x1b[0m"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHTMLEscapes(t *testing.T) {
	got := HTML(Segments([]rune(`Si a<b Entonces Escribir "<&>"`)))
	want := `<pre class="pasc"><code><span class="kw">Si</span> <span class="id">a</span>` +
		`<span class="op">&lt;</span><span class="id">b</span> <span class="kw">Entonces</span> ` +
		`<span class="kw">Escribir</span> <span class="st">&#34;&lt;&amp;&gt;&#34;</span></code></pre>`
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
	doc := HTMLDocument([]File{{Name: "<a>.psc", Segments: nil}})
	if !strings.Contains(doc, "<title>&lt;a&gt;.psc</title>") {
		t.Fatalf("title not escaped: %s", doc)
	}
}
