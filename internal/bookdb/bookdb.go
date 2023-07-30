package bookdb

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

var (
	bookDB   *sql.DB
	LogError *log.Logger
)

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		LogError.Println(err)
		return nil, err
	}

	initDBTable(db)

	return db, nil
}

func Init() error {
	LogError = log.New(os.Stderr, "BookDB Error: ", log.Ldate|log.Ltime|log.Lshortfile)

	var err error
	bookDB, err = ConnectDB()
	if err != nil {
		LogError.Println(err)
		return err
	}

	return nil
}

func initDBTable(db *sql.DB) error {
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

func InsertBook(book Book) error {
	stm, err := bookDB.Prepare(`INSERT INTO books(
		title, file_type, upload_date, publication_date,
		publisher, pages, description, author,
		cover_image, file_data) VALUES(?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		LogError.Println("Unable to prepare for book insertion:\n\t", err)
		return err
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
		LogError.Println("Unable to insert book:\n\t", err)
		return err
	}

	return nil
}
