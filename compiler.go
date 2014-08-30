package main

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"encoding/xml"
	"github.com/mdigger/bpool"
	"github.com/mdigger/epub3"
	"github.com/mdigger/metadata"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Списки имен предопределенных файлов для обработки.
var (
	metadataFiles = []string{"metadata.yaml", "metadata.yml", "metadata.json"}
	coverFiles    = []string{"cover.png", "cover.svg", "cover.jpeg", "cover.jpg", "cover.gif"}
	globalCSSFile = "global.css"
)

// Для отслеживания пустых строк между тегами.
var reMultiNewLines = regexp.MustCompile(`^\n{2,}$`)

// Компилятор публикации в формат EPUB.
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
	// Ищем файл с глобальными стилями
	var cssfile string
	if _, err := os.Stat(globalCSSFile); err == nil {
		cssfile = globalCSSFile
	}
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
		if finfo.IsDir() {
			// Полностью игнорируем каталоги, имя которых начинается с точки
			if filepath.Base(filename)[0] == '.' && len(filename) > 1 {
				return filepath.SkipDir
			}
			// Не обрабатываем отдельно каталоги
			return nil
		}
		// Игнорируем файлы, имя которых начинаются с точки
		if filepath.Base(filename)[0] == '.' {
			return nil
		}
		// Обрабатываем файлы в зависимости от расширения
		switch strings.ToLower(filepath.Ext(filename)) {
		case ".md", ".mdown", ".markdown": // Статья в формате Markdown: преобразуем и добавляем
			// Читаем файл и отделяем метаданные
			meta, data, err := metadata.ReadFile(filename)
			if err != nil {
				return err
			}
			// Определяем язык файла
			lang := meta.Lang()
			if lang == "" {
				lang = publang
			}
			meta["lang"] = lang
			// Вытаскиваем заголовок
			title := meta.Title()
			if title == "" {
				title = "* * *"
			}
			meta["title"] = title
			// Вычисляем, основной это текст или скрытый
			var ct epub.ContentType
			if meta.GetBool("hidden") {
				ct = epub.ContentTypeAuxiliary
			} else {
				ct = epub.ContentTypePrimary
			}
			// Добавляем глобальный стилевой файл публикации
			if cssfile != "" {
				meta["_globalcssfile_"] = cssfile
			}
			// Преобразуем из Markdown в HTML
			data = Markdown(data)
			// Разбираем получившийся HTML для последующей нормализации
			nodes, err := html.ParseFragment(bytes.NewReader(data), &html.Node{Type: html.ElementNode})
			if err != nil {
				return err
			}
			// Избавляемся от пустых строк между тегами и воссоздаем нормализованный XHTML.
			buf := bpool.Get()
			defer bpool.Put(buf)
			for _, node := range nodes {
				if node.Type == html.TextNode && reMultiNewLines.MatchString(node.Data) {
					buf.WriteByte('\n')
					continue
				}
				if err := html.Render(buf, node); err != nil {
					return err
				}
			}
			// Сохраняем получившийся HTML в том же самом описании метаданных, чтобы не плодить сущности
			meta["content"] = template.HTML(buf.String())
			buf.Reset() // Сбрасываем буфер
			// Осуществляем преобразование по шаблону для формирования полноценной страницы
			if err = templates.ExecuteTemplate(buf, "page", meta); err != nil {
				return err
			}
			// Изменяем расширение имени файла на .xhtml
			filename = filename[:len(filename)-len(filepath.Ext(filename))] + ".xhtml"
			// Получаем io.Writer для записи содержимого файла
			fileWriter, err := writer.Add(filename, ct)
			if err != nil {
				return err
			}
			// Добавляем в начало документа XML-заголовок
			if _, err := io.WriteString(fileWriter, xml.Header); err != nil {
				return err
			}
			// Записываем данные
			if _, err := buf.WriteTo(fileWriter); err != nil {
				return err
			}
			// Добавляем информацию о файле в оглавление
			nav = append(nav, &NavigationItem{
				Title:       title,
				Subtitle:    meta.Subtitle(),
				Filename:    filename,
				Level:       meta.GetInt("level"),
				ContentType: ct,
			})
			// Выводим информацию о файле
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
				// Обложка публикации
				properties = []string{"cover-image"}
				setCover = true
			}
			// Добавляем файл в публикацию
			if err := writer.AddFile(filename, filename, epub.ContentTypeMedia, properties...); err != nil {
				return err
			}
			// Выводим информацию о добавленном файле
			if properties == nil {
				properties = []string{"media"}
			}
			log.Printf("Add %s\t%q", filename, strings.Join(properties, ", "))

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
	tdata := map[string]interface{}{
		"lang":  publang,
		"title": "Оглавление",
		"toc":   nav,
	}
	if cssfile != "" {
		tdata["_globalcssfile_"] = cssfile
	}
	if err = templates.ExecuteTemplate(fileWriter, "toc", tdata); err != nil {
		return err
	}
	log.Printf("Generate %s %q", "_toc.xhtml", "nav")

	return nil
}
