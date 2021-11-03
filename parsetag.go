package reflectutils

import (
	"reflect"
	"regexp"
)

var aTagRE = regexp.MustCompile(`(\S+):"((?:[^"\\]|\\.)*)"(?:\s+|$)`)

type Tag struct {
	Tag   string
	Value string
}

func SplitTag(tag reflect.StructTag) []Tag {
	found := make([]Tag, 0, 5)
	s := string(tag)
	for len(s) > 0 {
		f := aTagRE.FindStringSubmatchIndex(s)
		if len(f) != 6 {
			break
		}
		tag := s[f[2]:f[3]]
		value := s[f[4]:f[5]]
		found = append(found, Tag{
			Tag:   tag,
			Value: value,
		})
		s = s[f[1]:]
	}
	if len(found) == 0 {
		return nil
	}
	return found
}
