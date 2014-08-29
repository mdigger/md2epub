package main

import (
	"github.com/mdigger/epub3"
)

type NavigationItem struct {
	Title       string
	Subtitle    string
	Filename    string
	ContentType epub.ContentType
}

type Navigaton []*NavigationItem
