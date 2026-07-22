package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type QuickfixItem struct {
	Filename string `json:"filename"`
	Lnum     int    `json:"lnum"`
	Text     string `json:"text"`
}

// getOrPrepareJdkSourceDirCLI locates JDK source or extracts src.zip into cache directory
func getOrPrepareJdkSourceDirCLI() (string, error) {
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

	return "", fmt.Errorf("JDK source zip or directory not found. Please set JAVA_HOME")
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
