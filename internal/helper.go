package internal

import (
	"fmt"
	"regexp"
	"strings"
)

func key(req string, patterns []*regexp.Regexp) string {
	splitted := strings.Split(req, " ")
	method := splitted[0]
	uri := strings.Split(splitted[1], "?")[0]

	for _, p := range patterns {
		if p.MatchString(uri) {
			uri = p.String()
			return fmt.Sprintf("%s %s", method, uri)
		}
	}
	return fmt.Sprintf("%s %s", method, uri)
}
