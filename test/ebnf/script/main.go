package main

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/soderasen-au/go-qlik/ebnf/script"
)

type treeShapeListener struct {
	*script.BaseScriptListener
}

func (s *treeShapeListener) VisitTerminal(node antlr.TerminalNode) {
	fmt.Printf("Terminal: %s: %\n", node.GetText())
}

func (s *treeShapeListener) VisitErrorNode(node antlr.ErrorNode) {
	fmt.Printf("Error: %s\n", node.GetText())
}

const (
	scriptInput string = `date(date#(stamp, 'YYY'), 'abc')`
)

func main() {
	input := antlr.NewInputStream(scriptInput)
	lexer := script.NewScriptLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := script.NewScriptParser(stream)
	tree := p.Expression()

	listener := &treeShapeListener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
}
