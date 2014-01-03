package tigertonic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func acceptJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if "" == accept {
		return true
	}
	return strings.Contains(accept, "*/*") || strings.Contains(accept, "application/json")
}

func errorName(err error) string {
	if namedError, ok := err.(NamedError); ok {
		return namedError.Name()
	}
	if httpEquivError, ok := err.(HTTPEquivError); ok && SnakeCaseHTTPEquivErrors {
		return strings.Replace(
			strings.ToLower(http.StatusText(httpEquivError.Status())),
			" ",
			"_",
			-1,
		)
	}
	t := reflect.TypeOf(err)
	if reflect.Ptr == t.Kind() {
		t = t.Elem()
	}
	if r, _ := utf8.DecodeRuneInString(t.Name()); unicode.IsLower(r) {
		return "error"
	}
	return t.String()
}

func errorStatus(err error) int {
	if httpEquivError, ok := err.(HTTPEquivError); ok {
		return httpEquivError.Status()
	}
	return http.StatusInternalServerError
}

func writeJSONError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorStatus(err))
	if jsonErr := json.NewEncoder(w).Encode(map[string]string{
		"description": err.Error(),
		"error":       errorName(err),
	}); nil != err {
		log.Println(jsonErr)
	}
}

func writePlaintextError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(errorStatus(err))
	fmt.Fprintf(w, "%s: %s", errorName(err), err)
}
