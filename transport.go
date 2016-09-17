package pasteburn

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	uuid "github.com/nu7hatch/gouuid"

	"golang.org/x/net/context"
)

// MakeTextAddHandler returns a handler that uses a Service to serve add requests
func MakeTextAddHandler(ctx context.Context, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			var req struct {
				Body string
				Key  string
			}

			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			d, err := NewDocument([]byte(req.Body), []byte(req.Key))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			err = s.PostDocument(ctx, d)

			json.NewEncoder(w).Encode(d)
		}
	}
}

// MakeTextViewHandler returns a handler that uses a Service to serve view requests
func MakeTextViewHandler(ctx context.Context, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {
			query := r.URL.Query()
			id, err := uuid.ParseHex(query.Get("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
}

// MakeMultiTextAddHandler returns a handler that uses a Service to serve add multidoc requests
func MakeMultiTextAddHandler(ctx context.Context, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			var req struct {
				Body  string
				Key   string
				Count string
			}

			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			count, err := strconv.ParseInt(req.Count, 10, 8)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			md, keys, err := NewMultiDoc([]byte(req.Body), []byte(req.Key), byte(count))

			if err := s.PostMultiDoc(ctx, md); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			res := struct {
				ID   string   `json:"id"`
				Keys [][]byte `json:"keys"`
			}{
				ID: md.ID.String(),
			}
			for _, key := range keys {
				res.Keys = append(res.Keys, key)
			}

			json.NewEncoder(w).Encode(&res)
		}
	}
}

// MakeMultiTextViewHandler returns a handler that uses a Service to serve view multidoc requests
func MakeMultiTextViewHandler(ctx context.Context, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {
			query := r.URL.Query()
			id, err := uuid.ParseHex(query.Get("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			key, err := base64.StdEncoding.DecodeString(query.Get("key"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			d, err := s.GetMultiDoc(ctx, *id, key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			json.NewEncoder(w).Encode(struct {
				ID   string `json:"id"`
				Body string `json:"body"`
			}{
				ID:   d.ID.String(),
				Body: string(d.Contents),
			})
		}
	}
}

// MakeImageAddHandler returns a handler that uses a Service to serve add requests
func MakeImageAddHandler(ctx context.Context, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {
			key := r.FormValue("key")

			file, _, err := r.FormFile("image")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			rawImage, err := ioutil.ReadAll(file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			d, err := NewDocument(rawImage, []byte(key))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			err = s.PostDocument(ctx, d)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			json.NewEncoder(w).Encode(d)
		}
	}
}

// MakeImageViewHandler returns a handler that uses a Service to serve view requests
func MakeImageViewHandler(ctx context.Context, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {
			query := r.URL.Query()
			id, err := uuid.ParseHex(query.Get("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
}
