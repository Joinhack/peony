package mole

import (
	"testing"
)

func TestCommentFunc(t *testing.T) {
	yyDebug = 1000
	cl := &CommentLexer{}
	cl.Comment = "@xx(\"\", a\t\r\n = [\"mmmmm\",\"i\",[]])"
	if i := yyParse(cl); i == 1 {
		panic(cl.Err)
	}
	t.Logf("%q", cl.Function)
}
