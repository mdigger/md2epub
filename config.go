package main

// Config описывает конфигурацию для публикации.
type Config struct {
	Lang     string   // Язык публикации по умолчанию
	Title    string   // Название публикации по умолчанию
	Metadata []string // Список имен файлов с метаинформацией
	Markdown []string // Список расширений файлов в формате Markdown
	Covers   []string // Список имен файлов с обложкой
	CSSFile  string   // Имя файла со стилем
}

// DefaultConfig описывает используемую по умолчанию конфигурацию.
var DefaultConfig = &Config{
	Lang:     "en",
	Title:    "Untitle",
	Metadata: []string{"metadata.yaml", "metadata.yml", "metadata.json"},
	Markdown: []string{".md", ".mdown", ",markdown"},
	Covers:   []string{"cover.png", "cover.svg", "cover.jpeg", "cover.jpg", "cover.gif"},
	CSSFile:  "style.css",
}
