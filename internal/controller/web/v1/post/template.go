package post

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.html
var templateFS embed.FS

type templates struct {
	tmpl *template.Template
}

func newTemplates() *templates {
	root := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int64) int64 { return a + b },
		"sub": func(a, b int64) int64 { return a - b },
	})
	tmpl := template.Must(root.ParseFS(templateFS, "templates/*.html"))

	return &templates{tmpl}
}

func (t templates) Render(wr io.Writer, name string, data any) error {
	return t.tmpl.ExecuteTemplate(wr, name, data)
}
