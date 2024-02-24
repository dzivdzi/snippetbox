package main

import (
	"net/http"

	"github.com/dzivdzi/snippetbox/ui"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Initialize the router
	router := httprouter.New()

	// Hander func which wraps our notFound() helper - we can assign the custom handler for Not Found responses.
	// We can also do the same for 405 - Method not allowed too
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// // Creates a file server which serves files out of the "./ui/static/" directory
	// // Note that the path given to the Dir function is relative to the project directory
	// // All requests for directories (with no index.html file) return a 404 Not Found response, instead of a directory listing or a redirect. This works for requests both with and without a trailing slash.
	// // The default behavior of http.FileServer isn't changed any other way, and index.html files work as per the standard library documentation.
	// router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Take the ui.Files embedded filesystem and convert it to a http.FS type so
	// that it satisfies the http.FileSystem interface. We then pass that to the
	// http.FileServer() function to create the file server handler.
	fileServer := http.FileServer(http.FS(ui.Files))
	// Our static files are contained in the "static" folder of the ui.Files
	// embedded filesystem. So, for example, our CSS stylesheet is located at
	// "static/css/main.css". This means that we now longer need to strip the
	// prefix from the request URL -- any requests that start with /static/ can
	// just be passed directly to the file server and the corresponding
	// static file will be served (so long as it exists).
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// Create a new middleware chain containing the middleware specific to our
	// dynamic application routes. For now, this chain will only contain the
	// LoadAndSave session middleware but we'll add more to it later.

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// because the alice ThenFunc()
	// method returns a http.Handler (rather than a http.HandlerFunc) we also
	// need to switch to registering the route using the router.Handler() method.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))

	// Protected (authenticated-only) application routes using a new "protected" middleware chanin which includes
	// the require authentication middleware
	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Return the 'standard' middleware chain followed by the servemux.
	return standard.Then(router)
}

/*
Embeds the standard http.FileSystem.
We then implement an Open() method on it — which gets called
each time our http.FileServer receives a request.
*/
// type neuteredFileSystem struct {
// 	fs http.FileSystem
// }

// /*
// Open() method we Stat() the requested file path and use the IsDir() method
// to check whether it's a directory or not. If it is a directory
// we then try to Open() any index.html file in it. If no index.html file exists,
// then this will return a os.ErrNotExist error
// (which in turn we return and it will be transformed into a 404 Not Found response by http.Fileserver).
// We also call Close() on the original file to avoid a file descriptor leak.
// Otherwise, we just return the file and let http.FileServer do its thing.
// */
// func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
// 	f, err := nfs.fs.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	s, _ := f.Stat()
// 	if s.IsDir() {
// 		index := filepath.Join(path, "index.html")
// 		if _, err := nfs.fs.Open(index); err != nil {
// 			closeErr := f.Close()
// 			if closeErr != nil {
// 				return nil, closeErr
// 			}
// 			return nil, err
// 		}
// 	}
// 	return f, nil
// }
