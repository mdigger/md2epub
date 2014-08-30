package main

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"fmt"
	"github.com/russross/blackfriday"
	"regexp"
)

var (
	extensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_FOOTNOTES |
		blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK |
		blackfriday.EXTENSION_HEADER_IDS |
		// blackfriday.EXTENSION_LAX_HTML_BLOCKS |
		blackfriday.EXTENSION_HARD_LINE_BREAK
	htmlFlags = 0 |
		blackfriday.HTML_USE_XHTML
)

func Markdown(data []byte) []byte {
	return blackfriday.Markdown(data, &htmlRender{
		Renderer: blackfriday.HtmlRenderer(htmlFlags, "", ""),
	}, extensions)
}

type htmlRender struct {
	blackfriday.Renderer
}

func (_ *htmlRender) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	fmt.Fprintf(out, "<sup><a rel=\"footnote\" href=\"#fn:%s\" epub:type=\"noteref\">%d</a></sup>",
		hashSlug(ref), id)
}

func (_ *htmlRender) Footnotes(out *bytes.Buffer, text func() bool) {
	text()
}

var reMultilines = regexp.MustCompile(`\n{2,}`)

func (_ *htmlRender) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	fmt.Fprintf(out, "\n<aside id=\"fn:%s\" epub:type=\"footnote\">\n%s</aside>\n",
		hashSlug(name), reMultilines.ReplaceAllLiteral(text, []byte("\n")))
}

func (_ *htmlRender) NormalText(out *bytes.Buffer, text []byte) {
	text = reMultilines.ReplaceAllLiteral(text, []byte("\n")) // Убираем пустые строки
	str := html.EscapeString(string(text))
	out.WriteString(str)
	// TODO: Smartypants
	// mark := 0
	// for i := 0; i < len(text); i++ {
	// 	if action := options.smartypants[text[i]]; action != nil {
	// 		if i > mark {
	// 			out.Write(text[mark:i])
	// 		}

	// 		previousChar := byte(0)
	// 		if i > 0 {
	// 			previousChar = text[i-1]
	// 		}
	// 		i += action(out, &smrt, previousChar, text[i:])
	// 		mark = i + 1
	// 	}
	// }

	// if mark < len(text) {
	// 	out.Write(text[mark:])
	// }
}
