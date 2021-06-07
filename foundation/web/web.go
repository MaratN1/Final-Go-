package web

import (
	"context"
	"github.com/dimfeld/httptreemux"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const KeyValues ctxKey = 1

type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	app := App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}

	return &app
}

func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		// Start or expand a distributed trace.
		ctx := r.Context()

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}
		ctx = context.WithValue(r.Context(), KeyValues, &v)

		err := handler(ctx, w, r)
		if err != nil {
			a.SignalShutdown()
			return
		}
	}

	a.ContextMux.Handle(method, path, h)
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}
