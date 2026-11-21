//go:build fapro

package demo

import (
	"embed"
	"html/template"
)

//go:embed pro_templates/*.html
var proFiles embed.FS

func parseProTemplates(tmpl *template.Template) error {
	_, err := tmpl.ParseFS(proFiles, "pro_templates/*.html")
	return err
}
