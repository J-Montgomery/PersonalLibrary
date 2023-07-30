package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type FSHandler404 = func(w http.ResponseWriter, r *http.Request)

type Book struct {
	Title           string
	Author          string
	Publisher       string
	FileType        string
	UploadDate      time.Time
	PublicationDate string
	Pages           int
	Description     string
	CoverImage      []byte
	FileData        []byte
}

var db *sql.DB

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

	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		LogError.Println(err)
		return
	}

	defer db.Close()

	var version string
	err = db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		LogError.Println(err)
		return
	}

	LogInfo.Println("Sqlite3 version: " + version)

	initializeDB(db)

}

func main() {

	//---------------------

	frontend := http.StripPrefix("/", CustomFileServer(http.Dir("frontend"), handlePageNotFound))
	http.Handle("/", frontend)

	http.HandleFunc("/api/v1/search", handleSearch)

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

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	book, err := makeBook(r)

	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Invalid Request"} ]`)
		return
	}
	stm, err := db.Prepare(`INSERT INTO books(
		title, file_type, upload_date, publication_date,
		publisher, pages, description, author,
		cover_image, file_data) VALUES(?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Unknown Failure"} ]`)
		LogError.Println("Unable to prepare for book insertion:\n\t", err)
		return
	}

	defer stm.Close()

	_, err = stm.Exec(
		book.Title,
		book.FileType,
		book.UploadDate.String(),
		book.PublicationDate,
		book.Publisher,
		book.Pages,
		book.Description,
		book.Author,
		book.CoverImage,
		book.FileData,
	)

	if err != nil {
		fmt.Fprintf(w, `[ {"Status": "Unknown Failure"} ]`)
		LogError.Println("Unable to insert book:\n\t", err)
		return
	}

	fmt.Fprintf(w, `[ {"Status": "Book successfully uploaded"} ]`)
}

func initializeDB(db *sql.DB) error {
	stm := `CREATE TABLE IF NOT EXISTS books (
		book_id INTEGER PRIMARY KEY,
		title TEXT,
		file_type TEXT,
		upload_date TEXT,
		publication_date TEXT,
		publisher TEXT,
		pages INTEGER,
		description TEXT,
		author TEXT,
		cover_image BLOB,
		file_data BLOB
	);`
	_, err := db.Exec(stm)

	if err != nil {
		LogError.Println("Error creating book table:\n\t", err)
	}

	return nil
}
