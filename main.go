package main

import (
	"database/sql"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var db *sql.DB
var default_hostname_base string

func main() {
	// Initialize new CLI app
	app := cli.NewApp()

	app.Name = "mac2hostname"
	app.Author = "Flavio Castelli"
	app.Email = "flavio@castelli.me"
	app.Usage = "Simple hostname generator"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{"port, p", "3000", "Port to listen on"},
		cli.StringFlag{"db, d", "mac2hostname.sqlite3", "Full path to the database file"},
		cli.StringFlag{"hostname_base, H", "machine", "String used to compose the final hostname"},
	}

	app.Action = func(c *cli.Context) {
		init_db(c.String("db"))
		defer db.Close()

		default_hostname_base = c.String("hostname_base")

		router := mux.NewRouter()
		router.HandleFunc("/mac2hostname", mac2hostname).Methods("GET")

		n := negroni.Classic()
		n.UseHandler(router)
		n.Run(":" + c.String("port"))
	}

	app.Run(os.Args)
}

func init_db(dbname string) {
	db, _ = sql.Open("sqlite3", dbname)

	db.Exec("CREATE TABLE IF NOT EXISTS machines(hostname_id INTEGER, hostname_base TEXT NOT NULL, mac TEXT UNIQUE NOT NULL, PRIMARY KEY(hostname_id, hostname_base))")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS mac_machines ON machines (mac)")

	db.SetMaxIdleConns(20)
}

func mac2hostname(w http.ResponseWriter, r *http.Request) {
	_, mac_param_found := r.URL.Query()["mac"]
	if mac_param_found {
		mac := r.URL.Query()["mac"][0]
		var hostname_base string
		if _, ok := r.URL.Query()["hostname_base"]; ok {
			hostname_base = r.URL.Query()["hostname_base"][0]
		} else {
			hostname_base = default_hostname_base
		}
		hostname, err := getHostname(strings.Replace(mac, "_", ":", -1), hostname_base)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal error - checkout logs on the server")
		} else {
			fmt.Fprintf(w, hostname)
		}
	} else {
		log.Println("mac param not provided")
		w.WriteHeader(400)
		fmt.Fprintf(w, "mac param not provided")
	}
}

func getHostname(mac string, hostname_base string) (string, error) {
	stmt, err := db.Prepare("SELECT hostname_id, hostname_base FROM machines WHERE mac = ?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var id, hostname string
	err = stmt.QueryRow(mac).Scan(&id, &hostname)
	switch {
	case err == sql.ErrNoRows:
		hostname, err = insertNewHost(mac, hostname_base)
		if err != nil {
			return "", err
		}
		return hostname, nil
	case err != nil:
		return "", err
	default:
		return hostname + id, nil
	}
}

func insertNewHost(mac string, hostname_base string) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	// find latest hostname_id allocated to this hostname_base group
	stmt, err := tx.Prepare("SELECT hostname_id FROM machines WHERE hostname_base = ? ORDER BY hostname_id DESC LIMIT 1")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(hostname_base).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		id = 1
	case err != nil:
		return "", err
	default:
		id += 1
	}

	stmt, err = tx.Prepare("INSERT INTO machines(mac, hostname_id, hostname_base) VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}

	_, err = stmt.Exec(mac, id, hostname_base)
	if err != nil {
		return "", err
	}
	tx.Commit()

	return hostname_base + strconv.Itoa(id), nil
}
