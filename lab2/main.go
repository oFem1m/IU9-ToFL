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
				// Если ячейка пуста
				if et.GetValue(prefix.Value, suffix) == '0' {
					word := prefix.Value + suffix
					// Проверяем наличие слова в словаре
					if et.CheckWord(word) {
						// Если слово принадлежит языку
						if et.Words[word] {
							et.Update(prefix.Value, suffix, '+')
						} else {
							et.Update(prefix.Value, suffix, '-')
						}
					} else { // Иначе спрашиваем
						if et.AskUserForWord(word) {
							if et.Words[word] {
								et.Update(prefix.Value, suffix, '+')
							} else {
								et.Update(prefix.Value, suffix, '-')
							}
						}
					}
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
					// Если префикс удалось добавить
					if et.AddPrefix(prefix) {
						// По необходимости задаём вопросы MAT, заполняем таблицу
						for _, suffix := range et.Suffixes {
							word := prefix.Value + suffix
							// Проверяем наличие слова в словаре
							if et.CheckWord(word) {
								// Если слово принадлежит языку
								if et.Words[word] {
									et.Update(prefix.Value, suffix, '+')
								} else {
									et.Update(prefix.Value, suffix, '-')
								}
							} else { // Иначе спрашиваем
								if et.AskUserForWord(word) {
									if et.Words[word] {
										et.Update(prefix.Value, suffix, '+')
									} else {
										et.Update(prefix.Value, suffix, '-')
									}
								}
							}
						}
					}
				}
			}
		}

		// Проверяем таблицу на полноту и приводим к полному виду
		et.CompleteTable()

		// Проверка, являются ли все префиксы главными
		if !et.AreAllPrefixesMain() {
			inconsistency := true
			for inconsistency {
				if et.InconsistencyTable(alphabet) {
					// Заполняем пустые значения таблицы
					for _, prefix := range et.Prefixes {
						for _, suffix := range et.Suffixes {
							// Если ячейка пуста
							if et.GetValue(prefix.Value, suffix) == '0' {
								word := prefix.Value + suffix
								// Проверяем наличие слова в словаре
								if et.CheckWord(word) {
									// Если слово принадлежит языку
									if et.Words[word] {
										et.Update(prefix.Value, suffix, '+')
									} else {
										et.Update(prefix.Value, suffix, '-')
									}
								} else { // Иначе спрашиваем
									if et.AskUserForWord(word) {
										if et.Words[word] {
											et.Update(prefix.Value, suffix, '+')
										} else {
											et.Update(prefix.Value, suffix, '-')
										}
									}
								}
							}
						}
					}
				} else {
					inconsistency = false
				}
			}
			// отправляем таблицу MAT
			response := et.AskUserForTable()
			// Если угадали, то конец меняем флаг, иначе - добавляем новые суффиксы
			if response == "OK" {
				IsDone = true
			} else {
				et.Words[response] = true
				for i := 0; i < len(response); i++ {
					et.AddSuffix(response[i:])
				}
			}
		}
	}
}
