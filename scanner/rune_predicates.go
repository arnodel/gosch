package scanner

func isDelimiter(r rune) bool {
	if isWhitespace(r) {
		return true
	}
	switch r {
	case '|', '(', ')', '"', ';', -1:
		return true
	}
	return false
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isSpecialInitial(r rune) bool {
	switch r {
	case '!', '$', '%', '&', '*', '/', ':', '<', '=', '>', '?', '^', '_', '~':
		return true
	default:
		return false
	}
}

func isInitial(r rune) bool {
	return isLetter(r) || isSpecialInitial(r)
}

func isSubsequent(r rune) bool {
	return isInitial(r) || isDigit(r) || isSpecialSubsequent(r)
}

func isSignSubsequent(r rune) bool {
	return isInitial(r) || r == '+' || r == '-' || r == '@'
}

func isDotSubsequent(r rune) bool {
	return isSignSubsequent(r) || r == '.'
}

func isSpecialSubsequent(r rune) bool {
	return r == '+' || r == '-' || r == '.' || r == '@'
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func isInlineWhitespace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
