package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/mdigger/epub3"
	"github.com/mdigger/metadata"
	"github.com/mdigger/uuid"
	"gopkg.in/yaml.v2"
)

// loadMetadata загружает или создает описание публикации.
func loadMetadata(config *Config) (*epub.Metadata, error) {
	// Инициализируем описание метаданных
	var pubmeta = &epub.Metadata{
		DC:   "http://purl.org/dc/elements/1.1/",
		Meta: make([]*epub.Meta, 0),
	}
	// Загружаем описание метаданных публикации
	for _, name := range config.Metadata {
		fi, err := os.Stat(name)
		if err != nil || fi.IsDir() {
			continue
		}
		// Читаем файл с описанием метаданных публикации
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}
		// Разбираем метаданные
		var metadata = make(metadata.Metadata)
		if err := yaml.Unmarshal(data, metadata); err != nil {
			return nil, err
		}
		// Переводим описание метаданных в метаданные публикации
		convertMetadata(metadata, pubmeta)
		break
	}
	// Устанавливаем язык, если его нет
	if len(pubmeta.Language) == 0 {
		pubmeta.Language.Add("", config.Lang)
	}
	// Добавляем заголовок, если его нет
	if len(pubmeta.Title) == 0 {
		pubmeta.Title.Add("", config.Title)
	}
	// Добавляем уникальный идентификатор, если его нет
	if len(pubmeta.Identifier) == 0 {
		pubmeta.Identifier.Add("uuid", "urn:uuid:"+uuid.New().String())
	}
	return pubmeta, nil
}

// convertMetadata конвертирует описание метаданных в формат метаданных публикации.
func convertMetadata(metadata metadata.Metadata, pubmeta *epub.Metadata) {
	// Добавляем язык
	if lang := metadata.Lang(); lang != "" {
		pubmeta.Language.Add("", lang)
	}
	// Добавляем заголовок
	if title := metadata.Title(); title != "" {
		pubmeta.Title.Add("title", title)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#title",
			Property: "title-type",
			Value:    "main",
		}, &epub.Meta{
			Refines:  "#title",
			Property: "display-seq",
			Value:    "1",
		})
	}
	// Добавляем подзаголовок
	if subtitle := metadata.Subtitle(); subtitle != "" {
		pubmeta.Title.Add("subtitle", subtitle)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#subtitle",
			Property: "title-type",
			Value:    "subtitle",
		}, &epub.Meta{
			Refines:  "#subtitle",
			Property: "display-seq",
			Value:    "2",
		})
	}
	// Добавляем название коллекции
	if collection := metadata.Get("collection"); collection != "" {
		pubmeta.Title.Add("collection", collection)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#collection",
			Property: "title-type",
			Value:    "collection",
		}, &epub.Meta{
			ID:       "collectionid",
			Property: "belongs-to-collection",
			Value:    collection,
		})
		// Добавляем порядковый номер в коллекции, если он есть
		if collectionNumber := metadata.Get("sequence"); collectionNumber != "" {
			pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
				Refines:  "#collection",
				Property: "group-position",
				Value:    collectionNumber,
			}, &epub.Meta{
				Refines:  "#collectionid",
				Property: "group-position",
				Value:    collectionNumber,
			})
		}
	}
	// Добавляем название редакции
	if edition := metadata.Get("edition"); edition != "" {
		pubmeta.Title.Add("edition", edition)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#edition",
			Property: "title-type",
			Value:    "edition",
		})
	}
	// Добавляем расширенное название публикации, если оно есть
	if fulltitle := metadata.Get("fulltitle"); fulltitle != "" {
		pubmeta.Title.Add("fulltitle", fulltitle)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#fulltitle",
			Property: "title-type",
			Value:    "expanded",
		})
	}
	// Добавляем авторов
	for _, author := range metadata.Authors() {
		pubmeta.Creator.Add("", author)
	}
	// Добавляем второстепенных авторов
	for _, author := range metadata.GetList("contributor") {
		pubmeta.Contributor.Add("", author)
	}
	// Добавляем информацию об издателях
	for _, author := range metadata.GetList("publisher") {
		pubmeta.Publisher.Add("", author)
	}
	// Добавляем уникальные идентификаторы
	for _, name := range []string{"uuid", "id", "identifier", "doi", "isbn", "issn"} {
		if value := metadata.Get(name); value != "" {
			switch name {
			case "uuid", "doi", "isbn", "issn":
				value = "urn:" + name + ":" + value
				// TODO: Добавить префиксы для других идентификаторов
			}
			pubmeta.Identifier.Add(name, value)
		}
	}
	// Добавляем краткое описание
	if description := metadata.Description(); description != "" {
		// Убираем новые строки и множественные пробелы
		description = strings.Join(strings.Fields(description), " ")
		pubmeta.Description.Add("description", description)

	}
	// Добавляем ключевые слова
	for _, keyword := range metadata.Keywords() {
		pubmeta.Subject.Add("", keyword)
	}
	// Добавляем описание сферы действия
	if coverage := metadata.Get("coverage"); coverage != "" {
		pubmeta.Coverage.Add("", coverage)
	}
	// Добавляем дату
	if date := metadata.Get("date"); date != "" {
		pubmeta.Date = &epub.Element{Value: date}
	}
	// Добавляем копирайты
	for _, name := range []string{"copyright", "rights"} {
		if rights := metadata.Get(name); rights != "" {
			pubmeta.Rights.Add(name, rights)
		}
	}
}
