package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

var learnerMode, server, port string
var counterTrueWords int

func main() {
	counterTrueWords = 0
	heuristicAdded := false
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
	eolAlphabet := ""

	maxLexemeSize, _ := SetModeForMAT(matMode)
	log.Printf("Максимальный размер лексеммы: %d", maxLexemeSize)

	// Время старта
	start := time.Now()

	IsDone := false

	// Инициализируем таблицу с картами префиксов и суффиксов
	prefixes := map[string]Prefix{
		epsilon: {Value: epsilon, IsMain: true},
	}
	suffixes := map[string]string{epsilon: epsilon}

	et := NewEquivalenceTable(prefixes, suffixes)
	useEol := true

	// Пока таблица не угадана
	for !IsDone {
		wordsToAsk := make(map[string]PrefixAndSuffixForWord)
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
					} else {
						// Иначе сохраняем для вопроса
						// Проверяем, существует ли уже такое слово в карте
						if _, exists := wordsToAsk[word]; !exists {
							// Если слова нет, создаем новую запись с пустым списком пар
							wordsToAsk[word] = PrefixAndSuffixForWord{
								Pairs: make([]Pair, 0),
							}
						}
						prefixSuffix := wordsToAsk[word]
						// Добавляем текущие префикс и суффикс в список пар для данного слова
						prefixSuffix.Pairs = append(wordsToAsk[word].Pairs, Pair{
							First:  prefix.Value,
							Second: suffix,
						})
						wordsToAsk[word] = prefixSuffix
					}
				}
			}
		}
		// Создаем дополнение таблицы для префиксов
		for _, oldPrefix := range et.Prefixes {
			// Для каждого символа алфавита
			for _, letter := range alphabet {
				if strings.Contains(eolAlphabet, string(letter)) && !useEol {
					continue
				}
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
							} else {
								// Проверяем, существует ли уже такое слово в карте
								if _, exists := wordsToAsk[word]; !exists {
									// Если слова нет, создаем новую запись с пустым списком пар
									wordsToAsk[word] = PrefixAndSuffixForWord{
										Pairs: make([]Pair, 0),
									}
								}
								prefixSuffix := wordsToAsk[word]
								// Добавляем текущие префикс и суффикс в список пар для данного слова
								prefixSuffix.Pairs = append(wordsToAsk[word].Pairs, Pair{
									First:  prefix.Value,
									Second: suffix,
								})
								wordsToAsk[word] = prefixSuffix
							}
						}
					}
				}
			}
		}
		et.AskForWordBatch(wordsToAsk)

		//wordsToAsk = make([]string, 0)
		wordsToAsk = make(map[string]PrefixAndSuffixForWord)
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
									// Проверяем, существует ли уже такое слово в карте
									if _, exists := wordsToAsk[word]; !exists {
										// Если слова нет, создаем новую запись с пустым списком пар
										wordsToAsk[word] = PrefixAndSuffixForWord{
											Pairs: make([]Pair, 0),
										}
									}
									prefixSuffix := wordsToAsk[word]
									// Добавляем текущие префикс и суффикс в список пар для данного слова
									prefixSuffix.Pairs = append(wordsToAsk[word].Pairs, Pair{
										First:  prefix.Value,
										Second: suffix,
									})
									wordsToAsk[word] = prefixSuffix
								}
							}
						}
					}

					et.AskForWordBatch(wordsToAsk)
					wordsToAsk = make(map[string]PrefixAndSuffixForWord)
				} else {
					inconsistency = false
				}
			}

			// отправляем таблицу MAT
			response, responseType := et.AskForTable()
			// Если угадали, то конец, меняем флаг, иначе - добавляем новые суффиксы
			if response == "true" {
				IsDone = true
			} else {
				if responseType == "true" {
					// fmt.Printf("Контрпример лернера: %s\n", response)
					et.Words[response] = true
				} else {
					// fmt.Printf("Контрпример мата: %s\n", response)

					et.Words[response] = false
				}
				for i := 0; i < len(response); i++ {
					et.AddSuffix(response[i:])
				}
				_, removedNumber := RemoveChars(eolAlphabet, response)
				if removedNumber > 0 {
					// fmt.Println("Используем eol")
					useEol = true
				} else {
					// fmt.Println("Не используем eol")
					useEol = false
				}
				//if heuristicAdded {
				//	fmt.Printf("Добавил контпример в префиксы: %s\n", response)
				//	prefix := Prefix{
				//		Value:  response,
				//		IsMain: true,
				//	}
				//	et.AddPrefix(prefix)
				//}
			}
		}
		// fmt.Printf("Количество угаданных слов: %d\n", counterTrueWords)

		counterEolAlphabets := 0
		if counterTrueWords > 5000 && !heuristicAdded {
			heuristicAdded = true
			OriginalWordsToAsk := make(map[string]PrefixAndSuffixForWord)
			for word := range et.Words {
				if et.Words[word] {
					OriginalWordsToAsk[word] = PrefixAndSuffixForWord{}
				}
			}
			eolFindFlag := false
			for length := len(alphabet) - 4; length > 0; length-- {
				if eolFindFlag {
					break
				}
				combinations := generateCombinations(alphabet, length)
				for _, subAlphabet := range combinations {
					NewWordsToAsk := make(map[string]PrefixAndSuffixForWord)
					// fmt.Printf("Проверка для подалфавита: %s\n", subAlphabet)
					emptyWord := false
					for word := range OriginalWordsToAsk {
						NewWord, _ := RemoveChars(subAlphabet, word)
						if NewWord == "" {
							emptyWord = true
							break
						}
						NewWordsToAsk[NewWord] = PrefixAndSuffixForWord{}
					}
					if !emptyWord {
						responseList := et.AskForWordBatch(NewWordsToAsk)
						// countingOfFalse := 0
						responseWithFalse := false
						for _, response := range responseList {
							if !response {
								responseWithFalse = true
							}
						}
						// fmt.Printf("Число ошибок: %d\n", countingOfFalse)
						if !responseWithFalse {
							if eolAlphabet != "" {
								eolAlphabet = Intersection(eolAlphabet, subAlphabet)
							} else {
								eolAlphabet = subAlphabet
							}

							fmt.Printf("Алфавит для eol: %s\n", eolAlphabet)
							counterEolAlphabets++
							eolFindFlag = true
							//break
						}
					}
				}
			}
			fmt.Printf("Количество найденных алфавитов: %d\n", counterEolAlphabets)
		}

	}
	// et.PrintTable()
	// Засекаем время
	finish := time.Since(start)
	fmt.Printf("Время выполнения программы: %s\n", finish)
}
