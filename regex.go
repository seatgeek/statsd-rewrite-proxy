package main

import "regexp"
import "strings"

import "fmt"

// Rule ...
type Rule struct {
	*regexp.Regexp
	name   string
	action string
}

// RuleResult ...
type RuleResult struct {
	Captures map[string]string
	Tags     []string
	name     string
	action   string
}

// NewMatchRule ...
func NewMatchRule(ruleString string, newPath string) *Rule {
	return &Rule{
		action: "match",
		Regexp: buildRegexp(ruleString),
		name:   newPath,
	}
}

// NewIgnoreRule ..
func NewIgnoreRule(ruleString string) *Rule {
	return &Rule{
		action: "ignore",
		Regexp: buildRegexp(ruleString),
	}
}

// NewPassRule ...
func NewPassRule(rulestring string) *Rule {
	return &Rule{
		action: "pass",
		Regexp: buildRegexp(rulestring),
	}
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
	logger.Debugf("Pattern: %s", reg)
	return regexp.MustCompile(reg)
}

// FindStringSubmatchMap add a new method to our new regular expression type
func (r *Rule) FindStringSubmatchMap(s string) *RuleResult {
	result := &RuleResult{
		action: r.action,
	}

	match := r.FindStringSubmatch(s)
	if match == nil {
		result.action = "miss"
		return result
	}

	if r.action == "skip" {
		return result
	}

	if r.action == "pass" {
		return result
	}

	result.Captures = make(map[string]string, 0)
	result.name = r.name

	for i, name := range r.SubexpNames() {
		if i == 0 {
			continue
		}

		result.Captures[name] = match[i]
		result.Tags = append(result.Tags, fmt.Sprintf("%s:%s", name, match[i]))
		result.name = strings.Replace(result.name, "{"+name+"}", match[i], -1)
	}

	return result
}
