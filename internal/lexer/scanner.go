package lexer

import (
	"fmt"
	"github.com/levpaul/glocks/internal/token"
	"go.uber.org/zap"
	"strconv"
)

var keywordMap = map[string]token.TokenType{
	"and":    token.AND,
	"class":  token.CLASS,
	"else":   token.ELSE,
	"false":  token.FALSE,
	"fun":    token.FUN,
	"for":    token.FOR,
	"if":     token.IF,
	"nil":    token.NIL,
	"or":     token.OR,
	"print":  token.PRINT,
	"return": token.RETURN,
	"super":  token.SUPER,
	"this":   token.THIS,
	"true":   token.TRUE,
	"var":    token.VAR,
	"while":  token.WHILE,
}

type Scanner struct {
	source string
	log    zap.SugaredLogger
	tokens []token.Token

	start   int
	current int
	line    int
}

func NewScanner(source string, log zap.SugaredLogger) *Scanner {
	return &Scanner{
		source: source,
		log:    log,
		tokens: []token.Token{},
		line:   1,
	}
}

// ScanTokens scans all text in source of scanner and returns them as tokens
func (s *Scanner) ScanTokens() []token.Token {
	for !s.isAtEnd() {
		// Reset start of current token being parsed
		s.start = s.current
		if err := s.scanToken(); err != nil {
			s.log.With("error", err).Errorf("Failed to scan token at line %d\n", s.line)
		}
	}

	s.tokens = append(s.tokens, token.Token{
		Type:    token.EOF,
		Lexeme:  "",
		Literal: nil,
		Line:    s.line,
	})

	return s.tokens
}

func (s *Scanner) scanToken() error {
	r := s.advance()
	switch r {
	case '(':
		s.addToken(token.LEFT_PAREN)
	case ')':
		s.addToken(token.RIGHT_PAREN)
	case '{':
		s.addToken(token.LEFT_BRACE)
	case '}':
		s.addToken(token.RIGHT_BRACE)
	case ',':
		s.addToken(token.COMMA)
	case '.':
		s.addToken(token.DOT)
	case '-':
		s.addToken(token.MINUS)
	case '+':
		s.addToken(token.PLUS)
	case ';':
		s.addToken(token.SEMICOLON)
	case '*':
		s.addToken(token.STAR)
	case '!':
		s.addToken(s.matchTern('=', token.BANG_EQUAL, token.BANG))
	case '=':
		s.addToken(s.matchTern('=', token.EQUAL_EQUAL, token.EQUAL))
	case '<':
		s.addToken(s.matchTern('=', token.LESS_EQUAL, token.LESS))
	case '>':
		s.addToken(s.matchTern('=', token.GREATER_EQUAL, token.GREATER))
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(token.SLASH)
		}
		//  === ignoring whitespace ===
	case ' ':
	case '\r':
	case '\t':
		//  === ignoring whitespace ===
	case '\n':
		s.line++
	case '"':
		s.scanString()
	default:
		switch {
		case isDigit(r):
			if err := s.scanNumber(); err != nil {
				return err
			}
		case isAlpha(r):
			s.scanIdentifier()
		default:
			return fmt.Errorf("unexpected character '%s' at line %d", r, s.line)
		}
	}

	return nil
}

func (s *Scanner) scanIdentifier() {
	for isAlpha(s.peek()) { // scan through initial digits
		s.advance()
	}
	identifier := s.source[s.start:s.current]

	tt, ok := keywordMap[identifier]
	if !ok {
		tt = token.IDENTIFIER
	}
	s.addToken(tt)
}

func (s *Scanner) scanNumber() error {
	for isDigit(s.peek()) { // scan through initial digits
		s.advance()
	}

	if s.peek() == '.' {
		s.advance()
		if isDigit(s.peek()) {
			s.advance()
			for isDigit(s.peek()) { // scan through decimal digits
				s.advance()
			}
		} else {
			return fmt.Errorf("invalid number detected at line %d", s.line)
		}
	}

	val, err := strconv.ParseFloat(s.source[s.start:s.current], 64)
	if err != nil {
		return fmt.Errorf("error converting number to float - %w", err)
	}
	s.addLiteralToken(token.NUMBER, val)
	return nil
}

func (s *Scanner) scanString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	s.advance() // skip last quote
	s.addLiteralToken(token.STRING, s.source[s.start+1:s.current-1])
}

// Returns rune from source at current index, without advancing the index
func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}
	return rune(s.source[s.current])
}

// Returns rune from source at current index + 1, without advancing the index
func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return rune(s.source[s.current+1])
}

// match retuns a bool based on whether the next character in the source matches the given r.
// In case a match is found, the scanner is advanced by one.
func (s *Scanner) match(r rune) bool {
	if s.isAtEnd() {
		return false
	}

	if rune(s.source[s.current]) != r {
		return false
	}

	// advance forward where next char is a match
	s.current++
	return true
}

// matchTern acts as a ternary operator, returning the positive or negative TokenType based on
// whether the next character in the source matches the given r. In case a match is found, the
// scanner is advanced by one.
func (s *Scanner) matchTern(r rune, posRes token.TokenType, negRes token.TokenType) token.TokenType {
	if s.match(r) {
		return posRes
	} else {
		return negRes
	}
}

func (s *Scanner) addToken(t token.TokenType) {
	s.addLiteralToken(t, nil)
}

func (s *Scanner) addLiteralToken(t token.TokenType, lit any) {
	s.tokens = append(s.tokens, token.Token{
		Type:    t,
		Lexeme:  s.source[s.start:s.current],
		Literal: lit,
		Line:    s.current,
	})
}

// advance reads the next rune (char) from the source and returns it, advancing the current index
func (s *Scanner) advance() rune {
	r := rune(s.source[s.current])
	s.current++
	return r
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
