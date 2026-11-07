package main

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}

func userIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(userIDKey).(string)
	return userID, ok
}

func WithUserID[T any](ctx context.Context, userID T) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
