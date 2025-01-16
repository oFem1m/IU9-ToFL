from tabulate import tabulate


def generate_equivalence_table(limit=40, output_file="equivalence_table2.txt"):
    table = []
    row = []
    for y in range(1, limit):  # Проходим по y
        x = 3 * y - 2  # Вычисляем x из уравнения x + 2 = 3y
        if x > limit:
            break  # Останавливаем, если превысили лимит

        # Представляем x и y в двоичном формате
        x_bin = bin(x)[2:]
        y_bin = bin(y)[2:]

        diff = x - y
        bin_diff = bin(diff)[2:]

        row.append(f"({x}, {y}) -> ({x_bin}, {y_bin}), (diff: {diff}, {bin_diff})")

        if len(row) == 4:
            table.append(row)
            row = []

    if row:
        table.append(row)  # Добавляем оставшиеся элементы

    # Форматируем таблицу с помощью tabulate
    formatted_table = tabulate(table, headers=["01 - 01", "00 - 10", "11 - 11", "10 - 00"], tablefmt="grid")

    with open(output_file, "w") as f:
        f.write(formatted_table)

    print(f"Таблица сохранена в {output_file}")


# Генерируем таблицу до x = 40 и сохраняем в файл
generate_equivalence_table(limit=995)
