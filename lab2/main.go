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
		for _, OldPrefix := range et.Prefixes {
			for _, letter := range alphabet {
				prefix := Prefix{
					Value:  OldPrefix.Value + string(letter),
					IsMain: false,
				}
				if !et.ContainsMainPrefix(prefix) {
					et.AddPrefix(prefix)
				}
			}
		}

		// Печать таблицы
		et.PrintTable()
		IsDone = true
	}

}
