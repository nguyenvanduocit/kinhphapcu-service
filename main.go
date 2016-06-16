package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"encoding/json"
	"flag"
	"fmt"
	"database/sql"
)
import (
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"strconv"
)

type Post struct {
	Id int `json:"id"`
	ChapterSlug string `json:"chapter_slug"`
	PoemVi string `json:"poem_vi"`
}

type Response struct{
	Success bool `json:"success"`
	Message string `json:"message"`
	Count int `json:"count"`
	Posts []*Post `json:"posts"`
}

type Server struct{
	db *sql.DB
	ip string
	port string
}

func NewServer(dbScheme string, ip string, port string)(*Server){
	db, err := sql.Open("mysql", dbScheme)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return &Server{
		db,
		ip,
		port,
	}
}

func (sv *Server)Listing(){

	address := fmt.Sprintf("%s:%s", sv.ip, sv.port)
	fmt.Println("Server is listen on ", address);
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/api/v1/post", sv.GetPost)
	router.HandleFunc("/api/v1/post/{id:[0-9]+}", sv.GetPost)

	router.HandleFunc("/", sv.Index)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("view/static"))))
	log.Fatal(http.ListenAndServe(address, router))
}

func (sv *Server)Index(w http.ResponseWriter, r *http.Request){
	templates, err := template.ParseFiles( "./view/index.html" );
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = templates.Execute(w,  nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (sv *Server)Stop(){
	sv.db.Close()
}

func (sv *Server)getPosts(limit int, chapter string, random bool)([]*Post, error){
	extrasQuery := " "
	var queryArgs []interface{}
	if(chapter != ""){
		extrasQuery += " WHERE `posts`.`chapter_slug` = ?"
		queryArgs = append(queryArgs, chapter)
	}
	if(random){
		extrasQuery += " ORDER BY RAND()"
	}
	if(limit > 0){
		extrasQuery += " LIMIT ?"
		queryArgs = append(queryArgs, limit)
	}
	statement, err := sv.db.Prepare("SELECT id, chapter_slug, poem_vi FROM `posts` " + extrasQuery)
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	rows, err := statement.Query(queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post;
		err = rows.Scan(&post.Id, &post.ChapterSlug,&post.PoemVi)
		fmt.Println(post.Id)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post);
	}
	return posts, nil

}

func (sv *Server)GetPost(w http.ResponseWriter, r *http.Request) {
	response := &Response{
		Success:false,
		Message:"Unknown error!",
	}
	count, _ := strconv.Atoi(r.FormValue("count"))
	chapter:= r.FormValue("chapter")
	random:= r.FormValue("random") == "true"
	posts, err := sv.getPosts(count, chapter, random)
	if (err != nil) {
		response.Message = err.Error()
	}
	response.Message= "Success"
	response.Success = true
	response.Posts = posts
	response.Count = len(posts)
	sv.SendResponse(w, r, response)
	return
}

func (sv *Server)SendResponse(w http.ResponseWriter, r *http.Request, response *Response) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
	return
}

func main() {
	var ip, port, dbScheme string
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip")
	flag.StringVar(&port, "port", "8181", "Port")
	flag.StringVar(&dbScheme, "db-scheme", "root:7facd974e4b@/sotaycuame", "Database scheme format username:password@/database_name")
	flag.Parse()
	sv := NewServer(dbScheme, ip, port);
	sv.Listing();
	defer sv.Stop();
}
