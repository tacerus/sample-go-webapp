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
	"net/http"
)

const (
	ERR_MISC  = 0 // internal issue
	ERR_CODE  = 1 // no or unexpected code value in session
	ERR_STATE = 2 // no or unexpected state value in session
	ERR_TOKEN = 3 // no or unexpected token value in session
	ERR_PARAM = 4 // missing query parameters
	ERR_ILLEG = 5 // operation on data not owned by requestor
)

func (app *WebApp) errorHandler(w http.ResponseWriter, r *http.Request, appErr int, text string) {
	p := newErrorData(text, app.getSessionId(r))

	switch appErr {
	case ERR_MISC:
		w.WriteHeader(http.StatusInternalServerError)
		app.templates.InternalError.ExecuteTemplate(w, "base", p)
	case ERR_STATE:
		w.WriteHeader(http.StatusBadRequest)
		app.templates.BadState.ExecuteTemplate(w, "base", p)
	}
}
