package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"fmt"
	"database/sql"
	"github.com/NYTimes/gziphandler"
)
import (
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"github.com/joho/godotenv"
	"os"
	"github.com/google/jsonapi"
	"github.com/nguyenvanduocit/kinhphapcu-service/midleware"
	"golang.org/x/net/http2"
)

type Item struct {
	Id int `jsonapi:"primary,items"`
	PoemVi string `jsonapi:"attr,poem_vi,omitempty"`
	YoutubeId string `jsonapi:"attr,youtube_id,omitempty"`
	Items *Chapter `jsonapi:"relation,chapter"`
}

type Chapter struct{
	Id int `jsonapi:"primary,chapters"`
	Slug string `jsonapi:"attr,slug"`
	Name string `jsonapi:"attr,name"`
	Items []*Item `jsonapi:"relation,items,omitempty"`
}

type Server struct{
	dbScheme string
	address string
}

func NewServer(dbScheme string, address string)(*Server){
	return &Server{
		dbScheme,
		address,
	}
}

func (sv *Server)NewDatabaseConnect() (*sql.DB, error){
	connection, err := sql.Open("mysql", sv.dbScheme)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func (sv *Server)Listing(){
	fmt.Println("Server is listen on ", sv.address);
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/api/v1/data", sv.HandleGetData)
	router.HandleFunc("/", sv.HandlerIndex)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("view/static"))))

	handler := gziphandler.GzipHandler(router)
	handler = middleware.AddApiHeader(handler)

	srv := &http.Server{
		Addr:    sv.address,
		Handler: handler,
	}

	http2.ConfigureServer(srv, nil)
	log.Panic(srv.ListenAndServe())
}

func (sv *Server)HandlerIndex(w http.ResponseWriter, r *http.Request){
	templates, err := template.ParseFiles( "./view/index.html" );
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = templates.Execute(w,  nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (sv *Server)HandleGetData(w http.ResponseWriter, r *http.Request) {
	chapters, err := sv.getData()
	if err != nil{
		http.Error(w, err.Error(), 500)
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/vnd.api+json")

	if err := jsonapi.MarshalManyPayload(w, chapters); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err.Error())
	}
	dbScheme := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DATABASE_USERNAME"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"))
	sv := NewServer(dbScheme, os.Getenv("ADDRESS"));
	sv.Listing();
}

//Database functions

func (sv *Server)getData()([]*Chapter, error){
	db, err := sv.NewDatabaseConnect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT c.`id`, c.`slug`,  c.`name` FROM `chapters` as c ORDER BY c.`id`")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chapters []*Chapter
	for rows.Next() {
		var chapter Chapter;
		if err := rows.Scan(&chapter.Id,&chapter.Slug, &chapter.Name); err != nil {
			return nil, err
		}
		chapter.Items, err = sv.getPostsByChapterId(chapter.Id)
		if err != nil{
			return nil, err
		}
		chapters = append(chapters, &chapter);
	}
	return chapters, nil

}

func (sv *Server)getPostsByChapterId(chapterId int)([]*Item, error){
	db, err := sv.NewDatabaseConnect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT c.`id`, c.`poem_vi`, c.`youtube_id` FROM `posts` as c")
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item;
		var youtubeId sql.NullString
		if err := rows.Scan(&item.Id,&item.PoemVi,&youtubeId); err != nil {
			return nil, err
		}
		item.YoutubeId = youtubeId.String
		items = append(items, &item);
	}
	return items, nil
}