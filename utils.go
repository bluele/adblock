package adblock

import (
	"bufio"
	"fmt"
	"github.com/bluele/adblock/regexp"
	"os"
	re "regexp"
	"strings"
)

var (
	HasAnyPrefix = createCheckStringSetFunc(strings.HasPrefix)
	ContainsAny  = createCheckStringSetFunc(strings.Contains)

	binaryOptions = []string{
		"script",
		"image",
		"stylesheet",
		"object",
		"xmlhttprequest",
		"object-subrequest",
		"subdocument",
		"document",
		"elemhide",
		"other",
		"background",
		"xbl",
		"ping",
		"dtd",
		"media",
		"third-party",
		"match-case",
		"collapse",
		"donottrack",
		"popup",
	}
	optionsSplitPat = fmt.Sprintf(",(?=~?(?:%v))", strings.Join(append(binaryOptions, "doman"), "|"))
	optionsSplitRe  = regexp.MustCompile(optionsSplitPat)

	escapeSpecialRegxp = re.MustCompile(`([.$+?{}()\[\]\\])`)
)

func createCheckStringSetFunc(checkFunc func(string, string) bool) func(string, ...string) bool {
	return func(s string, sets ...string) bool {
		for _, set := range sets {
			if checkFunc(s, set) {
				return true
			}
		}
		return false
	}
}

func splitOptions(option string) []string {
	return strings.Split(option, ",")
}

func parseDomainOption(text string) map[string]bool {
	domains := text[len("domain="):]

	parts := strings.Split(strings.Replace(domains, ",", "|", -1), "|")
	opts := make(map[string]bool, len(parts))
	for _, part := range parts {
		opts[strings.TrimLeft(part, "~")] = !strings.HasPrefix(part, "~")
	}
	return opts
}

// Convert AdBlock rule to a regular expression.
func ruleToRegexp(text string) (string, error) {
	if text == "" {
		return ".*", nil
	}

	// already regexp?
	length := len(text)
	if length >= 2 && text[:1] == "/" && text[length-1:] == "/" {
		// filter is a regular expression
		return text[1 : length-1], nil
	}

	rule := escapeSpecialRegxp.ReplaceAllStringFunc(text, func(src string) string {
		return fmt.Sprintf(`\%v`, src)
	})
	rule = strings.Replace(rule, "^", `(?:[^\\w\\d_\\\-.%]|$)`, -1)
	rule = strings.Replace(rule, "*", ".*", -1)

	length = len(rule)
	if rule[length-1] == '|' {
		rule = rule[:length-1] + "$"
	}

	if rule[:2] == "||" {
		if len(rule) > 2 {
			rule = `^(?:[^:/?#]+:)?(?://(?:[^/?#]*\\.)?)?` + rule[2:]
		}
	} else if rule[0] == '|' {
		rule = "^" + rule[1:]
	}

	rule = re.MustCompile(`(\|)[^$]`).ReplaceAllString(rule, `\|`)

	return rule, nil
}

func DomainVariants(domain string) []string {
	variants := []string{}
	parts := strings.Split(domain, ".")
	for i := len(parts); i > 1; i-- {
		p := parts[len(parts)-i:]
		variants = append(variants, strings.Join(p, "."))
	}
	return variants
}

func reverseStrings(input []string) []string {
	if len(input) == 0 {
		return input
	}
	return append(reverseStrings(input[1:]), input[0])
}

func sliceToMap(sl []string) map[string]interface{} {
	opts := make(map[string]interface{})
	for _, v := range sl {
		opts[v] = true
	}
	return opts
}

func DefaultOptions() map[string]interface{} {
	return sliceToMap(binaryOptions)
}

func CombinedRegex(rules []*Rule) *re.Regexp {
	regexes := []string{}
	for _, rule := range rules {
		regexes = append(regexes, rule.regexString)
	}
	rs := strings.Join(regexes, "|")
	if rs == "" {
		return nil
	}
	return re.MustCompile(rs)
}

func mapKeys(m map[string]interface{}) []string {
	keys := []string{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func isSuperSet(a, b []string, reverse bool) bool {
	var (
		mr map[string]interface{}
		sr []string
	)

	if !reverse {
		mr = sliceToMap(b)
		sr = a
	} else {
		mr = sliceToMap(a)
		sr = b
	}

	for _, key := range sr {
		_, ok := mr[key]
		if !ok {
			return false
		}
	}
	return true
}

func splitRuleData(iter []*Rule, pred func(*Rule) bool) ([]*Rule, []*Rule) {
	var yes, no []*Rule
	for _, v := range iter {
		if pred(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no
}

func splitBlackWhite(rules []*Rule) ([]*Rule, []*Rule) {
	return splitRuleData(rules, func(rule *Rule) bool {
		return !rule.isException
	})
}

func splitBlackWhiteDomain(rules []*Rule) (map[string][]*Rule, map[string][]*Rule) {
	blacklist, whitelist := splitBlackWhite(rules)
	return domainIndex(blacklist), domainIndex(whitelist)
}

func domainIndex(rules []*Rule) map[string][]*Rule {
	result := make(map[string][]*Rule)
	for _, rule := range rules {
		for domain, required := range rule.domainOptions {
			if required {
				result[domain] = append(result[domain], rule)
			}
		}
	}
	return result
}

func AnyTrueValue(mp map[string]bool) bool {
	for _, it := range mp {
		if it {
			return true
		}
	}
	return false
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	lines := []string{}
	for line := []byte{}; err == nil; line, _, err = reader.ReadLine() {
		sl := strings.TrimRight(string(line), "\n\r")
		if len(sl) == 0 {
			continue
		}
		lines = append(lines, sl)
	}

	return lines, nil
}
