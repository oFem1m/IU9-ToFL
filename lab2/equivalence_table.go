package main

import (
	"fmt"
)

// Prefix - Структура для хранения префикса и флага принадлежности к главной части таблицы
type Prefix struct {
	Value  string // Сам префикс
	IsMain bool   // Является ли частью главной таблицы
}

// EquivalenceTable - Структура таблицы эквивалентности
type EquivalenceTable struct {
	Prefixes []Prefix                   // Префиксы
	Suffixes []string                   // Суффиксы
	Table    map[string]map[string]rune // Таблица значений: префикс + суффикс -> rune
	Words    map[string]bool            // Словарь слов: слово -> принадлежность к языку
}

// NewEquivalenceTable - Создание новой таблицы
func NewEquivalenceTable(prefixes []Prefix, suffixes []string) *EquivalenceTable {
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

// GetValue - функция получения значения из таблицы
func (et *EquivalenceTable) GetValue(prefix string, suffix string) rune {
	return et.Table[prefix][suffix]
}

// AddPrefix - Добавление нового префикса
func (et *EquivalenceTable) AddPrefix(newPrefix Prefix) bool {
	// Если префикс уже существует, ничего не делаем
	for _, prefix := range et.Prefixes {
		if prefix.Value == newPrefix.Value {
			return false
		}
	}
	// Добавляем новый префикс в список
	et.Prefixes = append(et.Prefixes, newPrefix)
	et.Table[newPrefix.Value] = make(map[string]rune)

	// Инициализируем значения для каждого суффикса
	for _, suffix := range et.Suffixes {
		et.Update(newPrefix.Value, suffix, '0')
	}
	return true
}

// AddSuffix - Добавление нового суффикса
func (et *EquivalenceTable) AddSuffix(newSuffix string) bool {
	// Если суффикс уже существует, ничего не делаем
	for _, suffix := range et.Suffixes {
		if suffix == newSuffix {
			return false
		}
	}
	// Добавляем новый суффикс в список
	et.Suffixes = append(et.Suffixes, newSuffix)

	// Инициализируем значения для каждого префикса
	for _, prefix := range et.Prefixes {
		et.Table[prefix.Value][newSuffix] = '0'
	}
	return true
}

// Update - обновление значения в таблице и добавление нового слова в словарь
func (et *EquivalenceTable) Update(prefix, suffix string, value rune) {
	if _, exists := et.Table[prefix]; exists {
		et.Table[prefix][suffix] = value
		word := prefix + suffix
		switch value {
		case '+':
			et.Words[word] = true
		case '-':
			et.Words[word] = false
		}
	}
}

// ContainsMainPrefix проверяет, содержится ли строка среди главных префиксов
func (et *EquivalenceTable) ContainsMainPrefix(prefix Prefix) bool {
	for _, s := range et.Prefixes {
		if s.IsMain && (s.Value == prefix.Value) {
			return true
		}
	}
	return false
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
		if et.Table[prefix1][suffix] != et.Table[prefix2][suffix] {
			return false
		}
	}

	return true
}

// CompleteTable - Проверка таблицы на полноту и приведение её к полному виду
func (et *EquivalenceTable) CompleteTable() {
	for i, nonMainPrefix := range et.Prefixes {
		// Проверяем только неглавные префиксы
		if !nonMainPrefix.IsMain {
			isEquivalent := false

			// Сравниваем с каждым главным префиксом
			for _, mainPrefix := range et.Prefixes {
				if mainPrefix.IsMain && et.ArePrefixesEquivalent(nonMainPrefix.Value, mainPrefix.Value) {
					isEquivalent = true
					break
				}
			}

			// Если неглавный префикс не эквивалентен ни одному главному
			if !isEquivalent {
				// Обновляем флаг, что этот префикс теперь главный
				et.Prefixes[i].IsMain = true
			}
		}
	}
}

// PrintTable - Функция для вывода таблицы в консоль
func (et *EquivalenceTable) PrintTable() {
	// Печать суффиксов
	fmt.Print("   ")
	for _, suffix := range et.Suffixes {
		fmt.Printf("%s ", suffix)
	}
	fmt.Println()

	// Печать префиксов и значений таблицы
	for _, prefix := range et.Prefixes {
		if prefix.IsMain {
			fmt.Printf("%s(M) ", prefix.Value)
		} else {
			fmt.Printf("%s ", prefix.Value)
		}
		for _, suffix := range et.Suffixes {
			fmt.Printf("%c ", et.Table[prefix.Value][suffix])
		}
		fmt.Println()
	}

	// Печать списка слов с флагами принадлежности к языку
	fmt.Println("\nWords:")
	for word, belongs := range et.Words {
		fmt.Printf("%s: %v\n", word, belongs)
	}
}