package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	sourceList := []string{
		"https://v.firebog.net/hosts/lists.php?type=tick",
		"https://v.firebog.net/hosts/lists.php?type=nocross",
	}

	var err error
	defer func() {
		if err != nil {
			log.Fatal(err)
		}
	}()

	if len(os.Args) < 2 {
		fmt.Printf("no path provided to sqlite db")
		return
	}
	path := os.Args[1]

	var sources []string

	for _, url := range sourceList {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("unable to fetch adlists: %s", err)
			return
		}
		defer resp.Body.Close()

		reader := csv.NewReader(resp.Body)

		for {
			source, err := reader.Read()

			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("unable to parse adlists: %s", err)
				return
			}

			if len(source) > 0 {
				sources = append(sources, source[0])
			}
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		fmt.Printf("unable to open sqlite db: %s", err)
	}
	defer db.Close()

	for _, s := range sources {
		query := `
		INSERT OR IGNORE INTO adlist 
		(address)
		VALUES (?)
		`

		stmt, err := db.Prepare(query)
		if err != nil {
			fmt.Printf("unable to prepare sql query: %s", err)
			return
		}

		res, err := stmt.Exec(s)
		if err != nil {
			fmt.Printf("unable to execute sql query: %s", err)
			return
		}

		if _, err := res.RowsAffected(); err != nil {
			fmt.Printf("unable to check sql result: %s", err)
			return
		}
	}

	fmt.Printf("adlists updated successfully!")
}
