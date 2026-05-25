package blindfold

import (
	"strings"
	"testing"
)

func TestObfuscateDeobfuscate(t *testing.T) {
	ob := NewObfuscator("test-salt")
	input := "package main\n\nfunc hello() string {\n\treturn \"world\"\n}\n"
	obs := ob.Obfuscate(input)
	if obs == input {
		t.Fatal("obfuscation did nothing")
	}
	if strings.Contains(obs, "hello") {
		t.Fatal("original identifier 'hello' still visible after obfuscation")
	}
	deobs := ob.Deobfuscate(obs)
	if deobs != input {
		t.Fatalf("deobfuscation mismatch:\ninput:  %q\nobs:    %q\ndeobs:  %q", input, obs, deobs)
	}
}

func TestMultipleCalls(t *testing.T) {
	ob := NewObfuscator("salt")
	a := ob.Obfuscate("func foo() {}")
	b := ob.Obfuscate("func bar() {}")
	deobs := ob.Deobfuscate(b + a)
	if !strings.Contains(deobs, "foo") || !strings.Contains(deobs, "bar") {
		t.Fatal("deobfuscation lost identifiers after multiple calls")
	}
}

func TestKeywordPreserved(t *testing.T) {
	ob := NewObfuscator("test-kw")
	input := "func main() { return nil }"
	obs := ob.Obfuscate(input)
	if !strings.Contains(obs, "func") || !strings.Contains(obs, "return") || !strings.Contains(obs, "nil") {
		t.Fatal("keywords should be preserved in obfuscated output")
	}
	deobs := ob.Deobfuscate(obs)
	if deobs != input {
		t.Fatalf("deobfuscation failed: %s", deobs)
	}
}

func TestDeterministic(t *testing.T) {
	ob1 := NewObfuscator("same-salt")
	ob2 := NewObfuscator("same-salt")
	input := "func processData(input string) string { return input }"
	out1 := ob1.Obfuscate(input)
	out2 := ob2.Obfuscate(input)
	if out1 != out2 {
		t.Fatalf("expected deterministic obfuscation:\n%s\n%s", out1, out2)
	}
}

func TestMappings(t *testing.T) {
	ob := NewObfuscator("salt-map")
	ob.Obfuscate("func alpha() {}")
	ob.Obfuscate("func beta() {}")
	m := ob.Mappings()
	if len(m) < 2 {
		t.Fatalf("expected at least 2 mappings, got %d", len(m))
	}
	found := false
	for _, mapping := range m {
		if mapping.Original == "alpha" || mapping.Original == "beta" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected alpha or beta in mappings")
	}
}
