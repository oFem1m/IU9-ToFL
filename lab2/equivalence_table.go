package main

import (
	"fmt"
)

// Prefix - Структура для хранения префикса и флага принадлежности к главной части таблицы
type Prefix struct {
	Value  string // Сам префикс
	IsMain bool   // Является ли частью главной таблицы
}

// Suffix - Структура для хранения суффикса и флага принадлежности к главной части таблицы
type Suffix struct {
	Value  string // Сам суффикс
	IsMain bool   // Является ли частью главной таблицы
}

// EquivalenceTable - Структура таблицы эквивалентности
type EquivalenceTable struct {
	Prefixes []Prefix                   // Префиксы
	Suffixes []Suffix                   // Суффиксы
	Table    map[string]map[string]bool // Таблица значений: префикс -> суффикс -> bool
}

// NewEquivalenceTable - Создание новой таблицы эквивалентности
func NewEquivalenceTable(prefixes []Prefix, suffixes []Suffix) *EquivalenceTable {
	table := make(map[string]map[string]bool)

	for _, prefix := range prefixes {
		table[prefix.Value] = make(map[string]bool)
		for _, suffix := range suffixes {
			table[prefix.Value][suffix.Value] = false // Инициализируем значения false по умолчанию
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
	// Если префикс уже существует, ничего не делаем
	for _, prefix := range et.Prefixes {
		if prefix.Value == newPrefix.Value {
			return
		}
	}

	// Добавляем новый префикс в список
	et.Prefixes = append(et.Prefixes, newPrefix)
	et.Table[newPrefix.Value] = make(map[string]bool)

	// Инициализируем значения для каждого суффикса
	for _, suffix := range et.Suffixes {
		et.Table[newPrefix.Value][suffix.Value] = false
	}
}

// AddSuffix - Добавление нового суффикса
func (et *EquivalenceTable) AddSuffix(newSuffix Suffix) {
	// Если суффикс уже существует, ничего не делаем
	for _, suffix := range et.Suffixes {
		if suffix.Value == newSuffix.Value {
			return
		}
	}

	// Добавляем новый суффикс в список
	et.Suffixes = append(et.Suffixes, newSuffix)

	// Инициализируем значения для каждого префикса
	for _, prefix := range et.Prefixes {
		et.Table[prefix.Value][newSuffix.Value] = false
	}
}

// Update - Функция для обновления значения в таблице
func (et *EquivalenceTable) Update(prefix, suffix string, value bool) {
	if _, exists := et.Table[prefix]; exists {
		et.Table[prefix][suffix] = value
	}
}

// ContainsMainPrefix проверяет, содержится ли строка среди главных префиксов
func (et *EquivalenceTable) ContainsMainPrefix(prefix Prefix) bool {
	for _, s := range et.Suffixes {
		if s.IsMain && (s.Value == prefix.Value) {
			return true
		}
	}
	return false
}

// ContainsMainSuffix проверяет, содержится ли строка среди главных суффиксов
func (et *EquivalenceTable) ContainsMainSuffix(suffix Prefix) bool {
	for _, s := range et.Suffixes {
		if s.IsMain && (s.Value == suffix.Value) {
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
		if et.Table[prefix1][suffix.Value] != et.Table[prefix2][suffix.Value] {
			return false
		}
	}

	return true
}

// PrintTable - Функция для печати таблицы
func (et *EquivalenceTable) PrintTable() {
	// Печать суффиксов
	fmt.Print("   ")
	for _, suffix := range et.Suffixes {
		if suffix.IsMain {
			fmt.Printf("%s(M) ", suffix.Value)
		} else {
			fmt.Printf("%s ", suffix.Value)
		}
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
			fmt.Printf("%v ", et.Table[prefix.Value][suffix.Value])
		}
		fmt.Println()
	}
}
