// The Blackfriday Markdown HTML Render with EPUB support
package markdown

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/russross/blackfriday"
	"html"
	"strconv"
	"strings"
)

var (
	Extensions = 0
)

func init() {
	Extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	Extensions |= blackfriday.EXTENSION_TABLES
	Extensions |= blackfriday.EXTENSION_FENCED_CODE
	Extensions |= blackfriday.EXTENSION_AUTOLINK
	Extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	Extensions |= blackfriday.EXTENSION_LAX_HTML_BLOCKS
	Extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	// Extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	// Extensions |= blackfriday.EXTENSION_TAB_SIZE_EIGHT
	Extensions |= blackfriday.EXTENSION_FOOTNOTES
	Extensions |= blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK
	Extensions |= blackfriday.EXTENSION_HEADER_IDS
	Extensions |= blackfriday.EXTENSION_TITLEBLOCK
}

type htmlRender struct {
	lang         string
	title        string
	csslink      string
	headerCount  int
	currentLevel int
	toc          *bytes.Buffer
	smartypants  *smartypantsRenderer
}

func NewRender(lang, title, csslink string) *htmlRender {
	return &htmlRender{
		lang:         lang,
		title:        title,
		csslink:      csslink,
		headerCount:  0,
		currentLevel: 0,
		toc:          new(bytes.Buffer),
		smartypants:  smartypants(),
	}
}

func (self *htmlRender) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	count := 0
	for _, elt := range strings.Fields(lang) {
		if elt[0] == '.' {
			elt = elt[1:]
		}
		if len(elt) == 0 {
			continue
		}
		out.WriteString("<pre><code class=\"language-")
		out.WriteString(html.EscapeString(elt))
		out.WriteString("\">")
		count++
		break
	}
	if count == 0 {
		out.WriteString("<pre><code>")
	}
	out.WriteString(html.EscapeString(string(text)))
	out.WriteString("</code></pre>\n")
}

func (self *htmlRender) BlockQuote(out *bytes.Buffer, text []byte) {
	out.WriteString("<blockquote>\n")
	out.Write(text)
	out.WriteString("</blockquote>\n")
}

func (self *htmlRender) BlockHtml(out *bytes.Buffer, text []byte) {
	out.Write(text)
	out.WriteByte('\n')
}

func (self *htmlRender) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	marker := out.Len()
	if id != "" {
		out.WriteString(fmt.Sprintf("<h%d id=\"%s\">", level, id))
	} else {
		// headerCount is incremented in htmlTocHeader
		out.WriteString(fmt.Sprintf("<h%d id=\"toc_%d\">", level, self.headerCount))
	}
	tocMarker := out.Len()
	if !text() {
		out.Truncate(marker)
		return
	}
	self.TocHeader(out.Bytes()[tocMarker:], level)
	out.WriteString(fmt.Sprintf("</h%d>\n", level))
}

func (self *htmlRender) HRule(out *bytes.Buffer) {
	out.WriteString("<hr />\n")
}

func (self *htmlRender) List(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	if flags&blackfriday.LIST_TYPE_ORDERED != 0 {
		out.WriteString("<ol>\n")
	} else {
		out.WriteString("<ul>\n")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	if flags&blackfriday.LIST_TYPE_ORDERED != 0 {
		out.WriteString("</ol>\n")
	} else {
		out.WriteString("</ul>\n")
	}
}

func (self *htmlRender) ListItem(out *bytes.Buffer, text []byte, flags int) {
	out.WriteString("<li>")
	out.Write(text)
	out.WriteString("</li>\n")
}

func (self *htmlRender) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()
	out.WriteString("<p>")
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("</p>\n")
}

func (self *htmlRender) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	out.WriteString("<table>\n<thead>\n")
	out.Write(header)
	out.WriteString("</thead>\n<tbody>\n")
	out.Write(body)
	out.WriteString("</tbody>\n</table>\n")
}

func (self *htmlRender) TableRow(out *bytes.Buffer, text []byte) {
	out.WriteString("<tr>\n")
	out.Write(text)
	out.WriteString("\n</tr>\n")
}

func (self *htmlRender) TableHeaderCell(out *bytes.Buffer, text []byte, flags int) {
	switch flags {
	case blackfriday.TABLE_ALIGNMENT_LEFT:
		out.WriteString("<th align=\"left\">")
	case blackfriday.TABLE_ALIGNMENT_RIGHT:
		out.WriteString("<th align=\"right\">")
	case blackfriday.TABLE_ALIGNMENT_CENTER:
		out.WriteString("<th align=\"center\">")
	default:
		out.WriteString("<th>")
	}
	out.Write(text)
	out.WriteString("</th>")
}

func (self *htmlRender) TableCell(out *bytes.Buffer, text []byte, align int) {
	switch align {
	case blackfriday.TABLE_ALIGNMENT_LEFT:
		out.WriteString("<td align=\"left\">")
	case blackfriday.TABLE_ALIGNMENT_RIGHT:
		out.WriteString("<td align=\"right\">")
	case blackfriday.TABLE_ALIGNMENT_CENTER:
		out.WriteString("<td align=\"center\">")
	default:
		out.WriteString("<td>")
	}
	out.Write(text)
	out.WriteString("</td>")
}

func (self *htmlRender) Footnotes(out *bytes.Buffer, text func() bool) {
	text()
}

func (self *htmlRender) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	out.WriteString("<aside id=\"fn:")
	out.Write(slugify(name))
	out.WriteString("\" epub:type=\"footnote\">\n")
	out.Write(text)
	out.WriteString("</aside>\n")
}

func (self *htmlRender) TitleBlock(out *bytes.Buffer, text []byte) {
	text = bytes.TrimPrefix(text, []byte("% "))
	text = bytes.Replace(text, []byte("\n% "), []byte("\n"), -1)
	out.WriteString("<h1 class=\"title\">")
	out.Write(text)
	out.WriteString("\n</h1>")
}

func (self *htmlRender) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	// skipRanges := htmlEntity.FindAllIndex(link, -1)
	out.WriteString("<a href=\"")
	if kind == blackfriday.LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}
	// entityEscapeWithSkip(out, link, skipRanges)
	out.WriteString(html.EscapeString(string(link)))
	out.WriteString("\">")
	switch {
	case bytes.HasPrefix(link, []byte("mailto://")):
		out.WriteString(html.EscapeString(string(link[len("mailto://"):])))
	case bytes.HasPrefix(link, []byte("mailto:")):
		out.WriteString(html.EscapeString(string(link[len("mailto:"):])))
	default:
		out.WriteString(html.EscapeString(string(link)))
	}
	out.WriteString("</a>")
}

func (self *htmlRender) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("<code>")
	out.WriteString(html.EscapeString(string(text)))
	out.WriteString("</code>")
}

func (self *htmlRender) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong>")
	out.Write(text)
	out.WriteString("</strong>")
}

func (self *htmlRender) Emphasis(out *bytes.Buffer, text []byte) {
	if len(text) == 0 {
		return
	}
	out.WriteString("<em>")
	out.Write(text)
	out.WriteString("</em>")
}

func (self *htmlRender) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	out.WriteString("<img src=\"")
	out.WriteString(html.EscapeString(string(link)))
	out.WriteString("\" alt=\"")
	if len(alt) > 0 {
		out.WriteString(html.EscapeString(string(alt)))
	}
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		out.WriteString(html.EscapeString(string(title)))
	}
	out.WriteString("\" />")
}

func (self *htmlRender) LineBreak(out *bytes.Buffer) {
	out.WriteString("<br />")
}

func (self *htmlRender) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteString("<a href=\"")
	out.WriteString(html.EscapeString(string(link)))
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		out.WriteString(html.EscapeString(string(title)))
	}
	out.WriteString("\">")
	out.Write(content)
	out.WriteString("</a>")
}

func (self *htmlRender) RawHtmlTag(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (self *htmlRender) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong><em>")
	out.Write(text)
	out.WriteString("</em></strong>")
}

func (self *htmlRender) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.WriteString("<del>")
	out.Write(text)
	out.WriteString("</del>")
}

func (self *htmlRender) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	slug := slugify(ref)
	out.WriteString(`<sup><a rel="footnote" href="#fn:`)
	out.Write(slug)
	out.WriteString(`" epub:type="noteref">`)
	out.WriteString(strconv.Itoa(id))
	out.WriteString(`</a></sup>`)
}

func (self *htmlRender) Entity(out *bytes.Buffer, entity []byte) {
	if len(entity) > 3 {
		switch ent := string(entity[1 : len(entity)-1]); ent {
		case "amp", "lt", "gt", "quot", "apos":
			out.Write(entity)
			return
		default:
			if sym, ok := xml.HTMLEntity[ent]; ok {
				out.WriteString(sym)
				return
			}
		}
	}
	out.WriteString(html.EscapeString(string(entity)))
}

func (self *htmlRender) NormalText(out *bytes.Buffer, text []byte) {
	// out.WriteString(html.EscapeString(string(text)))
	smrt := smartypantsData{false, false}
	text = []byte(html.EscapeString(string(text)))
	mark := 0
	for i := 0; i < len(text); i++ {
		if action := self.smartypants[text[i]]; action != nil {
			if i > mark {
				out.Write(text[mark:i])
			}
			previousChar := byte(0)
			if i > 0 {
				previousChar = text[i-1]
			}
			i += action(out, &smrt, previousChar, text[i:])
			mark = i + 1
		}
	}
	if mark < len(text) {
		out.Write(text[mark:])
	}
}

func (self *htmlRender) DocumentHeader(out *bytes.Buffer) {
	out.WriteString("<!DOCTYPE html>\n")
	out.WriteString("<html xmlns=\"http://www.w3.org/1999/xhtml\" xmlns:epub=\"http://www.idpf.org/2007/ops\"")
	if self.lang != "" {
		out.WriteString(" xml:lang=\"")
		out.WriteString(self.lang)
		out.WriteString("\" lang=\"")
		out.WriteString(self.lang)
		out.WriteRune('"')
	}
	out.WriteString(">\n")
	out.WriteString("<head>\n")
	out.WriteString("  <title>")
	self.NormalText(out, []byte(self.title))
	out.WriteString("</title>\n")
	out.WriteString("  <meta charset=\"utf-8\" />\n")
	if self.csslink != "" {
		out.WriteString("  <link rel=\"stylesheet\" type=\"text/css\" href=\"")
		out.WriteString(html.EscapeString(self.csslink))
		out.WriteString("\" />")
	}
	out.WriteString("</head>\n")
	out.WriteString("<body>\n")
}

func (self *htmlRender) DocumentFooter(out *bytes.Buffer) {
	out.WriteString("</body>\n")
	out.WriteString("</html>\n")
}

func (self *htmlRender) GetFlags() int {
	return 0
}

func (self *htmlRender) TocHeader(text []byte, level int) {
	for level > self.currentLevel {
		switch {
		case bytes.HasSuffix(self.toc.Bytes(), []byte("</li>\n")):
			// this sublist can nest underneath a header
			size := self.toc.Len()
			self.toc.Truncate(size - len("</li>\n"))
		case self.currentLevel > 0:
			self.toc.WriteString("<li>")
		}
		if self.toc.Len() > 0 {
			self.toc.WriteByte('\n')
		}
		self.toc.WriteString("<ul>\n")
		self.currentLevel++
	}

	for level < self.currentLevel {
		self.toc.WriteString("</ul>")
		if self.currentLevel > 1 {
			self.toc.WriteString("</li>\n")
		}
		self.currentLevel--
	}
	self.toc.WriteString("<li><a href=\"#toc_")
	self.toc.WriteString(strconv.Itoa(self.headerCount))
	self.toc.WriteString("\">")
	self.headerCount++
	self.toc.Write(text)
	self.toc.WriteString("</a></li>\n")
}

// func (self *htmlRender) TocFinalize() {
// 	for self.currentLevel > 1 {
// 		self.toc.WriteString("</ul></li>\n")
// 		self.currentLevel--
// 	}
// 	if self.currentLevel > 0 {
// 		self.toc.WriteString("</ul>\n")
// 	}
// }
