package mid

import (
	"context"
	"errors"
	"gitlab.com/maratm/se1903Service/business/auth"
	"gitlab.com/maratm/se1903Service/foundation/web"
	"net/http"
	"strings"
)

//ErrForbidden is returned when an authenticated user does not have a
//sufficient role for an action
var ErrForbidden = web.NewRequestError(
	errors.New("u are not authorized for that action"),
	http.StatusForbidden,
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			//Expected: authorization: 'bearer <token>'.
			authStr := r.Header.Get("authorization")

			// Parse the authorization header.
			parts := strings.Split(authStr, " ")
			if len(parts) != 2 {
				err := errors.New("expected authorization header format: bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Add claims to the context so they can be retrieved later.
			ctx = context.WithValue(ctx, auth.Key, claims)

			// Call the next handler.
			return handler(ctx, w, r)
		}
		return h
	}
	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(roles ...string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, return failure.
			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return web.NewShutdownError("claims missing from context")
			}

			if !claims.Authorize(roles...) {
				return ErrForbidden
			}

			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
