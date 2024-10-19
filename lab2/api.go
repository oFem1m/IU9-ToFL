package main

import "fmt"

// AskUserForWord - Функция, которая спрашивает пользователя, является ли данная строка словом языка
func (et *EquivalenceTable) AskUserForWord(word string) bool {
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
}

// AskUserForTable - Функция, которая спрашивает пользователя, является ли данная таблица искомым автоматом
func (et *EquivalenceTable) AskUserForTable() string {
	et.PrintTable()
	var response string

	fmt.Print("Верна ли таблица выше? (true/<контрпример>): ")
	fmt.Scanln(&response)

	if response == "true" {
		return "OK"
	}
	return response
}
