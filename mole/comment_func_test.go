package mole

import (
	"testing"
)

func TestCommentFunc1(t *testing.T) {
	yyDebug = 0
	cl := &CommentLexer{}
	cl.Comment = "@xx(\"\", a\t\r\n = [\"mmmmm\",\"i\"],-112, 1.2, 0xabcdef,[[]])"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%d", *cl.Function.Args[2].Value.(*CommentIntValue))
	t.Logf("%f", *cl.Function.Args[3].Value.(*CommentFloatValue))
	t.Logf("%d", *cl.Function.Args[4].Value.(*CommentIntValue))
	t.Logf("%q", cl.Function)
}

func TestCommentFunc2(t *testing.T) {
	yyDebug = 0
	cl := &CommentLexer{}
	cl.Comment = "@xx( )"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%q", cl.Function)
}

func TestCommentFunc3(t *testing.T) {
	cl := &CommentLexer{}
	cl.Comment = "@xx"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%q", cl.Function)
}
