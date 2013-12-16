%{
package mole

import(
	//"fmt"
	"strconv"
)

type CommentValueType int

const (
	CommentStringType CommentValueType = iota
	CommentIntType
	CommentFloatType
	CommentArrayType
)

type CommentFunc struct {
	Name string
	Args []*CommentArg
}

type Value interface {
	ValueType() CommentValueType
}

type CommentStringValue string

type CommentArrayValue []Value

func(str *CommentArrayValue) ValueType() CommentValueType {
	return CommentArrayType
}

func(str *CommentStringValue) ValueType() CommentValueType {
	return CommentStringType
}

type CommentIntValue int64

func(str *CommentIntValue) ValueType() CommentValueType {
	return CommentIntType
}

type CommentFloatValue float64

func(str *CommentFloatValue) ValueType() CommentValueType {
	return CommentFloatType
}

type CommentArg struct {
	Name string
	Value Value
}

%}

%union {
	function *CommentFunc
	argument *CommentArg
	args []*CommentArg
	iconst int64
	fconst float64
	value Value
	values *CommentArrayValue
	s string
}

%token<s> str_const tok_name numstr_const
%token<iconst> tok_hex
%type<s> INTEXPR FLOATEXPR
%type<value> VALUE STRING
%type<values> VALUES ELEMENTS
%type<args> ARGS_ELEMENTS
%type<argument> ARGUMENT
%type<function> FUNCTION
%type<function> RESULT
%left '-'
%%

RESULT: FUNCTION {
	yylex.(*CommentLexer).Function = $1
}

FUNCTION: '@' tok_name '(' ARGS_ELEMENTS ')' {
	function := &CommentFunc{}
	function.Name = $2
	function.Args = $4
	$$ = function
} 
| '@' tok_name '('  ')' {
	function := &CommentFunc{}
	function.Name = $2
	function.Args = []*CommentArg{}
	$$ = function
}
| '@' tok_name {
	function := &CommentFunc{}
	function.Name = $2
	function.Args = []*CommentArg{}
	$$ = function
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
| INTEXPR {
	i, _ := strconv.Atoi($1)
	value := CommentIntValue(i)
	$$ = &value
}
| VALUES {
	$$ = $1
}

%%


