package main

import (
	"deathnote.owner.lalamilight/internal/models"
	"deathnote.owner.lalamilight/internal/validator"
	"errors"

	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func (app *application) main(w http.ResponseWriter, r *http.Request) {
	// Because httprouter matches the "/" path exactly, we can now remove the
	// manual check of r.URL.Path != "/" from this handler.
	notes, err := app.notes.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Notes = notes

	app.render(w, http.StatusOK, "main.tmpl", data)

}
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	notes, err := app.notes.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	//Template
	data := app.newTemplateData(r)
	data.Notes = notes
	app.render(w, http.StatusOK, "home.tmpl", data)
}
func (app *application) victim(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	// ByName() method to get the value of the "id" named
	// parameter from the slice and validate it as normal.
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	note, err := app.notes.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	//Template
	data := app.newTemplateData(r)
	data.Note = note

	app.render(w, http.StatusOK, "victim.tmpl", data)
}
func (app *application) create(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Initialize a new createSnippetForm instance and pass it to the template.
	data.Form = noteCreateForm{
		HowDie: "Heart Attack",
		Die:    40,
	}

	app.render(w, http.StatusOK, "create.tmpl", data)
}

type noteCreateForm struct {
	Fullname            string `form:"fullname"`
	HowDie              string `form:"howdie"`
	Die                 int    `form:"die"`
	validator.Validator `form:"-"`
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm() func adds any data in POST request bodies to the r.PostForm map.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var form noteCreateForm
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	/*
		expires value is automatically mapped to an int data type.
	*/

	form.CheckField(validator.NotBlank(form.Fullname), "fullname", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Fullname, 30), "fullname", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.HowDie), "howdie", "This field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Die, 40, 300, 608400), "die", "This field must equal 40 sec, 5 min or 1 week")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	//Insert
	id, err := app.notes.Insert(form.Fullname, form.HowDie, form.Die)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Victim successfully written!")
	http.Redirect(w, r, fmt.Sprintf("/note/view/%d", id), http.StatusSeeOther)
}
