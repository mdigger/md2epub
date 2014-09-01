package main

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"fmt"
	"github.com/russross/blackfriday"
	"hash/crc64"
	"regexp"
)

// Флаги для преобразования из Markdown в HTML.
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
		blackfriday.EXTENSION_HEADER_IDS
		// blackfriday.EXTENSION_LAX_HTML_BLOCKS |
		// blackfriday.EXTENSION_HARD_LINE_BREAK
	htmlFlags = 0 |
		blackfriday.HTML_USE_XHTML
)

// Markdown преобразует данные из формата Markdown в HTML.
func Markdown(data []byte) []byte {
	return blackfriday.Markdown(data, &htmlRender{
		Renderer: blackfriday.HtmlRenderer(htmlFlags, "", ""),
	}, extensions)
}

// htmlRender переопределяет преобразование некоторых видов информации из Markdwon
// в HTML. В тех случаях, когда переопределение не требуется, используется стандартный
// конвертер Markdown.
type htmlRender struct {
	blackfriday.Renderer
}

// FootnoteRef обрабатывает ссылки на сноски.
func (_ *htmlRender) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	fmt.Fprintf(out, "<sup><a rel=\"footnote\" href=\"#fn:%s\" epub:type=\"noteref\">%d</a></sup>",
		hashSlug(ref), id)
}

// Footnotes обрабатывает сноски.
func (_ *htmlRender) Footnotes(out *bytes.Buffer, text func() bool) {
	text()
}

// Поиск строк состоящих только из символов новой строки.
var reMultilines = regexp.MustCompile(`\n{2,}`)

// FootnoteItem обрабатывает содержимое сноски.
func (_ *htmlRender) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	fmt.Fprintf(out, "\n<aside id=\"fn:%s\" epub:type=\"footnote\">\n%s</aside>\n",
		hashSlug(name), reMultilines.ReplaceAllLiteral(text, []byte("\n")))
}

// NormalText выводит обычный текст. В данном случае, он предварительно обрабатывается
// и в нем происходят типографские замены.
func (_ *htmlRender) NormalText(out *bytes.Buffer, text []byte) {
	text = reMultilines.ReplaceAllLiteral(text, []byte("\n")) // Убираем пустые строки
	text = []byte(html.EscapeString(string(text)))
	runes := bytes.Runes(text)
	for i := 0; i < len(runes); i++ {
		switch c := runes[i]; {
		case c == '.' && len(runes) >= i+3 && runes[i+1] == '.' && runes[i+2] == '.':
			out.WriteString("&hellip;")
			i += 2
		case c == '(' && len(runes) >= i+3:
			if runes[i+2] == ')' {
				if runes[i+1] == 'c' || runes[i+1] == 'C' {
					out.WriteString("&copy;")
					i += 2
				} else if runes[i+1] == 'r' || runes[i+1] == 'R' {
					out.WriteString("&reg;")
					i += 2
				} else {
					out.WriteRune(c)
				}
			} else if len(runes) >= i+4 && runes[i+3] == ')' &&
				(runes[i+1] == 't' || runes[i+1] == 'T') &&
				(runes[i+2] == 'm' || runes[i+2] == 'M') {
				out.WriteString("&trade;")
				i += 3
			} else {
				out.WriteRune(c)
			}
		case c == '-' && len(runes) >= i+2:
			if runes[i+1] == '-' {
				out.WriteString("&mdash;")
				i += 1
			} else {
				out.WriteRune(c)
			}
		default:
			out.WriteRune(c)
		}

	}
}

// Словарь для генерации уникальных идентификаторов.
const dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_$"

// Хеш-функция, используемая для генерации уникальных идентификаторов.
var hash = crc64.New(crc64.MakeTable(crc64.ECMA))

// hashSlug возвращает уникальный идентификатор, сгенерированный на основании указанного имени.
func hashSlug(name []byte) []byte {
	hash.Reset()
	hash.Write([]byte(name))
	val := hash.Sum64()
	var a = make([]byte, 8)
	for i := 0; i < 8; i++ {
		a[i] = dictionary[val&63]
		val >>= 8
	}
	return a
}
