package main

import (
	"fmt"
	"hash/crc32"
	"io"
	"strconv"

	"gopkg.in/russross/blackfriday.v2"
)

var extensions = blackfriday.WithExtensions(blackfriday.Footnotes |
	blackfriday.CommonExtensions |
	blackfriday.NoEmptyLineBeforeBlock)
var render = blackfriday.WithRenderer(&htmlRender{
	blackfriday.NewHTMLRenderer(
		blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags |
				blackfriday.SmartypantsAngledQuotes,
		}),
})

// Markdown преобразует данные из формата Markdown в HTML.
func Markdown(data []byte) []byte {
	return blackfriday.Run(data, extensions, render)
}

// htmlRender переопределяет преобразование некоторых видов информации из Markdwon
// в HTML. В тех случаях, когда переопределение не требуется, используется стандартный
// конвертер Markdown.
type htmlRender struct {
	*blackfriday.HTMLRenderer
}

func slug(text []byte) string {
	return strconv.FormatUint(uint64(crc32.ChecksumIEEE(text)), 36)
}

// RenderNode переопределяет формирование сносок. Все остальное обрабатывается
// стандартным способом.
func (r *htmlRender) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
	case blackfriday.Link:
		if node.NoteID == 0 || !entering {
			break
		}
		fmt.Fprintf(w,
			"<sup><a rel=\"footnote\" href=\"#fn:%s\" epub:type=\"noteref\" id=\"fnref:%[1]s\">%d</a></sup>",
			slug(node.Destination), node.LinkData.NoteID)
		return blackfriday.GoToNext
	case blackfriday.List:
		if node.IsFootnotesList {
			// if entering {
			// 	io.WriteString(w, "\n<section epub:type=\"endnotes\">\n<hr/>\n<ol>\n")
			// } else {
			// 	io.WriteString(w, "</ol>\n</section>")
			// }
			return blackfriday.GoToNext
		}
	case blackfriday.Item:
		if node.ListData.RefLink != nil {
			if entering {
				fmt.Fprintf(w, "<aside id=\"fn:%s\" epub:type=\"footnote\">",
					slug(node.ListData.RefLink))
			} else {
				fmt.Fprintf(w,
					" <a href=\"#fnref:%s\" class=\"reversefootnote\" hidden=\"hidden\">&#8617;</a>",
					slug(node.ListData.RefLink))
				io.WriteString(w, "</aside>\n")
			}
			return blackfriday.GoToNext
		}
	}
	return r.HTMLRenderer.RenderNode(w, node, entering)
}
