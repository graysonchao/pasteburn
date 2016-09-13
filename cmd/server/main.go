package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/graysonchao/pasteburn"
)

func addHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
	} else {
		nbReader := r.Body
		nb, err := ioutil.ReadAll(nbReader)
		if err != nil {
			fmt.Fprintf(w, "Error! %s", err)
		}
		n := &pasteburn.Note{Body: nb}
		filename, _ := n.Save()
		fmt.Fprintf(w, filename)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
	} else {
		query := r.URL.Query()
		name := query.Get("name")

		n, err := pasteburn.LoadNote(string(name))
		if err != nil {
			log.Fatal(err)
		}

		log.Print(n.Body)

		fmt.Fprintf(w, "%s", n.Body)
	}
}

func initDb() error {
	db, err := bolt.Open("pasteburn.db", 0600, nil)
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Notes"))
		return err
	})
	return err
}

func main() {
	initDb()
	http.HandleFunc("/", addHandler)
	http.HandleFunc("/view", viewHandler)
	http.ListenAndServe("127.0.0.1:8080", nil)
}
