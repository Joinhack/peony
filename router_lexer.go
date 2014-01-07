package peony

import (
	"errors"
)

type RouterLexer struct {
	yyLexer
	RouterExpr   *RouterExpr
	Expr         string
	PrevPos      int
	Pos          int
	currentState int
	Err          string
}

const quotes = int('"')

func (c *RouterLexer) Parse(expr string) (*RouterExpr, error) {
	c.Expr = expr
	rs := yyParse(c)
	if rs != 0 {
		return nil, errors.New(c.Err)
	}
	return c.RouterExpr, nil
}

func (c *RouterLexer) Bool(s string) (bool, bool) {
	if len(s) != 4 {
		return false, false
	}
	if s == "true" {
		return true, true
	}
	if s == "false" {
		return false, true
	}
	return false, false

}

func (c *RouterLexer) Lex(lval *yySymType) int {
	length := len(c.Expr)
	//manual parse hex.
	var tmphex int64 = 0
	var strbuf []byte
	for c.Pos < length {
		chr := int(c.Expr[c.Pos])
		switch {
		case chr == quotes:
			//string start
			if c.currentState == 0 {
				c.PrevPos = c.Pos
				c.currentState = str_const
				strbuf = make([]byte, 0, length-c.Pos)
			} else if c.currentState == str_const {
				//string end
				c.currentState = 0
				lval.s = string(strbuf)
				c.Pos++
				return str_const
			}
		case ((chr >= 'a' && chr <= 'z') || (chr >= 'A' && chr <= 'Z')) && c.currentState == 0:
			//tok_var start
			c.currentState = tok_var
			c.PrevPos = c.Pos
		case c.currentState == tok_var:
			//collect a-z A-Z - _ to tok_var
			switch {
			case (chr >= '0' && chr <= '9'):
			case (chr >= 'a' && chr <= 'z'):
			case chr == '_' || chr == '-':
			default:
				//tok_var end
				c.currentState = 0
				s := c.Expr[c.PrevPos:c.Pos]
				if b, ok := c.Bool(s); ok {
					lval.b = b
					return tok_bool
				}
				lval.s = s
				return tok_var
			}
		case c.currentState == 0 && chr == '0':
			if c.Pos != length-1 {
				next := c.Expr[c.Pos+1]
				if next == 'x' || next == 'X' {
					c.PrevPos = c.Pos + 1
					c.currentState = tok_hex
					//skip x or X
					c.Pos++
					tmphex = 0
				}
			}
		case (chr >= '0' && chr <= '9') && c.currentState == 0:
			c.currentState = numstr_const
			c.PrevPos = c.Pos
		case c.currentState == numstr_const:
			if chr >= '0' && chr <= '9' {
				//collect number
				break
			} else {
				// number record end.
				// so float should be "numstr_const '.' numstr_const".
				c.currentState = 0
				lval.s = c.Expr[c.PrevPos:c.Pos]
				return numstr_const
			}
		case c.currentState == tok_hex:
			switch {
			case (chr >= '0' && chr <= '9'):
				tmphex *= 16
				tmphex += int64(chr - '0')
			case (chr >= 'a' && chr <= 'f'):
				tmphex *= 16
				//87 = ('a'-10), (chr - 87) = 10 + chr - 'a'
				tmphex += int64(chr - 87)
			case (chr >= 'A' && chr <= 'F'):
				tmphex *= 16
				//55 = ('A'-10), (chr - 55) = 10 + chr - 'A'
				tmphex += int64(chr - 55)
			default:
				lval.i = tmphex
				c.currentState = 0
				return tok_hex
			}
		case c.currentState == str_const:
			//collect all
			if chr == '\\' && c.Pos != length-1 {
				c.Pos++
				switch nextChr := c.Expr[c.Pos]; nextChr {
				case 'n':
					strbuf = append(strbuf, '\n')
				case 't':
					strbuf = append(strbuf, '\t')
				case 'b':
					strbuf = append(strbuf, '\b')
				case 'r':
					strbuf = append(strbuf, '\r')
				case '"':
					strbuf = append(strbuf, '"')
				case 'f':
					strbuf = append(strbuf, '\f')
				default:
					return -1
				}
				break
			}
			strbuf = append(strbuf, byte(chr))
		//when c.currentTok==0 ignore follow char
		case chr == ' ' || chr == '\t' || chr == '\r' || chr == '\n':
			break
		default:
			c.Pos++
			return chr
		}
		c.Pos++
	}
	return 0
}

func (c *RouterLexer) Error(s string) {
	c.Err = s
}
