package regexp

import (
	"github.com/bluele/adblock/regexp/pcre"
)

func Compile(pattern string) (*Regexp, error) {
	pre, err := pcre.Compile(pattern, pcre.CASELESS)
	if err != nil {
		return nil, err
	}
	re := &Regexp{}
	re.pre = &pre
	return re, nil
}

func MustCompile(pattern string) *Regexp {
	re, err := Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

type Regexp struct {
	pre *pcre.Regexp
}

func (re *Regexp) Match(target []byte) bool {
	return re.pre.Matcher(target, 0).Matches()
}

func (re *Regexp) MatchString(target string) bool {
	return re.Match([]byte(target))
}

func (re *Regexp) ReplaceAll(pattern, repl []byte) string {
	return string(re.pre.ReplaceAll(pattern, repl, 0))
}
