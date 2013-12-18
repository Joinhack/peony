
//line comment_func.y:2
package mole
import __yyfmt__ "fmt"
//line comment_func.y:2
		
import(
	//"fmt"
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

type CommentFunc struct {
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


//line comment_func.y:66
type yySymType struct {
	yys int
	function *CommentFunc
	argument *CommentArg
	args []*CommentArg
	iconst int64
	fconst float64
	value Value
	values *CommentArrayValue
	b bool
	s string
}

const str_const = 57346
const tok_name = 57347
const numstr_const = 57348
const tok_bool = 57349
const tok_hex = 57350

var yyToknames = []string{
	"str_const",
	"tok_name",
	"numstr_const",
	"tok_bool",
	"tok_hex",
	" -",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line comment_func.y:192




//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 24
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 55

var yyAct = []int{

	10, 17, 8, 19, 14, 13, 18, 34, 31, 25,
	33, 3, 20, 27, 23, 17, 9, 19, 14, 13,
	18, 28, 35, 7, 30, 29, 20, 17, 9, 19,
	14, 13, 18, 21, 22, 36, 5, 17, 20, 19,
	14, 13, 18, 32, 24, 4, 1, 2, 20, 6,
	26, 16, 11, 12, 15,
}
var yyPact = []int{

	1, -1000, -1000, 40, 25, 11, 21, -1000, -1000, 0,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 38, -8,
	-3, -1000, 23, 33, -9, 37, -6, -1000, -1000, -1000,
	-1000, 16, -1000, -1000, 33, -1000, -1000,
}
var yyPgo = []int{

	0, 54, 53, 0, 52, 51, 50, 49, 2, 47,
	46,
}
var yyR1 = []int{

	0, 10, 9, 9, 9, 7, 7, 8, 8, 5,
	5, 6, 6, 4, 1, 1, 2, 2, 3, 3,
	3, 3, 3, 3,
}
var yyR2 = []int{

	0, 1, 5, 4, 2, 3, 1, 3, 1, 3,
	2, 3, 1, 1, 1, 2, 4, 3, 1, 1,
	1, 1, 1, 1,
}
var yyChk = []int{

	-1000, -10, -9, 10, 5, 11, -7, 12, -8, 5,
	-3, -4, -2, 8, 7, -1, -5, 4, 9, 6,
	15, 12, 13, 14, 6, 17, -6, 16, -3, -8,
	-3, 17, 6, 16, 13, 6, -3,
}
var yyDef = []int{

	0, -2, 1, 0, 4, 0, 0, 3, 6, 0,
	8, 18, 19, 20, 21, 22, 23, 13, 0, 14,
	0, 2, 0, 0, 15, 0, 0, 10, 12, 5,
	7, 0, 17, 9, 0, 16, 11,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	11, 12, 3, 3, 13, 9, 17, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 14, 3, 3, 10, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 15, 3, 16,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 7, 8,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(c), uint(char))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		//line comment_func.y:91
		{
		yylex.(*CommentLexer).Function = yyS[yypt-0].function
	}
	case 2:
		//line comment_func.y:95
		{
		function := &CommentFunc{}
		function.Name = yyS[yypt-3].s
		function.Args = yyS[yypt-1].args
		yyVAL.function = function
	}
	case 3:
		//line comment_func.y:101
		{
		function := &CommentFunc{}
		function.Name = yyS[yypt-2].s
		function.Args = []*CommentArg{}
		yyVAL.function = function
	}
	case 4:
		//line comment_func.y:107
		{
		function := &CommentFunc{}
		function.Name = yyS[yypt-0].s
		function.Args = []*CommentArg{}
		yyVAL.function = function
	}
	case 5:
		//line comment_func.y:114
		{
		yyVAL.args = append(yyS[yypt-2].args, yyS[yypt-0].argument)
	}
	case 6:
		//line comment_func.y:117
		{
		yyVAL.args = []*CommentArg{yyS[yypt-0].argument}
	}
	case 7:
		//line comment_func.y:121
		{
		arg := &CommentArg{}
		arg.Name = yyS[yypt-2].s
		arg.Value = yyS[yypt-0].value
		yyVAL.argument = arg
	}
	case 8:
		//line comment_func.y:127
		{
		arg := &CommentArg{}
		arg.Value = yyS[yypt-0].value
		yyVAL.argument = arg
	}
	case 9:
		//line comment_func.y:133
		{
		yyVAL.values = yyS[yypt-1].values
	}
	case 10:
		//line comment_func.y:135
		{
		yyVAL.values = &CommentArrayValue{}
	}
	case 11:
		//line comment_func.y:139
		{
		vals := append(*yyS[yypt-2].values, yyS[yypt-0].value)
		yyVAL.values = &vals
	}
	case 12:
		//line comment_func.y:143
		{
		yyVAL.values = &CommentArrayValue{yyS[yypt-0].value}
	}
	case 13:
		//line comment_func.y:147
		{
		value := CommentStringValue(yyS[yypt-0].s)
		yyVAL.value = &value
	}
	case 14:
		//line comment_func.y:152
		{
		yyVAL.s = yyS[yypt-0].s
	}
	case 15:
		//line comment_func.y:155
		{
		yyVAL.s = "-" + yyS[yypt-0].s
	}
	case 16:
		//line comment_func.y:160
		{
		yyVAL.s = "-" + yyS[yypt-2].s + yyS[yypt-0].s
	}
	case 17:
		//line comment_func.y:163
		{
		yyVAL.s = yyS[yypt-2].s + "." + yyS[yypt-0].s
	}
	case 18:
		//line comment_func.y:167
		{
		yyVAL.value = yyS[yypt-0].value
	}
	case 19:
		//line comment_func.y:170
		{
		f, _ := strconv.ParseFloat(yyS[yypt-0].s, 64)
		value := CommentFloatValue(f)
		yyVAL.value = &value
	}
	case 20:
		//line comment_func.y:175
		{
		value := CommentIntValue(yyS[yypt-0].iconst)
		yyVAL.value = &value
	}
	case 21:
		//line comment_func.y:179
		{
		value := CommentBoolValue(yyS[yypt-0].b)
		yyVAL.value = &value	
	}
	case 22:
		//line comment_func.y:183
		{
		i, _ := strconv.Atoi(yyS[yypt-0].s)
		value := CommentIntValue(i)
		yyVAL.value = &value
	}
	case 23:
		//line comment_func.y:188
		{
		yyVAL.value = yyS[yypt-0].values
	}
	}
	goto yystack /* stack new state and value */
}
