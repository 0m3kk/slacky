package templates

import (
	"embed"
	"log"
	"text/template"

	"github.com/stoewer/go-strcase"
)

//go:embed *.tmpl
var embedded embed.FS

type Name string

const (
	StructTmpl     Name = "struct.tmpl"
	DispatcherTmpl Name = "dispatcher.tmpl"
)

var AllTemplates = []Name{
	StructTmpl,
	DispatcherTmpl,
}

var templates map[Name]*template.Template

var helperFunc = template.FuncMap{
	"snake": strcase.SnakeCase,
	"camel": strcase.UpperCamelCase,
}

func init() {
	templates = make(map[Name]*template.Template)
	for _, tmpl := range AllTemplates {
		t, err := template.New(string(tmpl)).Funcs(helperFunc).ParseFS(embedded, string(tmpl))
		if err != nil {
			log.Fatalln(err)
		}
		templates[tmpl] = t
	}
}

func GetTemplate(name Name) (*template.Template, bool) {
	t, ok := templates[name]
	return t, ok
}
