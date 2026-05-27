package verify

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path string
		want Language
	}{
		{"main.go", LangGo},
		{"app.ts", LangTypeScript},
		{"components/Button.tsx", LangTypeScript},
		{"utils.js", LangJavaScript},
		{"utils.jsx", LangJavaScript},
		{"script.py", LangPython},
		{"lib.rs", LangRust},
		{"README.md", LangUnknown},
		{"Dockerfile", LangUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := DetectLanguage(tt.path); got != tt.want {
				t.Errorf("DetectLanguage(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestParseFile_TypeScript(t *testing.T) {
	content := `
import { useState } from 'react';
import { Button } from '@/components/Button';
import { parse } from './utils';

export function testFunc() {
	console.log("hello");
	useState();
	fs.readFileSync("test.txt");
}

class TestClass {
	method() {}
}
`
	pf := ParseFile("test.ts", content)

	if pf.Language != LangTypeScript {
		t.Errorf("Language = %v, want %v", pf.Language, LangTypeScript)
	}

	if len(pf.Imports) != 3 {
		t.Errorf("Imports = %d, want 3", len(pf.Imports))
	}
	
	if len(pf.Functions) != 1 || pf.Functions[0] != "testFunc" {
		t.Errorf("Functions = %v, want [testFunc]", pf.Functions)
	}

	if len(pf.Classes) != 1 || pf.Classes[0] != "TestClass" {
		t.Errorf("Classes = %v, want [TestClass]", pf.Classes)
	}

	calls := []string{}
	for _, c := range pf.Calls {
		calls = append(calls, c.Name)
	}
	// "console.log", "useState", "fs.readFileSync"
	hasLog := false
	for _, c := range calls {
		if c == "console.log" {
			hasLog = true
		}
	}
	if !hasLog {
		t.Errorf("Calls missing console.log, got %v", calls)
	}
}

func TestParseFile_Python(t *testing.T) {
	content := `
import os
from datetime import datetime
import sys as system

def test_func():
	os.path.join("a", "b")
	print("hello")

class TestClass:
	pass
`
	pf := ParseFile("test.py", content)

	if pf.Language != LangPython {
		t.Errorf("Language = %v, want %v", pf.Language, LangPython)
	}

	if len(pf.Imports) != 3 {
		t.Errorf("Imports = %d, want 3", len(pf.Imports))
	}
	
	if len(pf.Functions) != 1 || pf.Functions[0] != "test_func" {
		t.Errorf("Functions = %v, want [test_func]", pf.Functions)
	}

	if len(pf.Classes) != 1 || pf.Classes[0] != "TestClass" {
		t.Errorf("Classes = %v, want [TestClass]", pf.Classes)
	}
}

func TestParseFile_Rust(t *testing.T) {
	content := `
use std::fs;
use crate::utils;

pub fn test_func() {
	println!("hello");
	fs::read_to_string("test.txt").unwrap();
}

struct TestStruct {
}

impl TestStruct {
}
`
	pf := ParseFile("test.rs", content)

	if pf.Language != LangRust {
		t.Errorf("Language = %v, want %v", pf.Language, LangRust)
	}

	if len(pf.Imports) != 2 {
		t.Errorf("Imports = %d, want 2", len(pf.Imports))
	}
	
	if len(pf.Functions) != 1 || pf.Functions[0] != "test_func" {
		t.Errorf("Functions = %v, want [test_func]", pf.Functions)
	}

	if len(pf.Classes) != 2 {
		t.Errorf("Classes = %v, want 2 elements", pf.Classes)
	}
}
