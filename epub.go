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
	"os"
	"path/filepath"
	"regexp"
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
		config:    config,
		writer:    writer,
		templates: templates,
		lang:      pubmeta.Language[0].Value, // Язык публикации
		nav:       make(Navigaton, 0),
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
		if err = pub.templates.ExecuteTemplate(fileWriter, "toc", tdata); err != nil {
			return err
		}
	}
	return nil
}

type EPUBCompiler struct {
	config    *Config            // Конфигурация параметров по умолчанию
	writer    *epub.Writer       // EPUB
	templates *template.Template // Шаблоны преобразования
	setCover  bool               // Флаг, что обложка уже добавлена
	setToc    bool               // Флаг, что файл с оглавлением уже добавлен
	cssfile   string             // Имя файла со стилем
	lang      string             // Язык публикации
	nav       Navigaton          // Оглавление
}

// walk вызывается на каждый файл и каталог в исходных данных.
func (pub *EPUBCompiler) walk(filename string, finfo os.FileInfo, err error) error {
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
	if ch := filepath.Base(filename)[0]; ch == '.' || ch == '~' {
		return nil
	}
	// Игнорируем описание метаданных публикации, т.к. уже разобрали его
	if isFilename(filename, pub.config.Metadata) {
		return nil
	}
	// Обрабатываем файлы в зависимости от расширения
	switch ext := filepath.Ext(filename); {
	case isFilename(ext, pub.config.Markdown):
		return pub.addMarkdown(filename)
	default:
		return pub.addMedia(filename)
	}
}

var reMultiNewLines = regexp.MustCompile(`^\n{2,}$`)

// addMarkdown добавляет Markdown файл в публикацию.
func (pub *EPUBCompiler) addMarkdown(filename string) error {
	// Читаем файл и отделяем метаданные
	meta, data, err := metadata.ReadFile(filename)
	if err != nil {
		return err
	}
	// Определяем язык файла
	lang := meta.Lang()
	if lang == "" {
		lang = pub.lang
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
	if pub.cssfile != "" {
		if rel, err := filepath.Rel(filepath.Dir(filename), pub.cssfile); err == nil {
			meta["_globalcssfile_"] = filepath.ToSlash(rel)
		} else {
			return err
		}
	}
	// Преобразуем из Markdown в HTML
	data = Markdown(data)
	// Разбираем получившийся HTML для последующей нормализации
	nodes, err := html.ParseFragment(bytes.NewReader(data), &html.Node{Type: html.ElementNode})
	if err != nil {
		return err
	}
	// Инициализируем внутренний пул для работы с информацией
	buf := bpool.Get()
	defer bpool.Put(buf)
	// Избавляемся от пустых строк между тегами и воссоздаем нормализованный XHTML
	for _, node := range nodes {
		if node.Type == html.TextNode && reMultiNewLines.MatchString(node.Data) {
			buf.WriteByte('\n')
			continue
		}
		// TODO: Убрать пустые строки во вложенных элементах
		if err := html.Render(buf, node); err != nil {
			return err
		}
	}
	// Сохраняем получившийся HTML в том же самом описании метаданных, чтобы не плодить сущности
	meta["content"] = template.HTML(buf.String())
	buf.Reset() // Сбрасываем буфер
	// Избавляемся от расширения файла
	filename = filename[:len(filename)-len(filepath.Ext(filename))]
	templateName := "page" // Название шаблона для преобразования
	properties := meta.GetQuickList("properties")
	for i, property := range properties {
		switch property {
		case "nav":
			templateName = "nav"
			pub.setToc = true // Файл с заголовком добавлен
		case "cover-image":
			properties[i] = "cover" // Смухлюем и поправим недопустимое
		}
	}
	// Осуществляем преобразование по шаблону для формирования полноценной страницы
	if err = pub.templates.ExecuteTemplate(buf, templateName, meta); err != nil {
		return err
	}
	// Добавляем расширение имени файла .xhtml
	filename += ".xhtml"
	// Добавляем информацию о файле в оглавление
	pub.nav = append(pub.nav, &NavigationItem{
		Title:       title,
		Subtitle:    meta.Subtitle(),
		Filename:    filename,
		Level:       meta.GetInt("level"),
		ContentType: ct,
	})
	// Получаем io.Writer для записи содержимого файла
	fileWriter, err := pub.writer.Add(filename, ct, properties...)
	if err != nil {
		return err
	}
	// Добавляем в начало документа XML-заголовок
	if _, err := io.WriteString(fileWriter, xml.Header); err != nil {
		return err
	}
	// Записываем данные
	_, err = buf.WriteTo(fileWriter)
	return err
}

func (pub *EPUBCompiler) addMedia(filename string) error {
	var properties []string
	switch {
	case !pub.setCover && isFilename(filename, pub.config.Covers):
		// Обложка публикации
		properties = []string{"cover-image"}
		pub.setCover = true // Обрабатываем только одну обложку
	}
	// Добавляем файл в публикацию
	return pub.writer.AddFile(filename, filename, epub.ContentTypeMedia, properties...)
}

// NavigationItem описывает ссылку из оглавления на файл
type NavigationItem struct {
	Title       string           // Заголовок
	Subtitle    string           // Подзаголовок
	Level       int              // Уровень заголовка
	Filename    string           // Имя файла
	ContentType epub.ContentType // Тип файла
}

// Navigaton описывает оглавление публикации
type Navigaton []*NavigationItem
