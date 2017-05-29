package main

import "regexp"
import "strings"

import "fmt"

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
