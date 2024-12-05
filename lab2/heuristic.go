package main

import (
	"strings"
)

// Функция для генерации всех возможных комбинаций длины k из строки s
func generateCombinations(s string, k int) []string {
	var result []string
	var backtrack func(start int, path []rune)

	backtrack = func(start int, path []rune) {
		if len(path) == k {
			result = append(result, string(path))
			return
		}
		for i := start; i < len(s); i++ {
			path = append(path, rune(s[i]))
			backtrack(i+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, []rune{})
	return result
}

// RemoveChars - удаляет все символы алфавита из слова
func RemoveChars(alphabet, word string) (string, int) {
	charMap := make(map[rune]bool)
	for _, char := range alphabet {
		charMap[char] = true
	}

	var result strings.Builder
	removedCount := 0

	for _, char := range word {
		if !charMap[char] {
			result.WriteRune(char)
		} else {
			removedCount++
		}
	}

	return result.String(), removedCount
}

// Intersection - функция для нахождения пересечения символов двух строк
func Intersection(str1, str2 string) string {
	// Создаем мапу для хранения символов первой строки
	charMap := make(map[rune]bool)
	for _, char := range str1 {
		charMap[char] = true
	}

	// Создаем буфер для результата
	var result strings.Builder

	// Проходим по символам второй строки
	for _, char := range str2 {
		// Если символ есть в мапе, добавляем его в результат
		if charMap[char] {
			result.WriteRune(char)
			// Удаляем символ из мапы, чтобы избежать дубликатов в результате
			delete(charMap, char)
		}
	}

	// Возвращаем результат в виде строки
	return result.String()
}
