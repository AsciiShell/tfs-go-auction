package template

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"gitlab.com/asciishell/tfs-go-auction/internal/errs"
)

type Templates map[string]*template.Template

const prefix = "internal/template/html/"

func NewTemplates() Templates {
	temps := make(Templates)
	pref, _ := filepath.Abs(prefix)
	pref += string(filepath.Separator)
	fmt.Printf("Prefix is: %s", pref)
	temps["all_lots"] = template.Must(template.ParseFiles(pref+"all_lots.html", pref+"base.html", pref+"lot_table.html"))
	temps["user_lots"] = template.Must(template.ParseFiles(pref+"user_lots.html", pref+"base.html", pref+"lot_table.html"))
	temps["lot_details"] = template.Must(template.ParseFiles(pref+"lot_details.html", pref+"base.html"))

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
