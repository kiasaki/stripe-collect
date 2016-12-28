package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

var t *template.Template

func init() {
	t = template.Must(template.ParseGlob("views/*"))
}

func main() {
	http.HandleFunc("/invoices", invoicesHandler)
	http.HandleFunc("/new-invoice", newInvoiceHandler)
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("started listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func invoicesHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	// Fetch
	// Display list
}

func newInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	// Create & Redirect if POST
	// Display form
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := t.ExecuteTemplate(w, "index", nil); err != nil {
		errorHandler(w, r, err)
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if err := t.ExecuteTemplate(w, "error", err); err != nil {
		w.Write([]byte("Internal Server Error"))
	}
}
