package ingress

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/evilmint/haargos-agent-golang/statistics"
)

type Ingress struct {
	Stats *statistics.Statistics
}

func NewIngress(stats *statistics.Statistics) *Ingress {
	return &Ingress{
		Stats: stats,
	}
}

func (i *Ingress) Run() error {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	defaultIngressPort := 8099

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uptime := i.Stats.GetUptime()

		lastConnection := i.Stats.GetLastSuccessfulConnection()
		renderTemplate(w, "index.html", map[string]string{
			"Title":              "Haargos",
			"Heading":            "Haargos main",
			"Uptime":             uptime,
			"DataSentInKb":       fmt.Sprintf("%d", i.Stats.GetDataSentInKB()),
			"FailedRequestCount": fmt.Sprintf("%d", i.Stats.GetFailedRequestCount()),
			"ObservationCount":   fmt.Sprintf("%d", i.Stats.GetObservationsSentCount()),
			"LastSuccessfulConnection": fmt.Sprintf("%d-%d-%d %d:%d:%d\n",
				lastConnection.Year(),
				lastConnection.Month(),
				lastConnection.Day(),
				lastConnection.Hour(),
				lastConnection.Hour(),
				lastConnection.Second()),
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
