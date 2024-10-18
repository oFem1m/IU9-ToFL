package main

func main() {
	alphabet := "ab"

	IsDone := false

	prefixes := []Prefix{
		{Value: "", IsMain: true},
	}
	suffixes := []string{""}

	et := NewEquivalenceTable(prefixes, suffixes)

	// Пока таблица не угадана
	for !IsDone {
		// Заполняем пустые значения таблицы
		for _, prefix := range et.Prefixes {
			for _, suffix := range et.Suffixes {
				if et.Table[prefix.Value][suffix] == '0' {
					et.AskUserForWord(prefix.Value, suffix)
				}
			}
		}
		// Создаем дополнение таблицы для префиксов
		for _, oldPrefix := range et.Prefixes {
			// Для каждого символа алфавита
			for _, letter := range alphabet {
				// Создаем новые префиксы на основе главных префиксов
				if oldPrefix.IsMain {
					prefix := Prefix{
						Value:  oldPrefix.Value + string(letter),
						IsMain: false,
					}
					// Предупреждаем дублирование
					if !et.ContainsMainPrefix(prefix) {
						et.AddPrefix(prefix)
						// Задаем вопросы MAT, заполняем таблицу
						for _, suffix := range et.Suffixes {
							et.AskUserForWord(prefix.Value, suffix)
						}
					}
				}
			}
		}

		// Проверяем таблицу на полноту и приводим к полному виду
		et.CompleteTable()

		// отправляем таблицу MAT
		response := et.AskUserForTable()

		// Если угадали, то конец меняем флаг, иначе - добавляем новые суффиксы
		if response == "OK" {
			IsDone = true
		} else {
			for i := 0; i < len(response); i++ {
				et.AddSuffix(response[i:])
			}
		}

		// Печать таблицы
		et.PrintTable()

	}

}
