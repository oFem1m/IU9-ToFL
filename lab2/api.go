package main

import "fmt"

// AskUserForWord - Функция, которая спрашивает пользователя, является ли данная строка словом языка
func (et *EquivalenceTable) AskUserForWord(prefix, suffix string) {
	word := prefix + suffix
	var response string

	fmt.Printf("Является ли '%s' словом языка? (+/-): ", word)
	fmt.Scanln(&response)

	value := false
	if response == "+" {
		value = true
	}

	et.Update(prefix, suffix, value)
}
