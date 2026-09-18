package config

import (
	"fmt"
	"strings"

	"github.com/agext/levenshtein"
)

// lowercaseKeys returns a copy of a resource specification with every mapping key
// lowercased, mirroring how mapstructure matches a specification field regardless of its
// case. On a collision the first key wins, as validating twice the same field would only
// duplicate problems.
func lowercaseKeys(value interface{}) interface{} {

	switch typed := value.(type) {

	case map[string]interface{}:
		lowercased := make(map[string]interface{}, len(typed))
		for _, key := range sortedKeys(typed) {
			normalized := strings.ToLower(key)
			if _, ok := lowercased[normalized]; ok {
				continue
			}
			lowercased[normalized] = lowercaseKeys(typed[key])
		}
		return lowercased

	case []interface{}:
		lowercased := make([]interface{}, len(typed))
		for i, item := range typed {
			lowercased[i] = lowercaseKeys(item)
		}
		return lowercased

	default:
		return value
	}
}

// suggest returns a hint naming the closest candidate to an unknown value, or an empty
// string when no candidate is close enough. A wrong suggestion is worse than none, so a
// candidate is only proposed when it is both close and unambiguous.
func suggest(value string, candidates []string) string {

	const maxDistance int = 2

	best := ""
	bestDistance := maxDistance + 1
	ambiguous := false

	for _, candidate := range candidates {
		distance := levenshtein.Distance(value, candidate, nil)

		switch {
		case distance < bestDistance:
			best, bestDistance, ambiguous = candidate, distance, false
		case distance == bestDistance:
			ambiguous = true
		}
	}

	if best == "" || ambiguous || bestDistance > maxDistance || bestDistance*3 > len(value) {
		return ""
	}

	return fmt.Sprintf(", did you mean %q?", best)
}
