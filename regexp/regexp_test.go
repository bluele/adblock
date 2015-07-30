package regexp_test

import (
	"github.com/bluele/adblock/regexp"
	"testing"
)

func TestMatch(t *testing.T) {
	re := regexp.MustCompile(`test`)
	if !re.Match([]byte(`test string`)) {
		t.Error(`not match`)
	}
	if re.ReplaceAll([]byte(`test_string`), []byte(`changed`)) != "changed_string" {
		t.Error(`not match`)
	}
}
