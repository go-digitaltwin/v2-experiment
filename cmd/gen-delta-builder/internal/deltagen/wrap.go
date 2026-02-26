package deltagen

import "strings"

const commentWidth = 77 // 80 columns minus "// " prefix

// wrapComment word-wraps a paragraph of text into Go comment lines. Each
// output line is prefixed with "// " and broken at word boundaries so the
// total line width stays within 80 columns.
//
// Blank lines in the input (empty strings) produce a bare "//" separator.
func wrapComment(paragraphs ...string) string {
	var b strings.Builder
	for i, para := range paragraphs {
		if i > 0 {
			b.WriteByte('\n')
		}
		if para == "" {
			b.WriteString("//")
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			b.WriteString("//")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if len(line)+1+len(w) > commentWidth {
				b.WriteString("// ")
				b.WriteString(line)
				b.WriteByte('\n')
				line = w
			} else {
				line += " " + w
			}
		}
		b.WriteString("// ")
		b.WriteString(line)
	}
	return b.String()
}
