package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
)

type H map[string]interface{}

type Invoice struct {
	Number         string `json:"number"`
	Client         string `json:"client"`
	Amount         int    `json:"amount"`
	Currency       string `json:"currency"`
	Status         string `json:"status"`
	StripeChargeId string `json:"stripe_charge_id,omitempty"`
}

func (i Invoice) FormattedAmount() string {
	return fmt.Sprintf("$%.2f %s", float64(i.Amount)/100, i.Currency)
}

var t *template.Template
var ctx context.Context
var jwtConfig *jwt.Config
var bucket *storage.BucketHandle
var notFoundCache = map[string]interface{}{}

func init() {
	// Load templates
	t = template.Must(template.ParseGlob("views/*"))

	// Setup Google Cloud Datastore client
	ctx = context.Background()
	conf, err := google.JWTConfigFromJSON(
		[]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")), storage.ScopeFullControl)
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig = conf
	client, err := storage.NewClient(
		ctx,
		option.WithTokenSource(conf.TokenSource(ctx)),
	)
	if err != nil {
		log.Fatal(err)
	}

	bucket = client.Bucket(os.Getenv("GOOGLE_BUCKET_ID"))
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

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		errorHandler(w, r, err)
	}
}

func invoicesHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	// Fetch
	// Display list
	w.Write([]byte("TODO"))
}

func newInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	// Create & Redirect if POST
	// Display form
	w.Write([]byte("TODO"))
}

func invoiceNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "invoice-not-found", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	invoiceId := r.URL.Query().Get("i")

	// Check the not found cache before
	if _, ok := notFoundCache[invoiceId]; ok {
		invoiceNotFoundHandler(w, r)
		return
	}

	// Try and fetch the invoice details/object
	rc, err := bucket.Object(invoiceId + ".json").NewReader(ctx)
	if err == storage.ErrObjectNotExist {
		// Avoid making this request next time
		notFoundCache[invoiceId] = nil
		invoiceNotFoundHandler(w, r)
		return
	} else if err != nil {
		errorHandler(w, r, err)
		return
	}

	// Fetch file contents
	contents, err := ioutil.ReadAll(rc)
	rc.Close()
	if err != nil {
		errorHandler(w, r, err)
		return
	}

	// Parse it from JSON to a struct
	var invoice = Invoice{}
	if err := json.Unmarshal(contents, &invoice); err != nil {
		errorHandler(w, r, err)
		return
	}

	// Generate PDF Signed URL for download
	var pdfUrl string
	pdfUrl, err = generateSignedUrl(invoiceId)
	if err != nil {
		errorHandler(w, r, err)
		return
	}

	// If user just posted CC info, process it
	if r.Method == "POST" {
		chargeInvoice(w, r, invoiceId, invoice)
		return
	}

	renderTemplate(w, r, "index", H{
		"invoice": invoice,
		"pdfUrl":  pdfUrl,
	})
}

func chargeInvoice(w http.ResponseWriter, r *http.Request, invoiceId string, invoice Invoice) {
	// statement_descriptor
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if err := t.ExecuteTemplate(w, "error", err); err != nil {
		w.Write([]byte("Internal Server Error"))
	}
}

func generateSignedUrl(invoiceId string) (string, error) {
	return storage.SignedURL(
		os.Getenv("GOOGLE_BUCKET_ID"),
		invoiceId+".pdf",
		&storage.SignedURLOptions{
			GoogleAccessID: jwtConfig.Email,
			PrivateKey:     jwtConfig.PrivateKey,
			Method:         "GET",
			Expires:        time.Now().Add(15 * time.Minute),
		},
	)
}
