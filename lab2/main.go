package main

import (
	"fmt"
	"time"
)

var learnerMode, server, port string

func main() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	alphabet := config.Alphabet
	epsilon := config.Epsilon
	learnerMode = config.LearnerMode
	server = config.ServerAddr
	port = config.ServerPort
	matMode := config.MatMode

	if learnerMode == "automatic" {
		SetModeForMAT(matMode)
	}

	// Время старта
	start := time.Now()

	IsDone := false

	// Инициализируем таблицу с картами префиксов и суффиксов
	prefixes := map[string]Prefix{
		epsilon: {Value: epsilon, IsMain: true},
	}
	suffixes := map[string]string{epsilon: epsilon}

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
					fmt.Println("inconsistency!")
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
			// Если угадали, то конец, меняем флаг, иначе - добавляем новые суффиксы
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
	et.PrintTable()
	// Засекаем время
	finish := time.Since(start)
	fmt.Printf("Время выполнения программы: %s\n", finish)
}
