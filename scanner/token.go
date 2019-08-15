package scanner

import (
	"fmt"
)

type Token struct {
	Type
	Lit []byte
	Pos
}

func (t *Token) String() string {
	if t == nil {
		return "nil"
	}
	return fmt.Sprintf("Token(type=%d, lit=%s, pos=%s", t.Type, t.Lit, t.Pos)
}

type Type int

const (
	INVALID Type = iota
	EOF
	TokCommentDatum

	TokIdentifier

	TokOpenParen
	TokCloseParen
	TokQuote
	TokBackQuote
	TokComma
	TokCommaAt
	TokDot
	TokNum
	TokString
	TokOpenVec
	TokOpenByteVec
	TokBoolean
	TokDirective
	TokCharacter
)

type Pos struct {
	Offset int
	Line   int
	Column int
}

func (p Pos) String() string {
	return fmt.Sprintf("Pos(offset=%d, line=%d, column=%d)", p.Offset, p.Line, p.Column)
}
