package web

import (
	"time"

	"github.com/alexedwards/scs/v2"
)

func newSessionManager() *scs.SessionManager {
	sm := scs.New()
	sm.Lifetime = 3 * time.Minute
	sm.Cookie.HttpOnly = true

	return sm
}
