package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Рекурсивная функция для генерации комбинаций
func generateCombinations(alphabet string, length int, prefix string, result *[]string) {
	if length == 0 {
		*result = append(*result, prefix)
		return
	}

	for _, char := range alphabet {
		generateCombinations(alphabet, length-1, prefix+string(char), result)
	}
}

// Генерация всех возможных строк длиной от 1 до maxLexemeSize
func generateStrings(alphabet string, maxLexemeSize int) []string {
	var result []string

	for length := 1; length <= maxLexemeSize; length++ {
		generateCombinations(alphabet, length, "", &result)
	}

	return result
}

// AskForEol - используется для поиска eol
func (et *EquivalenceTable) AskForEol(word string) bool {
	url := fmt.Sprintf("http://%s:%s/checkWord", server, port)

	requestBody, err := json.Marshal(map[string]string{
		"word": word,
	})
	if err != nil {
		log.Printf("Ошибка при формировании тела запроса: %v", err)
	}

	log.Printf("Отправка POST запроса на URL: %s с телом: %s", url, string(requestBody))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении ответа: %v", err)
	}

	log.Printf("Ответ от сервера: %s", string(body))

	var responseMap map[string]string
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		log.Printf("Ошибка при разборе ответа: %v", err)
	}

	// Обрабатываем ответ
	response := responseMap["response"]
	log.Printf("Результат разбора: %s", response)
	switch response {
	case "1":
		et.AddWord(word, true)
		return true
	case "0":
		et.AddWord(word, false)
		return false
	default:
		log.Printf("Неизвестный ответ от сервера: %s", response)
		return false
	}
}
