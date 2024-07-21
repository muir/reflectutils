# reflectutils - utility functions for working with Golang's reflect package

[![GoDoc](https://godoc.org/github.com/muir/reflectutils?status.png)](https://pkg.go.dev/github.com/muir/reflectutils)
![unit tests](https://github.com/muir/reflectutils/actions/workflows/go.yml/badge.svg)
[![report card](https://goreportcard.com/badge/github.com/muir/reflectutils)](https://goreportcard.com/report/github.com/muir/reflectutils)
[![codecov](https://codecov.io/gh/muir/reflectutils/branch/main/graph/badge.svg)](https://codecov.io/gh/muir/reflectutils)

Install:

	go get github.com/muir/reflectutils

---

Reflectutils is simply a repository for functions useful for working with Golang's reflect package.

Here's the highlights:

## Walking structures

```go
func WalkStructElements(t reflect.Type, f func(reflect.StructField) bool)
```

Recursively walking a struct with reflect has a pitfalls: 

1. It isn't recurse with respect to embeded structs
1. The `Index` field of `reflect.Structfield` of embedded structs is not relative to your starting point.

[WalkStructElements()](https://pkg.go.dev/github.com/muir/reflectutils#WalkStructElements) walks 
embedded elements and it updates `StructField.Index` so that it is
relative to the root struct that was passed in.

## Setting elements

```go
func MakeStringSetter(t reflect.Type, optArgs ...StringSetterArg) (func(target reflect.Value, value string) error, error)
```

[MakeStringSetter()](https://pkg.go.dev/github.com/muir/reflectutils#MakeStringSetter) 
returns a function that can be used to assing to `reflect.Value` given a
string value.  It can handle arrays and slices (splits strings on commas).

## Parsing struct tags

Use [SplitTag()](https://pkg.go.dev/github.com/muir/reflectutils#SplitTag) to break a struct
tag into it's elements and then use [Tag.Fill()](https://pkg.go.dev/github.com/muir/reflectutils#Tag.Fill)
to parse it into a struct.

For example:

```go
type TagInfo struct {
	Name	string	`pt:"0"`     // positional, first argument
	Train	bool	`pt:"train"` // boolean: true: "train", "train=true"; false: "!train", "train=false"
	Count	int	`pt:"count"` // integer value will be parsed
}

st := reflect.StructTag(`foo:"bar,!train,count=9"`)
var tagInfo TagInfo
err := GetTag(st, "foo").Fill(&tagInfo)

// tagInfo.Name will be "bar"
// tagInfo.Train will be false
// tagInfo.Count will be 9
```

## Type names

The `TypeName()` function exists to disambiguate between type names that are
versioned.  `reflect.Type.String()` will hides package versions.  This doesn't
matter unless you've, unfortunately, imported multiple versions of the same
package.  

## Default filler

The `FillInDefaultValues()` function will look at for a struct tag named "default"
and use that value to fill in values where no value has been set.

## Development status

Reflectutils is used by several packages.  Backwards compatability is expected.

