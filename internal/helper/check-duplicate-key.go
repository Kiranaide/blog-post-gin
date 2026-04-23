package helper

import "strings"

func IsDuplicateKeyError(err error, column string) bool {
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "duplicate") || strings.Contains(s, "unique") {
		if column == "" {
			return true
		}
		return strings.Contains(s, column)
	}
	return false
}
