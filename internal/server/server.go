package server

import (
	"net/http"
)

func StartServer() {
	http.HandleFunc("/reminders", remindersHandler)
	http.ListenAndServe(":8080", nil)
}

func remindersHandler(w http.ResponseWriter, r *http.Request) {
	// Respond with reminders list from SQLite
}
