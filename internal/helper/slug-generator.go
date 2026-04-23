package helper

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

func GenerateSlug(title string) string {
	var (
		nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s-]`)
		multipleSpaces  = regexp.MustCompile(`[\s-]+`)
	)

	s := strings.ToLower(title)
	s = nonAlphanumeric.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = multipleSpaces.ReplaceAllString(s, "-")
	return s
}

func GenerateUniqueSlug(title string) string {
	baseSlug := GenerateSlug(title)
	if baseSlug == "" {
		baseSlug = "slug"
	}
	base := make([]byte, 8)
	rand.Read(base)
	return baseSlug + "-" + hex.EncodeToString(base)
}
