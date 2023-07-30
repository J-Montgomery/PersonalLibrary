package main

import (
	"fmt"
	"net/http"
)

func main() {
	frontend := http.FileServer(http.Dir("frontend"))
	http.Handle("/", frontend)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
