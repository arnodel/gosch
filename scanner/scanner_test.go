package scanner

import (
	"fmt"
	"reflect"
	"testing"
)

type tok struct {
	tp        Type
	lit       string
	pos, l, c int
}

func (t tok) Token() *Token {
	line := t.l
	if line == 0 {
		line = 1
	}
	col := t.c
	if col == 0 {
		col = 1
	}
	return &Token{
		Type: t.tp,
		Lit:  []byte(t.lit),
		Pos:  Pos{Offset: t.pos, Line: line, Column: col},
	}
}

func tokenString(t *Token) string {
	return fmt.Sprintf("Type:%d, Lit:%q, Pos:%s", t.Type, t.Lit, t.Pos)
}

func TestScanner(t *testing.T) {
	tests := []struct {
		text string
		toks []tok
		err  string
	}{
		//
		// Tokens
		//
		{
			`hello there->2`,
			[]tok{
				{TokIdentifier, "hello", 0, 1, 1},
				{TokIdentifier, "there->2", 6, 1, 7},
				{EOF, "", 14, 1, 15},
			},
			"",
		},
		{
			`no-delim#`,
			[]tok{
				{TokIdentifier, "no-delim", 0, 1, 1},
				{INVALID, "", 8, 1, 9},
			},
			"expected delimiter or EOF",
		},
		{
			`|a token|hello`,
			[]tok{
				{TokIdentifier, "|a token|", 0, 1, 1},
				{TokIdentifier, "hello", 9, 1, 10},
				{EOF, "", 14, 1, 15},
			},
			"",
		},
		{
			`  |a \token\xff;|  `,
			[]tok{
				{TokIdentifier, `|a \token\xff;|`, 2, 1, 3},
				{EOF, "", 19, 1, 20},
			},
			"",
		},
		{
			`|a \invalid escape|`,
			[]tok{{INVALID, `|a \i`, 0, 1, 1}},
			"invalid escape sequence",
		},
		{
			`|not closed`,
			[]tok{{INVALID, `|not closed`, 0, 1, 1}},
			"unexpected EOF",
		},
		{
			`. .. .+.! .\`,
			[]tok{
				{TokDot, ".", 0, 1, 1},
				{TokIdentifier, "..", 2, 1, 3},
				{TokIdentifier, ".+.!", 5, 1, 6},
				{INVALID, `.\`, 10, 1, 11},
			},
			"invalid character after '.'",
		},
		{
			`+ -.ab +infty +.`,
			[]tok{
				{TokIdentifier, "+", 0, 1, 1},
				{TokIdentifier, "-.ab", 2, 1, 3},
				{TokIdentifier, "+infty", 7, 1, 8},
				{INVALID, "+.", 14, 1, 15},
			},
			"invalid character after '.'",
		},
		{
			`"hi \"there""bye`,
			[]tok{
				{TokString, `"hi \"there"`, 0, 1, 1},
				{INVALID, `"bye`, 12, 1, 13},
			},
			"unexpected EOF",
		},
		{
			"\"foo\\  \n  bar\" \"\\z\"",
			[]tok{
				{TokString, "\"foo\\  \n  bar\"", 0, 1, 1},
				{INVALID, `"\z`, 15, 2, 8},
			},
			"invalid escape sequence",
		},
		{
			`#!my-directive #\c; comment`,
			[]tok{
				{TokDirective, "#!my-directive", 0, 1, 1},
				{TokCharacter, `#\c`, 15, 1, 16},
				{EOF, "", 27, 1, 28},
			},
			"",
		},
		{
			`#\xff #\backspace #\b1`,
			[]tok{
				{TokCharacter, `#\xff`, 0, 1, 1},
				{TokCharacter, `#\backspace`, 6, 1, 7},
				{INVALID, `#\b1`, 18, 1, 19},
			},
			"invalid character",
		},
		{
			`a #|b#||||#  |#x`,
			[]tok{
				{TokIdentifier, `a`, 0, 1, 1},
				{TokIdentifier, `x`, 15, 1, 16},
				{EOF, "", 16, 1, 17},
			},
			"",
		},
		{
			`#true #f #; #u8(`,
			[]tok{
				{TokBoolean, `#true`, 0, 1, 1},
				{TokBoolean, `#f`, 6, 1, 7},
				{TokCommentDatum, `#;`, 9, 1, 10},
				{TokOpenByteVec, `#u8(`, 12, 1, 13},
				{EOF, "", 16, 1, 17},
			},
			"",
		},
		{
			"12.4 -1e6 #xff+abi #b1001 #o321",
			[]tok{
				{TokNum, "12.4", 0, 1, 1},
				{TokNum, "-1e6", 5, 1, 6},
				{TokNum, "#xff+abi", 10, 1, 11},
				{TokNum, "#b1001", 19, 1, 20},
				{TokNum, "#o321", 26, 1, 27},
				{EOF, "", 31, 1, 32},
			},
			"",
		},
		{
			"#i12 #x#Iff #e#B101 #b#x10",
			[]tok{
				{TokNum, "#i12", 0, 1, 1},
				{TokNum, "#x#Iff", 5, 1, 6},
				{TokNum, "#e#B101", 12, 1, 13},
				{INVALID, "#b#x", 20, 1, 21},
			},
			"invalid prefix",
		},
		{
			"('a ,,@#()",
			[]tok{
				{TokOpenParen, "(", 0, 1, 1},
				{TokQuote, "'", 1, 1, 2},
				{TokIdentifier, "a", 2, 1, 3},
				{TokComma, ",", 4, 1, 5},
				{TokCommaAt, ",@", 5, 1, 6},
				{TokOpenVec, "#(", 7, 1, 8},
				{TokCloseParen, ")", 9, 1, 10},
				{EOF, "", 10, 1, 11},
			},
			"",
		},
	}
	for i, test := range tests {
		name := fmt.Sprintf("Test %d", i+1)
		t.Run(name, func(t *testing.T) {
			scanner := New("test", []byte(test.text))
			for j, ts := range test.toks {
				next := scanner.Scan()
				if next == nil {
					t.Fatalf("Token %d: scan returns nil", j+1)
				}
				if !reflect.DeepEqual(next, ts.Token()) {
					t.Fatalf("Token %d: expected <%s>, got <%s>", j+1, tokenString(ts.Token()), tokenString(next))
				}
			}
			if scanner.ErrorMsg() != test.err {
				t.Fatalf("Wrong error message: expected %q, got %q", test.err, scanner.ErrorMsg())
			}
		})
	}
}
