package main

import (
	"strings"
)

// isFilename возвращает true, если имя файла совпадает с одним из указанных
// в переданном списке. В принципе, можно использовать и списки расширений файлов.
func isFilename(name string, list []string) bool {
	name = strings.ToLower(name)
	for _, item := range list {
		if item == name {
			return true
		}
	}
	return false
}
