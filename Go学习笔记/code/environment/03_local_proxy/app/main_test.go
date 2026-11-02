package main

import (
	"example.com/learning/left"
	"example.com/learning/right"
	"testing"
)

func TestSelectedModuleIsShared(t *testing.T) {
	if left.Value() != 13 || right.Value() != 13 {
		t.Fatalf("MVS did not select expected shared module")
	}
}
