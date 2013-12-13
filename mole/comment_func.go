
//line comment_func.y:2
package mole
import __yyfmt__ "fmt"
//line comment_func.y:2
		
import(
	//"fmt"
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


//line comment_func.y:57
type yySymType struct {
	yys int
	function *CommentFunc
	argument *CommentArg
	args []*CommentArg
	iconst int64
	fconst float64
	value Value
	values *CommentArrayValue
	s string
}

const str_const = 57346
const tok_name = 57347
const tok_float_const = 57348
const tok_int_const = 57349

var yyToknames = []string{
	"str_const",
	"tok_name",
	"tok_float_const",
	"tok_int_const",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line comment_func.y:147




//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 17
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 40

var yyAct = []int{

	10, 15, 8, 12, 13, 19, 15, 9, 12, 13,
	16, 21, 7, 17, 18, 16, 26, 22, 5, 25,
	24, 23, 15, 9, 12, 13, 15, 27, 12, 13,
	4, 16, 3, 1, 2, 16, 6, 20, 14, 11,
}
var yyPact = []int{

	24, -1000, -1000, 25, 9, 2, 3, -1000, -1000, -7,
	-1000, -1000, -1000, -1000, -1000, -1000, -3, -1000, 18, 22,
	5, -1000, -1000, -1000, -1000, -1000, 22, -1000,
}
var yyPgo = []int{

	0, 0, 39, 38, 37, 36, 2, 34, 33,
}
var yyR1 = []int{

	0, 8, 7, 7, 5, 5, 6, 6, 3, 3,
	4, 4, 2, 1, 1, 1, 1,
}
var yyR2 = []int{

	0, 1, 5, 4, 3, 1, 3, 1, 3, 2,
	3, 1, 1, 1, 1, 1, 1,
}
var yyChk = []int{

	-1000, -8, -7, 8, 5, 9, -5, 10, -6, 5,
	-1, -2, 6, 7, -3, 4, 13, 10, 11, 12,
	-4, 14, -1, -6, -1, 14, 11, -1,
}
var yyDef = []int{

	0, -2, 1, 0, 0, 0, 0, 3, 5, 0,
	7, 13, 14, 15, 16, 12, 0, 2, 0, 0,
	0, 9, 11, 4, 6, 8, 0, 10,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	9, 10, 3, 3, 11, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 12, 3, 3, 8, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 13, 3, 14,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 7,
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
		//line comment_func.y:79
		{
		yylex.(*CommentLexer).Function = yyS[yypt-0].function
	}
	case 2:
		//line comment_func.y:83
		{
		function := &CommentFunc{}
		function.Name = yyS[yypt-3].s
		function.Args = yyS[yypt-1].args
		yyVAL.function = function
	}
	case 3:
		//line comment_func.y:88
		{
		function := &CommentFunc{}
		function.Name = yyS[yypt-2].s
		function.Args = []*CommentArg{}
	}
	case 4:
		//line comment_func.y:94
		{
		yyVAL.args = append(yyS[yypt-2].args, yyS[yypt-0].argument)
	}
	case 5:
		//line comment_func.y:97
		{
		yyVAL.args = []*CommentArg{yyS[yypt-0].argument}
	}
	case 6:
		//line comment_func.y:101
		{
		arg := &CommentArg{}
		arg.Name = yyS[yypt-2].s
		arg.Value = yyS[yypt-0].value
		yyVAL.argument = arg
	}
	case 7:
		//line comment_func.y:107
		{
		arg := &CommentArg{}
		arg.Value = yyS[yypt-0].value
		yyVAL.argument = arg
	}
	case 8:
		//line comment_func.y:113
		{
		yyVAL.values = yyS[yypt-1].values
	}
	case 9:
		//line comment_func.y:115
		{
		yyVAL.values = &CommentArrayValue{}
	}
	case 10:
		//line comment_func.y:119
		{
		vals := append(*yyS[yypt-2].values, yyS[yypt-0].value)
		yyVAL.values = &vals
	}
	case 11:
		//line comment_func.y:123
		{
		yyVAL.values = &CommentArrayValue{yyS[yypt-0].value}
	}
	case 12:
		//line comment_func.y:127
		{
		value := CommentStringValue(yyS[yypt-0].s)
		yyVAL.value = &value
	}
	case 13:
		//line comment_func.y:132
		{
		yyVAL.value = yyS[yypt-0].value
	}
	case 14:
		//line comment_func.y:135
		{
		value := CommentFloatValue(yyS[yypt-0].fconst)
		yyVAL.value = &value
	}
	case 15:
		//line comment_func.y:139
		{
		value := CommentIntValue(yyS[yypt-0].iconst)
		yyVAL.value = &value
	}
	case 16:
		//line comment_func.y:143
		{
		yyVAL.value = yyS[yypt-0].values
	}
	}
	goto yystack /* stack new state and value */
}
