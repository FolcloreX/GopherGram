package scanner

import (
	"regexp"
	"strconv"
	"strings"
)

var chunker = regexp.MustCompile(`(\d+)|\D+`)

// Compare two strings using Natural Sort. Return true if str1 < str2.
func naturalLess(str1, str2 string) bool {
	chunks1 := chunker.FindAllString(str1, -1)
	chunks2 := chunker.FindAllString(str2, -1)

	for i := 0; i < len(chunks1) && i < len(chunks2); i++ {
		p1, p2 := chunks1[i], chunks2[i]

		n1, err1 := strconv.Atoi(p1)
		n2, err2 := strconv.Atoi(p2)

		if err1 == nil && err2 == nil {
			if n1 != n2 {
				return n1 < n2
			}
			continue
		}

		if strings.ToLower(p1) != strings.ToLower(p2) {
			return strings.ToLower(p1) < strings.ToLower(p2)
		}
	}

	return len(chunks1) < len(chunks2)
}
