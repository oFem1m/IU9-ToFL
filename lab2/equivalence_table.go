package main

import (
	"fmt"
)

// Prefix - Структура для хранения префикса и флага принадлежности к главной части таблицы
type Prefix struct {
	Value  string // Сам префикс
	IsMain bool   // Является ли частью главной таблицы
}

// EquivalenceTable - Структура таблицы классов эквивалентности
type EquivalenceTable struct {
	Prefixes map[string]Prefix          // Префиксы
	Suffixes map[string]string          // Суффиксы
	Table    map[string]map[string]rune // Таблица значений: префикс + суффикс -> rune
	Words    map[string]bool            // Словарь слов: слово -> принадлежность к языку
}

// Pair - структура пары строк
type Pair struct {
	First  string
	Second string
}

// PrefixAndSuffixForWord - хранит необходимые префиксы и суффиксы для слова
type PrefixAndSuffixForWord struct {
	Pairs []Pair
}

// NewEquivalenceTable - Создание новой таблицы
func NewEquivalenceTable(prefixes map[string]Prefix, suffixes map[string]string) *EquivalenceTable {
	table := make(map[string]map[string]rune)
	words := make(map[string]bool)

	for _, prefix := range prefixes {
		table[prefix.Value] = make(map[string]rune)
		for _, suffix := range suffixes {
			table[prefix.Value][suffix] = '0' // По умолчанию
		}
	}

	return &EquivalenceTable{
		Prefixes: prefixes,
		Suffixes: suffixes,
		Table:    table,
		Words:    words,
	}
}

// CheckWord - проверка наличия слова в словаре
func (et *EquivalenceTable) CheckWord(word string) bool {
	_, exists := et.Words[word]
	return exists
}

// AddWord - добавляет новое слово в словарь
func (et *EquivalenceTable) AddWord(word string, belonging bool) bool {
	_, exists := et.Words[word]
	if !exists {
		et.Words[word] = belonging
		return true
	}
	return false
}

// GetValue - функция получения значения из таблицы
func (et *EquivalenceTable) GetValue(prefix string, suffix string) rune {
	return et.Table[prefix][suffix]
}

// SetValue - функция внесения значения в таблицу
func (et *EquivalenceTable) SetValue(prefix string, suffix string, value rune) {
	// Проверка на существование префикса
	if _, exists := et.Prefixes[prefix]; !exists {
		return
	}

	// Проверка на существование суффикса
	if _, exists := et.Suffixes[suffix]; !exists {
		return
	}
	et.Table[prefix][suffix] = value
}

// AddPrefix - Добавление нового префикса
func (et *EquivalenceTable) AddPrefix(newPrefix Prefix) bool {
	if _, exists := et.Prefixes[newPrefix.Value]; exists {
		return false
	}
	et.Prefixes[newPrefix.Value] = newPrefix
	et.Table[newPrefix.Value] = make(map[string]rune)

	// Инициализируем значения для каждого суффикса
	for _, suffix := range et.Suffixes {
		et.Update(newPrefix.Value, suffix, '0')
	}
	return true
}

// AddSuffix - Добавление нового суффикса
func (et *EquivalenceTable) AddSuffix(newSuffix string) bool {
	if _, exists := et.Suffixes[newSuffix]; exists {
		return false
	}
	et.Suffixes[newSuffix] = newSuffix

	// Инициализируем значения для каждого префикса
	for _, prefix := range et.Prefixes {
		et.Update(prefix.Value, newSuffix, '0')
	}
	return true
}

// Update - обновление значения в таблице и добавление нового слова в словарь
func (et *EquivalenceTable) Update(prefix, suffix string, value rune) {
	if _, exists := et.Table[prefix]; exists {
		et.SetValue(prefix, suffix, value)
		currentPrefix := prefix
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
		switch value {
		case '+':
			et.Words[word] = true
		case '-':
			et.Words[word] = false
		}
	}
}

// AreAllPrefixesMain проверяет, являются ли все префиксы главными
func (et *EquivalenceTable) AreAllPrefixesMain() bool {
	for _, prefix := range et.Prefixes {
		if !prefix.IsMain {
			return false
		}
	}
	return true
}

// ArePrefixesEquivalent - Проверяет, эквивалентны ли два префикса
func (et *EquivalenceTable) ArePrefixesEquivalent(prefix1, prefix2 string) bool {
	// Если хотя бы один префикс отсутствует в таблице, они не эквивалентны
	_, exists1 := et.Table[prefix1]
	_, exists2 := et.Table[prefix2]
	if !exists1 || !exists2 {
		return false
	}

	// Сравниваем значения для всех суффиксов
	for _, suffix := range et.Suffixes {
		if et.GetValue(prefix1, suffix) != et.GetValue(prefix2, suffix) {
			return false
		}
	}

	return true
}

// CompleteTable - Приведение таблицы к полному виду
func (et *EquivalenceTable) CompleteTable() {
	for key, nonMainPrefix := range et.Prefixes {
		if !nonMainPrefix.IsMain {
			isEquivalent := false
			for _, mainPrefix := range et.Prefixes {
				if mainPrefix.IsMain && et.ArePrefixesEquivalent(nonMainPrefix.Value, mainPrefix.Value) {
					isEquivalent = true
					break
				}
			}
			if !isEquivalent {
				et.Prefixes[key] = Prefix{nonMainPrefix.Value, true}
			}
		}
	}
}

// InconsistencyTable - Проверка на противоречивость и исправление
func (et *EquivalenceTable) InconsistencyTable(alphabet string) bool {
	for _, prefix1 := range et.Prefixes {
		if !prefix1.IsMain {
			continue
		}

		for _, prefix2 := range et.Prefixes {
			if !prefix2.IsMain || prefix1.Value == prefix2.Value {
				continue
			}

			// Проверяем эквивалентность префиксов
			if et.ArePrefixesEquivalent(prefix1.Value, prefix2.Value) {
				// Ищем такие символы из алфавита и суффиксы v_k
				for _, suffix := range et.Suffixes {
					for _, letter := range alphabet { // Проходим по символам алфавита
						currentPrefix1 := prefix1.Value
						currentPrefix2 := prefix2.Value
						currentSuffix := suffix
						// Избавляемся от ε
						if currentPrefix1 == "ε" {
							currentPrefix1 = ""
						}
						if currentPrefix2 == "ε" {
							currentPrefix2 = ""
						}
						if currentSuffix == "ε" {
							currentSuffix = ""
						}

						word1 := currentPrefix1 + string(letter) + currentSuffix
						word2 := currentPrefix2 + string(letter) + currentSuffix

						flag1, ok1 := et.Words[word1]
						flag2, ok2 := et.Words[word2]

						if !ok1 {
							et.AskForWord(word1)
							flag1, ok1 = et.Words[word1]
						}
						if !ok2 {
							et.AskForWord(word2)
							flag2, ok2 = et.Words[word2]
						}

						// Проверяем на противоречие
						if flag1 != flag2 {
							// Найдено противоречие, добавляем новый суффикс a+v_k
							newSuffix := string(letter) + currentSuffix
							et.AddSuffix(newSuffix)
							return true // Возвращаем true, если было добавлено что-то новое
						}
					}
				}
			}
		}
	}
	return false // противоречий нет
}

// PrintTable - Функция для вывода таблицы в консоль
func (et *EquivalenceTable) PrintTable() {
	// Вывод суффиксов
	fmt.Print("   |")
	for _, suffix := range et.Suffixes {
		fmt.Printf("%s|", suffix)
	}
	fmt.Println()

	// Вывод префиксов и значений таблицы
	for _, prefix := range et.Prefixes {
		if prefix.IsMain {
			fmt.Printf("%s(M) ", prefix.Value)
		} else {
			fmt.Printf("%s ", prefix.Value)
		}
		for _, suffix := range et.Suffixes {
			fmt.Printf("%c ", et.GetValue(prefix.Value, suffix))
		}
		fmt.Println()
	}
}
