package mole

import (
	"testing"
)

func TestCommentFunc(t *testing.T) {
	yyDebug = 100
	cl := &CommentLexer{}
	cl.Comment = "@xx(\"\", a\t\r\n = [\"mmmmm\",\"i\"],-112, 1.2, 0xabcdef)"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%d", *cl.Function.Args[2].Value.(*CommentIntValue))
	t.Logf("%f", *cl.Function.Args[3].Value.(*CommentFloatValue))
	t.Logf("%d", *cl.Function.Args[4].Value.(*CommentIntValue))
	t.Logf("%q", cl.Function)
}
