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

		var isTokenSet string

		if i.Stats.GetHAAccessTokenSet() {
			isTokenSet = "Yes"
		} else {
			isTokenSet = "No"
		}

		renderTemplate(w, "index.html", map[string]string{
			"Title":   "Haargos",
			"Heading": "Haargos main",
			"Uptime":  uptime,
			"LastSuccessfulConnection": fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d\n",
				lastConnection.Year(),
				lastConnection.Month(),
				lastConnection.Day(),
				lastConnection.Hour(),
				lastConnection.Minute(),
				lastConnection.Second()),
			"HAAccessTokenSet":   isTokenSet,
			"FailedRequestCount": fmt.Sprintf("%d", i.Stats.GetFailedRequestCount()),
			"DataSentInKb":       fmt.Sprintf("%.1f", float32(i.Stats.GetDataSentInKB())/1024),
			"ObservationCount":   fmt.Sprintf("%d", i.Stats.GetObservationsSentCount()),
			"JobsProcessedCount": fmt.Sprintf("%d", i.Stats.GetJobsProcessedCount()),
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
