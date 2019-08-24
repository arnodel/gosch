package scanner

import (
	"unicode"
)

type numberScanner struct {
	radix      rune
	isDigit    runePredicate
	foundError bool
}

func newNumberScanner(radix rune) *numberScanner {
	radix = unicode.ToLower(radix)
	var isDig runePredicate
	switch radix {
	case 'd':
		isDig = isDigit
	case 'x':
		isDig = isHexDigit
	case 'b':
		isDig = isBinaryDigit
	case 'o':
		isDig = isOctalDigit
	default:
		panic("invalid radix")
	}
	return &numberScanner{
		radix:   radix,
		isDigit: isDig,
	}
}

func (n *numberScanner) error(l *Scanner, msg string) {
	if !n.foundError {
		l.errorf(msg)
		n.foundError = true
	}
}

func (n *numberScanner) uinteger(l *Scanner) bool {
	dc := accept(l, n.isDigit, -1)
	return dc > 0
}

func (n *numberScanner) scanI(l *Scanner) stateFn {
	if unicode.ToLower(l.next()) != 'i' {
		return l.errorf("expected i")
	}
	return checkDelimiter
}

func (n *numberScanner) scanComplex(l *Scanner) stateFn {
	ok, i := n.realOrI(l)
	if ok {
		switch l.next() {
		case '+', '-':
			if n.ureal(l) {
				return n.scanI
			}
			uinfnan, i := n.uinfnanOrI(l)
			if uinfnan {
				return n.scanI
			}
			ok = i
		case '@':
			ok = n.real(l)
		case 'i':
			ok = true
		default:
			l.backup()
		}
	}
	if n.foundError {
		return nil
	}
	if !ok && !i {
		return l.errorf("invalid number")
	}
	l.emit(TokNum)
	return checkDelimiter
}

func (n *numberScanner) real(l *Scanner) bool {
	ok, _ := n.realOrI(l)
	return ok
}

func (n *numberScanner) realOrI(l *Scanner) (bool, bool) {
	c := l.next()
	if c == '-' || c == '+' {
		if n.ureal(l) {
			return true, false
		}
		return n.uinfnanOrI(l)
	}
	l.backup()
	return n.ureal(l), false
}

func (n *numberScanner) uinfnanOrI(l *Scanner) (bool, bool) {
	if n.foundError {
		return false, false
	}
	nchars := acceptSeqCI(l, "inf.0")
	switch nchars {
	case 0:
		nchars = acceptSeqCI(l, "nan.0")
		return nchars == 5, false
	case 1:
		return false, true
	case 5:
		return true, false
	default:
		return false, false
	}
}

func (n *numberScanner) ureal(l *Scanner) bool {
	if n.foundError {
		return false
	}
	startsWithUinteger := n.uinteger(l)
	switch l.next() {
	case '/':
		if !startsWithUinteger {
			l.backup()
			return false
		}
		if !n.uinteger(l) {
			n.error(l, "invalid fraction")
			return false
		}
		return true
	case '.', 'e', 'E':
		l.backup()
		return n.decimal(l, startsWithUinteger)
	default:
		l.backup()
		return startsWithUinteger
	}
}

func (n *numberScanner) decimal(l *Scanner, startsWithNumber bool) bool {
	if n.foundError {
		return false
	}
	if n.radix != 'd' {
		return false
	}
	numberAfterDot := true
	if l.next() == '.' {
		numberAfterDot = n.uinteger(l)
	}
	if !startsWithNumber && !numberAfterDot {
		l.backup()
		return false
	}
	if unicode.ToLower(l.next()) != 'e' {
		l.backup()
		return true
	}
	l.accept("+-")
	nDigits := accept(l, n.isDigit, -1)
	if nDigits == 0 {
		n.error(l, "invalid exponent")
		return false
	}
	return true
}
