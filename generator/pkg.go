package generator

import (
	"text/template"

	"github.com/stoewer/go-strcase"
)

var helperFunc = template.FuncMap{
	"snake": strcase.SnakeCase,
	"camel": strcase.UpperCamelCase,
}
