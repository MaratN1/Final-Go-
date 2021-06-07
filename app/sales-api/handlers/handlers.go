package handlers

import (
	"gitlab.com/damanbolghanova/se1903Service/business/auth"
	"gitlab.com/damanbolghanova/se1903Service/business/mid"
	"gitlab.com/damanbolghanova/se1903Service/foundation/web"
	"log"
	"net/http"
	"os"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := check{
		logger: log,
	}

	app.Handle(http.MethodGet, "/readiness", check.readiness, mid.Authenticate(a), mid.Authorize(auth.RoleUser))

	return app
}
