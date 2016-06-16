package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"encoding/json"
	"io/ioutil"
	"flag"
	"fmt"
	"strings"
)

type Item struct{
	Id int `json:"id"`
	Poem_vi string `json:"poem_vi"`
}

func importChapter(fileName string, db *sql.DB){

	chapterFile, err := os.Open("data/" + fileName)
	if err != nil {
		panic(err.Error())
	}
	defer chapterFile.Close()
	var items []Item
	jsonParser := json.NewDecoder(chapterFile)
	err = jsonParser.Decode(&items);
	if err != nil {
		panic(err.Error())
	}

	chapterName := strings.Replace(fileName,".json","", -1)

	stmtIns, err := db.Prepare("INSERT INTO `posts` (id, chapter_slug, poem_vi) VALUES( ?, ?, ? )") // ? = placeholder
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(items); i++ {
		fmt.Println("Item:", items[i].Id)
		_, err = stmtIns.Exec(items[i].Id,chapterName, items[i].Poem_vi)
		if err != nil {
			panic(err.Error())
		}
	}
}

func createTable(db *sql.DB, name string,query string){
	_, err := db.Exec("DROP TABLE IF EXISTS `"+ name + "`")
	if err != nil {
		panic(err.Error())
	}
	_, err = db.Exec(query)
	if err != nil {
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

	createTable(db, "posts", "CREATE TABLE `posts` ( `id` int(11) unsigned NOT NULL AUTO_INCREMENT, `chapter_slug` varchar(255) DEFAULT NULL, `poem_vi` text, PRIMARY KEY (`id`) ) ENGINE=InnoDB AUTO_INCREMENT=424 DEFAULT CHARSET=utf8;")

	chapterFiles, _ := ioutil.ReadDir("./data")
	for _, f := range chapterFiles {
		fmt.Println("Import : " + f.Name())
		importChapter(f.Name(), db)
	}
}
