package main

import "regexp"
import "strings"

import "fmt"

// CaputeRule embed regexp.Regexp in a new type so we can extend it
type CaputeRule struct {
	*regexp.Regexp
	NewPath string
}

type regexRules []CaputeRule

// RuleResult ...
type RuleResult struct {
	Captures map[string]string
	NewPath  string
}

// NewRule ...
func NewRule(ruleString string, newPath string) *CaputeRule {
	rule := &CaputeRule{buildRegexp(ruleString), ruleString}
	return rule
}

func buildRegexp(rule string) *regexp.Regexp {
	regRule := make([]string, 0)

	chunks := strings.Split(rule, ".")
	for _, chunk := range chunks {
		// if the chunk contains markers, make it into a pattern match
		if chunk[:1] == "{" {
			chunk = fmt.Sprintf(`(?P<%s>[^\.]+)`, chunk[1:len(chunk)-1])
		}

		// if the chunk is a star, it means the rest
		if chunk[:1] == "*" {
			chunk = `.+?`
		}

		regRule = append(regRule, chunk)
	}

	reg := strings.Join(regRule, `\.+`)
	return regexp.MustCompile(reg)
}

// FindStringSubmatchMap add a new method to our new regular expression type
func (r *CaputeRule) FindStringSubmatchMap(s string) *RuleResult {
	result := RuleResult{}
	result.Captures = make(map[string]string, 0)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return &result
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 {
			continue
		}

		result.Captures[name] = match[i]
	}

	result.NewPath = r.ReplaceAllLiteralString(s, "")

	return &result
}
