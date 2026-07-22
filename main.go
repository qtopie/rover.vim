package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qtopie/sniphunt/pkg/search"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("rover - Vim plugin CLI for JDK search, document outline, and goto symbol")
		os.Exit(0)
	}

	subCmd := os.Args[1]
	args := os.Args[2:]

	switch subCmd {
	case "outline":
		runCmdOutline(args)
	case "jdk-search":
		runCmdJdkSearch(args)
	case "goto":
		runCmdGoto(args)
	case "jdk-clean":
		runCmdJdkClean()
	case "image-paste":
		runCmdImagePaste(args)
	case "preview":
		runCmdMarkdownPreview(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subCmd)
		os.Exit(1)
	}
}

func runCmdOutline(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: rover outline <filepath>\n")
		os.Exit(1)
	}
	filePath := args[0]
	srcContent, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var symbols []SymbolItem

	switch ext {
	case ".go":
		symbols, err = parseGoOutline(filePath, srcContent)
	case ".java":
		symbols, err = parseJavaOutline(filePath, srcContent)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported filetype: %s\n", ext)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing outline: %v\n", err)
		os.Exit(1)
	}

	var locList []QuickfixItem
	for _, s := range symbols {
		locList = append(locList, QuickfixItem{
			Filename: filePath,
			Lnum:     s.Line,
			Text:     s.Text,
		})
	}

	jsonOutput(locList)
}

func runCmdJdkSearch(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: rover jdk-search <query>\n")
		os.Exit(1)
	}
	query := strings.Join(args, " ")

	srcDir, err := getOrPrepareJdkSourceDirCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "JDK source resolution error: %v\n", err)
		os.Exit(1)
	}

	s := search.NewSearcher()
	s.Extensions = []string{".java"}

	ctx := context.Background()
	matchChan, errChan := s.Search(ctx, srcDir, query)

	var qfList []QuickfixItem
	for match := range matchChan {
		textSnippet := strings.TrimSpace(string(match.Text))
		qfList = append(qfList, QuickfixItem{
			Filename: match.Path,
			Lnum:     match.LineNum,
			Text:     textSnippet,
		})
		if len(qfList) >= 200 {
			break
		}
	}

	if err := <-errChan; err != nil && len(qfList) == 0 {
		fmt.Fprintf(os.Stderr, "Sniphunt search error: %v\n", err)
		os.Exit(1)
	}

	jsonOutput(qfList)
}

func runCmdGoto(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: rover goto <symbol>\n")
		os.Exit(1)
	}
	symbol := strings.Join(args, " ")

	searchDirs := []string{}
	if cwd, err := os.Getwd(); err == nil {
		searchDirs = append(searchDirs, cwd)
	}
	if jdkDir, err := getOrPrepareJdkSourceDirCLI(); err == nil {
		searchDirs = append(searchDirs, jdkDir)
	}

	var qfList []QuickfixItem
	s := search.NewSearcher()
	s.Extensions = []string{".go", ".java"}

	ctx := context.Background()

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

	jsonOutput(qfList)
}

func runCmdJdkClean() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		userCacheDir = os.TempDir()
	}
	destDir := filepath.Join(userCacheDir, "rover", "jdk_src")
	if err := os.RemoveAll(destDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clean JDK cache: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("JDK source cache cleaned successfully.")
}

func jsonOutput(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "JSON encoding error: %v\n", err)
		os.Exit(1)
	}
}
