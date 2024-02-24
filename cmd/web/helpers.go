package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/justinas/nosurf"
)

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

/*
The serverError helper writes an error message and stack trace to the
errorLog, then sends a generic 500 Internal Server Error response to the user.
The serverError() helper we use the debug.Stack() function
to get a stack trace for the current goroutine and append it to the
log message. Being able to see the execution path of the
application via the stack trace can be helpful when you’re trying to
debug errors.
*/
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	/*
		Output writes the output for a logging event.
		The string s contains the text to print after
		the prefix specified by the flags of the Logger.
		A newline is appended if the last character of
		s is not already a newline. Calldepth is used
		to recover the PC and is provided for generality,
		although at the moment on all pre-defined paths it will be 2.
	*/
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

/*
The clientError helper sends a specific status code and corresponding
description to the user. We'll use this later to send responses
like 400 "Bad Request" when there's a problem with the request that the user sent.
In the clientError() helper we use the http.StatusText()
function to automatically generate a human-friendly text
representation of a given HTTP status code. For example,
http.StatusText(400) will return the string "Bad Request".
*/
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

/*
notFound helper. This is simply a convenience wrapper around clientError
which sends a 404 Not Found response to the user.
*/
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// Return true if the request is from an auth user, otherwise, return false
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.tmpl'). If no entry exists in the cache with the
	// provided name, then create a new error and call the serverError() helper
	// method that we made earlier and return.
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}
	// Initialize a new buffer
	buf := new(bytes.Buffer)

	// Write the template to the buffer instead of straight to the ResponseWriter.
	// If there is an error, call out serverError()

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write out the provided HTTP status code
	w.WriteHeader(status)

	buf.WriteTo(w)

	// // Execute the template set and write the response body. Again, if there
	// // is any error we call the the serverError() helper.
	// err := ts.ExecuteTemplate(w, "base", data)
	// if err != nil {
	// 	app.serverError(w, err)
	// }
}
