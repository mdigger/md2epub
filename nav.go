package main

import (
	"github.com/mdigger/epub3"
	"html/template"
)

type NavigationItem struct {
	Title       string
	Subtitle    string
	Filename    string
	ContentType epub.ContentType
}

type Navigaton []*NavigationItem

type TOCItem struct {
	ID    string
	Text  template.HTML
	Level int
}
