package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/graysonchao/pasteburn"
	uuid "github.com/nu7hatch/gouuid"
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

		rbody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}
		var req struct {
			Body string
			Key  string
		}
		if err := json.Unmarshal(rbody, &req); err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}

		nbody := []byte(req.Body)
		key := []byte(req.Key)
		if err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}

		n, err := pasteburn.MakeNote(nbody, key)
		if err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}
		if err := n.Save(); err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}

		json.NewEncoder(w).Encode(n)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
	} else {
		query := r.URL.Query()
		key := []byte(query.Get("key"))

		uuid, err := uuid.ParseHex(query.Get("name"))
		if err != nil {
			log.Fatal(err)
		}

		n, err := pasteburn.LoadNote(*uuid, key)
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
