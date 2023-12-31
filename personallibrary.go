package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"personallibrary/internal/bookdb"
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
	bookList []Book
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
	initBooklist()
}

func main() {
	frontend := http.StripPrefix("/", CustomFileServer(http.Dir("frontend"), handlePageNotFound))
	http.Handle("/", frontend)

	http.HandleFunc("/api/v1/search", handleSearch)
	http.HandleFunc("/api/v1/upload", handleUpload)
	http.HandleFunc("/api/v1/info", handleInfo)

	LogInfo.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handlePageNotFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404.html", http.StatusSeeOther)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	results, err := bookdb.SearchBooks(query)
	if err != nil {
		LogError.Println("Searching books failed:\n\t", err)
	}

	books_json, err := json.Marshal(results)
	if err != nil {
		LogError.Println("Searching books failed:\n\t", err)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(books_json))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Uploading book")

	w.Header().Set("Content-Type", "application/json")

	r.ParseMultipartForm(10 << 20)
	book, err := makeBook(r)

	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
		return
	}

	file, handler, err := r.FormFile("book-file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	book.FileData = fileBytes

	// err = bookdb.InsertBook(bookdb.Book(book))
	// if err != nil {
	// 	fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
	// 	return
	// }

	fmt.Fprintf(w, `[ {"Status": "%s successfully uploaded"} ]`, book.Title)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	// FileServer mangles query parameters randomly. title -> q, apparently
	title := r.URL.Query().Get("q")
	book, err := bookdb.GetBook(title)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
		return
	}

	book_info_json, err := json.Marshal(book)
	if err != nil {
		LogError.Println("Converting book info to json failed:\n\t", err)
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
	}

	fmt.Fprint(w, string(book_info_json))
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
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	log.Printf("Error reading body: %v", err)
	// }
	// fmt.Println("Body:", string(body))
	// fmt.Println(r.PostFormValue("book-title"))

	var book Book

	description := r.PostFormValue("book-description")
	if description == "" {
		description = "No description available"
	}

	pages := r.PostFormValue("book-pages")

	book.UploadDate = time.Now()
	book.Description = description
	book.FileType = "epub"

	book.Pages, _ = strconv.Atoi(pages)
	book.Title = r.PostFormValue("book-title")
	book.Author = r.PostFormValue("book-author")
	book.Publisher = r.PostFormValue("book-publisher")
	book.PublicationDate = r.PostFormValue("book-publication-date")

	book.Description = r.PostFormValue("book-description")
	// book.CoverImage, _ = r.PostFormValue("book-cover")

	//book.FileData, _ = getBase64ParamBytes(r, "book-data")

	return book, nil
}

func initBooklist() {
	bookList = make([]Book, 3)
	bookList[0] = Book{
		Title:       "The Go Programming Language",
		Description: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
	}

	bookList[1] = Book{
		Title:       "The C Programming Language",
		Description: "The original K&R C book by Brian W. Kernighan and Dennis M. Ritchie.",
	}

	bookList[2] = Book{
		Title:       "The Rust Programming Language",
		Description: "A crabby introduction to Rust.",
	}

	for _, book := range bookList {
		fmt.Println("Inserting book: ", book.Title)
		bookdb.InsertBook(bookdb.Book(book))
	}
}
