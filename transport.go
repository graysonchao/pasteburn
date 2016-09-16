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

// MakeTextAddHandler returns a handler that uses a Service to serve add requests
func MakeTextAddHandler(ctx context.Context, s Service) handler {
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

			d, err := NewDocument([]byte(req.Body), []byte(req.Key))
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			err = s.PostDocument(ctx, d)

			json.NewEncoder(w).Encode(d)
		}
	}
}

// MakeTextViewHandler returns a handler that uses a Service to serve view requests
func MakeTextViewHandler(ctx context.Context, s Service) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id, err := uuid.ParseHex(query.Get("id"))
		if err != nil {
			panic(err)
		}

		key := []byte(query.Get("key"))

		d, err := s.GetDocument(ctx, *id, key)

		json.NewEncoder(w).Encode(&struct {
			id   string
			body string
		}{
			id:   d.ID.String(),
			body: string(d.Contents),
		})
	}
}

// MakeImageAddHandler returns a handler that uses a Service to serve add requests
func MakeImageAddHandler(ctx context.Context, s Service) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
		} else {
			key := r.FormValue("key")

			file, _, err := r.FormFile("image")
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			rawImage, err := ioutil.ReadAll(file)
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			d, err := NewDocument(rawImage, []byte(key))
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			err = s.PostDocument(ctx, d)
			if err != nil {
				fmt.Fprintf(w, "Error! %s", err)
			}

			json.NewEncoder(w).Encode(d)
		}
	}
}

// MakeImageViewHandler returns a handler that uses a Service to serve view requests
func MakeImageViewHandler(ctx context.Context, s Service) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id, err := uuid.ParseHex(query.Get("id"))
		if err != nil {
			panic(err)
		}

		key := []byte(query.Get("key"))

		d, err := s.GetDocument(ctx, *id, key)

		rawData := d.Contents
		contentType := http.DetectContentType(rawData)

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", string(len(d.Contents)))

		w.Write(rawData)
	}
}
