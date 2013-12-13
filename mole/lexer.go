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
				goto GOON
			case (chr >= 'a' && chr <= 'z'):
				goto GOON
			case chr == '_' || chr == '-':
				goto GOON
			default:
				c.currentState = 0
				lval.s = c.Comment[c.PrevPos:c.Pos]
				return tok_name
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
	GOON:
		c.Pos++
	}
	return 0
}

func (c *CommentLexer) Error(s string) {
	c.Err = s
}
