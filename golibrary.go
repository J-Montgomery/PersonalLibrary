package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golibrary/internal/bookdb"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type FSHandler404 = func(w http.ResponseWriter, r *http.Request)
type Book bookdb.Book

var (
	LogInfo  *log.Logger
	LogWarn  *log.Logger
	LogError *log.Logger
)

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
				LogError.Println("Path not found: " + uri)
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

func init() {
	initLoggers()
	bookdb.Init()
}

func main() {
	frontend := http.StripPrefix("/", CustomFileServer(http.Dir("frontend"), handlePageNotFound))
	http.Handle("/", frontend)

	http.HandleFunc("/api/v1/search", handleSearch)
	http.HandleFunc("/api/v1/upload", handleUpload)

	LogInfo.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handlePageNotFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404.html", http.StatusSeeOther)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	books := make([]Book, 3)
	books[0] = Book{
		Title:       "The Go Programming Language",
		Description: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
	}

	books[1] = Book{
		Title:       "The C Programming Language",
		Description: "The original K&R C book by Brian W. Kernighan and Dennis M. Ritchie.",
	}

	books[2] = Book{
		Title:       "The Rust Programming Language",
		Description: "A crabby introduction to Rust.",
	}

	books_json, err := json.Marshal(books)
	if err != nil {
		LogError.Println("Converting books to json failed:\n\t", err)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(books_json))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	book, err := makeBook(r)

	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
		return
	}

	err = bookdb.InsertBook(bookdb.Book(book))
	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
		return
	}

	fmt.Fprintf(w, `[ {"Status": "Book successfully uploaded"} ]`)
}

func initLoggers() {
	LogInfo = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogWarn = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogError = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func getBase64ParamBytes(r *http.Request, param string) ([]byte, error) {
	var value = r.URL.Query().Get(param)
	return base64.StdEncoding.DecodeString(value)
}

func getBase64Param(r *http.Request, param string) (string, error) {
	value, err := getBase64ParamBytes(r, param)
	return string(value), err
}

func makeBook(r *http.Request) (Book, error) {
	fmt.Println("Book params were:", r.URL.Query())

	var book Book

	description, err := getBase64Param(r, "description")
	if err != nil {
		description = "No description available"
	}

	pages, _ := getBase64Param(r, "pages")

	book.UploadDate = time.Now()
	book.Description = description
	book.FileType = "epub"

	book.Pages, _ = strconv.Atoi(pages)
	book.Title, _ = getBase64Param(r, "title")
	book.Author, _ = getBase64Param(r, "author")
	book.Publisher, _ = getBase64Param(r, "publisher")
	book.PublicationDate, _ = getBase64Param(r, "publisher")

	book.Description, _ = getBase64Param(r, "description")
	book.CoverImage, _ = getBase64ParamBytes(r, "cover")
	book.FileData, _ = getBase64ParamBytes(r, "file")

	return book, nil
}
