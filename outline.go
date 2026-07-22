package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/govim/govim"
	"github.com/qtopie/sniphunt/pkg/search"
)

// showDocumentOutline displays the AST symbol outline for current buffer in Vim's Location List
func showDocumentOutline(g govim.Govim, flags govim.CommandFlags, args ...string) error {
	rawPath, err := g.ChannelExpr("expand('%:p')")
	if err != nil || len(rawPath) == 0 {
		return showErrMsg(g, "Failed to get current buffer path")
	}

	var filePath string
	_ = json.Unmarshal(rawPath, &filePath)
	if filePath == "" {
		return showErrMsg(g, "Please save the buffer to a file first")
	}

	rawLines, err := g.ChannelExpr("getline(1, '$')")
	if err != nil || len(rawLines) == 0 {
		return showErrMsg(g, "Failed to read buffer content")
	}

	var lines []string
	if err := json.Unmarshal(rawLines, &lines); err != nil {
		return showErrMsg(g, "Failed to parse buffer lines: %v", err)
	}

	srcContent := []byte(strings.Join(lines, "\n"))

	var symbols []SymbolItem
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		symbols, err = parseGoOutline(filePath, srcContent)
	case ".java":
		symbols, err = parseJavaOutline(filePath, srcContent)
	default:
		return showErrMsg(g, "Outline is currently supported for .go and .java files (current extension: %s)", ext)
	}

	if err != nil {
		return showErrMsg(g, "Failed to parse document outline: %v", err)
	}

	if len(symbols) == 0 {
		return showMsg(g, "No outline symbols found in %s", filepath.Base(filePath))
	}

	var locList []QuickfixItem
	for _, s := range symbols {
		locList = append(locList, QuickfixItem{
			Filename: filePath,
			Lnum:     s.Line,
			Text:     s.Text,
		})
	}

	locJSON, err := json.Marshal(locList)
	if err != nil {
		return showErrMsg(g, "Failed to encode outline list: %v", err)
	}

	cmd := fmt.Sprintf("call setloclist(0, %s, 'r') | lopen", string(locJSON))
	if err := g.ChannelEx(cmd); err != nil {
		return showErrMsg(g, "Failed to open location list window: %v", err)
	}

	return showMsg(g, "Outline loaded %d symbols for %s", len(symbols), filepath.Base(filePath))
}

// gotoSymbolLocation jumps to class/variable/symbol declaration using AST & Sniphunt
func gotoSymbolLocation(g govim.Govim, flags govim.CommandFlags, args ...string) error {
	var symbol string
	if len(args) > 0 {
		symbol = strings.Join(args, " ")
	} else {
		rawWord, err := g.ChannelExpr("expand('<cword>')")
		if err == nil && len(rawWord) > 0 {
			_ = json.Unmarshal(rawWord, &symbol)
		}
	}

	if symbol == "" {
		return showErrMsg(g, "Usage: :RoverGoto <symbol_name> (or place cursor on a symbol)")
	}

	// Search current directory and JDK source cache
	searchDirs := []string{}
	if cwd, err := os.Getwd(); err == nil {
		searchDirs = append(searchDirs, cwd)
	}
	if jdkDir, err := getOrPrepareJdkSourceDir(g); err == nil {
		searchDirs = append(searchDirs, jdkDir)
	}

	var qfList []QuickfixItem
	s := search.NewSearcher()
	s.Extensions = []string{".go", ".java"}

	ctx := context.Background()

	// Try declaration regex patterns: e.g. "class PriorityQueue", "struct SymbolItem", "interface Govim", etc.
	declPatterns := []string{
		fmt.Sprintf(`(class|interface|enum|type|struct)\s+%s\b`, symbol),
		fmt.Sprintf(`(func|def|void|int|boolean|String|public|private)\s+.*?\b%s\b`, symbol),
		fmt.Sprintf(`\b%s\b`, symbol),
	}

	for _, pattern := range declPatterns {
		for _, dir := range searchDirs {
			matchChan, errChan := s.Search(ctx, dir, pattern)
			for match := range matchChan {
				textSnippet := strings.TrimSpace(string(match.Text))
				qfList = append(qfList, QuickfixItem{
					Filename: match.Path,
					Lnum:     match.LineNum,
					Text:     textSnippet,
				})
				if len(qfList) >= 100 {
					break
				}
			}
			_ = <-errChan
			if len(qfList) > 0 {
				break
			}
		}
		if len(qfList) > 0 {
			break
		}
	}

	if len(qfList) == 0 {
		return showMsg(g, "No declaration found for symbol '%s'", symbol)
	}

	qfJSON, err := json.Marshal(qfList)
	if err != nil {
		return showErrMsg(g, "Failed to encode search results: %v", err)
	}

	// If single match, jump directly. If multiple matches, open quickfix list.
	if len(qfList) == 1 {
		cmd := fmt.Sprintf("edit +%d %s", qfList[0].Lnum, qfList[0].Filename)
		return g.ChannelEx(cmd)
	}

	cmd := fmt.Sprintf("call setqflist(%s, 'r') | copen", string(qfJSON))
	if err := g.ChannelEx(cmd); err != nil {
		return showErrMsg(g, "Failed to open quickfix window: %v", err)
	}

	return showMsg(g, "Found %d declaration matches for '%s'", len(qfList), symbol)
}
