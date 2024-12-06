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
    Возвращаем:
      action[state][terminal] = список действий [('shift', s) или ('reduce', rule_i) или ('accept', None)]
      goto_table[state][nonterminal] = список состояний
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
                    print_step(stack, pos, lookahead, act, new_stack)

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
                            print_step(stack, pos, lookahead, act, new_stack2)
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
                    print_step(stack, pos, lookahead, act, stack)

        return results

def print_step(stack, pos, lookahead, action, new_stack):
    """
    Выводит информацию о шаге разбора.
    """
    print(f"Step: {pos}")
    print(f"Stack: {stack}")
    print(f"Lookahead: {lookahead}")
    print(f"Action: {action}")
    print(f"New Stack: {new_stack}")
    print("-" * 40)


class State:
    def __init__(self, name):
        self.name = name
    def __repr__(self):
        return f"State({self.name})"

class Transition:
    def __init__(self, current_state, input_symbol, stack_top, next_state, stack_action):
        # input_symbol: либо терминал, либо 'ε'
        # stack_top: ожидаемый символ стека (состояние)
        # stack_action: строка вида:
        #   'push:X' - положить состояние X
        #   'popN:X' - снять N состояний и положить X
        #   'pop' - снять одно состояние
        #   'ε' - ничего не делать
        self.current_state = current_state
        self.input_symbol = input_symbol
        self.stack_top = stack_top
        self.next_state = next_state
        self.stack_action = stack_action

    def __repr__(self):
        return f"Transition({self.current_state.name}, {self.input_symbol}, {self.stack_top}, {self.next_state.name}, {self.stack_action})"


class PushdownAutomaton:
    def __init__(self, states, alphabet, stack_alphabet, transitions, start_state, accept_states):
        self.states = states
        self.alphabet = alphabet
        self.stack_alphabet = stack_alphabet
        self.transitions = transitions
        self.start_state = start_state
        self.accept_states = accept_states


def build_pda_from_lr(action, goto_table, grammar):
    """
    Строим PDA по таблицам action/goto LR-анализатора.
    Идея:
    - Состояния PDA соответствуют LR-состояниям.
    - Стековые символы - тоже состояния (мы храним в стеке номера состояний).
    - Начальный стек содержит начальное состояние.
    - Для shift:
        читаем терминал input_symbol
        stack_action = "push:next_state"
        Переход: (current_state, input_symbol, stack_top=<current_state> -> next_state, stack_action)
    - Для reduce:
        нужно снять |body| символов стека, затем перейти в состояние goto.
        Переход по ε: stack_action='popN:next_state' где next_state - состояние после goto
    - Для accept:
        Переход по ε из текущего состояния в accept state без изменений.
        Или можно считать, что accept state - это состояние, у которого нет переходов дальше, просто принимаем.

    В случае конфликтов (несколько вариантов action) - добавляем все переходы.
    """

    # Определяем множество состояний
    # Состояния = числа от 0 до len(action)
    # Создадим объекты State
    states = [State(f"q{i}") for i in range(len(action))]
    # Алфавит входа - terminals + '$'
    alphabet = set(grammar.terminals) | {'$'}
    # stack_alphabet = множество состояний (названия?)
    # В данном случае стек символы = те же состояния.
    stack_alphabet = set(s.name for s in states)

    transitions = []
    accept_states = set()

    # Принимающее состояние можно создать дополнительное. Но в LR-парсере
    # accept - это действие в одном из состояний при входном символе '$'.
    # Можно просто считать, что состояние, в котором есть accept-действие,
    # переходит в специальное состояние-допустимости.
    # Создадим отдельное состояние принятия:
    accept_state = State("q_accept")
    states.append(accept_state)
    accept_states.add(accept_state)

    # Начальное состояние PDA будет соответствовать начальному LR-состоянию (обычно это 0).
    start_state = states[0]

    # Построим переходы
    # action[state][term] и goto_table[state][NT]
    # Для shift: (q_s, term, stack_top="q_s") -> q_ts, stack_action="push:q_ts"
    # Для reduce: нужно взять правило A->α и сделать pop(len(α)) затем push состояние goto[state_of_pop][A].
    #   Это ε-переход, stack_action="popN:q_goto"
    # Для accept: ε-переход в q_accept.

    # Для работы с rules создадим список для быстрого доступа:
    rule_list = grammar.rules

    for s_i in range(len(states)-1): # не берем q_accept
        # shift/reduce/accept actions
        for term, acts in action[s_i].items():
            for act_type, val in acts:
                if act_type == 'shift':
                    # shift в состояние val
                    next_st = states[val]
                    # Для shift: читаем symbol=term, stack_top=states[s_i], stack_action='push:q_val'
                    transitions.append(Transition(
                        current_state=states[s_i],
                        input_symbol=term,
                        stack_top=states[s_i].name,
                        next_state=next_st,
                        stack_action=f"push:{next_st.name}"
                    ))
                elif act_type == 'reduce':
                    # reduce по правилу val
                    (A,B) = rule_list[val]
                    # длина B - сколько pop
                    # после попа goto от состояния, на котором окажемся.
                    pop_len = len(B)

                    for s_j in range(len(states)-1):
                        # Для каждого состояния проверяем goto_table
                        if A in goto_table[s_j] and len(goto_table[s_j][A])>0:
                            for q_goto in goto_table[s_j][A]:
                                q_goto_st = states[q_goto]
                                transitions.append(Transition(
                                    current_state=states[s_i],
                                    input_symbol='ε',
                                    stack_top=states[s_i].name,
                                    next_state=q_goto_st,
                                    stack_action=f"popN:{pop_len}:{q_goto_st.name}"
                                ))

                elif act_type == 'accept':
                    # accept: переход в q_accept
                    transitions.append(Transition(
                        current_state=states[s_i],
                        input_symbol='ε',
                        stack_top=states[s_i].name,
                        next_state=accept_state,
                        stack_action='ε'
                    ))

        # goto действия отразим как epsilon переходы?
        # В LR парсере goto делается после reduce, уже учтено выше при reduce.

    pda = PushdownAutomaton(states, alphabet, stack_alphabet, transitions, start_state, accept_states)
    return pda

def parse_with_pda(pda, input_tokens):
    """
    Недетерминированный разбор по PDA:
    - Стек изначально содержит start_state.name (как символ стека).
    - Начинаем из start_state.
    - Для каждого шага:
      смотрим на переходы:
        - Если input_symbol совпадает с текущим токеном или 'ε', и stack_top совпадает,
          то можем сделать переход.
        - По stack_action корректируем стек.
      Если несколько переходов - ветвимся (backtracking).

    Возвращаем True, если хотя бы один путь ведет к состоянию принятия.
    """
    input_tokens = input_tokens + ['$']
    initial_config = (pda.start_state, tuple([pda.start_state.name]), 0)  # (current_state, stack, pos)
    # очередь для бэктрекинга
    queue = deque([initial_config])
    visited = set() # чтобы ограничить бесконечные циклы

    while queue:
        current_state, stack, pos = queue.pop()
        if current_state in pda.accept_states and pos == len(input_tokens):
            # достигли акцепта
            return True

        # Находим подходящие переходы
        current_input = input_tokens[pos] if pos < len(input_tokens) else None

        # На вершине стека:
        top = stack[-1] if stack else None

        # Подбираем переходы
        # переход подходит, если:
        #   (transition.input_symbol == current_input or 'ε')
        #   (transition.stack_top == top or 'ε')
        for t in pda.transitions:
            if t.current_state == current_state:
                # Проверяем символ входа
                if t.input_symbol == 'ε':
                    # нет потребности читать вход
                    can_read = True
                else:
                    can_read = (current_input == t.input_symbol)

                # Проверяем стек
                if t.stack_top == top or t.stack_top == 'ε':
                    if can_read:
                        # Применяем переход
                        new_stack = list(stack)
                        # stack_action
                        # варианты:
                        # 'push:X'
                        # 'pop'
                        # 'popN:N:X'
                        # 'ε'
                        sa = t.stack_action
                        if sa.startswith('push:'):
                            sym = sa.split(':',1)[1]
                            new_stack.append(sym)
                        elif sa == 'pop':
                            if new_stack:
                                new_stack.pop()
                            else:
                                continue
                        elif sa.startswith('popN:'):
                            # popN:count:state
                            parts = sa.split(':')
                            count = int(parts[1])
                            if len(new_stack) < count:
                                continue
                            # pop count times
                            for _ in range(count):
                                new_stack.pop()
                            # затем push new state
                            if len(parts) > 2:
                                new_sym = parts[2]
                                new_stack.append(new_sym)
                        elif sa == 'ε':
                            # ничего не делаем
                            pass

                        new_stack = tuple(new_stack)
                        new_pos = pos
                        if t.input_symbol != 'ε' and t.input_symbol is not None:
                            new_pos = pos + 1

                        new_config = (t.next_state, new_stack, new_pos)
                        if new_config not in visited:
                            visited.add(new_config)
                            queue.append(new_config)

    return False


if __name__ == "__main__":
    # Грамматика:
    grammar_rules = [
        "S -> a S b",
        "S -> c"
    ]

    # 1. Создаем грамматику
    grammar = Grammar(grammar_rules)
    # 2. Удаление левой рекурсии (упрощенно)
    # grammar.remove_left_recursion()
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

    parser = PDA(action, goto_table, grammar)
    # 7. Парсим по таблице:
    tokens = list("aacbb")  # входные токены
    parses = parser.parse_all(tokens)

    # Выведем все результаты
    print("Возможные пути парсинга:")
    for p in parses:
        print(p)

    # Попробуем распарсить строку по PDA
    pda = build_pda_from_lr(action, goto_table, grammar)
    tokens = list("aacbb")
    res = parse_with_pda(pda, tokens)
    print("Результат парсинга", res)
