package main

import "net/http"

func main() {
	db := initDB()

	http.HandleFunc("/register", registerHandler(db))
	http.HandleFunc("/login", loginHandler(db))
	http.Handle("/hello", requireAuth(db, helloHandler))

	http.ListenAndServe(":8080", nil)
}
