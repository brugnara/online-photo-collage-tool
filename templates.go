package main

import "html/template"

var tpls *template.Template

func init() {
	tpls = template.Must(
		template.New(
			"",
		).Funcs(
			template.FuncMap{},
		).ParseGlob("./templates/*.gohtml"))
}
