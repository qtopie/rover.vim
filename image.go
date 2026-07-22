package main

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"golang.design/x/clipboard"
)

func runCmdImagePaste(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: rover image-paste <md_file_path>\n")
		os.Exit(1)
	}

	mdFilePath := args[0]
	b := clipboard.Read(clipboard.FmtImage)
	if b == nil {
		fmt.Fprintf(os.Stderr, "no image in system clipboard.\n")
		os.Exit(1)
	}

	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode image: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{Lossless: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode webp: %v\n", err)
		os.Exit(1)
	}

	assertsDir := strings.ToLower(strings.TrimSuffix(mdFilePath, filepath.Ext(mdFilePath))) + ".assets"
	err = os.MkdirAll(assertsDir, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create assets dir: %v\n", err)
		os.Exit(1)
	}

	pictureName := generatePictureFileName()
	file := filepath.Join(assertsDir, pictureName)
	err = os.WriteFile(file, buf.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save image %s: %v\n", file, err)
		os.Exit(1)
	}

	relPath := filepath.Join(filepath.Base(assertsDir), pictureName)
	fmt.Printf("![](%s)\n", relPath)
}

func generatePictureFileName() string {
	t := time.Now()
	return fmt.Sprintf("img-%d%02d%02d%02d%02d%02d.webp", t.Year(), t.Month(),
		t.Day(), t.Hour(), t.Minute(), t.Second())
}
