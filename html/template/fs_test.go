package template_test

import (
	"embed"
	"html/template"
	"strings"
	"testing"

	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/html/template"
)

//go:embed all:testdata/*.html
var fsys embed.FS

func TestFS_Parse(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var (
			is = is.NewRelaxed(t)
		)

		tt, err := template.ParseFS(fsys, "testdata/base.html", "testdata/home.html")
		is.NoErr(err) // template.ParseFS

		var sb strings.Builder
		err = tt.Execute(&sb, nil)
		is.NoErr(err) // Template.Execute

		html := `<a>A</a><b>2</b>
<c>C</c>`
		is.Equal(strings.TrimSpace(sb.String()), html)
	})

	t.Run("OK", func(t *testing.T) {
		var (
			is = is.NewRelaxed(t)
		)

		fs := NewFS(fsys)

		tt, err := fs.Parse("testdata/base.html", "testdata/home.html")
		is.NoErr(err) // FS.Parse

		var sb strings.Builder
		err = tt.Execute(&sb, nil)
		is.NoErr(err) // Template.Execute

		html := `<a>A</a><b>2</b>
<c>C</c>`
		is.Equal(strings.TrimSpace(sb.String()), html)
	})
}
