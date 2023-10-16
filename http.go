package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func startHTTPServer(db *sql.DB, config Configuration) {
	router := mux.NewRouter()
	router.HandleFunc("/income", func(w http.ResponseWriter, r *http.Request) {
		incomeData, err := getIncomeData(db, config.DB.Table)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(incomeData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
