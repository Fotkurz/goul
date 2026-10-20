package web

import (
	"embed"
	"html/template"
)

//go:embed static/*
var StaticFS embed.FS

//go:embed templates/*
var TemplateFS embed.FS

func ParseTemplates() (*template.Template, error) {
	return template.ParseFS(TemplateFS, "templates/*.html")
}
