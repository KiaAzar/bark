package bark

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var barkPath string = filepath.Join(os.Getenv("HOME"), ".bark")
var databaseFilename string = "database"
var databaseFullFilename string = filepath.Join(barkPath, databaseFilename)

func InitializeDatabase() (err error) {
	err = os.MkdirAll(barkPath, os.ModePerm)
	if err != nil {
		return
	}

	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	sqlStmt := `
	create table if not exists bookmarks (
		uuid text not null unique primary key,
		added_ts integer not null unique,
		archived_ts integer unique,
		title text,
		url text not null
	);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return
	}

	return
}

func GetBookmarks() (bookmarks []Bookmark, err error) {
	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	sqlStmt := `
	select
		uuid,
		added_ts,
		url,
		title
	from
		bookmarks
	where
		archived_ts is null
	`
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uuid string
		var added_ts int64
		var url string
		var title string
		err = rows.Scan(&uuid, &added_ts, &url, &title)
		if err != nil {
			return
		}

		hostname := GetHostFromURL(url)

		bookmarks = append(bookmarks, Bookmark{uuid, added_ts, url, title, hostname})
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func GetBookmarkByUUID(UUID string) (err error, bookmark Bookmark) {
	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	var uuid string
	var added_ts int64
	var url string
	var title string

	sqlStmt := `
	select
		uuid,
		added_ts,
		url,
		title
	from
		bookmarks
	where
		uuid=?
	`
	err = db.QueryRow(sqlStmt, UUID).Scan(&uuid, &added_ts, &url, &title)
	if err != nil {
		return
	}

	hostname := GetHostFromURL(url)

	bookmark = Bookmark{uuid, added_ts, url, title, hostname}

	return
}

func GetArchivedBookmarks() (bookmarks []Bookmark, err error) {
	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	sqlStmt := `
	select
		uuid,
		added_ts,
		archived_ts,
		url,
		title
	from
		bookmarks
	where
		archived_ts is not null
	order by
		archived_ts
	`
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uuid string
		var added_ts int64
		var archived_ts int64
		var url string
		var title string
		err = rows.Scan(&uuid, &added_ts, &archived_ts, &url, &title)
		if err != nil {
			return
		}

		hostname := GetHostFromURL(url)

		bookmarks = append(bookmarks, Bookmark{uuid, added_ts, url, title, hostname})
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func AddBookmark(url string) (title string, err error) {
	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	new_uuid := uuid.New().String()
	added_ts := time.Now().Unix()
	title, err = GetPageTitle(url)
	if err != nil {
		return
	}

	_, err = db.Exec("insert into bookmarks(uuid, added_ts, url, title) values(?, ?, ?, ?)", new_uuid, added_ts, url, title)

	return
}

func ArchiveBookmark(uuid string) (err error) {
	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	archived_ts := time.Now().Unix()
	_, err = db.Exec("update bookmarks set archived_ts=? where uuid=?", archived_ts, uuid)

	return
}

func DeleteBookmark(uuid string) (err error) {
	db, err := sql.Open("sqlite3", databaseFullFilename)
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec("delete from bookmarks where uuid=?", uuid)

	return
}
