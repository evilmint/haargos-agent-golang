package ingress

import (
	"fmt"
	"io"
	"net/http"
)

type Ingress struct {
}

func NewIngress() *Ingress {
	return &Ingress{}
}

func (i *Ingress) Run() error {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)

	defaultIngressPort := 8099

	err := http.ListenAndServe(fmt.Sprintf(":%d", defaultIngressPort), nil)

	if err != nil {
		return err
	}

	return nil
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "<html><head></head><body><h1>This is my website! <b>test</b> <a href=\"/hello\">hello</a></h1></body></html>\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP!\n")
}
