package main

import (
	"github.com/mdigger/epub3"
)

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
