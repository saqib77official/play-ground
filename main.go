package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type NameInput struct {
	Name string `json:"name"`
}

type NameRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Open or create SQLite database
	db, err := sql.Open("sqlite", "./names.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS names (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// POST /name → save name
	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { // <-- FIXED HERE
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var input NameInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO names (name) VALUES (?)", input.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Name saved successfully"))
	})

	// GET /names → return all names
	http.HandleFunc("/names", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name FROM names")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []NameRecord
		for rows.Next() {
			var n NameRecord
			rows.Scan(&n.ID, &n.Name)
			results = append(results, n)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// Serve frontend files like index.html
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
