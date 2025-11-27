/*
   Copyright (C) 2025  SUSE LLC <georg.pfuetzenreuter@suse.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package web

import (
	"html/template"
)

type PageData struct {
	Subject   string
	Error     template.HTML
	SessionId string
}

type Templates struct {
	BadState      *template.Template
	Index         *template.Template
	InternalError *template.Template
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
