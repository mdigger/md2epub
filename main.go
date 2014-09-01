package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Разбираем входящие параметры
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	// Путь к каталогу с исходниками публикации
	sourcePath := flag.Arg(0)
	var outputFilename string // Имя результирующего файла с публикацией
	if flag.NArg() > 1 {
		outputFilename = flag.Arg(1)
	} else {
		// Убираем слеш в конце пути, если он там указан
		if sourcePath[len(sourcePath)-1] == '/' {
			sourcePath = sourcePath[:len(sourcePath)-1]
		}
		// Добавляем расширение файла публикации
		outputFilename = sourcePath + ".epub"
	}
	// Запускаем компиляцию исходников
	if err := Compile(sourcePath, outputFilename, DefaultConfig); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
