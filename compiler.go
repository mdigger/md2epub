package main

import (
	"encoding/xml"
	"github.com/mdigger/epub3"
	"github.com/mdigger/md2epub/markdown"
	"github.com/mdigger/metadata"
	"github.com/russross/blackfriday"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	metadataFiles = []string{"metadata.yaml", "metadata.yml", "metadata.json"}
	coverFiles    = []string{"cover.png", "cover.svg", "cover.jpeg", "cover.jpg", "cover.gif"}
)

// Компилятор публикации
func compiler(sourcePath, outputFilename string) error {
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
	// Загружаем описание метаданных публикации
	var pubMetadata *epub.Metadata
	for _, name := range metadataFiles {
		fi, err := os.Stat(name)
		if err != nil || fi.IsDir() {
			continue
		}
		if pubMetadata, err = loadMetadata(name); err != nil {
			return err
		}
		break
	}
	// Если описания не найдено, то инициализируем пустое описание
	if pubMetadata == nil {
		pubMetadata = defaultMetada()
	}
	// Добавляем язык, если он не определен
	if len(pubMetadata.Language) == 0 {
		pubMetadata.Language.Add("", defaultLang)
	}
	// Вытаскиваем язык публикации
	publang := pubMetadata.Language[0].Value
	// Создаем упаковщик в формат EPUB
	writer, err := epub.Create(filepath.Join(currentPath, outputFilename))
	if err != nil {
		return err
	}
	defer writer.Close()
	// Добавляем метаданные в публикацию
	writer.Metadata = pubMetadata
	// Оглавление
	nav := make(Navigaton, 0)
	// Флаг для избежания двойной обработки обложки: после его установки
	// новые попадающиеся обложки игнорируются.
	var setCover bool
	// Определяем функция для обработки перебора файлов и каталогов
	walkFn := func(filename string, finfo os.FileInfo, err error) error {
		// Игнорируем, если открытие файла произошло с ошибкой
		if err != nil {
			return nil
		}
		// Не обрабатываем отдельно каталоги
		if finfo.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(filename)) {
		case ".md", ".mdown", ".markdown": // Статья в формате Markdown: преобразуем и добавляем
			// Читаем файл и отделяем метаданные
			meta, data, err := metadata.ReadFile(filename)
			if err != nil {
				return err
			}
			lang := meta.Lang()
			if lang == "" {
				lang = publang
			}
			title := meta.Title()
			if title == "" {
				title = "* * *"
			}
			// Инициализируем HTML-преобразователь из формата Markdown
			mdRender := markdown.NewRender(lang, title, "")
			// Преобразуем из Markdown в HTML
			data = blackfriday.Markdown(data, mdRender, markdown.Extensions)
			// Изменяем расширение имени файла на .xhtml
			filename = filename[:len(filename)-len(filepath.Ext(filename))] + ".xhtml"
			// Добавляем в основной список чтения, если имя файла не начинается с подчеркивания
			spine := filepath.Base(filename)[0] != '_'
			fileWriter, err := writer.Add(filename, spine)
			if err != nil {
				return err
			}
			// Добавляем в начало документа XML-заголовок
			if _, err := io.WriteString(fileWriter, xml.Header); err != nil {
				return err
			}
			if _, err := fileWriter.Write(data); err != nil {
				return err
			}
			// Добавляем информацию о файле в оглавление
			nav = append(nav, &NavigationItem{
				Title:    title,
				Subtitle: meta.Subtitle(),
				Filename: filename,
				Spine:    spine,
			})
			log.Printf("Add %s %q", filename, title)

		case ".jpg", ".jpe", ".jpeg", ".png", ".gif", ".svg",
			".mp3", ".mp4", ".aac", ".m4a", ".m4v", ".m4b", ".m4p", ".m4r",
			".css", ".js", ".javascript",
			".json",
			".otf", ".woff",
			".pls", ".smil", ".smi", ".sml": // Иллюстрации и другие файлы
			var properties []string
			// Специальная обработка
			switch {
			case !setCover && isFilename(filename, coverFiles):
				// Обложка
				properties = []string{"cover-image"}
				setCover = true
			}
			if err := writer.AddFile(filename, filename, false, properties...); err != nil {
				return err
			}
			if properties != nil {
				log.Printf("Add %s\t%q", filename, strings.Join(properties, ", "))
			} else {
				log.Printf("Add %s", filename)
			}

		default: // Другое — игнорируем
			if !isFilename(filename, metadataFiles) {
				log.Printf("Ignore %s", filename)
			}
		}

		return nil
	}
	// Перебираем все файлы и подкаталоги в исходном каталоге
	if err := filepath.Walk(".", walkFn); err != nil {
		return err
	}
	// Добавляем оглавление
	fileWriter, err := writer.Add("toc.xhtml", false, "nav")
	if err != nil {
		return err
	}
	// Добавляем в начало документа XML-заголовок
	if _, err := io.WriteString(fileWriter, xml.Header); err != nil {
		return err
	}
	// Преобразуем по шаблону и записываем в публикацию.
	err = tnav.Execute(fileWriter, map[string]interface{}{
		"lang":  publang,
		"title": "Оглавление",
		"toc":   nav,
	})
	if err != nil {
		return err
	}
	log.Printf("Generate %s %q", "toc.xhtml", "nav")

	return nil
}
