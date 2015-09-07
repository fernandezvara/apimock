package apimock

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// NewAPIMock returns a new instance of the Mock API
func NewAPIMock(cors bool, log *logrus.Logger, apiType string) *APIMock {
	return &APIMock{
		CORSEnabled: cors,
		Log:         log,
		Type:        verifyType(apiType),
	}
}

// APIMock is the main struct that
type APIMock struct {
	CORSEnabled bool
	Log         *logrus.Logger
	Type        string
	URIMocks    []*URIMock
	URL         string
	server      *httptest.Server
}

// URIMock represents a API call and its response
type URIMock struct {
	Method   string
	URI      string
	Response interface{}
}

// ErrorMessage is the struct to format error messages returned by API
type ErrorMessage struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// Start runs the APIMock server
func (a *APIMock) Start() {
	router := a.createRouter()
	a.server = httptest.NewServer(router)
	a.URL = a.server.URL

	a.Log.WithFields(logrus.Fields{"service": "apiMock"}).Infoln("Listening: " + a.server.Listener.Addr().String())
	a.Log.WithFields(logrus.Fields{"service": "apiMock"}).Infoln("API: Started.")
}

// Stop finishes listening
func (a *APIMock) Stop() {
	a.URL = ""
	a.server.Close()
	a.Log.WithFields(logrus.Fields{"service": "apiMock"}).Infoln("API: Stopped.")
}

// AddMock adds a new mock route/Response
func (a *APIMock) AddMock(uriMock *URIMock) {
	a.URIMocks = append(a.URIMocks, uriMock)
}

func (a *APIMock) createRouter() *mux.Router {
	r := mux.NewRouter()

	for _, mock := range a.URIMocks {
		lMethod := mock.Method
		lURI := mock.URI
		lResponse := mock.Response
		a.Log.WithFields(logrus.Fields{
			"method": lMethod,
			"route":  lURI,
		}).Info("Registering HTTP route")
		wrap := func(w http.ResponseWriter, r *http.Request) {
			a.Log.WithFields(logrus.Fields{"service": "apiMock", "method": r.Method, "uri": r.RequestURI, "ip": r.RemoteAddr}).Info("HTTP request received")
			if a.CORSEnabled {
				writeCorsHeaders(w, r)
			}
			if a.Type == "json" {
				json.NewEncoder(w).Encode(lResponse)
			}
			if a.Type == "xml" {
				xml.NewEncoder(w).Encode(lResponse)
			}
		}
		wrapOptions := func(w http.ResponseWriter, r *http.Request) {
			a.Log.WithFields(logrus.Fields{"service": "apiMock", "method": "OPTIONS", "uri": r.RequestURI, "ip": r.RemoteAddr}).Info("HTTP request received")
			if a.CORSEnabled {
				writeCorsHeaders(w, r)
			}
			w.WriteHeader(http.StatusOK)
		}
		// add the new route
		r.Path(mock.URI).Methods(mock.Method).HandlerFunc(wrap)
		r.Path(mock.URI).Methods("OPTIONS").HandlerFunc(wrapOptions)
	}

	return r
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

func verifyType(t string) string {
	if t == "json" || t == "xml" {
		return t
	}
	panic("not allowed API Type!")
}
