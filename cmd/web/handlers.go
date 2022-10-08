package main

import (
	"errors"   // didnt find snippet error
	"fmt"      // redirect sprintf view
	"net/http" // routing, statuses, response objects
	"strconv"  // convert string to int in viewSnippet

	"github.com/julienschmidt/httprouter"   // parameter in viewSnippet
	"person.mmaliar.com/internal/models"    // only for the custom error lol
	"person.mmaliar.com/internal/validator" // embedded into form struct
)

// snippetCreateForm defines fields which the decoder will fill. The result
// will be passed into the templateData struct
type snippetCreateForm struct {
	validator.Validator `form:"-"` // decoder to fill, automatic type conversion
	Title               string     `form:"title"`   // title of snippet
	Content             string     `form:"content"` // content of snippet
	Expires             int        `form:"expires"` // days until expiration of snippet
}

// home renders home.tmpl using latest snippets into http response
func (app *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets // fill template data with new snippets
	app.render(w, http.StatusOK, "home.tmpl.html", data)

}

// snippetView extracts snippet ID from URL, tries to get the snippet
// and renders the "view.tmpl" with the snippet if appropriate
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	data := app.newTemplateData(r)
	data.Snippet = snippet
	app.render(w, http.StatusOK, "view.tmpl.html", data)
}

// snippetCreate renders "createtmpl" with empty form templateData
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{
		Title:   "",
		Content: "",
		Expires: 365,
	}
	// Get template from cache, fill it with data, try to execute template to buff
	// if it works go ahead and do it with the right status
	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

// snippetCreatePost decodes form into snippetFormPost struct, and validates form fields.
// If the fields aren't valid, it renders the "create.tmpl" with an error message.
// If the fields are valid, it inserts the snippet into the database and redirects
// into the snippetView URL to view the newly created snippet
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	var form snippetCreateForm

	// decode into form
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7, or 365")
	if !form.Valid() { // if map is not empty
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data) // resend with error
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
