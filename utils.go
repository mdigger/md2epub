package main

import (
	"strings"
)

func isFilename(name string, list []string) bool {
	name = strings.ToLower(name)
	for _, item := range list {
		if item == name {
			return true
		}
	}
	return false
}
