package main

import "fmt"

// AskUserForWord - Функция, которая спрашивает пользователя, является ли данная строка словом языка
func (et *EquivalenceTable) AskUserForWord(prefix, suffix string) {
	word := prefix + suffix
	var response string

	fmt.Printf("Является ли '%s' словом языка? (+/-): ", word)
	fmt.Scanln(&response)

	value := '-'
	if response == "+" {
		value = '+'
	}

	et.Update(prefix, suffix, value)
}

// AskUserForTable - Функция, которая спрашивает пользователя, является ли данная таблица искомым авотматом
func (et *EquivalenceTable) AskUserForTable() string {
	et.PrintTable()
	var response string

	fmt.Print("Верна ли таблица выше? (+/<контрпример>): ")
	fmt.Scanln(&response)

	if response == "+" {
		return "OK"
	}
	return response
}
