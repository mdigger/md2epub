package main

type NavigationItem struct {
	Title    string
	Subtitle string
	Filename string
	Spine    bool
}

type Navigaton []*NavigationItem
