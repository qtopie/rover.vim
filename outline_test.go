package main

import (
	"testing"
)

func TestParseGoOutline(t *testing.T) {
	goCode := `package sample

type User struct {
	ID   int
	Name string
}

type Service interface {
	GetUser(id int) (*User, error)
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func (u *User) GetName() string {
	return u.Name
}
`

	symbols, err := parseGoOutline("sample.go", []byte(goCode))
	if err != nil {
		t.Fatalf("parseGoOutline error: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatalf("expected symbols, got none")
	}

	foundStruct := false
	foundInterface := false
	foundFunc := false
	foundMethod := false

	for _, s := range symbols {
		if s.Kind == KindStruct && s.Name == "User" {
			foundStruct = true
		}
		if s.Kind == KindInterface && s.Name == "Service" {
			foundInterface = true
		}
		if s.Kind == KindFunction && s.Name == "NewUser" {
			foundFunc = true
		}
		if s.Kind == KindMethod && s.Name == "GetName" {
			foundMethod = true
		}
	}

	if !foundStruct {
		t.Errorf("expected to find struct User")
	}
	if !foundInterface {
		t.Errorf("expected to find interface Service")
	}
	if !foundFunc {
		t.Errorf("expected to find function NewUser")
	}
	if !foundMethod {
		t.Errorf("expected to find method GetName")
	}
}

func TestParseJavaOutline(t *testing.T) {
	javaCode := `package com.example;

public class PriorityQueue<E> {
    private int size = 0;
    
    public PriorityQueue() {
    }

    public boolean offer(E e) {
        return true;
    }
}
`

	symbols, err := parseJavaOutline("PriorityQueue.java", []byte(javaCode))
	if err != nil {
		t.Fatalf("parseJavaOutline error: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatalf("expected symbols, got none")
	}

	foundClass := false
	foundField := false
	foundMethod := false

	for _, s := range symbols {
		if s.Kind == KindClass && s.Name == "PriorityQueue" {
			foundClass = true
		}
		if s.Kind == KindField && s.Name == "size" {
			foundField = true
		}
		if s.Kind == KindMethod && s.Name == "offer" {
			foundMethod = true
		}
	}

	if !foundClass {
		t.Errorf("expected to find class PriorityQueue")
	}
	if !foundField {
		t.Errorf("expected to find field size")
	}
	if !foundMethod {
		t.Errorf("expected to find method offer")
	}
}
