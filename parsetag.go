package reflectutils

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var aTagRE = regexp.MustCompile(`(\S+):"((?:[^"\\]|\\.)*)"(?:\s+|$)`)

// Tag represents a single element of a struct tag.  For example for the
// field S in the struct below, there would be to Tags: one for json and
// one for xml.  The value for the json one would be "s,omitempty".
//
//	type Foo struct {
//		S string `json:"s,omitempty" xml:"s_thing"`
//	}
//
type Tag struct {
	Tag   string
	Value string
}

type Tags []Tag

// TagSet is a simple transformation of an array of Tag into an
// indexted structure so that lookup is efficient.
type TagSet struct {
	tags  []Tag
	index map[string]int
}

// SplitTag breaks apart a reflect.StructTag into an array of annotated key/value pairs.
// Tags are expected to be in the conventional format.  What does "contentional"
// mean?  `name:"values,value=value" name2:"value"`.  See https://flaviocopes.com/go-tags/
// a light introduction.
func SplitTag(tag reflect.StructTag) Tags {
	found := make([]Tag, 0, 5)
	s := string(tag)
	for len(s) > 0 {
		f := aTagRE.FindStringSubmatchIndex(s)
		if len(f) != 6 {
			break
		}
		tag := s[f[2]:f[3]]
		value := s[f[4]:f[5]]
		found = append(found, mkTag(tag, value))
		s = s[f[1]:]
	}
	if len(found) == 0 {
		return nil
	}
	return found
}

func mkTag(tag, value string) Tag {
	return Tag{
		Tag:   tag,
		Value: value,
	}
}

func (s TagSet) Get(tag string) Tag {
	t, _ := s.Lookup(tag)
	return t
}

func (s TagSet) Lookup(tag string) (Tag, bool) {
	if i, ok := s.index[tag]; ok {
		return s.tags[i], true
	}
	return mkTag("", ""), false
}

func (t Tags) Set() TagSet {
	index := make(map[string]int)
	for i, tag := range t {
		index[tag.Tag] = i
	}
	return TagSet{
		tags:  t,
		index: index,
	}
}

// Fill unpacks struct tags into a struct based on tags of the desitnation struct.
//
// 	type MyTags struct {
//		Name	string	`pt:"0"`
//		Flag	bool	`pt:"flag"`
//		Int	int	`pt:"intValue"`
//	}
//
// The above will fill the Name field by grabbing the first element of the tag.
// It will fill Flag by noticing if any of the following are present in the
// comma-separated list of tag elements: "flag", "!flag" (sets to false), "flag=true",
// "flag=false", "flag=0", "flag=1", "flag=t", "flag=f", "flag=T", "flag=F".
// It will set Int by looking for "intValue=280" (set to 280).
func (tag Tag) Fill(model interface{}, opts ...FillOptArg) error {
	opt := fillOpt{
		tag: "pt",
	}
	for _, f := range opts {
		f(&opt)
	}
	v := reflect.ValueOf(model)
	if !v.IsValid() || v.IsNil() || v.Type().Kind() != reflect.Ptr || v.Type().Elem().Kind() != reflect.Struct {
		return errors.Errorf("Fill target must be a pointer to a struct, not %T", model)
	}
	kv := make(map[string]string)
	elements := strings.Split(tag.Value, ",")
	for _, element := range elements {
		if eq := strings.IndexByte(element, '='); eq != -1 {
			kv[element[0:eq]] = element[eq+1:]
		} else {
			if strings.HasPrefix(element, "!") {
				kv[element[1:]] = "f"
			} else {
				kv[element] = "t"
			}
		}
	}
	var count int
	var walkErr error
	WalkStructElements(v.Type(), func(f reflect.StructField) bool {
		tag := f.Tag.Get(opt.tag)
		name := f.Name
		if tag == "-" {
			return false
		}
		count++
		i, err := strconv.Atoi(tag)
		var value string
		if err == nil {
			// positional!
			if i >= len(elements) {
				return true
			}
			value = elements[i]
		} else {
			if tag != "" {
				name = tag
			}
			value = kv[name]
		}
		if value == "" {
			return true
		}
		set, err := MakeStringSetter(f.Type)
		if err != nil {
			walkErr = errors.Wrapf(err, "Cannot set %s", f.Type)
			return true
		}
		err = set(v.Elem().FieldByIndex(f.Index), value)
		if err != nil {
			walkErr = errors.Wrap(err, f.Name)
		}
		return true
	})
	return walkErr
}

type FillOptArg func(*fillOpt)

type fillOpt struct {
	tag string
}

func WithTag(tag string) FillOptArg {
	return func(o *fillOpt) {
		o.tag = tag
	}
}
