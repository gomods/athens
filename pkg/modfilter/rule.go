package modfilter

type ruleNode struct {
	next map[string]ruleNode
	rule Rule
}
