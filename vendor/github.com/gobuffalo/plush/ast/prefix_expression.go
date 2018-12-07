package ast

import (
	"bytes"
)

type PrefixExpression struct {
	TokenAble
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	if pe.Right != nil {
		out.WriteString(pe.Right.String())
	} else {
		out.WriteString("<pe.Right == nil>")
	}
	out.WriteString(")")

	return out.String()
}
