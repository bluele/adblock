package adblock_test

import (
	"github.com/bluele/adblock"
	"testing"
)

func TestDomainVariants(t *testing.T) {
	variants := adblock.DomainVariants("foo.bar.example.com")
	if len(variants) != 3 {
		t.Errorf("variants should be %v.", len(variants))
	}
	if variants[0] != "foo.bar.example.com" {
		t.Errorf("%v should be %v.", variants[0], "foo.bar.example.com")
	}
	if variants[1] != "bar.example.com" {
		t.Errorf("%v should be %v.", variants[1], "bar.example.com")
	}
	if variants[2] != "example.com" {
		t.Errorf("%v should be %v.", variants[2], "example.com")
	}
}
