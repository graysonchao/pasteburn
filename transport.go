package pasteburn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	uuid "github.com/nu7hatch/gouuid"

	"golang.org/x/net/context"
)

type handler func(w http.ResponseWriter, r *http.Request)

// MakeAddHandler returns a handler that uses a Service to serve add requests
func MakeAddHandler(ctx context.Context, s Service) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
		} else {

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}
			var req struct {
				Body string
				Key  string
			}
			if err := json.Unmarshal(body, &req); err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			d, err := MakeDocumentRandomID([]byte(req.Body), []byte(req.Key))
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			err = s.PostDocument(ctx, d)

			json.NewEncoder(w).Encode(d)
		}
	}
}

// MakeViewHandler returns a handler that uses a Service to serve view requests
func MakeViewHandler(ctx context.Context, s Service) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id, err := uuid.ParseHex(query.Get("name"))
		if err != nil {
			panic(err)
		}

		key := []byte(query.Get("key"))

		d, err := s.GetDocument(ctx, *id, key)

		json.NewEncoder(w).Encode(d)
	}
}
