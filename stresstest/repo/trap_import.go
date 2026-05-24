package main

import "stresstest/pkg/nonexistent"

func useImport() {
	_ = nonexistent.Foo
}
