package main

import (
	"encoding/xml"
	"github.com/mdigger/epub3"
	"github.com/mdigger/metadata"
	"io"
	"os"
	"path/filepath"
)

func Compile(sourcePath, outputFilename string, config *Config) error {
	// Делаем исходный каталог текущим, чтобы не вычислять относительный путь. По окончании
	// обработки восстанавливаем исходный каталог обратно.
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(sourcePath); err != nil {
		return err
	}
	defer os.Chdir(currentPath)
	// Загружаем и разбираем метаданные публикации
	pubmeta, err := loadMetadata(config)
	if err != nil {
		return err
	}
	// Создаем упаковщик в формат EPUB
	writer, err := epub.Create(filepath.Join(currentPath, outputFilename))
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.Metadata = pubmeta
	// Инициализируем компилятор
	pub := &EPUBCompiler{
		config: config,
		writer: writer,
		lang:   pubmeta.Language[0].Value, // Язык публикации
		nav:    make(Navigaton, 0),
	}
	// Ищем файл со стилем
	if _, err := os.Stat(config.CSSFile); err == nil {
		pub.cssfile = config.CSSFile
	}
	// Перебираем все файлы и подкаталоги в исходном каталоге
	if err := filepath.Walk(".", pub.walk); err != nil {
		return err
	}
	// Генерируем оглавление, если его не добавили в виде файла
	if !pub.setToc {
		// Добавляем оглавление как скрытый (вспомогательный) файл
		fileWriter, err := writer.Add("_toc.xhtml", epub.ContentTypeAuxiliary, "nav")
		if err != nil {
			return err
		}
		// Добавляем в начало документа XML-заголовок
		if _, err := io.WriteString(fileWriter, xml.Header); err != nil {
			return err
		}
		// Преобразуем по шаблону и записываем в публикацию.
		tdata := metadata.Metadata{
			"lang":  pub.lang,
			"title": "Оглавление",
			"toc":   pub.nav,
		}
		// Добавляем ссылку на стилевой файл, если он определен
		if pub.cssfile != "" {
			// Здесь не нужен относительный путь, т.к. они на одном уровне
			tdata["_globalcssfile_"] = pub.cssfile
		}
		// Преобразуем по шаблону
		if err = templates.ExecuteTemplate(fileWriter, "toc", tdata); err != nil {
			return err
		}
	}
	return nil
}
