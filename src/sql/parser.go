package sql

import (
	"fmt"
	"io"
)

// SelectStatement represents a SQL SELECT statement.
type SelectStatement struct {
	Fields    []string
	TableName string
}

type InsertStatement struct {
	MapValues map[string]interface{}
	Values    []interface{}
	TableName string
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a SQL SELECT statement.
func (p *Parser) Parse() (interface{}, error) {

	//Find the first token
	tok, lit := p.scanIgnoreWhitespace()

	switch tok {
	case SELECT:
		p.unscan() //We unscan to get back to what we've already parsed
		return p.ParseSelect()
	case INSERT:
		p.unscan()
		return p.ParseInsert()
	default:
		return nil, fmt.Errorf("found %q, expected SELECT/INSERT/CREATE", lit)
	}

	// Return the successfully parsed statement.
	return nil, fmt.Errorf("Could not parse the statement")
}

//ParseInsert parses the given string and returns an instance of InsertStatement
func (p *Parser) ParseInsert() (InsertStatement, error) {
	stmt := &InsertStatement{}

	if tok, lit := p.scanIgnoreWhitespace(); tok != INSERT {
		return *stmt, fmt.Errorf("found %q, expected INSERT", lit)
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != INTO {
		return *stmt, fmt.Errorf("found %q, expected INTO", lit)
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != IDENT {
		return *stmt, fmt.Errorf("found %q, table name cannot be a keyword", lit)
	} else {
		stmt.TableName = lit
	}

	if tok, lit := p.scanIgnoreWhitespace(); tok != VALUES {
		return *stmt, fmt.Errorf("found %q, expected VALUES", lit)
	}

	return *stmt, nil
}

//ParseSelect parses the given string and returns an instance of SelectStatement
func (p *Parser) ParseSelect() (SelectStatement, error) {
	stmt := &SelectStatement{}
	// First token should be a "SELECT" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != SELECT {
		return *stmt, fmt.Errorf("found %q, expected SELECT", lit)
	}

	// Next we should loop over all our comma-delimited fields.
	for {
		// Read a field.
		tok, lit := p.scanIgnoreWhitespace()
		if tok != IDENT && tok != ASTERISK {
			return *stmt, fmt.Errorf("found %q, expected field", lit)
		}
		stmt.Fields = append(stmt.Fields, lit)

		// If the next token is not a comma then break the loop.
		if tok, _ := p.scanIgnoreWhitespace(); tok != COMMA {
			p.unscan()
			break
		}
	}

	// Next we should see the "FROM" keyword.
	if tok, lit := p.scanIgnoreWhitespace(); tok != FROM {
		return *stmt, fmt.Errorf("found %q, expected FROM", lit)
	}

	// Finally we should read the table name.
	tok, lit := p.scanIgnoreWhitespace()
	if tok != IDENT {
		return *stmt, fmt.Errorf("found %q, expected table name", lit)
	}
	stmt.TableName = lit

	// Return the successfully parsed statement.
	return *stmt, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
