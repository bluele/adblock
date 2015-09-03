package adblock

import (
	"github.com/bluele/adblock/regexp"
	re "regexp"
	"strings"
)

type Rule struct {
	raw         string
	text        string
	regexString string
	regex       *regexp.Regexp
	isComment   bool
	isHTMLRule  bool
	isException bool

	options       map[string]bool
	domainOptions map[string]bool
	rawOptions    []string

	optionsKeys []string
}

func (rule *Rule) OptionsKeys() []string {
	opts := []string{}
	for opt, _ := range rule.options {
		if opt != "match-case" {
			opts = append(opts, opt)
		}
	}
	if rule.domainOptions != nil && len(rule.domainOptions) >= 0 {
		opts = append(opts, "domain")
	}
	return opts
}

func (rule *Rule) HasOption(key string) bool {
	if key == "domain" {
		return rule.domainOptions != nil && len(rule.domainOptions) >= 0
	}
	_, ok := rule.options[key]
	return ok
}

func (rule *Rule) DomainOptions() map[string]bool {
	return rule.domainOptions
}

func NewRule(text string) (*Rule, error) {
	rule := &Rule{}
	rule.raw = text
	text = strings.TrimSpace(text)
	rule.isComment = HasAnyPrefix(text, "!", "[Adblock")
	if rule.isComment {
		rule.isHTMLRule = false
		rule.isException = false
	} else {
		rule.isHTMLRule = ContainsAny(text, "##", "#@#")
		rule.isException = strings.HasPrefix(text, "@@")
		if rule.isException {
			text = text[2:]
		}
	}

	rule.options = make(map[string]bool)
	if !rule.isComment && strings.Contains(text, "$") {
		var option string
		parts := strings.SplitN(text, "$", 2)
		length := len(parts)
		if length > 0 {
			text = parts[0]
		}
		if length > 1 {
			option = parts[1]
		}

		rule.rawOptions = splitOptions(option)
		for _, opt := range rule.rawOptions {
			if strings.HasPrefix(opt, "domain=") {
				rule.domainOptions = parseDomainOption(opt)
			} else {
				rule.options[strings.TrimLeft(opt, "~")] = !strings.HasPrefix(opt, "~")
			}
		}
	} else {
		rule.rawOptions = []string{}
		rule.domainOptions = make(map[string]bool)
	}

	rule.optionsKeys = rule.OptionsKeys()
	rule.text = text

	if rule.isComment || rule.isHTMLRule {
		rule.regexString = ""
	} else {
		var err error
		rule.regexString, err = ruleToRegexp(text)
		if err != nil {
			return nil, err
		}
	}

	return rule, nil
}

func (rule *Rule) MatchingSupported(options map[string]interface{}) bool {
	if rule.isComment {
		return false
	}
	if rule.isHTMLRule {
		return false
	}
	if options == nil {
		options = map[string]interface{}{}
	}
	keys := mapKeys(options)
	if !isSuperSet(rule.OptionsKeys(), keys) {
		return false
	}

	return true
}

// Returns if this rule matches the URL.
func (rule *Rule) MatchURL(u string, options map[string]interface{}) bool {
	// TODO ignroe options
	for opt, _ := range rule.options {
		if opt == "match-case" {
			continue
		}
		if _, ok := options[opt]; !ok {
			// TODO 見直し
			// panic("Rule requires option " + opt)
			return false
		}

		// check if this rule has an option.
		v, ok := options[opt]
		if ok {
			bl, ok := v.(bool)
			if ok {
				rv, ok := rule.options[opt]
				if ok {
					if bl != rv {
						return false
					}
				}
			}
		}
	}

	if len(rule.DomainOptions()) > 0 {
		if _, ok := options["domain"]; !ok {
			panic("Rule requires option domain")
		}
	}

	// check domain exists
	v, ok := options["domain"]
	if ok {
		sv := v.(string)
		if !rule.domainMatches(sv) {
			return false
		}
	}

	return rule.urlMatches(u)
}

func (rule *Rule) urlMatches(u string) bool {
	if rule.regex == nil {
		rule.regex = regexp.MustCompile(rule.regexString)
	}
	return rule.regex.MatchString(u)
}

func (rule *Rule) domainMatches(domain string) bool {
	for _, dm := range DomainVariants(domain) {
		if bl, ok := rule.domainOptions[dm]; ok {
			return bl
		}
	}
	for _, bl := range rule.domainOptions {
		if bl {
			return false
		}
	}
	return true
}

func (rule *Rule) Text() string {
	return rule.text
}

func (rule *Rule) Raw() string {
	return rule.raw
}

func (rule *Rule) RegexString() string {
	return rule.regexString
}

func (rule *Rule) Regex() *regexp.Regexp {
	return rule.regex
}

func (rule *Rule) IsComment() bool {
	return rule.isComment
}

func (rule *Rule) IsException() bool {
	return rule.isException
}

func (rule *Rule) IsHTMLRule() bool {
	return rule.isHTMLRule
}

type Rules struct {
	rules     []*Rule
	opt       *RulesOption
	blacklist []*Rule
	whitelist []*Rule

	blacklistRe *re.Regexp
	whitelistRe *re.Regexp

	blacklistWithOptions []*Rule
	whitelistWithOptions []*Rule

	blacklistRequireDomain map[string][]*Rule
	whitelistRequireDomain map[string][]*Rule
}

type RulesOption struct {
	Supports              []string
	CheckUnsupportedRules bool
}

func NewRules(ruleStrs []string, opt *RulesOption) (*Rules, error) {
	rls := &Rules{}
	if opt == nil {
		rls.opt = &RulesOption{}
	} else {
		rls.opt = opt
	}

	if rls.opt.Supports == nil {
		rls.opt.Supports = append(binaryOptions, "domain")
	}

	params := sliceToMap(rls.opt.Supports)
	for _, ruleStr := range ruleStrs {
		rule, err := NewRule(ruleStr)
		if err != nil {
			return nil, err
		}
		if rule.regexString != "" && rule.MatchingSupported(params) {
			rls.rules = append(rls.rules, rule)
		}
	}

	advancedRules, basicRules := splitRuleData(rls.rules, func(rule *Rule) bool {
		if (rule.options != nil && len(rule.options) > 0) || (rule.domainOptions != nil && len(rule.domainOptions) > 0) {
			return true
		}
		return false
	})

	domainRequiredRules, NonDomainRules := splitRuleData(advancedRules, func(rule *Rule) bool {
		return rule.HasOption("domain") && AnyTrueValue(rule.domainOptions)
	})

	rls.blacklist, rls.whitelist = splitBlackWhite(basicRules)

	rls.blacklistRe = CombinedRegex(rls.blacklist)
	rls.whitelistRe = CombinedRegex(rls.whitelist)

	rls.blacklistWithOptions, rls.whitelistWithOptions = splitBlackWhite(NonDomainRules)
	rls.blacklistRequireDomain, rls.whitelistRequireDomain = splitBlackWhiteDomain(domainRequiredRules)

	return rls, nil
}

func NewRulesFromFile(path string, opt *RulesOption) (*Rules, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	return NewRules(lines, opt)
}

func (rules *Rules) ShouldBlock(u string, options map[string]interface{}) bool {
	if rules.IsWhiteListed(u, options) {
		return false
	}
	if rules.IsBlackListed(u, options) {
		return true
	}
	return false
}

func (rules *Rules) IsWhiteListed(u string, options map[string]interface{}) bool {
	return rules.matches(u, options, rules.whitelistRe, rules.whitelistRequireDomain, rules.whitelistWithOptions)
}

func (rules *Rules) IsBlackListed(u string, options map[string]interface{}) bool {
	return rules.matches(u, options, rules.blacklistRe, rules.blacklistRequireDomain, rules.blacklistWithOptions)
}

func (rules *Rules) matches(u string, options map[string]interface{}, generalRe *re.Regexp, domainRequiredRules map[string][]*Rule, rulesWithOptions []*Rule) bool {
	if generalRe != nil && generalRe.MatchString(u) {
		return true
	}

	rls := []*Rule{}
	isrcDomain, ok := options["domain"]
	srcDomain, ok2 := isrcDomain.(string)
	if ok && ok2 && len(domainRequiredRules) > 0 {
		for _, domain := range DomainVariants(srcDomain) {
			if vs, ok := domainRequiredRules[domain]; ok {
				rls = append(rls, vs...)
			}
		}
	}

	rls = append(rls, rulesWithOptions...)

	if !rules.opt.CheckUnsupportedRules {
		for _, rule := range rls {
			if rule.MatchingSupported(options) {
				if rule.MatchURL(u, options) {
					return true
				}
			}
		}
	}

	return false
}

func (rules *Rules) BlackList() []*Rule {
	return rules.blacklist
}

func (rules *Rules) WhiteList() []*Rule {
	return rules.whitelist
}
