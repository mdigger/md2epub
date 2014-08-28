package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(0)
	// Разбираем входящие параметры
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	sourcePath := flag.Arg(0)
	var outputFilename string // Имя результирующего файла с публикацией
	if flag.NArg() > 1 {
		outputFilename = flag.Arg(1)
	} else {
		outputFilename = filepath.Base(sourcePath) + ".epub"
	}
	if err := compiler(sourcePath, outputFilename); err != nil {
		log.Fatal(err)
	}
}
