package utils

import (
	"errors"

	"github.com/tailscale/hujson"
)

func GetObjectMemberName(input hujson.ObjectMember) (string, error) {
	name, ok := input.Name.Value.(hujson.Literal)
	if !ok {
		return "", errors.New("name value was not a literal")
	}
	return name.String(), nil
}

func GetPatternFromHujson(
	input hujson.ValueTrimmed,
	matchesPattern func(s string) bool,
) []string {
	switch value := input.(type) {
	case hujson.Literal:
		stringValue := value.String()
		if !matchesPattern(stringValue) {
			return []string{}
		}

		return []string{stringValue}
	case *hujson.Object:
		var matches []string
		for _, member := range value.Members {
			matches = append(matches, GetPatternFromHujson(member.Value.Value, matchesPattern)...)
			matches = append(matches, GetPatternFromHujson(member.Name.Value, matchesPattern)...)
		}
		return matches
	case *hujson.Array:
		var matches []string
		for _, member := range value.Elements {
			matches = append(matches, GetPatternFromHujson(member.Value, matchesPattern)...)
		}
		return matches
	default:
		return []string{}
	}
}
