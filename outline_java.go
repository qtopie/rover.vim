package main

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

// parseJavaOutline parses Java source code into structured symbols using Tree-Sitter
func parseJavaOutline(filename string, src []byte) ([]SymbolItem, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	var symbols []SymbolItem
	traverseJavaNode(rootNode, src, 0, &symbols)
	return symbols, nil
}

func traverseJavaNode(node *sitter.Node, src []byte, indent int, symbols *[]SymbolItem) {
	if node == nil {
		return
	}

	nodeType := node.Type()

	switch nodeType {
	case "class_declaration":
		name := getChildContentByName(node, "name", src)
		line := int(node.StartPoint().Row) + 1
		prefix := strings.Repeat("   ", indent)
		*symbols = append(*symbols, SymbolItem{
			Name:   name,
			Kind:   KindClass,
			Line:   line,
			Indent: indent,
			Text:   fmt.Sprintf("%s[class] %s", prefix, name),
		})
		traverseChildren(node, src, indent+1, symbols)

	case "interface_declaration":
		name := getChildContentByName(node, "name", src)
		line := int(node.StartPoint().Row) + 1
		prefix := strings.Repeat("   ", indent)
		*symbols = append(*symbols, SymbolItem{
			Name:   name,
			Kind:   KindInterface,
			Line:   line,
			Indent: indent,
			Text:   fmt.Sprintf("%s[interface] %s", prefix, name),
		})
		traverseChildren(node, src, indent+1, symbols)

	case "enum_declaration":
		name := getChildContentByName(node, "name", src)
		line := int(node.StartPoint().Row) + 1
		prefix := strings.Repeat("   ", indent)
		*symbols = append(*symbols, SymbolItem{
			Name:   name,
			Kind:   KindEnum,
			Line:   line,
			Indent: indent,
			Text:   fmt.Sprintf("%s[enum] %s", prefix, name),
		})
		traverseChildren(node, src, indent+1, symbols)

	case "method_declaration", "constructor_declaration":
		name := getChildContentByName(node, "name", src)
		line := int(node.StartPoint().Row) + 1
		prefix := strings.Repeat("   ", indent)
		*symbols = append(*symbols, SymbolItem{
			Name:   name,
			Kind:   KindMethod,
			Line:   line,
			Indent: indent,
			Text:   fmt.Sprintf("%s[method] %s()", prefix, name),
		})

	case "field_declaration":
		// Field declaration may contain variable_declarator
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "variable_declarator" {
				name := getChildContentByName(child, "name", src)
				if name == "" {
					// Fallback to first identifier inside declarator
					name = getFirstIdentifierContent(child, src)
				}
				line := int(child.StartPoint().Row) + 1
				prefix := strings.Repeat("   ", indent)
				*symbols = append(*symbols, SymbolItem{
					Name:   name,
					Kind:   KindField,
					Line:   line,
					Indent: indent,
					Text:   fmt.Sprintf("%s[field] %s", prefix, name),
				})
			}
		}

	default:
		traverseChildren(node, src, indent, symbols)
	}
}

func traverseChildren(node *sitter.Node, src []byte, indent int, symbols *[]SymbolItem) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		traverseJavaNode(child, src, indent, symbols)
	}
}

func getChildContentByName(node *sitter.Node, childName string, src []byte) string {
	child := node.ChildByFieldName(childName)
	if child != nil {
		return child.Content(src)
	}
	return getFirstIdentifierContent(node, src)
}

func getFirstIdentifierContent(node *sitter.Node, src []byte) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		c := node.Child(i)
		if c.Type() == "identifier" {
			return c.Content(src)
		}
	}
	return ""
}
