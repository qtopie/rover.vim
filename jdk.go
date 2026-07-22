package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/govim/govim"
	"github.com/qtopie/sniphunt/pkg/search"
)

type QuickfixItem struct {
	Filename string `json:"filename"`
	Lnum     int    `json:"lnum"`
	Text     string `json:"text"`
}

// searchJdkSource handles the :JdkSearch and :RoverJdkSearch commands in Vim.
func searchJdkSource(g govim.Govim, flags govim.CommandFlags, args ...string) error {
	if len(args) == 0 {
		return showErrMsg(g, "Usage: :JdkSearch <query>")
	}

	query := strings.Join(args, " ")

	// 1. Locate JDK source directory
	srcDir, err := getOrPrepareJdkSourceDir(g)
	if err != nil {
		return showErrMsg(g, "JDK source resolution error: %v", err)
	}

	// 2. Perform search using sniphunt
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
		return showErrMsg(g, "Sniphunt search error: %v", err)
	}

	if len(qfList) == 0 {
		return showMsg(g, "No JDK source matches found for: %s", query)
	}

	// 3. Convert quickfix items to JSON and send to Vim setqflist() & copen
	qfJSON, err := json.Marshal(qfList)
	if err != nil {
		return showErrMsg(g, "Failed to encode quickfix list: %v", err)
	}

	cmd := fmt.Sprintf("call setqflist(%s, 'r') | copen", string(qfJSON))
	err = g.ChannelEx(cmd)
	if err != nil {
		return showErrMsg(g, "Failed to open quickfix window: %v", err)
	}

	return showMsg(g, "Found %d JDK source matches for '%s'", len(qfList), query)
}

// getOrPrepareJdkSourceDir locates JDK source or extracts src.zip into cache directory
func getOrPrepareJdkSourceDir(g govim.Govim) (string, error) {
	// First check if user defined g:rover_jdk_source_path or g:jdk_source_path in Vim
	var customPath string
	rawPath, err := g.ChannelExpr("get(g:, 'rover_jdk_source_path', get(g:, 'jdk_source_path', ''))")
	if err == nil && len(rawPath) > 0 {
		var s string
		if json.Unmarshal(rawPath, &s) == nil && s != "" {
			customPath = s
		}
	}

	if customPath != "" {
		st, err := os.Stat(customPath)
		if err == nil {
			if st.IsDir() {
				return customPath, nil
			}
			if strings.HasSuffix(customPath, ".zip") {
				return extractZipToCache(customPath)
			}
		}
	}

	// Next, search standard JDK locations & JAVA_HOME
	candidateZips := []string{}
	candidateDirs := []string{}

	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		candidateZips = append(candidateZips,
			filepath.Join(javaHome, "lib", "src.zip"),
			filepath.Join(javaHome, "src.zip"),
		)
		candidateDirs = append(candidateDirs,
			filepath.Join(javaHome, "src"),
		)
	}

	// Common system JDK locations
	if runtime.GOOS == "linux" {
		matches, _ := filepath.Glob("/usr/lib/jvm/*/lib/src.zip")
		candidateZips = append(candidateZips, matches...)
		matches2, _ := filepath.Glob("/usr/lib/jvm/*/src.zip")
		candidateZips = append(candidateZips, matches2...)
	} else if runtime.GOOS == "darwin" {
		matches, _ := filepath.Glob("/Library/Java/JavaVirtualMachines/*/Contents/Home/lib/src.zip")
		candidateZips = append(candidateZips, matches...)
	}

	// Check existing uncompressed directory
	for _, d := range candidateDirs {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d, nil
		}
	}

	// Check candidate zip files
	for _, z := range candidateZips {
		if st, err := os.Stat(z); err == nil && !st.IsDir() {
			return extractZipToCache(z)
		}
	}

	// Check if cached JDK source directory already exists
	userCacheDir, cacheErr := os.UserCacheDir()
	if cacheErr == nil {
		extractedDir := filepath.Join(userCacheDir, "rover", "jdk_src")
		if st, err := os.Stat(extractedDir); err == nil && st.IsDir() {
			return extractedDir, nil
		}
	}

	return "", fmt.Errorf("JDK source zip or directory not found. Please set g:rover_jdk_source_path or JAVA_HOME")
}

// extractZipToCache extracts a JDK src.zip file into ~/.cache/rover/jdk_src
func extractZipToCache(zipPath string) (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		userCacheDir = os.TempDir()
	}

	destDir := filepath.Join(userCacheDir, "rover", "jdk_src")
	markerFile := filepath.Join(destDir, ".extracted_ok")

	zipStat, err := os.Stat(zipPath)
	if err != nil {
		return "", err
	}

	// If marker exists and is newer than zip file, skip extraction
	if markerStat, err := os.Stat(markerFile); err == nil {
		if markerStat.ModTime().After(zipStat.ModTime()) {
			return destDir, nil
		}
	}

	_ = os.RemoveAll(destDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir: %w", err)
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip %s: %w", zipPath, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		targetPath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), destDir) {
			continue // prevent ZipSlip vulnerability
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return "", err
		}
	}

	// Touch marker file
	os.WriteFile(markerFile, []byte("OK"), 0644)
	return destDir, nil
}

// removeJdkCacheDir removes extracted JDK source cache directory from disk
func removeJdkCacheDir() error {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		userCacheDir = os.TempDir()
	}
	destDir := filepath.Join(userCacheDir, "rover", "jdk_src")
	return os.RemoveAll(destDir)
}

// cleanJdkCache removes extracted JDK source cache directory and notifies Vim
func cleanJdkCache(g govim.Govim, flags govim.CommandFlags, args ...string) error {
	if err := removeJdkCacheDir(); err != nil {
		return showErrMsg(g, "Failed to clean JDK source cache: %v", err)
	}
	return showMsg(g, "JDK source cache cleaned successfully.")
}

// checkAndAutoCleanJdkCache cleans up cache if no JDK source buffers remain open in Vim
func checkAndAutoCleanJdkCache(g govim.Govim) error {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		userCacheDir = os.TempDir()
	}
	destDir := filepath.Join(userCacheDir, "rover", "jdk_src")

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return nil
	}

	rawBufs, err := g.ChannelExpr("getbufinfo({'bufloaded': 1})")
	if err != nil || len(rawBufs) == 0 {
		return nil
	}

	var bufInfos []struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(rawBufs, &bufInfos) == nil {
		for _, b := range bufInfos {
			if strings.HasPrefix(b.Name, destDir) {
				return nil // At least one JDK source buffer is still open/loaded
			}
		}
		// No JDK source buffers are open anymore, auto-clean cache!
		_ = os.RemoveAll(destDir)
	}
	return nil
}


