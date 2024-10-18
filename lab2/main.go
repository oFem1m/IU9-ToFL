package main

func main() {
	alphabet := "ab"

	IsDone := false

	prefixes := []Prefix{
		{Value: "", IsMain: true},
	}
	suffixes := []Suffix{
		{Value: "", IsMain: true},
	}

	et := NewEquivalenceTable(prefixes, suffixes)

	et.AskUserForWord("", "")

	// Пока таблица не угадана
	for !IsDone {
		// Создаем дополнение таблицы для префиксов
		for _, oldPrefix := range et.Prefixes {
			// Для каждого символа алфавита
			for _, letter := range alphabet {
				// Создаем новые префиксы
				prefix := Prefix{
					Value:  oldPrefix.Value + string(letter),
					IsMain: false,
				}
				// Предупреждаем дублирование
				if !et.ContainsMainPrefix(prefix) {
					et.AddPrefix(prefix)
					// Задаем вопросы MAT, заполняем таблицу
					for _, suffix := range et.Suffixes {
						et.AskUserForWord(prefix.Value, suffix.Value)
					}
				}
			}
		}

		// Печать таблицы
		et.PrintTable()
		IsDone = true
	}

}
