package utils

import (
	"regexp"
	"strings"
)

var match1stCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	match1stCap.FindAllString(str, -1)
	str = match1stCap.ReplaceAllString(str, "${1}_${2}")
	str = matchAllCap.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(str)
}

func ToCap(str string) string {
	if str == "" {
		return ""
	}
	cap := strings.ToUpper(str[:1])
	return cap + str[1:]
}
