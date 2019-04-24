package template

import (
	"html/template"
	"net/http"

	"gitlab.com/asciishell/tfs-go-auction/internal/errs"
)

type Templates map[string]*template.Template

const prefix = "internal/template/html/"

func NewTemplates() Templates {
	temps := make(Templates)
	temps["all_lots"] = template.Must(template.ParseFiles(prefix+"all_lots.html", prefix+"base.html", prefix+"lot_table.html"))
	temps["user_lots"] = template.Must(template.ParseFiles(prefix+"user_lots.html", prefix+"base.html", prefix+"lot_table.html"))
	temps["lot_details"] = template.Must(template.ParseFiles(prefix+"lot_details.html", prefix+"base.html"))

	return temps
}

func (t Templates) Render(w http.ResponseWriter, name string, viewModel interface{}) {
	tmpl, ok := t[name]
	if !ok {
		http.Error(w, errs.NewErrorStr("can't find template").StringJSON(), http.StatusInternalServerError)
		return
	}
	err := tmpl.ExecuteTemplate(w, "base", viewModel)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusInternalServerError)
	}
}
