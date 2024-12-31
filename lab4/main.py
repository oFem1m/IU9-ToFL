class RegexParserError(Exception):
    pass

class Token:
    def __init__(self, ttype, value=None):
        self.ttype = ttype
        self.value = value

    def __repr__(self):
        return f"Token({self.ttype}, {self.value})"

class Lexer:
    def __init__(self, text):
        self.text = text
        self.pos = 0

    def peek(self):
        if self.pos < len(self.text):
            return self.text[self.pos]
        return None

    def advance(self):
        self.pos += 1

    def tokenize(self):
        tokens = []
        while self.pos < len(self.text):
            ch = self.peek()

            if ch == '(':
                # Проверяем конструкции (?:  (?=  (?N)
                self.advance()
                nxt = self.peek()
                if nxt == '?':
                    # Это либо незахватывающая, либо опережающая проверка, либо (?N)
                    self.advance()
                    nxt2 = self.peek()
                    if nxt2 == ':':
                        # (?: )
                        self.advance()
                        tokens.append(Token('NONCAP_OPEN'))
                    elif nxt2 == '=':
                        # (?= )
                        self.advance()
                        tokens.append(Token('LOOKAHEAD_OPEN'))
                    elif nxt2 and nxt2.isdigit():
                        # (?N)
                        self.advance()  # перешли на цифру N
                        val = int(nxt2)
                        tokens.append(Token('EXPR_REF_OPEN', val))
                        # Закрывающая скобка будет обработана в парсере
                    else:
                        raise RegexParserError("Некорректный синтаксис после (?")
                else:
                    # Захватывающая группа ( ... )
                    tokens.append(Token('CAP_OPEN'))
            elif ch == ')':
                tokens.append(Token('CLOSE'))
                self.advance()
            elif ch == '|':
                tokens.append(Token('ALT'))
                self.advance()
            elif ch == '*':
                tokens.append(Token('STAR'))
                self.advance()
            elif ch and 'a' <= ch <= 'z':
                tokens.append(Token('CHAR', ch))
                self.advance()
            else:
                raise RegexParserError(f"Неизвестный символ: {ch}")
        return tokens


# Узлы AST
class GroupNode:
    def __init__(self, group_id, node):
        self.group_id = group_id
        self.node = node

    def __repr__(self):
        return f"GroupNode({self.group_id}, {self.node})"

class NonCapGroupNode:
    def __init__(self, node):
        self.node = node

    def __repr__(self):
        return f"NonCapGroupNode({self.node})"

class LookaheadNode:
    def __init__(self, node):
        self.node = node

    def __repr__(self):
        return f"LookaheadNode({self.node})"

class ConcatNode:
    def __init__(self, nodes):
        self.nodes = nodes

    def __repr__(self):
        return f"ConcatNode({self.nodes})"

class AltNode:
    def __init__(self, branches):
        self.branches = branches

    def __repr__(self):
        return f"AltNode({self.branches})"

class StarNode:
    def __init__(self, node):
        self.node = node

    def __repr__(self):
        return f"StarNode({self.node})"

class CharNode:
    def __init__(self, ch):
        self.ch = ch

    def __repr__(self):
        return f"CharNode('{self.ch}')"

class ExprRefNode:
    def __init__(self, ref_id):
        self.ref_id = ref_id

    def __repr__(self):
        return f"ExprRefNode({self.ref_id})"

class Parser:
    def __init__(self, tokens):
        self.tokens = tokens
        self.pos = 0
        self.group_count = 0
        self.max_groups = 9
        self.in_lookahead = False

        # Сохраняем определения групп, чтобы проверять корректность ссылок
        # group_id -> AST подграмматики
        self.groups_ast = {}

    def current_token(self):
        if self.pos < len(self.tokens):
            return self.tokens[self.pos]
        return None

    def eat(self, ttype=None):
        tok = self.current_token()
        if tok is None:
            raise RegexParserError("Неожиданный конец выражения")
        if ttype is not None and tok.ttype != ttype:
            raise RegexParserError(f"Ожидается {ttype}, найдено {tok.ttype}")
        self.pos += 1
        return tok

    def parse(self):
        node = self.parse_alternation()
        if self.current_token() is not None:
            # Если что-то осталось непрочитанное, синтаксическая ошибка
            raise RegexParserError("Лишние символы после корректного выражения")
        # Проверяем синтаксическую корректность, но не будем запрещать forward refs
        self.check_references(node, defined_groups=set())
        return node

    def parse_alternation(self):
        # alternation: concatenation ('|' concatenation)*
        branches = [self.parse_concatenation()]
        while self.current_token() and self.current_token().ttype == 'ALT':
            self.eat('ALT')
            if self.current_token() is None or self.current_token().ttype in ['CLOSE', 'ALT']:
                raise RegexParserError("Пустая альтернатива запрещена")
            branches.append(self.parse_concatenation())
        if len(branches) == 1:
            return branches[0]
        return AltNode(branches)

    def parse_concatenation(self):
        # concatenation: repetition+
        nodes = []
        while self.current_token() and self.current_token().ttype not in ['CLOSE', 'ALT']:
            nodes.append(self.parse_repetition())
        if len(nodes) == 1:
            return nodes[0]
        return ConcatNode(nodes)

    def parse_repetition(self):
        # repetition: base ('*')?
        node = self.parse_base()
        while self.current_token() and self.current_token().ttype == 'STAR':
            self.eat('STAR')
            node = StarNode(node)
        return node

    def parse_base(self):
        tok = self.current_token()
        if tok is None:
            raise RegexParserError("Неожиданный конец при ожидании базового выражения")

        if tok.ttype == 'CAP_OPEN':
            # ( ... )
            self.eat('CAP_OPEN')
            self.group_count += 1
            if self.group_count > self.max_groups:
                raise RegexParserError("Превышено число групп захвата > 9")
            group_id = self.group_count
            node = self.parse_alternation()
            self.eat('CLOSE')
            # Сохраняем AST группы
            self.groups_ast[group_id] = node
            return GroupNode(group_id, node)

        elif tok.ttype == 'NONCAP_OPEN':
            # (?: ... )
            self.eat('NONCAP_OPEN')
            node = self.parse_alternation()
            self.eat('CLOSE')
            return NonCapGroupNode(node)

        elif tok.ttype == 'LOOKAHEAD_OPEN':
            # (?= ... )
            if self.in_lookahead:
                raise RegexParserError("Вложенные опережающие проверки запрещены")
            self.eat('LOOKAHEAD_OPEN')
            old_look = self.in_lookahead
            self.in_lookahead = True
            node = self.parse_alternation()
            self.in_lookahead = old_look
            self.eat('CLOSE')
            return LookaheadNode(node)

        elif tok.ttype == 'EXPR_REF_OPEN':
            # (?N)
            ref_id = tok.value
            self.eat('EXPR_REF_OPEN')
            self.eat('CLOSE')
            return ExprRefNode(ref_id)

        elif tok.ttype == 'CHAR':
            ch = tok.value
            self.eat('CHAR')
            return CharNode(ch)

        else:
            raise RegexParserError(f"Некорректный токен: {tok}")

    def check_references(self, node, defined_groups):
        # Проверка корректности:
        # 1) ExprRefNode(ref_id): ref_id должна быть определена (ref_id in defined_groups)
        # 2) GroupNode: после неё добавить её номер
        # 3) LookaheadNode: внутри не может быть групп захвата, других lookahead
        # PS:
        # Не запрещаем forward references, рекурсия допустима
        # Просто проверим отсутствия lookahead внутри lookahead или групп захвата в нём.
        if isinstance(node, CharNode):
            return defined_groups

        elif isinstance(node, ExprRefNode):
            # Не выдаём ошибку при forward ref
            return defined_groups

        elif isinstance(node, GroupNode):
            # Внутри группы сначала проверяем содержимое
            new_defined = self.check_references(node.node, defined_groups)
            # После конца группы эта группа считается определённой
            new_defined = set(new_defined)
            new_defined.add(node.group_id)
            return new_defined

        elif isinstance(node, NonCapGroupNode):
            return self.check_references(node.node, defined_groups)

        elif isinstance(node, LookaheadNode):
            # Проверить, что внутри нет групп захвата и нет других lookahead
            self.check_no_cap_and_lookahead(node.node, inside_lookahead=True)
            # Ссылки на группы должны быть из уже определённых
            return self.check_references(node.node, defined_groups)

        elif isinstance(node, StarNode):
            return self.check_references(node.node, defined_groups)

        elif isinstance(node, ConcatNode):
            cur_defined = defined_groups
            for child in node.nodes:
                cur_defined = self.check_references(child, cur_defined)
            return cur_defined

        elif isinstance(node, AltNode):
            # Изначально было пересечение, теперь делаем объединение,
            # чтобы ситуации вроде (a|(bb))(a|(?2)) были корректными.
            all_defs = []
            for branch in node.branches:
                branch_defs = self.check_references(branch, defined_groups)
                all_defs.append(branch_defs)
            union_defs = set()
            for d in all_defs:
                union_defs.update(d)
            return union_defs

        else:
            raise RegexParserError("Неизвестный тип узла AST при проверке ссылок")

    def check_no_cap_and_lookahead(self, node, inside_lookahead):
        """
        Проверяем, что внутри lookahead нет захватывающих групп и lookahead.
        """
        if isinstance(node, GroupNode) and inside_lookahead:
            raise RegexParserError("Внутри опережающей проверки не допускаются захватывающие группы")
        if isinstance(node, LookaheadNode) and inside_lookahead:
            raise RegexParserError("Внутри опережающей проверки не допускаются другие опережающие проверки")

        if isinstance(node, (NonCapGroupNode, LookaheadNode, StarNode, ConcatNode, AltNode)):
            # Рекурсивно проверяем для детей
            if isinstance(node, NonCapGroupNode):
                self.check_no_cap_and_lookahead(node.node, inside_lookahead)
            elif isinstance(node, LookaheadNode):
                self.check_no_cap_and_lookahead(node.node, inside_lookahead)
            elif isinstance(node, StarNode):
                self.check_no_cap_and_lookahead(node.node, inside_lookahead)
            elif isinstance(node, ConcatNode):
                for n in node.nodes:
                    self.check_no_cap_and_lookahead(n, inside_lookahead)
            elif isinstance(node, AltNode):
                for b in node.branches:
                    self.check_no_cap_and_lookahead(b, inside_lookahead)

# ------------------------------------------------------------
# Построение каркасной КС-грамматики
# ------------------------------------------------------------

class CFGBuilder:
    def __init__(self, groups_ast):
        # groups_ast: {group_id: node}
        self.groups_ast = groups_ast
        self.group_nonterm = {}
        self.noncap_index = 1
        self.star_index = 1

    def build(self, node):
        start = 'S'
        rules = {}

        # Регистрируем G1 для группы 1
        main_nt = self.node_to_cfg(node, rules)

        # Добавляем правило для S -> main_nt (группа 1)
        rules[start] = [[main_nt]]

        # Убедимся, что все группы зарегистрированы
        for group_id, ast in self.groups_ast.items():
            if group_id not in self.group_nonterm:
                nt = f"G{group_id}"
                self.group_nonterm[group_id] = nt
                self.node_to_cfg(ast, rules, start_symbol=nt)

        return start, rules

    def node_to_cfg(self, node, rules, start_symbol=None):
        """
        Преобразует узел AST в нетерминал CFG.
        Возвращает имя нетерминала.
        rules: dict {NonTerminal: [ [symbols], [symbols] ... ]}
        """
        if isinstance(node, CharNode):
            # Терминальный символ
            # Если нам дали start_symbol, используем его как имя нетерминала, иначе генерим
            nt = start_symbol if start_symbol else self.fresh_nt('CHAR')
            rules.setdefault(nt, []).append([node.ch])
            return nt

        elif isinstance(node, GroupNode):
            nt = self.group_nonterm.get(node.group_id)
            if nt is None:
                nt = f"G{node.group_id}"
                self.group_nonterm[node.group_id] = nt
            # Строим внутреннее правило для содержимого группы
            sub_nt = self.node_to_cfg(node.node, rules)
            # G{group_id} -> sub_nt
            rules.setdefault(nt, []).append([sub_nt])
            return nt

        elif isinstance(node, NonCapGroupNode):
            # Генерируем новый нетерминал для незахватывающей группы
            nt = start_symbol if start_symbol else self.fresh_nt('N')
            sub_nt = self.node_to_cfg(node.node, rules)
            # Правило: nt -> sub_nt
            rules.setdefault(nt, []).append([sub_nt])
            return nt

        elif isinstance(node, LookaheadNode):
            # Опережающую проверку не представляем в CFG (по условию), заменим ее на ε
            nt = start_symbol if start_symbol else self.fresh_nt('LA')
            rules.setdefault(nt, []).append([])
            return nt

        elif isinstance(node, ConcatNode):
            nt = start_symbol if start_symbol else self.fresh_nt('C')
            # Конкатенация: nodes - список узлов
            # Преобразуем каждый узел в нетерминал, потом nt -> seq
            # Если узел терминальный, node_to_cfg вернёт нетерминал с одним правилом
            seq_nts = [self.node_to_cfg(ch, rules) for ch in node.nodes]
            rules.setdefault(nt, []).append(seq_nts)
            return nt

        elif isinstance(node, AltNode):
            nt = start_symbol if start_symbol else self.fresh_nt('A')
            # Альтернатива: для каждой ветви генерируем правило
            for branch in node.branches:
                br_nt = self.node_to_cfg(branch, rules)
                # nt -> br_nt
                rules.setdefault(nt, []).append([br_nt])
            return nt

        elif isinstance(node, StarNode):
            # Звёздочка: X* означает 0 или более повторений X
            # Создаём нетерминал для звёздочки
            nt = start_symbol if start_symbol else self.fresh_nt('R')
            sub_nt = self.node_to_cfg(node.node, rules)
            # R -> ε | R sub_nt
            rules.setdefault(nt, []).append([])          # ε
            rules[nt].append([nt, sub_nt])
            return nt

        elif isinstance(node, ExprRefNode):
            # Ссылка на выражение группы
            ref_id = node.ref_id
            if ref_id not in self.group_nonterm:
                self.group_nonterm[ref_id] = f"G{ref_id}"
                if ref_id not in self.groups_ast:
                    raise RegexParserError(f"Ссылка на несуществующую группу {ref_id}")
                # Строим правила для группы ref_id
                sub_nt = self.node_to_cfg(self.groups_ast[ref_id], rules)
                nt = self.group_nonterm[ref_id]
                # sub_nt уже построен выше:
                rules.setdefault(nt, []).append([sub_nt])
            return self.group_nonterm[ref_id]

        else:
            raise RegexParserError("Неизвестный тип узла при построении грамматики")

    def fresh_nt(self, prefix):
        # Генерируем уникальный нетерминал
        # Можно использовать простые счётчики
        if prefix == 'N':
            name = f"N{self.noncap_index}"
            self.noncap_index += 1
            return name
        elif prefix == 'R':
            name = f"R{self.star_index}"
            self.star_index += 1
            return name
        else:
            # Для простоты: добавим общий счётчик
            name = f"{prefix}{self.noncap_index + self.star_index}"
            self.noncap_index += 1
            return name


def main():
    text = input().strip()
    try:
        if text == "":
            raise RegexParserError("Пустая строка")
        # Разбиваем на лексеммы
        lexer = Lexer(text)
        tokens = lexer.tokenize()

        # Строим дерево
        parser = Parser(tokens)
        ast = parser.parse()

        # Построим КС-грамматику
        builder = CFGBuilder(parser.groups_ast)
        start_symbol, rules = builder.build(ast)
        print("Выражение корректно синтаксически и удовлетворяет ограничениям.")
        print("Построенная КС-грамматика (каркас):")
        print("Начальный нетерминал:", start_symbol)
        for nt in rules:
            for rhs in rules[nt]:
                rhs_str = " ".join(rhs) if rhs else "ε"
                print(f"{nt} -> {rhs_str}")
        print("=" * 60)
        for token in tokens:
            print(token)
        print("=" * 60)
        print(ast)

    except RegexParserError as e:
        print("Ошибка:", e)

if __name__ == "__main__":
    main()
