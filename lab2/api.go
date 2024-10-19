package main

import "fmt"

// AskUserForWord - Функция, которая спрашивает пользователя, является ли данная строка словом языка
func (et *EquivalenceTable) AskUserForWord(word string) bool {
	var response string

	fmt.Printf("Является ли '%s' словом языка? (+/-): ", word)
	fmt.Scanln(&response)

	switch response {
	case "+":
		et.AddWord(word, true)
		return true
	case "-":
		et.AddWord(word, false)
		return true
	}
	return false
}

// AskUserForTable - Функция, которая спрашивает пользователя, является ли данная таблица искомым автоматом
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
