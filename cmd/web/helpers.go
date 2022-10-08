package main

import (
	"bytes"         // buffer
	"errors"        // errors.As to check the decoder error
	"fmt"           // error from string
	"net/http"      // http error helpers
	"runtime/debug" // generate stack trace
	"time"          // add current year to templateData initial

	"github.com/go-playground/form/v4" // decoder error
)

// serverError prints the stack and the error in the errorlog, and then
// writes an internal server error to the http response
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError writes a status into the http response as an error
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound runs clientError with the http status not found
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// render tries to pull the template from the cache, and then tries to write it to the buffer
// with data, and if that succeeds it writes it to the http response
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page] // we try to get the page in the template cache
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page) // custom error
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

// newTemplateData initializes the HtML template data with current year
// set to the year
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
		Flash:       app.sessionManager.PopString(r.Context(), "flash"),
	}
}

// decodePostForm parses the request form and tries to decode it into
// the destination dst. If this fails it panics.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm() // parse form
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) { // weird error have to check
			panic(err) // panic if we cant put it in the form
		}

		return err
	}
	return nil
}
