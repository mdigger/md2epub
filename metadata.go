package main

import (
	"fmt"
	"github.com/mdigger/epub3"
	"github.com/mdigger/metadata"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"
)

const defaultLang = "en"

func defaultMetada() *epub.Metadata {
	return &epub.Metadata{
		DC:   "http://purl.org/dc/elements/1.1/",
		Meta: make([]*epub.Meta, 0),
	}
}
func loadMetadata(name string) (pubmeta *epub.Metadata, err error) {
	// Читаем файл с описанием метаданных публикации
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	// Разбираем метаданные
	meta := make(metadata.Metadata)
	if err := yaml.Unmarshal(data, meta); err != nil {
		return nil, err
	}
	// Инициализируем описание метаданных
	pubmeta = defaultMetada()
	// Инициализируем вывод
	tab := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
	fmt.Fprintln(tab, strings.Repeat("—", 80))
	// Добавляем язык
	lang := meta.Lang()
	if lang == "" {
		lang = defaultLang
	}
	pubmeta.Language.Add("", lang)
	fmt.Fprintf(tab, "Lang:\t%s\n", lang)
	// Добавляем заголовок
	if title := meta.Title(); title != "" {
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
		fmt.Fprintf(tab, "Title:\t%s\n", title)
	}
	// Добавляем подзаголовок
	if subtitle := meta.Subtitle(); subtitle != "" {
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
		fmt.Fprintf(tab, "Subtitle:\t%s\n", subtitle)
	}
	// Добавляем название коллекции
	if collection := meta.Get("collection"); collection != "" {
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
		if collectionNumber := meta.Get("sequence"); collectionNumber != "" {
			pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
				Refines:  "#collection",
				Property: "group-position",
				Value:    collectionNumber,
			}, &epub.Meta{
				Refines:  "#collectionid",
				Property: "group-position",
				Value:    collectionNumber,
			})
			fmt.Fprintf(tab, "Collection:\t%s (#%s)\n", collection, collectionNumber)
		} else {
			fmt.Fprintf(tab, "Collection:\t%s\n", collection)
		}
	}
	// Добавляем название редакции
	if edition := meta.Get("edition"); edition != "" {
		pubmeta.Title.Add("edition", edition)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#edition",
			Property: "title-type",
			Value:    "edition",
		})
		fmt.Fprintf(tab, "Edition:\t%s\n", edition)
	}
	// Добавляем расширенное название публикации, если оно есть
	if fulltitle := meta.Get("fulltitle"); fulltitle != "" {
		pubmeta.Title.Add("fulltitle", fulltitle)
		pubmeta.Meta = append(pubmeta.Meta, &epub.Meta{
			Refines:  "#fulltitle",
			Property: "title-type",
			Value:    "expanded",
		})
		fmt.Fprintf(tab, "Full title:\t%s\n", fulltitle)
	}
	// Добавляем авторов
	for i, author := range meta.Authors() {
		pubmeta.Creator.Add("", author)
		if i == 0 {
			fmt.Fprintf(tab, "Author:\t%s\n", author)
		} else {
			fmt.Fprintf(tab, "\t%s\n", author)
		}
	}
	// Добавляем второстепенных авторов
	for i, author := range meta.GetList("contributor") {
		pubmeta.Contributor.Add("", author)
		if i == 0 {
			fmt.Fprintf(tab, "Contributor:\t%s\n", author)
		} else {
			fmt.Fprintf(tab, "\t%s\n", author)
		}
	}
	// Добавляем информацию об издателях
	for i, author := range meta.GetList("publisher") {
		pubmeta.Publisher.Add("", author)
		if i == 0 {
			fmt.Fprintf(tab, "Publisher:\t%s\n", author)
		} else {
			fmt.Fprintf(tab, "\t%s\n", author)
		}
	}
	// Добавляем уникальные идентификаторы
	for _, name := range []string{"uuid", "id", "identifier", "doi", "isbn", "issn"} {
		if value := meta.Get(name); value != "" {
			switch name {
			case "uuid", "doi", "isbn", "issn":
				value = "urn:" + name + ":" + value
				// TODO: Добавить префиксы для других идентификаторов
			}
			pubmeta.Identifier.Add(name, value)
			fmt.Fprintf(tab, "%s:\t%s\n", strings.ToUpper(name), value)
		}
	}
	// Добавляем краткое описание
	if description := meta.Description(); description != "" {
		pubmeta.Description.Add("description", description)
	}
	// Добавляем ключевые слова
	if keywords := meta.Keywords(); len(keywords) > 0 {
		for _, keyword := range keywords {
			pubmeta.Subject.Add("", keyword)
		}
		fmt.Fprintf(tab, "Keywords:\t%s\n", strings.Join(keywords, ", "))
	}
	// Добавляем описание сферы действия
	if coverage := meta.Get("coverage"); coverage != "" {
		pubmeta.Coverage.Add("", coverage)
	}
	// Добавляем дату
	if date := meta.Get("date"); date != "" {
		pubmeta.Date = &epub.Element{
			Value: date,
		}
	}
	// Добавляем копирайты
	for _, name := range []string{"copyright", "rights"} {
		if rights := meta.Get(name); rights != "" {
			pubmeta.Rights.Add(name, rights)
			fmt.Fprintf(tab, "%s:\t%s\n", strings.Title(name), rights)
		}
	}
	fmt.Fprintln(tab, strings.Repeat("—", 80))
	tab.Flush()
	return pubmeta, err
}
