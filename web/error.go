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
