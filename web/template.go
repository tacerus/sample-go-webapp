package web

import (
	"html/template"
)

type PageData struct {
	Subject string
	Error        template.HTML
	SessionId    string
}

type Templates struct {
	BadState             *template.Template
	Index *template.Template
	InternalError        *template.Template

}

func newErrorData(text string, id string) *PageData {
	p := PageData{}

	if text != "" {
		p.Error = template.HTML(text)
	}

	if id != "" {
		p.SessionId = id
	}

	return &p
}

func readTemplates(assetDir string) *Templates {
	tmpldir := assetDir + "/templates/"
	tmpls := new(Templates)

	tmpls.BadState = template.Must(template.ParseFiles(tmpldir+"bad_state.html", tmpldir+"top.html", tmpldir+"base.html"))
	tmpls.Index = template.Must(template.ParseFiles(tmpldir+"index.html", tmpldir+"top.html", tmpldir+"base.html"))
	tmpls.InternalError = template.Must(template.ParseFiles(tmpldir+"internal_error.html", tmpldir+"top.html", tmpldir+"base.html"))

	return tmpls
}
