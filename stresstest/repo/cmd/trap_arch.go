package main

import "stresstest/pkg/secret"

func leakInternal() {
	secret.Do()
}
