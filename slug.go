package main

import (
	"hash/crc64"
)

// Словарь для генерации уникальных идентификаторов.
const dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_$"

// Хеш-функция, используемая для генерации уникальных идентификаторов.
var hash = crc64.New(crc64.MakeTable(crc64.ECMA))

// hashSlug возвращает уникальный идентификатор, сгенерированный на основании указанного имени.
func hashSlug(name []byte) []byte {
	hash.Reset()
	hash.Write([]byte(name))
	val := hash.Sum64()
	var a = make([]byte, 8)
	for i := 0; i < 8; i++ {
		a[i] = dictionary[val&63]
		val >>= 8
	}
	return a
}
