%{
package peony

import (
	"strconv"
)


//e.g. <str:name>
//<str(len=10):name>
//<str(maxlen=30):name>
//<str(maxlen=30, minlen=10):name>
//<int(len=20):name>
//<float:name>
//<re(\w{10}):name>


type ExprValueType int

type ExprType int

const (
	ExprStringType ExprValueType = iota
	ExprIntType
	ExprBoolType
	ExprFloatType
	ExprArrayType
)

const (
	FuncExprType ExprType = iota
	StaticExprType
)

type ExprValue interface {
	Type() ExprValueType
}

type ExprIntValue int64

type ExprStringValue string

type ExprFloatValue float64

type ExprBoolValue bool

type ExprArrayValue []ExprValue

func (v *ExprIntValue) Type() ExprValueType {
	return ExprValueType(1)
}

func (v *ExprStringValue) Type() ExprValueType {
	return ExprValueType(2)
}

func (v *ExprFloatValue) Type() ExprValueType {
	return ExprValueType(3)
}

func (v *ExprBoolValue) Type() ExprValueType {
	return ExprValueType(4)
}

func (v *ExprArrayValue) Type() ExprValueType {
	return ExprValueType(5)
}

type ExprArg struct {
	Name string
	Value ExprValue
}

type RouterExpr interface {
	Type() ExprType
}

type FuncExpr struct {
	RouterExpr
	VarName string
	Name string
	Args []*ExprArg
}

func (expr *FuncExpr) Type() ExprType {
	return ExprType(1)
}

type StaticExpr struct {
	RouterExpr
	Value string
}

func (expr *StaticExpr) Type() ExprType {
	return ExprType(2)
}

%}

%union {
	s string
	i int64
	float float64
	v ExprValue
	b bool
	exprs []RouterExpr
	expr RouterExpr
	funcExpr *FuncExpr
	arg  *ExprArg
	args []*ExprArg
	exprval ExprValue
}

%token<s> tok_var numstr_const str_const static_const

%token<i> tok_hex

%token<b> tok_bool

%type<s> FLOATEXPR INTEXPR

%type<exprs> EXPRS

%type<expr> EXPR

%type<arg> ARG

%type<args> ARGS

%type<exprval> VALUE

%type<funcExpr> EXPR_FUNC

%left '-'

%%

EXPRS: EXPRS EXPR {
	$$ = append($1, $2)
}
| EXPR {
	$$ = []RouterExpr{$1}
}

EXPR: '<' EXPR_FUNC ':' tok_var '>' {
	$$ = $2
	$2.VarName = $4
	$$ = $2
}
| static_const {
	expr := StaticExpr{Value: $1}
	$$ = &expr
}

EXPR_FUNC: tok_var {
	$$ = &FuncExpr{Name: $1}
}
| tok_var '(' ARGS ')' {
	$$ = &FuncExpr{Name: $1, Args: $3}
}

ARGS: ARGS ',' ARG {
	$$ = append($1, $3)
}
| ARG {
	$$ = []*ExprArg{$1}
}

ARG: tok_var '=' VALUE  {
	
}
| VALUE {
	
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

VALUE: INTEXPR {
	i, _ := strconv.Atoi($1)
	value := ExprIntValue(i)
	$$ = &value
}
| FLOATEXPR {
	f, _ := strconv.ParseFloat($1, 64)
	value := ExprFloatValue(f)
	$$ = &value
}
| str_const {
	value := ExprStringValue($1)
	$$ = &value	
} 
| tok_bool {
	value := ExprBoolValue($1)
	$$ = &value
}


%%
