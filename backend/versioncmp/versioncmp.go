package versioncmp

import (
	"strconv"
	"strings"
	"unicode"
)

type segment struct {
	number int
	suffix string
}

// Compare compares two version strings.
// Returns:
// -1 if a < b
//
//	0 if a == b
//	1 if a > b
func Compare(a, b string) int {
	left := parse(a)
	right := parse(b)

	maxLen := len(left)
	if len(right) > maxLen {
		maxLen = len(right)
	}

	for i := 0; i < maxLen; i++ {
		ls := segment{}
		rs := segment{}

		if i < len(left) {
			ls = left[i]
		}
		if i < len(right) {
			rs = right[i]
		}

		if ls.number < rs.number {
			return -1
		}
		if ls.number > rs.number {
			return 1
		}

		if ls.suffix == rs.suffix {
			continue
		}

		// Stable release is considered newer than suffixed prerelease.
		if ls.suffix == "" {
			return 1
		}
		if rs.suffix == "" {
			return -1
		}

		if ls.suffix < rs.suffix {
			return -1
		}
		return 1
	}

	return 0
}

func parse(raw string) []segment {
	clean := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(raw, "v"), "V"))
	if clean == "" {
		return nil
	}

	parts := strings.Split(clean, ".")
	out := make([]segment, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			out = append(out, segment{})
			continue
		}

		i := 0
		for i < len(part) && unicode.IsDigit(rune(part[i])) {
			i++
		}

		seg := segment{}
		if i > 0 {
			n, err := strconv.Atoi(part[:i])
			if err == nil {
				seg.number = n
			}
		}

		if i < len(part) {
			seg.suffix = strings.ToLower(strings.TrimSpace(part[i:]))
		}

		out = append(out, seg)
	}

	return out
}
