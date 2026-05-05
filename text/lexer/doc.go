// Package lexer is a small toolkit for hand-written state-function
// parsers: a UTF-8-aware [Cursor] over a string source, a text-shaped
// emit buffer, and a generic [StateFn] machine driver.
//
// The API speaks runes and strings; UTF-8 encoding is an
// implementation detail.
//
// Typical use embeds *[Cursor] in the caller's parser state:
//
//	type parser struct {
//	    *lexer.Cursor
//	    // ...additional state
//	}
//
//	func stateStart(p *parser) (lexer.StateFn[*parser], error) {
//	    r, ok := p.Peek()
//	    if !ok {
//	        return nil, nil
//	    }
//	    // ...
//	}
//
//	err := lexer.Run(&parser{Cursor: lexer.New(line)}, stateStart)
package lexer
