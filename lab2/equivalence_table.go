package main

import (
	"fmt"
)

// Prefix - Структура для хранения префикса и флага принадлежности к главной части таблицы
type Prefix struct {
	Value  string // Сам префикс
	IsMain bool   // Является ли частью главной таблицы
}

//// Suffix - Структура для хранения суффикса
//type Suffix struct {
//	Value string // Сам суффикс
//}

// EquivalenceTable - Структура таблицы эквивалентности
type EquivalenceTable struct {
	Prefixes []Prefix                   // Префиксы
	Suffixes []string                   // Суффиксы
	Table    map[string]map[string]rune // Таблица значений: префикс -> суффикс -> bool
}

// NewEquivalenceTable - Создание новой таблицы
func NewEquivalenceTable(prefixes []Prefix, suffixes []string) *EquivalenceTable {
	table := make(map[string]map[string]rune)

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
	}
}

// AddPrefix - Добавление нового префикса
func (et *EquivalenceTable) AddPrefix(newPrefix Prefix) {
	// Добавляем новый префикс в список
	et.Prefixes = append(et.Prefixes, newPrefix)
	et.Table[newPrefix.Value] = make(map[string]rune)

	// Инициализируем значения для каждого суффикса
	for _, suffix := range et.Suffixes {
		et.Table[newPrefix.Value][suffix] = '0'
	}
}

// AddSuffix - Добавление нового суффикса
func (et *EquivalenceTable) AddSuffix(newSuffix string) {
	// Если суффикс уже существует, ничего не делаем
	for _, suffix := range et.Suffixes {
		if suffix == newSuffix {
			return
		}
	}

	// Добавляем новый суффикс в список
	et.Suffixes = append(et.Suffixes, newSuffix)

	// Инициализируем значения для каждого префикса
	for _, prefix := range et.Prefixes {
		et.Table[prefix.Value][newSuffix] = '0'
	}
}

// Update - обновление значения в таблице
func (et *EquivalenceTable) Update(prefix, suffix string, value rune) {
	if _, exists := et.Table[prefix]; exists {
		et.Table[prefix][suffix] = value
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
				fmt.Printf("Префикс '%s' не эквивалентен ни одному главному префиксу. Добавляем его в основную часть.\n", nonMainPrefix.Value)
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
}
