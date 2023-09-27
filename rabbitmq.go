package rabbitmq

import "strings"

const STAR = "*"
const HASH = "#"
const SEPR = "."

func MatchKey(required, check string) bool {
	return matchParts(
		strings.Split(required, SEPR),
		strings.Split(check, SEPR),
	)
}

func matchParts(needle, stack []string) bool {
	if len(needle) == 0 {
		return (len(stack) == 0)
	}
	if len(stack) == 0 {
		return false
	}

	switch needle[0] {
	case stack[0]:
		return matchParts(needle[1:], stack[1:])
	case STAR:
		return matchParts(needle[1:], stack[1:])
	case HASH:
		return matchPartsHash(needle[1:], stack[1:])
	default:
		return false
	}
}

func matchPartsHash(needle, stack []string) bool {
	for i := len(stack); i >= 0; i-- {
		if matchParts(needle, stack[i:]) {
			return true
		}
	}
	return false
}
