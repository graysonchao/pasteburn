package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/graysonchao/pasteburn"
	log "github.com/sirupsen/logrus"
)

func initDb() error {
	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Notes"))
		return err
	})
	return err
}

func addHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}
		var req struct {
			ID   string
			Body string
			Key  string
		}
		if err := json.Unmarshal(body, &req); err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}
		n := &pasteburn.Note{ID: req.ID, Body: req.Body}
		n.Save([]byte(req.Key))
		json.NewEncoder(w).Encode(n)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
	} else {
		query := r.URL.Query()
		name := query.Get("name")
		key := query.Get("key")

		n, err := pasteburn.LoadNote([]byte(name), []byte(key))
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(n)
	}
}

func main() {
	initDb()
	http.HandleFunc("/api/create", addHandler)
	http.HandleFunc("/api/view", viewHandler)
	http.ListenAndServe("127.0.0.1:8080", nil)
}
