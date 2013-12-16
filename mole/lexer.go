package mole

import (
	"errors"
)

type CommentLexer struct {
	yyLexer
	Function     *CommentFunc
	Comment      string
	PrevPos      int
	Pos          int
	currentState int
	Err          string
}

const quotes = int('"')

func (c *CommentLexer) Parse(comment string) (*CommentFunc, error) {
	c.Comment = comment
	rs := yyParse(c)
	if rs != 0 {
		return nil, errors.New(c.Err)
	}
	return c.Function, nil
}

func (c *CommentLexer) Lex(lval *yySymType) int {
	length := len(c.Comment)
	//manual parse hex.
	var tmphex int64 = 0
	for c.Pos < length {
		chr := int(c.Comment[c.Pos])
		switch {
		case chr == quotes:
			//string start
			if c.currentState == 0 {
				c.PrevPos = c.Pos
				c.currentState = str_const
			} else if c.currentState == str_const {
				//string end
				c.currentState = 0
				lval.s = c.Comment[c.PrevPos+1 : c.Pos]
				c.Pos++
				return str_const
			}
		case ((chr >= 'a' && chr <= 'z') || (chr >= 'A' && chr <= 'Z')) && c.currentState == 0:
			//tok_name start
			c.currentState = tok_name
			c.PrevPos = c.Pos
		case c.currentState == tok_name:
			//collect a-z A-Z - _ to tok_name
			switch {
			case (chr >= '0' && chr <= '9'):
			case (chr >= 'a' && chr <= 'z'):
			case chr == '_' || chr == '-':
			default:
				//tok_name end
				c.currentState = 0
				lval.s = c.Comment[c.PrevPos:c.Pos]
				return tok_name
			}
		case c.currentState == 0 && chr == '0':
			if c.Pos != len(c.Comment)-1 {
				next := c.Comment[c.Pos+1]
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
				lval.s = c.Comment[c.PrevPos:c.Pos]
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
				lval.iconst = tmphex
				c.currentState = 0
				return tok_hex
			}
		case c.currentState == str_const:
			//collect all
			break
		//when c.currentTok==0 ignore follow char
		case chr == ' ' || chr == '\t' || chr == '\r' || chr == '\n':
			break
		default:
			c.Pos++
			return chr
		}
		c.Pos++
	}
	//support method like "@Mapper"
	if c.Pos == length && c.currentState == tok_name {
		c.currentState = 0
		return tok_name
	}
	return 0
}

func (c *CommentLexer) Error(s string) {
	c.Err = s
}
