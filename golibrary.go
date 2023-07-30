package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

type FSHandler404 = func(w http.ResponseWriter, r *http.Request)

func CustomFileServer(root http.FileSystem, handler404 FSHandler404) http.Handler {

	fs := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		uri := r.URL.Path
		if !strings.HasPrefix(uri, "/") {
			uri = "/" + uri
			r.URL.Path = uri
		}
		uri = path.Clean(uri)

		f, err := root.Open(uri)
		if err != nil {
			if os.IsNotExist(err) {
				handler404(w, r)
				return
			}
		}

		if err == nil {
			f.Close()
		}

		fs.ServeHTTP(w, r)
	})
}

func main() {

	frontend := http.StripPrefix("/", CustomFileServer(http.Dir("frontend"), handlePageNotFound))
	http.Handle("/", frontend)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handlePageNotFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404.html", http.StatusSeeOther)
}
