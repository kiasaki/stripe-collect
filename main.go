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

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
)

type H map[string]interface{}

type Invoice struct {
	Number         string `json:"number"`
	Client         string `json:"client"`
	ClientEmail    string `json:"client_email"`
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

	// Setup Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

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
		chargeInvoice(w, r, invoiceId, invoice, pdfUrl)
		return
	}

	// Render success when invoice is paid
	if invoice.Status == "paid" {
		renderTemplate(w, r, "success", H{
			"invoice": invoice,
			"pdfUrl":  pdfUrl,
		})
		return
	}

	renderTemplate(w, r, "index", H{
		"invoice": invoice,
		"pdfUrl":  pdfUrl,
	})
}

func chargeInvoice(w http.ResponseWriter, r *http.Request, invoiceId string, invoice Invoice, pdfUrl string) {
	if err := r.ParseForm(); err != nil {
		errorHandler(w, r, err)
		return
	}

	// Setup Stripe charge params
	params := &stripe.ChargeParams{
		Amount:    uint64(invoice.Amount),
		Currency:  stripe.Currency(invoice.Currency),
		Statement: "Invoice " + invoice.Number,
		Email:     invoice.ClientEmail,
	}
	params.SetSource(&stripe.CardParams{
		Number: r.FormValue("number"),
		Month:  r.FormValue("exp_month"),
		Year:   r.FormValue("exp_year"),
		CVC:    r.FormValue("cvc"),
	})

	// Create Stripe charge
	ch, err := charge.New(params)
	if err != nil {
		message := "An unknown error occured handling your payment."
		if e, ok := err.(*stripe.Error); ok {
			message = e.Msg
		}
		renderTemplate(w, r, "index", H{
			"invoice": invoice,
			"pdfUrl":  pdfUrl,
			"error":   message,
		})
		return
	}

	// Set invoice as paid
	invoice.Status = "paid"
	invoice.StripeChargeId = ch.ID[3:]

	// JSON encode invoice
	contentsString, err := json.MarshalIndent(invoice, "", "  ")
	if err != nil {
		errorHandler(w, r, err)
		return
	}

	// Upload updated invoice
	wc := bucket.Object(invoiceId + ".json").NewWriter(ctx)
	if _, err := fmt.Fprintf(wc, string(contentsString)); err != nil {
		errorHandler(w, r, err)
		return
	}
	if err := wc.Close(); err != nil {
		errorHandler(w, r, err)
		return
	}

	http.Redirect(w, r, "/?i="+invoiceId, http.StatusFound)
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
