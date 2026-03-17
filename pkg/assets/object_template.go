package assets

import (
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/tnozicka/k8s-controller-lib/pkg/helpers"
)

type ObjectTemplate[T runtime.Object] struct {
	tmpl    *template.Template
	decoder runtime.Decoder
}

func ParseObjectTemplate[T runtime.Object](
	name,
	tmplString string,
	funcMap template.FuncMap,
	decoder runtime.Decoder,
) (ObjectTemplate[T], error) {
	tmpl, err := template.New(name).Funcs(funcMap).Parse(tmplString)
	if err != nil {
		return ObjectTemplate[T]{}, fmt.Errorf("can't parse template %q: %w", name, err)
	}

	return ObjectTemplate[T]{
		tmpl:    tmpl,
		decoder: decoder,
	}, nil
}

func ParseObjectTemplateOrDie[T runtime.Object](
	name string,
	tmplString string,
	funcMap template.FuncMap,
	decoder runtime.Decoder,
) ObjectTemplate[T] {
	return helpers.Must(ParseObjectTemplate[T](name, tmplString, funcMap, decoder))
}

func (t *ObjectTemplate[T]) RenderObject(inputs any) (T, string, error) {
	return RenderAndDecode[T](t.tmpl, inputs, t.decoder)
}

type ObjectTemplateItem[T runtime.Object] struct {
	Template ObjectTemplate[T]
	Data     map[string]any
}

func (i *ObjectTemplateItem[T]) Render() (T, string, error) {
	return i.Template.RenderObject(i.Data)
}
