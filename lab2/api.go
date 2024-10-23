package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AskForWord - Спрашивает, является ли данная строка словом языка
func (et *EquivalenceTable) AskForWord(word string) bool {
	if learnerMode == "manual" {
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
			fmt.Printf("Ошибка при формировании тела запроса: %v\n", err)
			return false
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Printf("Ошибка при отправке запроса: %v\n", err)
			return false
		}
		defer resp.Body.Close()

		// Читаем ответ
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Ошибка при чтении ответа: %v\n", err)
			return false
		}

		var responseMap map[string]string
		err = json.Unmarshal(body, &responseMap)
		if err != nil {
			fmt.Printf("Ошибка при разборе ответа: %v\n", err)
			return false
		}

		// Обрабатываем ответ
		response := responseMap["response"]
		switch response {
		case "1":
			et.AddWord(word, true)
			return true
		case "0":
			et.AddWord(word, false)
			return true
		default:
			fmt.Printf("Неизвестный ответ от сервера: %s\n", response)
			return false
		}
	}
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
		mainPrefixes := []string{}
		nonMainPrefixes := []string{}
		suffixes := []string{}
		tableData := []string{}

		// Собираем данные префиксов и суффиксов
		for _, prefix := range et.Prefixes {
			if prefix.IsMain {
				mainPrefixes = append(mainPrefixes, prefix.Value)
			} else {
				nonMainPrefixes = append(nonMainPrefixes, prefix.Value)
			}
		}

		for _, suffix := range et.Suffixes {
			suffixes = append(suffixes, suffix)
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
			fmt.Printf("Ошибка при формировании тела запроса: %v\n", err)
			return "ERROR", "ERROR"
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Printf("Ошибка при отправке запроса: %v\n", err)
			return "ERROR", "ERROR"
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Ошибка при чтении ответа: %v\n", err)
			return "ERROR", "ERROR"
		}

		// Обрабатываем ответ от сервера
		var responseStruct struct {
			Response string `json:"response"`
			Type     *bool  `json:"type"`
		}
		err = json.Unmarshal(body, &responseStruct)
		if err != nil {
			fmt.Printf("Ошибка при разборе ответа: %v\n", err)
			return "ERROR", "ERROR"
		}

		// Возвращаем ответ в зависимости от типа
		if responseStruct.Type == nil {
			return "true", "" // Автомат угадан
		} else if *responseStruct.Type {
			return responseStruct.Response, "true"
		} else {
			return responseStruct.Response, "false"
		}
	}
}

// SetModeForMAT - выбор одного из режимов MAT: easy, medium, hard
func SetModeForMAT(mode string) bool {
	url := fmt.Sprintf("http://%s:%s/generate", server, port)
	requestBody, err := json.Marshal(map[string]string{
		"mode": mode,
	})
	if err != nil {
		fmt.Printf("Ошибка при формировании тела запроса: %v\n", err)
		return false
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Ошибка при отправке запроса: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	return true
}
