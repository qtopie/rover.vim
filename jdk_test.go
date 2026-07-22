package main

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qtopie/sniphunt/pkg/search"
)

func TestExtractZipToCache(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "src.zip")

	// Create dummy zip file containing a Java source file
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zw := zip.NewWriter(zf)
	f, err := zw.Create("java/util/PriorityQueue.java")
	if err != nil {
		t.Fatalf("failed to add file to zip: %v", err)
	}

	javaContent := `package java.util;

public class PriorityQueue<E> extends AbstractQueue<E> implements java.io.Serializable {
    public PriorityQueue() {
        this(DEFAULT_INITIAL_CAPACITY, null);
    }
}
`
	_, _ = f.Write([]byte(javaContent))
	zw.Close()
	zf.Close()

	destDir, err := extractZipToCache(zipPath)
	if err != nil {
		t.Fatalf("extractZipToCache returned error: %v", err)
	}

	extractedFile := filepath.Join(destDir, "java", "util", "PriorityQueue.java")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file %s: %v", extractedFile, err)
	}

	if !strings.Contains(string(content), "class PriorityQueue") {
		t.Errorf("expected content to contain 'class PriorityQueue', got: %s", string(content))
	}
}

func TestSniphuntJdkSearch(t *testing.T) {
	tmpDir := t.TempDir()

	javaFile := filepath.Join(tmpDir, "java", "util", "PriorityQueue.java")
	_ = os.MkdirAll(filepath.Dir(javaFile), 0755)

	javaCode := `package java.util;

public class PriorityQueue<E> {
    private static final int DEFAULT_INITIAL_CAPACITY = 11;
    public boolean offer(E e) {
        if (e == null) throw new NullPointerException();
        return true;
    }
}
`
	if err := os.WriteFile(javaFile, []byte(javaCode), 0644); err != nil {
		t.Fatalf("failed to write java file: %v", err)
	}

	s := search.NewSearcher()
	s.Extensions = []string{".java"}

	ctx := context.Background()
	matchChan, errChan := s.Search(ctx, tmpDir, "PriorityQueue")

	var matches []search.Match
	for match := range matchChan {
		matches = append(matches, match)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("search returned error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatalf("expected matches for PriorityQueue, got none")
	}

	foundClassDef := false
	for _, m := range matches {
		if strings.Contains(string(m.Text), "public class PriorityQueue") {
			foundClassDef = true
			if m.LineNum != 3 {
				t.Errorf("expected line number 3, got %d", m.LineNum)
			}
		}
	}

	if !foundClassDef {
		t.Errorf("expected to find line defining 'public class PriorityQueue'")
	}
}

func TestSystemJdkSourceExtractionAndSearch(t *testing.T) {
	zipPath := "/usr/lib/jvm/openjdk-17/lib/src.zip"
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skip("system JDK src.zip not found, skipping integration test")
	}

	destDir, err := extractZipToCache(zipPath)
	if err != nil {
		t.Fatalf("failed to extract system src.zip: %v", err)
	}

	s := search.NewSearcher()
	s.Extensions = []string{".java"}

	ctx := context.Background()
	matchChan, errChan := s.Search(ctx, destDir, "class PriorityQueue")

	var matches []search.Match
	for m := range matchChan {
		matches = append(matches, m)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("search error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatalf("expected matches for 'class PriorityQueue' in JDK source, got none")
	}

	t.Logf("Found %d matches for 'class PriorityQueue' in system JDK source!", len(matches))
}

func TestCleanJdkCache(t *testing.T) {
	userCacheDir, _ := os.UserCacheDir()
	destDir := filepath.Join(userCacheDir, "rover", "jdk_src")
	_ = os.MkdirAll(destDir, 0755)

	dummyFile := filepath.Join(destDir, "test.txt")
	_ = os.WriteFile(dummyFile, []byte("dummy"), 0644)

	_ = os.RemoveAll(destDir)
	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		t.Errorf("expected cache dir to be removed")
	}
}
