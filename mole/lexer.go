package mole

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
			c.currentState = tok_name
			c.PrevPos = c.Pos
		case c.currentState == tok_name:
			//collect a-z A-Z - _
			switch {
			case (chr >= '0' && chr <= '9'):
			case (chr >= 'a' && chr <= 'z'):
			case chr == '_' || chr == '-':
			default:
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
				break
			} else {
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
				tmphex += 10 + int64(chr-'a')
			case (chr >= 'A' && chr <= 'F'):
				tmphex *= 16
				tmphex += 10 + int64(chr-'A')
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
	return 0
}

func (c *CommentLexer) Error(s string) {
	c.Err = s
}
