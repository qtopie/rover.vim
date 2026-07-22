package main

type SymbolKind string

const (
	KindClass     SymbolKind = "class"
	KindInterface SymbolKind = "interface"
	KindEnum      SymbolKind = "enum"
	KindStruct    SymbolKind = "struct"
	KindMethod    SymbolKind = "method"
	KindFunction  SymbolKind = "function"
	KindField     SymbolKind = "field"
	KindVariable  SymbolKind = "variable"
	KindConstant  SymbolKind = "constant"
)

type SymbolItem struct {
	Name   string     `json:"name"`
	Kind   SymbolKind `json:"kind"`
	Line   int        `json:"line"`
	Indent int        `json:"indent"`
	Text   string     `json:"text"`
}
