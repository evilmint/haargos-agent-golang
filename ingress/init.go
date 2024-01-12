package ingress

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type Ingress struct {
}

func NewIngress() *Ingress {
	return &Ingress{}
}

func (i *Ingress) Run() error {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	defaultIngressPort := 8099

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index.html", map[string]string{
			"Title":   "Haargos",
			"Heading": "Haargos main",
		})
	})

	err := http.ListenAndServe(fmt.Sprintf(":%d", defaultIngressPort), nil)

	if err != nil {
		return err
	}

	return nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := filepath.Join("templates", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
