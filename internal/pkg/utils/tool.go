package utils

import (
	"regexp"
	"strings"
)

// 将title转换为slug
func Slugify(title string) string {
	re, _ := regexp.Compile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(title, "-"))
}
