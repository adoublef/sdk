package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
)

const (
	root   = "/"
	parent = ".."
)

// Path is the absolute pathname given to a file
type Path struct {
	Dir  string // dirname
	Base string // filename
}

func (p Path) String() string {
	return path.Join(p.Dir, p.Base)
}

func (p *Path) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return fmt.Errorf("os: unable to decode: %w", err)
	}
	*p, err = ParsePath(s)
	return err
}

func (p Path) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// ParsePath
func ParsePath(s string) (Path, error) {
	if s == "" {
		return Path{}, ErrPathLen0
	}
	if len(s) > 255 {
		return Path{}, ErrPathLen255
	}
	if strings.Contains(s, parent) || !(strings.HasPrefix(s, root)) {
		return Path{}, ErrPathAbs
	}

	dir, base := path.Split(s)
	if base == "" {
		return Path{}, ErrBaseLen0
	}
	return Path{Dir: dir, Base: base}, nil
}

// MustPath panics if an error occurs
func MustPath(s string) Path {
	p, err := ParsePath(s)
	if err != nil {
		panic(err)
	}
	return p
}

func IsNil(p Path) bool {
	return p.Base == ""
}

// Join is a wrapper for the stdlib path.Join function
func Join(elem ...string) string {
	return path.Join(elem...)
}

// Dir returns all but the last element of path
func Dir(s string) string {
	return path.Dir(s)
}

var (
	ErrPathLen0   = errors.New("os: pathname must be provided")
	ErrPathLen255 = errors.New("os: pathname exceeds 255 characters")
	ErrPathAbs    = errors.New("os: absolute path not provided")
	ErrBaseLen0   = errors.New("os: basename must be provided")
)
