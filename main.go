package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	_ "reflect"
	"time"
)

/*
var data = `{
		"consul": "http://consul.liveoak.us.int:3000",
		"nomad": "http://consul.liveoak.us.int:4646"
	}`
*/

const index_file = "index.tmpl"
const db_name = "shortcuts"
const db_file = db_name + ".db"
var search_cut = ""

func init(){
        db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
        if err != nil {
                log.Fatal(err)
        }
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		db, err := tx.CreateBucketIfNotExists([]byte(db_name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_ = db
		return nil
	})
        db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db_name))
		err := b.Delete([]byte("/jobs"))
		if err != nil {
			log.Fatal(err)
		}
                return nil
        })
}

func addShortcut(shortcut string, dest string){
	if shortcut == "/" {
		fmt.Println("empty shortcut won't work")
		return
	}
	db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
        	if err != nil {
			fmt.Println("Failed to open db")
                	log.Fatal(err)
        }
	defer db.Close()
	search_cut = ""
	db.View(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(db_name))
		v := b.Get([]byte(shortcut))
		if v != nil {
			fmt.Println("search before assign : ", search_cut)
			search_cut = string(v)
			fmt.Println("search cut after assign : ", search_cut)
		}
		return err
	})
	if search_cut == "" {
		fmt.Println("adding new shortcut")
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(db_name))
			err := b.Put([]byte(shortcut), []byte(dest))
			if err != nil {
				fmt.Println("something failed on put", err)
			}
			return err
		})
	}
}
/*
func find(s string) error{
        db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
                if err != nil {
                        log.Fatal(err)
                }   
        defer db.Close()
        db.View(func(tx *bolt.Tx) error{
                b := tx.Bucket([]byte(db_name))
                v := b.Get([]byte(s))
                if v != nil {
			search_cut = string(v)
                        return err
                }   
                return err
        })  
	return err
}
*/

func readShortcuts() map[string]string{
        db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
                if err != nil {
                        log.Fatal(err)
                }
	defer db.Close()
	
	shortcuts := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db_name))
		c := b.Cursor()
		for k,v := c.First(); k != nil; k,v = c.Next(){
			shortcuts[string(k)] = string(v)
		}
		return nil
	})
	return shortcuts
}

func jsonMap() map[string]string{
	data, err := ioutil.ReadFile("db.json")
	if err != nil {
		log.Fatal(err)
	}
	d := map[string]string{}
	if err := json.Unmarshal([]byte(data), &d); err != nil {
		log.Fatal(err)
	}
	return d
}

func main(){

	//mux := http.NewServeMux()
	indexHandler := indexHandler()
	log.Fatal(http.ListenAndServe(":80",indexHandler))
}
//fallback http.Handler
func indexHandler() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		d := readShortcuts()
		//d := jsonMap()
		fmt.Println("d : ", d)
		path := r.URL.Path
		if dest, ok := d[path]; ok {
			fmt.Println("redirecting")
			http.Redirect(w,r,dest,http.StatusFound)
			return
		}
		
		fmt.Println("Path : ", path, " not in data")
		switch r.Method {
			case "GET":
				t, err := template.ParseFiles(index_file)
				if err != nil {
					fmt.Println("error with template")
					log.Fatal(err)
				}
				t.Execute(w,d)
			case "POST":
				r.ParseForm()
				shortcut := r.Form["shortcut"][0]
				destination := r.Form["destination"][0]
				addShortcut("/"+shortcut,destination)
				t, err := template.ParseFiles(index_file)
                                if err != nil {
                                        log.Fatal(err)
                                }
				d := readShortcuts()
				t.Execute(w,d)
		}
	}
}
