package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// AskForWord - Спрашивает, является ли данная строка словом языка
func (et *EquivalenceTable) AskForWord(word string) bool {
	if learnerMode == "manual" {
		// Ручной режим, без изменений
		var response string
		fmt.Printf("Является ли '%s' словом языка? (1/0): ", word)
		fmt.Scanln(&response)

		switch response {
		case "1":
			et.AddWord(word, true)
			return true
		case "0":
			et.AddWord(word, false)
			return true
		}
		return false
	} else {
		url := fmt.Sprintf("http://%s:%s/checkWord", server, port)

		requestBody, err := json.Marshal(map[string]string{
			"word": word,
		})
		if err != nil {
			log.Printf("Ошибка при формировании тела запроса: %v", err)
			return false
		}

		// log.Printf("Отправка POST запроса на URL: %s с телом: %s", url, string(requestBody))

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			log.Printf("Ошибка при отправке запроса: %v", err)
			return false
		}
		defer resp.Body.Close()

		// Читаем ответ
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Ошибка при чтении ответа: %v", err)
			return false
		}

		// log.Printf("Ответ от сервера: %s", string(body))

		var responseMap map[string]string
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			log.Printf("Ошибка при разборе ответа: %v", err)
			return false
		}

		// Обрабатываем ответ
		response := responseMap["response"]
		// log.Printf("Результат разбора: %s", response)
		switch response {
		case "1":
			et.AddWord(word, true)
			return true
		case "0":
			et.AddWord(word, false)
			return true
		default:
			log.Printf("Неизвестный ответ от сервера: %s", response)
			return false
		}
	}
}

// AskForWordBatch - Спрашивает, является ли каждое слово в wordsToAsk словом языка
func (et *EquivalenceTable) AskForWordBatch(wordsToAsk map[string]PrefixAndSuffixForWord) []bool {
	url := fmt.Sprintf("http://%s:%s/check-word-batch", server, port)

	// Собираем список слов для отправки на сервер
	words := make([]string, 0, len(wordsToAsk))
	for word := range wordsToAsk {
		words = append(words, word)
	}

	// Формируем тело запроса
	type WordsRequest struct {
		Words []string `json:"wordList"`
	}
	requestBody := WordsRequest{
		Words: words,
	}

	// Сериализуем запрос в JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Errorf("Ошибка при формировании тела запроса: %v", err)
	}

	// Отправляем POST запрос на сервер
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Errorf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Декодируем ответ сервера
	type BoolResponse struct {
		Bools []bool `json:"responseList"`
	}
	var response BoolResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Errorf("Ошибка при декодировании JSON: %v", err)
	}

	// Проверка на количество слов и полученных результатов
	if len(response.Bools) != len(words) {
		fmt.Errorf("Некорректное количество ответов: ожидалось %d, получено %d", len(words), len(response.Bools))
	}

	// Обрабатываем каждый ответ
	for i, word := range words {
		belonging := response.Bools[i] // Получаем результат для текущего слова (true/false)

		// Добавляем слово в словарь таблицы эквивалентности
		et.AddWord(word, belonging)

		// Обновляем значения в таблице по всем парам префикс/суффикс для этого слова
		for _, pair := range wordsToAsk[word].Pairs {
			if belonging {
				et.Update(pair.First, pair.Second, '+')
			} else {
				et.Update(pair.First, pair.Second, '-')
			}
		}
	}
	return response.Bools
}

// AskForTable - Спрашивает, является ли данная таблица искомым автоматом
func (et *EquivalenceTable) AskForTable() (string, string) {
	if learnerMode == "manual" {
		// Ручной режим
		et.PrintTable()
		var response, response_type string
		fmt.Print("Верна ли таблица выше? (true/false): ")
		fmt.Scanln(&response)
		if response == "true" {
			return "true", ""
		} else {
			fmt.Print("Введите контрпример: ")
			fmt.Scanln(&response)
			fmt.Print("Введите тип контрпримера (true - принадлежит МАТУ, но не Лернеру; false - Лернеру, но не МАТу): ")
			fmt.Scanln(&response_type)
		}

		return response, response_type
	} else {
		mainPrefixes := []string{"ε"} // Добавляем ε как первый главный префикс
		nonMainPrefixes := []string{}
		suffixes := []string{"ε"} // Добавляем ε как первый суффикс
		tableData := []string{}

		// Собираем данные префиксов
		for _, prefix := range et.Prefixes {
			if prefix.Value != "ε" { // Пропускаем ε, так как он уже добавлен
				if prefix.IsMain {
					mainPrefixes = append(mainPrefixes, prefix.Value)
				} else {
					nonMainPrefixes = append(nonMainPrefixes, prefix.Value)
				}
			}
		}

		// Собираем суффиксы
		for _, suffix := range et.Suffixes {
			if suffix != "ε" { // Пропускаем ε, так как он уже добавлен
				suffixes = append(suffixes, suffix)
			}
		}

		// Собираем значения таблицы
		for _, prefix := range mainPrefixes {
			for _, suffix := range suffixes {
				val := et.GetValue(prefix, suffix)
				if val == '+' {
					tableData = append(tableData, "1")
				} else {
					tableData = append(tableData, "0")
				}
			}
		}
		for _, prefix := range nonMainPrefixes {
			for _, suffix := range suffixes {
				val := et.GetValue(prefix, suffix)
				if val == '+' {
					tableData = append(tableData, "1")
				} else {
					tableData = append(tableData, "0")
				}
			}
		}

		url := fmt.Sprintf("http://%s:%s/checkTable", server, port)
		requestBody, err := json.Marshal(map[string]string{
			"main_prefixes":     strings.Join(mainPrefixes, " "),
			"non_main_prefixes": strings.Join(nonMainPrefixes, " "),
			"suffixes":          strings.Join(suffixes, " "),
			"table":             strings.Join(tableData, " "),
		})
		if err != nil {
			log.Printf("Ошибка при формировании тела запроса: %v", err)
			return "ERROR", "ERROR"
		}

		// log.Printf("Отправка POST запроса на URL: %s с телом: %s", url, string(requestBody))

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Errorf("Ошибка при отправке запроса: %v", err)
			return "ERROR", "ERROR"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Errorf("Ошибка при чтении ответа: %v", err)
			return "ERROR", "ERROR"
		}

		// log.Printf("Ответ от сервера: %s", string(body))

		var responseStruct struct {
			Response string `json:"response"`
			Type     *bool  `json:"type"`
		}
		err = json.Unmarshal(body, &responseStruct)
		if err != nil {
			fmt.Errorf("Ошибка при разборе ответа: %v", err)
			return "ERROR", "ERROR"
		}

		// Возвращаем ответ в зависимости от типа
		if responseStruct.Type == nil {
			// log.Printf("Таблица подтверждена.")
			return "true", "" // Автомат угадан
		} else if *responseStruct.Type {
			// log.Printf("Контрпример: %s, тип: true", responseStruct.Response)
			return responseStruct.Response, "true"
		} else {
			// log.Printf("Контрпример: %s, тип: false", responseStruct.Response)
			return responseStruct.Response, "false"
		}
	}
}

// SetModeForMAT - выбор одного из режимов MAT: easy, medium, hard
func SetModeForMAT(mode string) (int, int) {
	url := fmt.Sprintf("http://%s:%s/generate", server, port)
	requestBody, err := json.Marshal(map[string]string{
		"mode": mode,
	})
	if err != nil {
		fmt.Errorf("Ошибка при формировании тела запроса: %v\n", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Errorf("Ошибка при отправке запроса: %v\n", err)
	}
	defer resp.Body.Close()

	type GenerateResponse struct {
		MaxLexemeSize     int `json:"maxLexemeSize"`
		MaxBracketNesting int `json:"maxBracketNesting"`
	}
	var response GenerateResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Errorf("Ошибка при декодировании JSON: %v", err)
	}

	return response.MaxLexemeSize, response.MaxBracketNesting
}
