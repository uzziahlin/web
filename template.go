package web

import (
	"bytes"
	"context"
	"html/template"
)

func ServerWithTemplateEngine(t TemplateEngine) Option {
	return func(httpServer *DefaultHttpServer) {
		httpServer.t = t
	}
}

type TemplateEngine interface {
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}

type GoTemplateEngine struct {
	Tpl *template.Template
}

func (g GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	buffer := bytes.Buffer{}

	err := g.Tpl.ExecuteTemplate(&buffer, tplName, data)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (g GoTemplateEngine) LoadGlob(pattern string) error {

	var err error

	g.Tpl, err = g.Tpl.ParseGlob(pattern)

	if err != nil {
		return err
	}

	return nil
}
