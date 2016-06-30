package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"encoding/json"
	"flag"
	"fmt"
	"database/sql"
	"github.com/NYTimes/gziphandler"
)
import (
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"strconv"
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

type Response struct{
	Success bool `json:"success"`
	Message string `json:"message"`
	Count int `json:"count"`
	Result interface{} `json:"result"`
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

	router.HandleFunc("/api/v1/posts", sv.GetPost)
	router.HandleFunc("/api/v1/post/{id:[0-9]+}", sv.GetPost)

	router.HandleFunc("/api/v1/chapters", sv.GetChapter)

	router.HandleFunc("/", sv.Index)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("view/static"))))
	gzipWrapper := gziphandler.GzipHandler(router)
	log.Fatal(http.ListenAndServe(address, gzipWrapper))
}

func (sv *Server)Stop(){
	sv.db.Close()
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


func (sv *Server)getChapters()([]*Chapter, error){
	statement, err := sv.db.Prepare("SELECT c.`id`, c.`slug`,  c.`name`, count(p.`id`) as total FROM `chapters` as c INNER JOIN `posts` AS p ON `c`.`id` = `p`.`chapter_id` GROUP By c.`id` ORDER BY c.`id`")
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
		if err := rows.Scan(&chapter.Id,&chapter.Slug, &chapter.Name, &chapter.Total); err != nil {
			return nil, err
		}
		chapters = append(chapters, &chapter);
	}
	return chapters, nil

}

func (sv *Server)getPosts(limit int, chapter int, random bool)([]*Item, error){
	extrasQuery := " "
	var queryArgs []interface{}
	if(chapter > 0){
		extrasQuery += " WHERE `posts`.`chapter_id` = ?"
		queryArgs = append(queryArgs, chapter)
	}
	if(random){
		extrasQuery += " ORDER BY RAND()"
	}
	if(limit > 0){
		extrasQuery += " LIMIT ?"
		queryArgs = append(queryArgs, limit)
	}
	statement, err := sv.db.Prepare("SELECT id, poem_vi FROM `posts` " + extrasQuery)
	if err != nil {
		return nil, err
	}
	defer statement.Close()

	rows, err := statement.Query(queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Item
	for rows.Next() {
		var post Item;
		err = rows.Scan(&post.Id,&post.PoemVi)
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
	chapter, _ := strconv.Atoi(r.FormValue("chapter"))
	random:= r.FormValue("random") == "true"
	posts, err := sv.getPosts(count, chapter, random)
	if (err != nil) {
		response.Message = err.Error()
	}else{
		response.Message= "Success"
		response.Success = true
		response.Result = posts
		response.Count = len(posts)
	}
	sv.SendResponse(w, r, response)
	return
}

func (sv *Server)GetChapter(w http.ResponseWriter, r *http.Request){
	response := &Response{
		Success:false,
		Message:"Unknown error!",
	}
	chapters, err := sv.getChapters()
	if (err != nil) {
		response.Message = err.Error()
	}else{
		response.Message= "Success"
		response.Success = true
		response.Result = chapters
		response.Count = len(chapters)
	}
	sv.SendResponse(w, r, response)
	return
}

func (sv *Server)SendResponse(w http.ResponseWriter, r *http.Request, response *Response) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
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
