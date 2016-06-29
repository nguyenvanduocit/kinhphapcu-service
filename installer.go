package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"encoding/json"
	"io/ioutil"
	"flag"
	"fmt"
)

type Item struct {
	Id int `json:"id"`
	PoemVi string `json:"poem_vi"`
}

type Chapter struct{
	Id int `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Total int `json:"total"`
	Items []*Item `json:"items"`
}

func importChapter(fileName string, db *sql.DB){

	chapterFile, err := os.Open("data/" + fileName)
	if err != nil {
		panic(err.Error())
	}
	defer chapterFile.Close()
	var chapter Chapter
	jsonParser := json.NewDecoder(chapterFile)
	err = jsonParser.Decode(&chapter);
	if err != nil {
		panic(err.Error())
	}

	insChapter, err := db.Prepare("INSERT INTO `chapters` (id, slug, name) VALUES( ?, ?, ? )") // ? = placeholder
	if err != nil {
		panic(err.Error())
	}
	_, err = insChapter.Exec(chapter.Id,chapter.Slug, chapter.Name)
	if err != nil {
		panic(err.Error())
	}
	insPost, err := db.Prepare("INSERT INTO `posts` (id, chapter_id, poem_vi) VALUES( ?, ?, ? )") // ? = placeholder
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(chapter.Items); i++ {
		fmt.Println("Item:", chapter.Items[i].Id)
		_, err = insPost.Exec(chapter.Items[i].Id,chapter.Id, chapter.Items[i].PoemVi)
		if err != nil {
			panic(err.Error())
		}
	}
}

func createTable(db *sql.DB, name string,query string){

	if _, err := db.Exec("SET foreign_key_checks = 0"); err != nil {
		panic(err.Error())
	}

	if _, err := db.Exec("DROP TABLE IF EXISTS `"+ name + "`"); err != nil {
		panic(err.Error())
	}

	if _, err := db.Exec(query); err != nil {
		panic(err.Error())
	}
	fmt.Println("Table " + name + " created")
}

func main() {
	var dbScheme string
	flag.StringVar(&dbScheme, "db-scheme", "root:7facd974e4b@/sotaycuame", "Database scheme format username:password@/database_name")
	flag.Parse()

	db, err := sql.Open("mysql", dbScheme)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	createTable(db, "chapters", "CREATE TABLE `chapters` ( `id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(255) DEFAULT NULL, `slug` varchar(255) DEFAULT NULL, PRIMARY KEY (`id`) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	createTable(db, "posts", "CREATE TABLE `posts` ( `id` int(11) unsigned NOT NULL AUTO_INCREMENT, `chapter_id` int(11) unsigned, `poem_vi` text, PRIMARY KEY (`id`), FOREIGN KEY (chapter_id) REFERENCES chapters(id) ) ENGINE=InnoDB AUTO_INCREMENT=424 DEFAULT CHARSET=utf8;")

	chapterFiles, _ := ioutil.ReadDir("./data")
	for _, f := range chapterFiles {
		fmt.Println("Import : " + f.Name())
		importChapter(f.Name(), db)
	}
}
