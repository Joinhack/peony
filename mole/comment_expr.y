%{
package mole

import(
	"strconv"
)

type CommentValueType int

const (
	CommentStringType CommentValueType = iota
	CommentIntType
	CommentBoolType
	CommentFloatType
	CommentArrayType
)

type CommentExpr struct {
	Name string
	Args []*CommentArg
}

type Value interface {
	ValueType() CommentValueType
}

type CommentStringValue string

type CommentBoolValue bool

type CommentIntValue int64

type CommentFloatValue float64

type CommentArrayValue []Value

func(str *CommentArrayValue) ValueType() CommentValueType {
	return CommentArrayType
}

func(str *CommentStringValue) ValueType() CommentValueType {
	return CommentStringType
}


func(str *CommentIntValue) ValueType() CommentValueType {
	return CommentIntType
}

func(str *CommentFloatValue) ValueType() CommentValueType {
	return CommentFloatType
}

func(str *CommentBoolValue) ValueType() CommentValueType {
	return CommentBoolType
}

type CommentArg struct {
	Name string
	Value Value
}

%}

%union {
	expr *CommentExpr
	argument *CommentArg
	args []*CommentArg
	iconst int64
	fconst float64
	value Value
	values *CommentArrayValue
	b bool
	s string
}

%token<s> str_const tok_name numstr_const
%token<b> tok_bool
%token<iconst> tok_hex
%type<s> INTEXPR FLOATEXPR
%type<value> VALUE STRING
%type<values> VALUES ELEMENTS
%type<args> ARGS_ELEMENTS
%type<argument> ARGUMENT
%type<expr> EXPR
%type<expr> RESULT
%left '-'
%%

RESULT: EXPR {
	yylex.(*CommentLexer).Expr = $1
}

EXPR: '@' tok_name '(' ARGS_ELEMENTS ')' {
	expr := &CommentExpr{}
	expr.Name = $2
	expr.Args = $4
	$$ = expr
} 
| '@' tok_name '('  ')' {
	expr := &CommentExpr{}
	expr.Name = $2
	expr.Args = []*CommentArg{}
	$$ = expr
}
| '@' tok_name {
	expr := &CommentExpr{}
	expr.Name = $2
	expr.Args = []*CommentArg{}
	$$ = expr
}

ARGS_ELEMENTS: ARGS_ELEMENTS ',' ARGUMENT {
	$$ = append($1, $3)
}
| ARGUMENT {
	$$ = []*CommentArg{$1}
}

ARGUMENT: tok_name '=' VALUE {
	arg := &CommentArg{}
	arg.Name = $1
	arg.Value = $3
	$$ = arg
} 
| VALUE {
	arg := &CommentArg{}
	arg.Value = $1
	$$ = arg
}

VALUES: '[' ELEMENTS ']' {
	$$ = $2
} | '[' ']' {
	$$ = &CommentArrayValue{}
}

ELEMENTS: ELEMENTS ',' VALUE {
	vals := append(*$1, $3)
	$$ = &vals
}
| VALUE {
	$$ = &CommentArrayValue{$1}
}

STRING: str_const {
	value := CommentStringValue($1)
	$$ = &value
}

INTEXPR: numstr_const {
	$$ = $1
}
| '-' numstr_const {
	$$ = "-" + $2
}


FLOATEXPR: '-' numstr_const '.' numstr_const {
	$$ = "-" + $2 + $4
} 
| numstr_const '.' numstr_const {
	$$ = $1 + "." + $3
}

VALUE: STRING {
	$$ = $1
}
| FLOATEXPR {
	f, _ := strconv.ParseFloat($1, 64)
	value := CommentFloatValue(f)
	$$ = &value
}
| tok_hex {
	value := CommentIntValue($1)
	$$ = &value
}
| tok_bool {
	value := CommentBoolValue($1)
	$$ = &value	
}
| INTEXPR {
	i, _ := strconv.Atoi($1)
	value := CommentIntValue(i)
	$$ = &value
}
| VALUES {
	$$ = $1
}

%%


