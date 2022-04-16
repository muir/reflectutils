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
	// Tags is a read-only value
	Tags  Tags
	index map[string]int
}

// SplitTag breaks apart a reflect.StructTag into an array of annotated key/value pairs.
// Tags are expected to be in the conventional format.  What does "contentional"
// mean?  `name:"values,value=value" name2:"value"`.  See https://flaviocopes.com/go-tags/
// a light introduction.
func SplitTag(tags reflect.StructTag) Tags {
	found := make([]Tag, 0, 5)
	s := string(tags)
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

func (t Tags) Set() TagSet {
	index := make(map[string]int)
	for i, tag := range t {
		index[tag.Tag] = i
	}
	return TagSet{
		Tags:  t,
		index: index,
	}
}

func (s TagSet) Get(tag string) Tag {
	t, _ := s.Lookup(tag)
	return t
}

func (s TagSet) Lookup(tag string) (Tag, bool) {
	if i, ok := s.index[tag]; ok {
		return s.Tags[i], true
	}
	return mkTag("", ""), false
}

func GetTag(tags reflect.StructTag, tag string) Tag {
	t, _ := LookupTag(tags, tag)
	return t
}

func LookupTag(tags reflect.StructTag, tag string) (Tag, bool) {
	value, ok := tags.Lookup(tag)
	return Tag{
		Tag:   tag,
		Value: value,
	}, ok
}

// Fill unpacks struct tags into a struct based on tags of the desitnation struct.
// This is very meta.  It is using struct tags to control parsing of struct tags.
// The tag being parsed is the receiver (tag).  The model that controls the parsing
// is the function parameter (model).  The parsing may be adjusted based on the opts.
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
//
// When filling an array value, the default character to split upon is
// comma, but other values can be set with "split=X" to split on X.
// Special values of X are "quote", "space", and "none"
//
// For bool values (and *bool, etc) an antonym can be specified:
//
//	MyBool	bool	`pt:"mybool,!other"`
//
// So, then "mybool" maps to true, "!mybool" maps to false,
// "other" maps to false and "!other" maps to true.
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
	// Break apart the tag into a list of elements (split on ",") and
	// key/values (kv) when the elements have values (split on "=").  If
	// an element doesn't have a value from =, then it gets a value of
	// "t" (true) unless the element name starts with "!" in which case,
	// the "!" is discarded and the value is "f" (false)
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
	// Now walk over the input model that controls the parsing.
	var count int
	var walkErr error
	WalkStructElements(v.Type(), func(f reflect.StructField) bool {
		tag := f.Tag.Get(opt.tag)
		if tag == "-" {
			return false
		}
		count++
		parts := strings.Split(tag, ",")
		var value string
		isBool := NonPointer(f.Type).Kind() == reflect.Bool
		if len(parts) > 0 && parts[0] != "" {
			i, err := strconv.Atoi(parts[0])
			if err == nil {
				// positional!
				if i >= len(elements) {
					return true
				}
				value = elements[i]
			} else {
				if isBool {
					for _, p := range parts {
						if len(p) > 0 && p[0] == '!' {
							if v, ok := kv[p[1:]]; ok {
								value = v
								switch value {
								case "f":
									value = "t"
								case "t":
									value = "f"
								}
							}
						} else if v, ok := kv[p]; ok {
							value = v
							break
						}
					}
				} else {
					value = kv[parts[0]]
				}
			}
		} else {
			value = kv[f.Name]
		}
		if value == "" {
			return true
		}
		var sso []StringSetterArg
		if len(parts) > 1 {
			for _, part := range parts[1:] {
				if strings.HasPrefix(part, "split=") {
					splitOn := part[len("split="):]
					switch splitOn {
					case "quote":
						splitOn = `"`
					case "space":
						splitOn = " "
					}
					sso = append(sso, WithSplitOn(splitOn))
				}
			}
		}
		set, err := MakeStringSetter(f.Type, sso...)
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

// WithTag overrides the tag used by Tag.Fill.  The default is "pt".
func WithTag(tag string) FillOptArg {
	return func(o *fillOpt) {
		o.tag = tag
	}
}
