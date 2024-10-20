package main

import (
	"fmt"
	"strings"
)

var manualMode = true
var server, port string

func main() {
	var alphabet string

	fmt.Print("Введите символы алфавита одной строкой: ")
	fmt.Scanln(&alphabet)
	var response string
	epsilon := ""
	fmt.Print("Использовать ε в роли пустой строки? +/-: ")
	fmt.Scanln(&response)
	if response == "+" {
		epsilon = "ε"
	}

	fmt.Print("Использовать Лернер в ручном режиме? (Без программы MAT) +/-: ")
	fmt.Scanln(&response)
	if response == "-" {
		manualMode = false
		fmt.Print("Введите ip-адрес и порт сервера MAT в формате <адрес>:<порт> ")
		fmt.Scanln(&response)
		// Разделяем строку на адрес и порт
		parts := strings.Split(response, ":")
		server = parts[0]
		port = parts[1]
	}

	IsDone := false

	prefixes := []Prefix{
		{Value: epsilon, IsMain: true},
	}
	suffixes := []string{epsilon}

	et := NewEquivalenceTable(prefixes, suffixes)

	// Пока таблица не угадана
	for !IsDone {
		// Заполняем пустые значения таблицы
		for _, prefix := range et.Prefixes {
			for _, suffix := range et.Suffixes {
				// Если ячейка пуста
				if et.GetValue(prefix.Value, suffix) == '0' {
					currentPrefix := prefix.Value
					currentSuffix := suffix
					var word string
					// Избавляемся от ε
					if currentPrefix == "ε" && currentSuffix == "ε" {
						word = "ε"
					} else if currentPrefix == "ε" {
						currentPrefix = ""
						word = currentPrefix + currentSuffix
					} else if currentSuffix == "ε" {
						currentSuffix = ""
						word = currentPrefix + currentSuffix
					} else {
						word = currentPrefix + currentSuffix
					}
					// Проверяем наличие слова в словаре
					if et.CheckWord(word) {
						// Если слово принадлежит языку
						if et.Words[word] {
							et.Update(prefix.Value, suffix, '+')
						} else {
							et.Update(prefix.Value, suffix, '-')
						}
					} else { // Иначе спрашиваем
						if et.AskForWord(word) {
							if et.Words[word] {
								et.Update(prefix.Value, suffix, '+')
							} else {
								et.Update(prefix.Value, suffix, '-')
							}
						}
					}
				}
			}
		}
		// Создаем дополнение таблицы для префиксов
		for _, oldPrefix := range et.Prefixes {
			// Для каждого символа алфавита
			for _, letter := range alphabet {
				// Создаем новые префиксы на основе главных префиксов
				if oldPrefix.IsMain {
					currentPrefix := oldPrefix.Value
					if currentPrefix == "ε" {
						currentPrefix = ""
					}
					prefix := Prefix{
						Value:  currentPrefix + string(letter),
						IsMain: false,
					}
					// Если префикс удалось добавить
					if et.AddPrefix(prefix) {
						// По необходимости задаём вопросы MAT, заполняем таблицу
						for _, suffix := range et.Suffixes {
							// Убираем ε-суффикс
							currentSuffix := suffix
							if currentSuffix == "ε" {
								currentSuffix = ""
							}
							word := prefix.Value + currentSuffix

							// Проверяем наличие слова в словаре
							if et.CheckWord(word) {
								// Если слово принадлежит языку
								if et.Words[word] {
									et.Update(prefix.Value, suffix, '+')
								} else {
									et.Update(prefix.Value, suffix, '-')
								}
							} else { // Иначе спрашиваем
								if et.AskForWord(word) {
									if et.Words[word] {
										et.Update(prefix.Value, suffix, '+')
									} else {
										et.Update(prefix.Value, suffix, '-')
									}
								}
							}
						}
					}
				}
			}
		}

		// Проверяем таблицу на полноту и приводим к полному виду
		et.CompleteTable()

		// Проверка, являются ли все префиксы главными
		if !et.AreAllPrefixesMain() {
			inconsistency := true
			for inconsistency {
				if et.InconsistencyTable(alphabet) {
					// Заполняем пустые значения таблицы
					for _, prefix := range et.Prefixes {
						for _, suffix := range et.Suffixes {
							// Если ячейка пуста
							if et.GetValue(prefix.Value, suffix) == '0' {
								currentPrefix := prefix.Value
								currentSuffix := suffix
								var word string
								// Избавляемся от ε
								if currentPrefix == "ε" && currentSuffix == "ε" {
									word = "ε"
								} else if currentPrefix == "ε" {
									currentPrefix = ""
									word = currentPrefix + currentSuffix
								} else if currentSuffix == "ε" {
									currentSuffix = ""
									word = currentPrefix + currentSuffix
								} else {
									word = currentPrefix + currentSuffix
								}
								// Проверяем наличие слова в словаре
								if et.CheckWord(word) {
									// Если слово принадлежит языку
									if et.Words[word] {
										et.Update(prefix.Value, suffix, '+')
									} else {
										et.Update(prefix.Value, suffix, '-')
									}
								} else { // Иначе спрашиваем
									if et.AskForWord(word) {
										if et.Words[word] {
											et.Update(prefix.Value, suffix, '+')
										} else {
											et.Update(prefix.Value, suffix, '-')
										}
									}
								}
							}
						}
					}
				} else {
					inconsistency = false
				}
			}
			// отправляем таблицу MAT
			response := et.AskForTable()
			// Если угадали, то конец меняем флаг, иначе - добавляем новые суффиксы
			if response == "true" {
				IsDone = true
			} else {
				et.Words[response] = true
				for i := 0; i < len(response); i++ {
					et.AddSuffix(response[i:])
				}
			}
		}
	}
}
