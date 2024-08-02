package main

import (
	"bytes"
	"html/template"
	"io"
)

func writeTemplate(w io.Writer, data interface{}, templateFile string) error {
	var rendered bytes.Buffer
	funcMap := template.FuncMap{
		"revIndex": func(index, length int) (revIndex int) { return (length - 1) - index },
	}
	tmpls, err := template.New("page.gohtml").Funcs(funcMap).ParseGlob("templates/*.gohtml")
	if err != nil {
		return err
	}
	if err := tmpls.ExecuteTemplate(&rendered, templateFile, data); err != nil {
		return err
	}

	_, err = w.Write(rendered.Bytes())
	return err
}
