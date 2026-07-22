package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed _template/preview.html
var tpl string

type MarkdownContent struct {
	Content string
}

func runCmdMarkdownPreview(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: rover preview <md_file_path>\n")
		os.Exit(1)
	}

	contentBytes, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read file: %v\n", err)
		os.Exit(1)
	}

	buf := renderMarkdown(string(contentBytes))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(buf.Bytes())
	})

	openBrowser("http://127.0.0.1:7070")
	fmt.Println("Markdown preview server started on http://127.0.0.1:7070")
	if err := http.ListenAndServe("127.0.0.1:7070", nil); err != nil {
		log.Fatal(err)
	}
}

func renderMarkdown(content string) bytes.Buffer {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithParagraphTransformers(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		panic(err)
	}
	return buf
}

func openBrowser(url string) (err error) {
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
