# Исключение для ошибок
struct RegexParserError <: Exception
    msg::String
end

# Тип для представления токена
struct Token
    ttype::String
    value::Union{Nothing, Int, String}
end

function Base.show(io::IO, t::Token)
    print(io, "Token(", t.ttype, ", ", t.value, ")")
end

# Лексер
mutable struct Lexer
    text::String
    pos::Int
end

# Конструктор
function Lexer(text::String)
    return Lexer(text, 1)
end

# Получить текущий символ
function peek(lexer::Lexer)
    return lexer.pos <= length(lexer.text) ? lexer.text[lexer.pos] : nothing
end

# Продвинуть указатель
function advance!(lexer::Lexer)
    lexer.pos += 1
end

# Основной метод для токенизации
function tokenize(lexer::Lexer)
    tokens = Token[]
    while lexer.pos <= length(lexer.text)
        ch = peek(lexer)

        if ch == '('
            advance!(lexer)
            nxt = peek(lexer)
            if nxt == '?'
                advance!(lexer)
                nxt2 = peek(lexer)
                if nxt2 == ':'
                    advance!(lexer)
                    push!(tokens, Token("NONCAP_OPEN", nothing))
                elseif nxt2 == '='
                    advance!(lexer)
                    push!(tokens, Token("LOOKAHEAD_OPEN", nothing))
                elseif isnumeric(nxt2)
                    advance!(lexer)
                    val = Base.parse(Int, string(nxt2))
                    push!(tokens, Token("EXPR_REF_OPEN", val))
                else
                    throw(RegexParserError("Некорректный синтаксис после (?"))
                end
            else
                push!(tokens, Token("CAP_OPEN", nothing))
            end
        elseif ch == ')'
            push!(tokens, Token("CLOSE", nothing))
            advance!(lexer)
        elseif ch == '|'
            push!(tokens, Token("ALT", nothing))
            advance!(lexer)
        elseif ch == '*'
            push!(tokens, Token("STAR", nothing))
            advance!(lexer)
        elseif ch >= 'a' && ch <= 'z'
            push!(tokens, Token("CHAR", string(ch)))
            advance!(lexer)
        else
            throw(RegexParserError("Неизвестный символ: $ch"))
        end
    end
    return tokens
end

# Узлы AST

# Узел группы (захватывающей)
struct GroupNode
    group_id::Int
    node
end

function Base.show(io::IO, node::GroupNode)
    print(io, "GroupNode(", node.group_id, ", ", node.node, ")")
end

# Узел незахватывающей группы
struct NonCapGroupNode
    node
end

function Base.show(io::IO, node::NonCapGroupNode)
    print(io, "NonCapGroupNode(", node.node, ")")
end

# Узел опережающей проверки
struct LookaheadNode
    node
end

function Base.show(io::IO, node::LookaheadNode)
    print(io, "LookaheadNode(", node.node, ")")
end

# Узел конкатенации
struct ConcatNode
    nodes::Vector
end

function Base.show(io::IO, node::ConcatNode)
    print(io, "ConcatNode(", node.nodes, ")")
end

# Узел альтернативы
struct AltNode
    branches::Vector
end

function Base.show(io::IO, node::AltNode)
    print(io, "AltNode(", node.branches, ")")
end

# Узел повторения (*)
struct StarNode
    node
end

function Base.show(io::IO, node::StarNode)
    print(io, "StarNode(", node.node, ")")
end

# Узел символа
struct CharNode
    ch::String
end

function Base.show(io::IO, node::CharNode)
    print(io, "CharNode('", node.ch, "')")
end

# Узел ссылки на выражение (?N)
struct ExprRefNode
    ref_id::Int
end

function Base.show(io::IO, node::ExprRefNode)
    print(io, "ExprRefNode(", node.ref_id, ")")
end

# ------------------------------------------------
# Парсер
# ------------------------------------------------
mutable struct Parser
    tokens::Vector{Token}
    pos::Int
    group_count::Int
    max_groups::Int
    in_lookahead::Bool
    groups_ast::Dict{Int, Any}  # Определение групп: group_id -> AST
end

# Конструктор
function Parser(tokens::Vector{Token})
    return Parser(tokens, 1, 0, 9, false, Dict())
end

function current_token(parser::Parser)
    parser.pos <= length(parser.tokens) ? parser.tokens[parser.pos] : nothing
end

function eat!(parser::Parser, ttype::String = nothing)
    tok = current_token(parser)
    if isnothing(tok)
        throw(RegexParserError("Неожиданный конец выражения"))
    end
    if !isnothing(ttype) && tok.ttype != ttype
        throw(RegexParserError("Ожидается $ttype, найдено $(tok.ttype)"))
    end
    parser.pos += 1
    return tok
end

# Основной метод парсинга
function parse_ast(parser::Parser)
    node = parse_alternation(parser)
    if !isnothing(current_token(parser))
        throw(RegexParserError("Лишние символы после корректного выражения"))
    end
    return node
end

function parse_alternation(parser::Parser)
    branches = Any[parse_concatenation(parser)]
    while !isnothing(current_token(parser)) && current_token(parser).ttype == "ALT"
        eat!(parser, "ALT")
        push!(branches, parse_concatenation(parser))
    end
    return length(branches) == 1 ? branches[1] : AltNode(branches)
end

function parse_concatenation(parser::Parser)
    nodes = Any[]
    while !isnothing(current_token(parser)) && !(current_token(parser).ttype in ["CLOSE", "ALT"])
        push!(nodes, parse_repetition(parser))
    end
    return length(nodes) == 1 ? nodes[1] : ConcatNode(nodes)
end


function parse_repetition(parser::Parser)
    node = parse_base(parser)
    while !isnothing(current_token(parser)) && current_token(parser).ttype == "STAR"
        eat!(parser, "STAR")
        node = StarNode(node)
    end
    return node
end

function parse_base(parser::Parser)
    tok = current_token(parser)
    if isnothing(tok)
        throw(RegexParserError("Неожиданный конец при ожидании базового выражения"))
    end

    if tok.ttype == "CAP_OPEN"
        eat!(parser, "CAP_OPEN")
        parser.group_count += 1
        if parser.group_count > parser.max_groups
            throw(RegexParserError("Превышено число групп захвата > 9"))
        end
        group_id = parser.group_count
        node = parse_alternation(parser)
        eat!(parser, "CLOSE")
        parser.groups_ast[group_id] = node
        return GroupNode(group_id, node)

    elseif tok.ttype == "NONCAP_OPEN"
        eat!(parser, "NONCAP_OPEN")
        node = parse_alternation(parser)
        eat!(parser, "CLOSE")
        return NonCapGroupNode(node)

    elseif tok.ttype == "LOOKAHEAD_OPEN"
        if parser.in_lookahead
            throw(RegexParserError("Вложенные опережающие проверки запрещены"))
        end
        eat!(parser, "LOOKAHEAD_OPEN")
        old_lookahead = parser.in_lookahead
        parser.in_lookahead = true

        # Вызов проверки на захватывающие группы внутри lookahead
        node = parse_alternation(parser)
        check_no_cap_and_lookahead(parser, node, true)

        parser.in_lookahead = old_lookahead
        eat!(parser, "CLOSE")
        return LookaheadNode(node)

    elseif tok.ttype == "EXPR_REF_OPEN"
        ref_id = tok.value
        eat!(parser, "EXPR_REF_OPEN")
        eat!(parser, "CLOSE")
        return ExprRefNode(ref_id)

    elseif tok.ttype == "CHAR"
        eat!(parser, "CHAR")
        return CharNode(tok.value)

    else
        throw(RegexParserError("Некорректный токен: $tok"))
    end
end


# Проверка корректности ссылок на группы и вложенности
function check_references(parser::Parser, node, defined_groups::Set{Int})
    if node isa CharNode
        return defined_groups

    elseif node isa ExprRefNode
        # Не проверяем forward references (допустимы)
        return defined_groups

    elseif node isa GroupNode
        # Проверяем содержимое группы
        new_defined = check_references(parser, node.node, defined_groups)
        # Добавляем текущую группу в множество определённых групп
        push!(new_defined, node.group_id)
        return new_defined

    elseif node isa NonCapGroupNode
        return check_references(parser, node.node, defined_groups)

    elseif node isa LookaheadNode
        # Проверяем, что внутри lookahead нет групп захвата или других lookahead
        check_no_cap_and_lookahead(parser, node.node, true)
        return check_references(parser, node.node, defined_groups)

    elseif node isa StarNode
        return check_references(parser, node.node, defined_groups)

    elseif node isa ConcatNode
        cur_defined = defined_groups
        for child in node.nodes
            cur_defined = check_references(parser, child, cur_defined)
        end
        return cur_defined

    elseif node isa AltNode
        all_defs = []
        for branch in node.branches
            push!(all_defs, check_references(parser, branch, defined_groups))
        end
        # Объединяем все определения групп из ветвей альтернативы
        union_defs = Set{Int}()
        for d in all_defs
            union!(union_defs, d)
        end
        return union_defs

    else
        throw(RegexParserError("Неизвестный тип узла AST при проверке ссылок"))
    end
end

# Проверка отсутствия групп захвата и вложенных lookahead внутри lookahead
function check_no_cap_and_lookahead(parser::Parser, node, inside_lookahead::Bool)
    if node isa GroupNode && inside_lookahead
        throw(RegexParserError("Внутри опережающей проверки не допускаются захватывающие группы"))
    elseif node isa LookaheadNode && inside_lookahead
        throw(RegexParserError("Внутри опережающей проверки не допускаются другие опережающие проверки"))
    end

    if node isa NonCapGroupNode
        check_no_cap_and_lookahead(parser, node.node, inside_lookahead)
    elseif node isa LookaheadNode
        check_no_cap_and_lookahead(parser, node.node, inside_lookahead)
    elseif node isa StarNode
        check_no_cap_and_lookahead(parser, node.node, inside_lookahead)
    elseif node isa ConcatNode
        for n in node.nodes
            check_no_cap_and_lookahead(parser, n, inside_lookahead)
        end
    elseif node isa AltNode
        for b in node.branches
            check_no_cap_and_lookahead(parser, b, inside_lookahead)
        end
    end
end


# Построитель КС-грамматики (CFG)
mutable struct CFGBuilder
    groups_ast::Dict{Int, Any}        # {group_id: AST узел}
    group_nonterm::Dict{Int, String}
    noncap_index::Int
    star_index::Int
end

# Конструктор
function CFGBuilder(groups_ast::Dict{Int, Any})
    return CFGBuilder(groups_ast, Dict{Int, String}(), 1, 1)
end

# Построение грамматики
function build(builder::CFGBuilder, node)
    start = "S"              # Начальный нетерминал
    rules = Dict{String, Vector{Vector{String}}}()  # Правила CFG

    # Регистрируем основной нетерминал для группы 1
    main_nt = node_to_cfg(builder, node, rules)

    # Добавляем правило S -> main_nt
    rules[start] = [[main_nt]]

    # Убедимся, что все группы зарегистрированы
    for (group_id, ast) in builder.groups_ast
        if group_id ∉ keys(builder.group_nonterm)
            nt = "G$group_id"
            builder.group_nonterm[group_id] = nt
            node_to_cfg(builder, ast, rules, start_symbol=nt)
        end
    end

    return start, rules
end

# Преобразование AST в нетерминал CFG
function node_to_cfg(builder::CFGBuilder, node, rules::Dict{String, Vector{Vector{String}}};
                     start_symbol::Union{String, Nothing} = nothing)
    if node isa CharNode
        # Терминальный символ
        nt = isnothing(start_symbol) ? fresh_nt(builder, "CHAR") : start_symbol
        push!(get!(rules, nt, []), [node.ch])
        return nt

    elseif node isa GroupNode
        # Захватывающая группа
        nt = get(builder.group_nonterm, node.group_id, nothing)
        if isnothing(nt)
            nt = "G$(node.group_id)"
            builder.group_nonterm[node.group_id] = nt
        end
        sub_nt = node_to_cfg(builder, node.node, rules)
        push!(get!(rules, nt, []), [sub_nt])
        return nt

    elseif node isa NonCapGroupNode
        # Незахватывающая группа
        nt = isnothing(start_symbol) ? fresh_nt(builder, "N") : start_symbol
        sub_nt = node_to_cfg(builder, node.node, rules)
        push!(get!(rules, nt, []), [sub_nt])
        return nt

    elseif node isa LookaheadNode
        # Lookahead заменяем на ε
        nt = isnothing(start_symbol) ? fresh_nt(builder, "LA") : start_symbol
        push!(get!(rules, nt, []), [])
        return nt

    elseif node isa ConcatNode
        # Конкатенация узлов
        nt = isnothing(start_symbol) ? fresh_nt(builder, "C") : start_symbol
        seq_nts = [node_to_cfg(builder, child, rules) for child in node.nodes]
        push!(get!(rules, nt, []), seq_nts)
        return nt

    elseif node isa AltNode
        # Альтернатива (ветви)
        nt = isnothing(start_symbol) ? fresh_nt(builder, "A") : start_symbol
        for branch in node.branches
            br_nt = node_to_cfg(builder, branch, rules)
            push!(get!(rules, nt, []), [br_nt])
        end
        return nt

    elseif node isa StarNode
        # Повторение (X*)
        nt = isnothing(start_symbol) ? fresh_nt(builder, "R") : start_symbol
        sub_nt = node_to_cfg(builder, node.node, rules)
        push!(get!(rules, nt, []), [])              # ε
        push!(rules[nt], [nt, sub_nt])              # R -> R sub_nt
        return nt

    elseif node isa ExprRefNode
        # Ссылка на выражение группы
        ref_id = node.ref_id
        if ref_id ∉ keys(builder.group_nonterm)
            builder.group_nonterm[ref_id] = "G$ref_id"
            if ref_id ∉ keys(builder.groups_ast)
                throw(RegexParserError("Ссылка на несуществующую группу $ref_id"))
            end
            node_to_cfg(builder, builder.groups_ast[ref_id], rules)
        end
        return builder.group_nonterm[ref_id]

    else
        throw(RegexParserError("Неизвестный тип узла при построении грамматики"))
    end
end

# Генерация имени нетерминала
function fresh_nt(builder::CFGBuilder, prefix::String)
    if prefix == "N"
        name = "N$(builder.noncap_index)"
        builder.noncap_index += 1
        return name
    elseif prefix == "R"
        name = "R$(builder.star_index)"
        builder.star_index += 1
        return name
    else
        # Общий счётчик для других типов
        name = "$(prefix)$(builder.noncap_index + builder.star_index)"
        builder.noncap_index += 1
        return name
    end
end

function main()
    text = String(strip(readline()))

    try
        if isempty(text)
            throw(RegexParserError("Пустая строка"))
        end

        # Разбиваем на лексемы
        lexer = Lexer(text)
        tokens = tokenize(lexer)

        # Строим AST
        parser = Parser(tokens)
        ast = parse_ast(parser)

        # Строим КС-грамматику
        builder = CFGBuilder(parser.groups_ast)
        start_symbol, rules = build(builder, ast)

        println("Выражение корректно синтаксически и удовлетворяет ограничениям.")
        println("Построенная КС-грамматика (каркас):")
        println("Начальный нетерминал: ", start_symbol)

        # Печатаем правила грамматики
        for (nt, rhs_list) in rules
            for rhs in rhs_list
                rhs_str = isempty(rhs) ? "ε" : join(rhs, " ")
                println("$nt -> $rhs_str")
            end
        end

        println("=" ^ 20) 
        for token in tokens
            println(token)
        end

        println("=" ^ 20)
        println(ast)

    catch e
        if e isa RegexParserError
            println("Ошибка: ", e.msg)
        else
            rethrow(e) 
        end
    end
end

if abspath(PROGRAM_FILE) == @__FILE__
    main()
end

