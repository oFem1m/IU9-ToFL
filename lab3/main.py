"""
Задача: Реализовать PDA (pushdown automaton) на основе LR(0) + реализовать парсинг по полученному PDA.

Данный код демонстрирует:
1. Считывание и преобразование входной CFG.
2. Возможную обработку левой рекурсии (подготовка грамматики).
3. Построение LR(0)-автомата:
   - Генерация множества ситуаций (items)
   - Построение NFA
   - Построение DFA из NFA (замыкание по e-переходам)
4. Формирование управляющей таблицы (action/goto).
5. Создание PDA на основе LR(0)-автомата.
6. Реализацию парсера по PDA, способного обрабатывать недетерминизм
   (при конфликтах будут рассматриваться разные пути).

Использование:
- Определить грамматику в `grammar_rules`.
- Запустить. Пример приведен для простой грамматики.
"""

from collections import defaultdict, deque


class Grammar:
    def __init__(self, rules):
        """
        rules: список строк вида:
           'S -> A B', 'A -> a', ...
        """
        self.raw_rules = rules
        self.rules = []  # список (head, [body_symbols])
        self.nonterminals = set()
        self.terminals = set()
        self.start_symbol = None
        self._parse_rules()

    def _parse_rules(self):
        for line in self.raw_rules:
            line = line.strip()
            if not line:
                continue
            if '->' not in line:
                continue
            lhs, rhs = line.split('->', 1)
            lhs = lhs.strip()
            if self.start_symbol is None:
                self.start_symbol = lhs
            rhs_symbols = rhs.strip().split()
            self.rules.append((lhs, rhs_symbols))

        # Определим терминалы и нетерминалы
        # Предполагаем: нетерминалы - большие буквы, терминалы - маленькие буквы
        for head, body in self.rules:
            self.nonterminals.add(head)
            for sym in body:
                if sym.islower():
                    self.terminals.add(sym)
                else:
                    self.nonterminals.add(sym)

    def remove_left_recursion(self):
        """
        Очень примитивное удаление непосредственной левой рекурсии.
        Можно будет улучшить.
        """
        # Алгоритм:
        # Для каждого нетерминала A:
        #   Разбить правила A->Aα|β на A->βA' и A'->αA'|ε
        changed = True
        while changed:
            changed = False
            new_rules = []
            to_add = []
            for nt in self.nonterminals:
                # Правила для нетерминала nt
                current_rules = [(h, b) for (h, b) in self.rules if h == nt]
                # Проверим, есть ли левая рекурсия
                left_rec = [r for r in current_rules if r[1][0] == nt]
                if not left_rec:
                    # нет левой рекурсии
                    for r in current_rules:
                        if r not in new_rules:
                            new_rules.append(r)
                    continue
                # Удаляем текущие правила для nt из списка
                for r in current_rules:
                    self.rules.remove(r)
                # Устранение левой рекурсии
                # nt -> nt α | β
                # Вводим nt'
                nt_new = nt + "'"
                while nt_new in self.nonterminals:
                    nt_new += "'"
                self.nonterminals.add(nt_new)
                beta_rules = [r for r in current_rules if r[1][0] != nt]
                alpha_rules = [r for r in current_rules if r[1][0] == nt]
                betas = [r[1] for r in beta_rules]
                alphas = [r[1][1:] for r in alpha_rules]  # убираем первый символ nt
                # nt -> β nt'
                for b in betas:
                    new_rules.append((nt, b + [nt_new]))
                # nt' -> α nt' | ε
                for a in alphas:
                    new_rules.append((nt_new, a + [nt_new]))
                new_rules.append((nt_new, ['ε']))
                changed = True
            # Добавляем остальные правила (неизмененные)
            other_rules = [r for r in self.rules if r not in new_rules]
            self.rules = new_rules + other_rules

        # Удаляем ε из терминалов
        if 'ε' in self.terminals:
            self.terminals.remove('ε')

    def augment_grammar(self):
        """
        Добавляем искусственный стартовый символ S' -> S
        """
        new_start = self.start_symbol + "'"
        while new_start in self.nonterminals:
            new_start += "'"
        self.nonterminals.add(new_start)
        self.rules.insert(0, (new_start, [self.start_symbol]))
        self.start_symbol = new_start

    def get_rules_for(self, nonterminal):
        return [r for r in self.rules if r[0] == nonterminal]


class LR0Item:
    def __init__(self, head, body, dot_pos=0, rule_index=None):
        self.head = head
        self.body = body
        self.dot_pos = dot_pos
        self.rule_index = rule_index  # индекс правила в списке грамматики (для удобства)

    def __eq__(self, other):
        return (self.head, self.body, self.dot_pos) == (other.head, other.body, other.dot_pos)

    def __hash__(self):
        return hash((self.head, tuple(self.body), self.dot_pos))

    def __repr__(self):
        b = self.body[:]
        b.insert(self.dot_pos, '•')
        return f"{self.head} -> {' '.join(b)}"

    def next_symbol(self):
        if self.dot_pos < len(self.body):
            return self.body[self.dot_pos]
        return None

    def is_complete(self):
        return self.dot_pos == len(self.body)


def closure(items, grammar):
    """Находим замыкание для множества LR(0)-ситуаций"""
    changed = True
    closure_set = set(items)
    while changed:
        changed = False
        new_items = set()
        for it in closure_set:
            X = it.next_symbol()
            if X and X in grammar.nonterminals:
                # Добавляем ситуации для всех правил X -> ...
                for (h, b) in grammar.get_rules_for(X):
                    new_item = LR0Item(h, b, 0)
                    if new_item not in closure_set:
                        new_items.add(new_item)
        if new_items:
            closure_set |= new_items
            changed = True
    return closure_set


def goto(items, symbol, grammar):
    """Функция GOTO для LR(0)"""
    moved = [LR0Item(it.head, it.body, it.dot_pos + 1, it.rule_index)
             for it in items if it.next_symbol() == symbol]
    return closure(moved, grammar) if moved else set()


def build_lr0_automaton(grammar):
    """
    Строим LR(0)-автомат (DFA) из грамматики:
    1. Начальное состояние - closure(S' -> • S)
    2. Для каждого состояния и символа считаем goto.
    """
    # индекс правил для удобства
    for i, (h, b) in enumerate(grammar.rules):
        pass

    start_item = LR0Item(grammar.rules[0][0], grammar.rules[0][1], 0, 0)
    I0 = closure([start_item], grammar)
    states = [I0]
    transitions = []
    done = False
    while True:
        new_states_found = False
        for i, st in enumerate(states):
            # определим все символы, по которым есть переход
            symbols = set(it.next_symbol() for it in st if it.next_symbol())
            for sym in symbols:
                g = goto(st, sym, grammar)
                if g and g not in states:
                    states.append(g)
                    if (i, sym, len(states) -1 ) not in transitions:
                        transitions.append((i, sym, len(states) - 1))
                    new_states_found = True
                elif g:
                    # состояние уже известно
                    j = states.index(g)
                    if (i, sym, j) not in transitions:
                        transitions.append((i, sym, j))
        if not new_states_found:
            break

    # возвращаем детерминированный автомат
    return states, transitions


def build_parsing_table(states, transitions, grammar):
    """
    Строим управляющую таблицу (action/goto) для LR(0).
    Если есть конфликты - таблица будет недетерминированной.
    Представим таблицу в виде:
      action[state][terminal] = [('shift', next_state), ('reduce', rule_i), ...]
    Аналогично для goto.
    """
    # Создадим индекс правил для reduce
    rule_index = {}
    for i, (h, b) in enumerate(grammar.rules):
        rule_index[(h, tuple(b))] = i

    action = defaultdict(lambda: defaultdict(list))
    goto_table = defaultdict(lambda: defaultdict(list))

    # Собираем переходы по терминалам и нетерминалам
    # Состояния - items наборы
    # Если item вида A->α•, добавляем reduce
    for s_i, st in enumerate(states):
        for it in st:
            if it.is_complete():
                # reduce по правилу it
                i = rule_index[(it.head, tuple(it.body))]
                # Если это стартовое правило S'->S, то accept
                if it.head == grammar.start_symbol and it.is_complete():
                    action[s_i]['$'].append(('accept', None))
                else:
                    # reduce по всем терминалам?
                    # В LR(0) не учитываем follow, добавляем reduce на все терминалы.
                    # Но для корректности здесь можно ограничиться терминалами + $
                    for term in list(grammar.terminals) + ['$']:
                        action[s_i][term].append(('reduce', i))

    # transitions: (from_state, symbol, to_state)
    # Если symbol - терминал -> shift в action
    # Если symbol - нетерминал -> goto
    for (fs, sym, ts) in transitions:
        if sym in grammar.terminals:
            action[fs][sym].append(('shift', ts))
        else:
            goto_table[fs][sym].append(ts)

    return action, goto_table

def print_parsing_table(action, goto_table, grammar):
    """
    Выводит управляющую таблицу в удобном для понимания виде.
    """
    # Выводим заголовок таблицы
    print("State\t| Action\t\t\t\t\t| Goto")
    print("-" * 80)

    # Получаем все состояния
    states = set(action.keys()).union(set(goto_table.keys()))
    states = sorted(states)

    # Получаем все терминалы и нетерминалы
    terminals = sorted(grammar.terminals) + ['$']
    nonterminals = sorted(grammar.nonterminals)

    # Выводим строки для каждого состояния
    for state in states:
        action_str = "\t".join([f"{term}: {', '.join([f'{act[0]} {act[1]}' for act in action[state][term]])}" for term in terminals if action[state][term]])
        goto_str = "\t".join([f"{nt}: {', '.join(map(str, goto_table[state][nt]))}" for nt in nonterminals if goto_table[state][nt]])
        print(f"{state}\t\t| {action_str}\t| {goto_str}")


class PDA:
    """
    PDA на основе LR(0)-автомата.
    Состояния PDA будут соответствовать состояниям LR(0)-DFA.
    Стек будет содержать состояния.
    Вход - последовательность терминалов.
    Таблицы action/goto определяют переходы.
    """

    def __init__(self, action, goto_table, grammar):
        self.action = action
        self.goto = goto_table
        self.grammar = grammar

    def parse_all(self, tokens):
        """
        Недетерминированный парсер:
        При возникновении конфликтов будем запускать несколько ветвей разбора.

        tokens - список терминалов. Добавим в конец '$'.

        Возвращает список всех возможных деревьев разбора (или просто успешных результатов).
        Для простоты в дереве разбора будем возвращать структуру: (head, [children])
        где children либо терминалы, либо подобные структуры.
        """
        tokens = tokens + ['$']
        # Будем делать бэктрекинг:
        # Элемент стека для поиска: (stack, pos_in_tokens, partial_tree)
        # stack: список состояний
        # partial_tree: стек для построения дерева (будем хранить пары (symbol, children))
        # При reduce будем заменять несколько символов на один нетерминал.

        # Для удобства: на стеке будем хранить (state, node), где node - уже построенная часть дерева.
        # node для состояния - это None, для shift терминала - (token), для reduce - (head, [children])

        initial_stack = [(0, None)]
        queue = deque()
        # Будем хранить варианты: (stack, pos, forest)
        # forest - стек с синтаксическими узлами (один к одному со стеками состояний)
        # но можно хранить вместе, чтобы после reduce было проще восстановить
        queue.append((initial_stack, 0))
        results = []

        while queue:
            stack, pos = queue.pop()
            state = stack[-1][0]
            lookahead = tokens[pos] if pos < len(tokens) else None

            # Получаем список возможных действий
            possible_actions = []
            if lookahead in self.action[state]:
                possible_actions += self.action[state][lookahead]
            # Если нет lookahead в action, может быть reduce на epsilon?
            # LR(0) не предполагает epsilon reduce без lookahead, но если бы...

            if not possible_actions and lookahead is None:
                # конец входа и нет действий - неудача
                continue

            if not possible_actions and lookahead not in self.action[state]:
                # нет действий для этого lookahead - ошибка
                continue

            # Для каждого действия создадим новую ветку
            for act in possible_actions:
                atype = act[0]
                if atype == 'shift':
                    next_state = act[1]
                    # shift: добавляем терминал в дерево
                    new_stack = stack[:]
                    new_stack.append((next_state, (lookahead, [])))  # терминал как (symbol, [])
                    queue.append((new_stack, pos + 1))

                elif atype == 'reduce':
                    rule_i = act[1]
                    head, body = self.grammar.rules[rule_i]
                    # снимаем len(body) элементов со стека
                    if body == ['ε']:
                        # Epsilon правило
                        children = []
                        new_stack = stack[:]
                    else:
                        if len(stack) < len(body) + 1:
                            # ошибка, некорректное состояние стека
                            continue
                        popped = stack[-len(body):]
                        new_stack = stack[:-len(body)]
                        children = [n for (st, n) in popped]

                    # Формируем новый узел дерева
                    new_node = (head, children)

                    # теперь goto от new_stack[-1][0] по head
                    prev_state = new_stack[-1][0]
                    if head in self.goto[prev_state]:
                        possible_goto = self.goto[prev_state][head]
                        # может быть несколько переходов goto
                        for g_st in possible_goto:
                            new_stack2 = new_stack[:]
                            new_stack2.append((g_st, new_node))
                            queue.append((new_stack2, pos))
                    else:
                        # нет перехода, ошибка
                        continue

                elif atype == 'accept':
                    # Принятие разбора: последнее правило должно собрать всё дерево
                    # На вершине стека должен быть (S', [S]) - дерево
                    if len(stack) == 2 and stack[-1][1] is not None:
                        # stack[-1][1] - узел дерева для стартового символа
                        results.append(stack[-1][1])
                    else:
                        # Возможно нужна проверка, что это точно стартовое правило
                        if stack[-1][1] is not None:
                            results.append(stack[-1][1])

        return results


if __name__ == "__main__":
    # Пример грамматики:
    # S -> A A
    # A -> a A | b
    # grammar_rules = [
    #     "S -> A A",
    #     "A -> a A",
    #     "A -> b"
    # ]

    grammar_rules = [
        "S -> a S b",
        "S -> c"
    ]

    # 1. Создаем грамматику
    grammar = Grammar(grammar_rules)
    # 2. Удаление левой рекурсии (упрощенный)
    grammar.remove_left_recursion()
    # 3. Добавление стартового символа
    grammar.augment_grammar()

    # 4. Строим LR(0)-автомат
    print("LR(0)-автомат:")
    states, transitions = build_lr0_automaton(grammar)
    print("Состояния:")
    for state in states:
        print(state)
    print("Переходы:")
    for transition in transitions:
        print(transition)

    # 5. Строим парсинг-таблицу
    action, goto_table = build_parsing_table(states, transitions, grammar)
    print_parsing_table(action, goto_table, grammar)

    # 6. Создаем PDA
    pda = PDA(action, goto_table, grammar)

    # 7. Парсим вход:
    # К примеру, строка: a a b b соответствует цепочке A A => (a A)(b)
    # Попробуем: a b b (по сути S->A A->aA A->b| A->b)
    tokens = list("aacbb")  # входные токены
    parses = pda.parse_all(tokens)

    # Выведем все результаты
    print("Возможные пути парсинга:")
    for p in parses:
        print(p)
