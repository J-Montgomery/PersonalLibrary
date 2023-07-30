package main

import (
	"fmt"
	"net/http"
)

func main() {
	frontend := http.FileServer(http.Dir("frontend"))
	http.Handle("/", frontend)

	http.HandleFunc("/api/v1/search", handleSearch)
	http.HandleFunc("/api/v1/checkout", handleCheckout)
	http.HandleFunc("/api/v1/return", handleReturn)
	http.HandleFunc("/api/v1/info", handleInfo)
	http.HandleFunc("/api/v1/upload", handleUpload)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Search API called")
	return
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Checkout API called")
	return
}

func handleReturn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Return API called")
	return
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Info API called")
	return
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Upload API called")
	return
}
