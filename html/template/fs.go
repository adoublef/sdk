package template

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"path/filepath"
)

// An FS provides access to a file system that produces a safe HTML document templates.
type FS struct {
	fsys  fs.FS
	funcs template.FuncMap
}

// Parse parses the named files and associates the resulting templates with t
func (fsys *FS) Parse(filenames ...string) (Template, error) {
	t, err := template.New(filepath.Base(filenames[0])).Funcs(fsys.funcs).ParseFS(fsys.fsys, filenames...)
	if err != nil {
		return nil, errors.Join(ErrParseTemplate, err)
	}
	return t, nil
}

// MustParse will panic if unable to parse files
func (fsys *FS) MustParse(filenames ...string) Template {
	t, err := fsys.Parse(filenames...)
	if err != nil {
		panic(err)
	}
	return t
}

// Funcs adds the elements of the argument map to the template's function map.
func (fsys *FS) Funcs(funcs ...template.FuncMap) *FS {
	for _, f := range funcs {
		maps.Copy(fsys.funcs, f)
	}
	return fsys
}

// NewFS allocates a new file system for templates
func NewFS(fsys fs.FS) *FS {
	return &FS{fsys: fsys, funcs: make(template.FuncMap)}
}

type Template interface {
	// Execute applies a parsed template to the specified data object, writing the output to wr
	Execute(wr io.Writer, data any) error
	// ExecuteTemplate applies the template associated with t that has the given
	// name to the specified data object and writes the output to wr. 
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

var (
	ErrParseTemplate = errors.New("template: parse template files")
)
