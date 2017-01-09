package main

import "regexp"
import "strings"

import "fmt"

// CaputeRule embed regexp.Regexp in a new type so we can extend it
type CaputeRule struct {
	*regexp.Regexp
	name string
}

// RuleResult ...
type RuleResult struct {
	Captures map[string]string
	Tags     []string
	name     string
}

// NewRule ...
func NewRule(ruleString string, newPath string) *CaputeRule {
	rule := &CaputeRule{}
	rule.Regexp = buildRegexp(ruleString)
	rule.name = newPath
	return rule
}

func buildRegexp(rule string) *regexp.Regexp {
	regRule := make([]string, 0)

	chunks := strings.Split(rule, ".")
	for _, chunk := range chunks {
		// if the chunk contains markers, make it into a pattern match
		if chunk[:1] == "{" {
			chunk = fmt.Sprintf(`(?P<%s>[^\.]+)`, chunk[1:len(chunk)-1])
		} else if chunk[:1] == "*" { // Stars will just glob anything
			chunk = `.+?`
		} else { // litterals will be, well, litterals and just escape for safe regexp processing
			chunk = regexp.QuoteMeta(chunk)
		}

		regRule = append(regRule, chunk)
	}

	reg := strings.Join(regRule, `\.+`)
	logger.Infof(reg)
	return regexp.MustCompile(reg)
}

// FindStringSubmatchMap add a new method to our new regular expression type
func (r *CaputeRule) FindStringSubmatchMap(s string) *RuleResult {
	result := RuleResult{}
	result.Captures = make(map[string]string, 0)
	result.name = r.name

	match := r.FindStringSubmatch(s)
	if match == nil {
		return &result
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 {
			continue
		}

		// logger.Infof("name: %s", result.name)
		result.Captures[name] = match[i]
		result.Tags = append(result.Tags, fmt.Sprintf("%s:%s", name, match[i]))
		result.name = strings.Replace(result.name, "{"+name+"}", match[i], -1)
	}
	// logger.Infof("name: %s", result.name)

	return &result
}
