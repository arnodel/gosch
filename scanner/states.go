package scanner

import "unicode"

func scanToken(l *Scanner) stateFn {
	switch c := l.next(); {
	case isInitial(c):
		return scanIdentifierSubsequent
	case isWhitespace(c):
		l.ignore()
	default:
		switch c {
		case '|':
			return scanIdentifierInBars(l)
		case '(':
			l.emit(TokOpenParen)
		case ')':
			l.emit(TokCloseParen)
		case '#':
			return scanHashToken
		case '"':
			return scanString
		case '\'':
			l.emit(TokQuote)
		case '`':
			l.emit(TokBackQuote)
		case ',':
			if l.accept("@") {
				l.emit(TokCommaAt)
			} else {
				l.emit(TokComma)
			}
		case ';':
			for {
				c := l.next()
				if c == '\n' || c == -1 {
					l.ignore()
					break
				}
			}
		case '.':
			c = l.next()
			switch {
			case isDelimiter(c):
				l.backup()
				l.emit(TokDot)
			case isDotSubsequent(c):
				return scanIdentifierSubsequent
			case isDigit(c):
				return scanPostDecimalPoint
			default:
				return l.errorf("invalid character after '.'")
			}
		case '+', '-':
			c = l.next()
			switch {
			case isDelimiter(c):
				l.backup()
				l.emit(TokIdentifier)
			case c == '.':
				if !isDotSubsequent(l.next()) {
					return l.errorf("invalid character after '.'")
				}
				fallthrough
			case isSignSubsequent(c):
				return scanIdentifierSubsequent
			default:
				return l.errorf("invalid character after '%c'", c)
			}
		case -1:
			l.emit(EOF)
			return nil
		default:
			return l.errorf("illegal character")
		}
	}
	return scanToken
}

func scanIdentifierSubsequent(l *Scanner) stateFn {
	accept(l, isSubsequent, -1)
	lit := string(l.lit())
	tp := INVALID
	if lit[0] == '+' || lit[0] == '-' {
		tp = identifierExceptions[lit[1:]]
	}
	if tp == INVALID {
		tp = TokIdentifier
	}
	l.emit(tp)
	return checkDelimiter
}

var identifierExceptions = map[string]Type{
	"i":      TokNum,
	"inf.0":  TokNum,
	"nan.0":  TokNum,
	"inf.0i": TokNum,
	"nan.0i": TokNum,
}

func scanIdentifierInBars(l *Scanner) stateFn {
	for {
		c := l.next()
		switch c {
		case -1:
			return unexpectedEOF
		case '|':
			l.emit(TokIdentifier)
			return scanToken
		case '\\':
			if !acceptCharEscape(l, '|', false) {
				return l.errorf("invalid escape sequence")
			}
		}
	}
}

func checkDelimiter(l *Scanner) stateFn {
	if isDelimiter(l.peek()) {
		return scanToken
	}
	return l.errorf("expected delimiter or EOF")
}

func unexpectedEOF(l *Scanner) stateFn {
	return l.errorf("unexpected EOF")
}

func scanHashToken(l *Scanner) stateFn {
	c := l.next()
	switch c {
	case -1:
		return l.errorf("unexpected EOF")
	case '|':
		return scanNestedComment
	case '!':
		return scanDirective
	case '\\':
		return scanCharacter
	case '(':
		l.emit(TokOpenVec)
	case ';':
		l.emit(TokCommentDatum)
	case 'x', 'd', 'b', 'o', 'i', 'e':
		l.backup()
		return scanNumber
	default:
		if !isInitial(c) {
			return l.errorf("invalid character following #")
		}
		accept(l, isSubsequent, -1)
		w := string(l.lit()[1:])
		switch w {
		case "t", "true", "f", "false":
			l.emit(TokBoolean)
			return checkDelimiter
		case "u8":
			if l.next() == '(' {
				l.emit(TokOpenByteVec)
				return scanToken
			}
			return l.errorf("expected '('")
		default:
			return l.errorf("invalid # word")
		}
	}
	return scanToken
}

func scanNumber(l *Scanner) stateFn {
	panic("unimplemented")
}

func scanNestedComment(l *Scanner) stateFn {
	depth := 1
	const opening = 1
	const closing = 2
	state := 0
	for depth > 0 {
		switch c := l.next(); c {
		case -1:
			return unexpectedEOF
		case '#':
			if state == closing {
				depth--
				state = 0
			} else {
				state = opening
			}
		case '|':
			if state == opening {
				depth++
				state = 0
			} else {
				state = closing
			}
		}
	}
	l.ignore()
	return scanToken
}

func scanDirective(l *Scanner) stateFn {
	if !isInitial(l.next()) {
		return l.errorf("invalid directive")
	}
	accept(l, isSubsequent, -1)
	l.emit(TokDirective)
	return checkDelimiter
}

func scanCharacter(l *Scanner) stateFn {
	c := l.next()
	if c == -1 {
		return unexpectedEOF
	}
	d := l.next()
	switch {
	case isDelimiter(d):
		l.backup()
		l.emit(TokCharacter)
		return scanToken
	case (c == 'x' || c == 'X') && isHexDigit(d):
		accept(l, isHexDigit, -1)
	case isInitial(c) && isSubsequent(d):
		accept(l, isSubsequent, -1)
	default:
		return l.errorf("invalid character")
	}
	l.emit(TokCharacter)
	return checkDelimiter
}

func scanString(l *Scanner) stateFn {
	for {
		switch c := l.next(); c {
		case -1:
			return unexpectedEOF
		case '"':
			l.emit(TokString)
			return scanToken
		case '\\':
			if !acceptCharEscape(l, '"', true) {
				return l.errorf("invalid escape sequence")
			}
		}
	}
}

func scanPostDecimalPoint(l *Scanner) stateFn {
	return nil
}

type runePredicate func(rune) bool

func accept(l *Scanner, p runePredicate, max int) int {
	for i := 0; i != max; i++ {
		if !p(l.next()) {
			l.backup()
			return i
		}
	}
	return max
}

func acceptSeqCI(l *Scanner, seq string) int {
	for i, c := range seq {
		if unicode.ToLower(l.next()) != c {
			l.backup()
			return i
		}
	}
	return len(seq)
}

func acceptCharEscape(l *Scanner, delim rune, multiline bool) bool {
	c := l.next()
	switch c {
	case 'x':
		return accept(l, isHexDigit, -1) > 0 && l.next() == ';'
	case 'a', 'b', 't', 'n', 'r', '\\', delim:
		return true
	case ' ', '\t':
		accept(l, isInlineWhitespace, -1)
		if l.next() != '\n' {
			return false
		}
		fallthrough
	case '\n':
		accept(l, isInlineWhitespace, -1)
		return true
	default:
		return false
	}
}
