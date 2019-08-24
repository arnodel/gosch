package scanner

import (
	"testing"
)

func Test_numberScanner_scanComplex(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		radix  rune
		errMsg string
	}{
		{
			"simple decimal uinteger",
			"1234",
			'd',
			"",
		},
		{
			"simple octal uinteger",
			"3417",
			'o',
			"",
		},
		{
			"simple binary uinteger",
			"10011",
			'b',
			"",
		},
		{
			"simple hex uinteger",
			"1bfA78",
			'x',
			"",
		},
		{
			"invalid decimal uinteger",
			"12a34",
			'd',
			"expected delimiter or EOF",
		},
		{
			"simple octal uinteger",
			"12384",
			'o',
			"expected delimiter or EOF",
		},
		{
			"simple binary uinteger",
			"1012",
			'b',
			"expected delimiter or EOF",
		},
		{
			"simple hex uinteger",
			"1bfg78",
			'x',
			"expected delimiter or EOF",
		},
		{
			"fraction",
			"123/456",
			'd',
			"",
		},
		{
			"simple complex",
			"12+4i",
			'o',
			"",
		},
		{
			"hex complex",
			"ff-12abi",
			'x',
			"",
		},
		{
			"binary complex",
			"-10/11+11/100i",
			'b',
			"",
		},
		{
			"simple decimal",
			"3.1416",
			'd',
			"",
		},
		{
			"decimal without digits",
			".",
			'd',
			"invalid number",
		},
		{
			"decimal with exponent",
			".35e10",
			'd',
			"",
		},
		{
			"no digits after exponent",
			"1.2Exx",
			'd',
			"invalid exponent",
		},
		{
			"binary decimal (invalid)",
			"1.1",
			'b',
			"invalid number",
		},
		{
			"invalid fraction",
			"12/89",
			'o',
			"invalid fraction",
		},
		{
			"fraction without numerator",
			"/12",
			'd',
			"invalid number",
		},
		{
			"+i",
			"+i",
			'd',
			"",
		},
		{
			"2i",
			"2i",
			'd',
			"",
		},
		{
			"@",
			"-2/3@-3.56",
			'd',
			"",
		},
		{
			"nan",
			"+nan.0",
			'x',
			"",
		},
		{
			"inf",
			"-inf.0",
			'b',
			"",
		},
		{
			"something that starts like inf",
			"+in2",
			'd',
			"invalid number",
		},
		{
			"infi",
			"+inf.0i",
			'd',
			"",
		},
		{
			"img inf",
			"2/3+inf.0i",
			'd',
			"",
		},
		{
			"img i",
			"5-i",
			'd',
			"",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newNumberScanner(tt.radix)
			scanner := New("test", []byte(tt.text))
			next := n.scanComplex(scanner)
			for next != nil {
				next = next(scanner)
			}

			if tt.errMsg != scanner.ErrorMsg() {
				t.Errorf("expected error %q, got %q", tt.errMsg, scanner.ErrorMsg())
			} else if tt.errMsg == "" && scanner.pos.Offset != len(tt.text) {
				t.Error("text not consumed")
			}
		})
	}
}
