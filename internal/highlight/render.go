package highlight

import (
	"html"
	"strings"
)

// ansiStyles uses the basic 16-color palette, so the output adapts to the
// terminal's own theme. Plain, identifiers and punctuation keep the default
// foreground color.
var ansiStyles = map[Class]string{
	Keyword:  "1;35", /* bold magenta */
	Operator: "33",   /* yellow */
	Function: "34",   /* blue */
	Builtin:  "36",   /* cyan */
	Type:     "1;36", /* bold cyan */
	String:   "32",   /* green */
	Number:   "33",   /* yellow */
	Boolean:  "1;33", /* bold yellow */
	Comment:  "3;90", /* italic gray */
	Invalid:  "4;31", /* underlined red */
}

// ANSI renders segments with terminal color escape sequences. Each colored
// line is reset before its line break, so pagers such as "less -R" and
// line-based tools like grep show every line with its own colors.
func ANSI(segments []Segment) string {
	var b strings.Builder
	for _, s := range segments {
		style, ok := ansiStyles[s.Class]
		if !ok {
			b.WriteString(s.Text)
			continue
		}
		for i, line := range strings.Split(s.Text, "\n") {
			if i > 0 {
				b.WriteByte('\n')
			}
			if line == "" {
				continue
			}
			b.WriteString("\x1b[" + style + "m" + line + "\x1b[0m")
		}
	}
	return b.String()
}

// htmlClasses are the CSS classes of each segment class. Plain text is not
// wrapped in a span.
var htmlClasses = map[Class]string{
	Keyword:     "kw",
	Operator:    "op",
	Punctuation: "pu",
	Identifier:  "id",
	Function:    "fn",
	Builtin:     "bi",
	Type:        "ty",
	String:      "st",
	Number:      "nu",
	Boolean:     "bo",
	Comment:     "co",
	Invalid:     "er",
}

// HTML renders segments as a <pre> element whose spans carry the classes
// styled by Stylesheet.
func HTML(segments []Segment) string {
	var b strings.Builder
	b.WriteString(`<pre class="pasc"><code>`)
	for _, s := range segments {
		text := html.EscapeString(s.Text)
		if class, ok := htmlClasses[s.Class]; ok {
			b.WriteString(`<span class="` + class + `">` + text + `</span>`)
		} else {
			b.WriteString(text)
		}
	}
	b.WriteString("</code></pre>")
	return b.String()
}

// Stylesheet styles the output of HTML, with light and dark variants.
const Stylesheet = `.pasc {
  --fg: #1f2328; --bg: #f6f8fa; --kw: #8250df; --op: #953800; --fn: #0550ae;
  --bi: #0a7c86; --ty: #0a7c86; --st: #116329; --nu: #953800; --co: #6e7781; --er: #cf222e;
  color: var(--fg); background: var(--bg);
  padding: 1rem; border-radius: 6px; overflow-x: auto; tab-size: 4;
  font: 14px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
@media (prefers-color-scheme: dark) {
  .pasc {
    --fg: #e6edf3; --bg: #161b22; --kw: #d2a8ff; --op: #ffa657; --fn: #79c0ff;
    --bi: #56d4dd; --ty: #56d4dd; --st: #7ee787; --nu: #ffa657; --co: #8b949e; --er: #ff7b72;
  }
}
.pasc .kw { color: var(--kw); font-weight: 600; }
.pasc .op, .pasc .nu { color: var(--nu); }
.pasc .fn { color: var(--fn); }
.pasc .bi { color: var(--bi); }
.pasc .ty { color: var(--ty); font-weight: 600; }
.pasc .st { color: var(--st); }
.pasc .bo { color: var(--nu); font-weight: 600; }
.pasc .co { color: var(--co); font-style: italic; }
.pasc .er { color: var(--er); text-decoration: wavy underline; }
`

// File is one highlighted source file of an HTML document.
type File struct {
	Name     string
	Segments []Segment
}

// HTMLDocument renders a standalone page showing each file in turn.
func HTMLDocument(files []File) string {
	title := "PSeInt"
	if len(files) == 1 {
		title = files[0].Name
	}

	var b strings.Builder
	b.WriteString("<!doctype html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>" + html.EscapeString(title) + "</title>\n")
	b.WriteString("<style>\nbody { margin: 2rem auto; max-width: 60rem; padding: 0 1rem; font-family: system-ui, sans-serif; background: #fff; color: #1f2328; }\n")
	b.WriteString("@media (prefers-color-scheme: dark) { body { background: #0d1117; color: #e6edf3; } }\n")
	b.WriteString("h2 { font-size: 1rem; font-family: ui-monospace, monospace; }\n")
	b.WriteString(Stylesheet)
	b.WriteString("</style>\n</head>\n<body>\n")
	for _, file := range files {
		if len(files) > 1 {
			b.WriteString("<h2>" + html.EscapeString(file.Name) + "</h2>\n")
		}
		b.WriteString(HTML(file.Segments) + "\n")
	}
	b.WriteString("</body>\n</html>\n")
	return b.String()
}
