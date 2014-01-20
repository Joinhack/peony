package mole

import (
	"testing"
)

func TestCommentFunc1(t *testing.T) {
	yyDebug = 0
	cl := &CommentLexer{}
	cl.Comment = "@xx(\"\", a\t\r\n = [\"mmmmm\",\"i\"],-112, 1.2, 0xabcdef,true,[[]])"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%d", *cl.Expr.Args[2].Value.(*CommentIntValue))
	t.Logf("%f", *cl.Expr.Args[3].Value.(*CommentFloatValue))
	t.Logf("%d", *cl.Expr.Args[4].Value.(*CommentIntValue))
	t.Logf("%q", *cl.Expr.Args[5].Value.(*CommentBoolValue))
	t.Logf("%q", cl.Expr)
}

func TestCommentFunc2(t *testing.T) {
	yyDebug = 0
	cl := &CommentLexer{}
	cl.Comment = "@xx( )"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%q", cl.Expr)
}

func TestCommentFunc3(t *testing.T) {
	cl := &CommentLexer{}
	cl.Comment = "@xx"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%q", cl.Expr)
}
