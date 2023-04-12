package lexer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
var reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
var reIdentifier = regexp.MustCompile(`^[_\d\w]+`)

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

type Lexer struct {
	chunk         string // 源代码
	chunkName     string // 源文件名
	line          int    // 当前行号
	nextToken     string
	nextTokenKind int
	nextTokenLine int
}

func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{chunk, chunkName, 1, "", 0, 0}
}

func (lex *Lexer) NextToken() (line, kind int, token string) {
	if lex.nextTokenLine > 0 {
		line = lex.nextTokenLine
		kind = lex.nextTokenKind
		token = lex.nextToken
		lex.line = lex.nextTokenLine
		lex.nextTokenLine = 0
		return
	}

	lex.skipWhiteSpaces()

	if len(lex.chunk) == 0 {
		return lex.line, TOKEN_EOF, "EOF"
	}

	switch lex.chunk[0] {
	case ';':
		lex.next(1)
		return lex.line, TOKEN_SEP_SEMI, ";"
	case ',':
		lex.next(1)
		return lex.line, TOKEN_SEP_COMMA, ","
	case '(':
		lex.next(1)
		return lex.line, TOKEN_SEP_LPAREN, "("
	case ')':
		lex.next(1)
		return lex.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		lex.next(1)
		return lex.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		lex.next(1)
		return lex.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		lex.next(1)
		return lex.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		lex.next(1)
		return lex.line, TOKEN_OP_ADD, "+"
	case '-':
		lex.next(1)
		return lex.line, TOKEN_OP_MINUS, "-"
	case '*':
		lex.next(1)
		return lex.line, TOKEN_OP_MUL, "*"
	case '^':
		lex.next(1)
		return lex.line, TOKEN_OP_POW, "^"
	case '%':
		lex.next(1)
		return lex.line, TOKEN_OP_MOD, "%"
	case '&':
		lex.next(1)
		return lex.line, TOKEN_OP_BAND, "&"
	case '|':
		lex.next(1)
		return lex.line, TOKEN_OP_BOR, "|"
	case '#':
		lex.next(1)
		return lex.line, TOKEN_OP_LEN, "#"
	case ':':
		if lex.test("::") {
			lex.next(2)
			return lex.line, TOKEN_SEP_LABEL, "::"
		} else {
			lex.next(1)
			return lex.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if lex.test("//") {
			lex.next(2)
			return lex.line, TOKEN_OP_IDIV, "//"
		} else {
			lex.next(1)
			return lex.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if lex.test("~=") {
			lex.next(2)
			return lex.line, TOKEN_OP_NE, "~="
		} else {
			lex.next(1)
			return lex.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if lex.test("==") {
			lex.next(2)
			return lex.line, TOKEN_OP_EQ, "=="
		} else {
			lex.next(1)
			return lex.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if lex.test("<<") {
			lex.next(2)
			return lex.line, TOKEN_OP_SHL, "<<"
		} else if lex.test("<=") {
			lex.next(2)
			return lex.line, TOKEN_OP_LE, "<="
		} else {
			lex.next(1)
			return lex.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if lex.test(">>") {
			lex.next(2)
			return lex.line, TOKEN_OP_SHR, ">>"
		} else if lex.test(">=") {
			lex.next(2)
			return lex.line, TOKEN_OP_GE, ">="
		} else {
			lex.next(1)
			return lex.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if lex.test("...") {
			lex.next(3)
			return lex.line, TOKEN_VARARG, "..."
		} else if lex.test("..") {
			lex.next(2)
			return lex.line, TOKEN_OP_CONCAT, ".."
		} else if len(lex.chunk) == 1 || !isDigit(lex.chunk[1]) {
			lex.next(1)
			return lex.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if lex.test("[[") || lex.test("[=") {
			return lex.line, TOKEN_STRING, lex.scanLongString()
		} else {
			lex.next(1)
			return lex.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return lex.line, TOKEN_STRING, lex.scanShortString()
	}

	c := lex.chunk[0]
	if c == '.' || isDigit(c) {
		token := lex.scanNumber()
		return lex.line, TOKEN_NUMBER, token
	}

	if c == '_' || isLatter(c) {
		token := lex.scanIdentifier()
		if kind, found := keywords[token]; found {
			return line, kind, token // keyword
		} else {
			return line, TOKEN_IDENTIFIER, token
		}
	}

	lex.error("unexpected symbol near %q", c)
	return
}

func (lex *Lexer) LookAhead() int {
	if lex.nextTokenLine > 0 {
		return lex.nextTokenKind
	}

	currentLine := lex.line
	line, kind, token := lex.NextToken()
	lex.line = currentLine
	lex.nextTokenLine = line
	lex.nextTokenKind = kind
	lex.nextToken = token
	return kind
}

func (lex *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := lex.NextToken()
	if kind != _kind {
		lex.error("syntax error near '%s'", token)
	}
	return line, token
}

func (lex *Lexer) NextIdentifier() (line int, token string) {
	return lex.NextTokenOfKind(TOKEN_IDENTIFIER)
}

func (lex *Lexer) Line() int {
	return lex.line
}

func isLatter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (lex *Lexer) scanIdentifier() string {
	return lex.scan(reIdentifier)
}

func (lex *Lexer) scanNumber() string {
	return lex.scan(reNumber)
}

func (lex *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(lex.chunk); token != "" {
		lex.next(len(token))
		return token
	}
	panic("unreachable!")
}

func (lex *Lexer) skipWhiteSpaces() {
	for len(lex.chunk) > 0 {
		if lex.test("--") {
			lex.skipComment()
		} else if lex.test("\r\n") || lex.test("\n\r") {
			lex.next(2)
			lex.line += 1
		} else if isNewLine(lex.chunk[0]) {
			lex.next(1)
			lex.line += 1
		} else if isWhiteSpace(lex.chunk[0]) {
			lex.next(1)
		} else {
			break
		}
	}
}

func (lex *Lexer) test(s string) bool {
	return strings.HasPrefix(lex.chunk, s)
}

func (lex *Lexer) next(n int) {
	lex.chunk = lex.chunk[n:]
}

func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

func (lex *Lexer) skipComment() {
	lex.next(2)        // skip --
	if lex.test("[") { // long comment ?
		if reOpeningLongBracket.FindString(lex.chunk) != "" {
			lex.scanLongString()
			return
		}
	}

	// short comment
	for len(lex.chunk) > 0 && !isNewLine(lex.chunk[0]) {
		lex.next(1)
	}
}

func (lex *Lexer) scanLongString() string {
	openingLongBracket := reOpeningLongBracket.FindString(lex.chunk)
	if openingLongBracket == "" {
		lex.error("invalid long string delimiter near '%s'",
			lex.chunk[0:2])
	}

	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIdx := strings.Index(lex.chunk, closingLongBracket)
	if closingLongBracketIdx < 0 {
		lex.error("unfinished long string or comment")
	}

	str := lex.chunk[len(openingLongBracket):closingLongBracketIdx]
	lex.next(closingLongBracketIdx + len(closingLongBracket))

	str = reNewLine.ReplaceAllString(str, "\n")
	lex.line += strings.Count(str, "\n")
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}

	return str
}

func (lex *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	err = fmt.Sprintf("%s:%d: %s", lex.chunkName, lex.line, err)
	panic(err)
}

func (lex *Lexer) scanShortString() string {
	if str := reShortStr.FindString(lex.chunk); str != "" {
		lex.next(len(str))
		str = str[1 : len(str)-1]
		if strings.Contains(str, `\`) {
			lex.line += len(reNewLine.FindAllString(str, -1))
			str = lex.escape(str)
		}
		return str
	}
	lex.error("unfinished string")
	return ""
}

func (lex *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			lex.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				lex.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				lex.error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		lex.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}
