package constants

import (
	"net/http"

	"github.com/rs/cors"
)

// Email and Mobile Regex defines a regular expression pattern for validating email addresses
// and mobile
const (
	EmailRegex  = "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"
	MobileRegex = "^([+]\\d{2})?\\d{10}$"
)

// CorsOptions defines the CORS (Cross-Origin Resource Sharing) configuration.
var CorsOptions = cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowCredentials: true,
	AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch},
	AllowedHeaders:   []string{"*"},
}
