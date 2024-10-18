package main

import (
	"database/sql"
	"forum/models"
	"forum/web"
	"github.com/gorilla/mux"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

func init() {
	var err error

	models.Db, err = sql.Open("sqlite", "db/forum.db")
	if err != nil {
		log.Println("Error connecting to the database:", err)
		return
	}
}

func main() {
	certFilePath := "localhost+1.pem"
	keyFilePath := "localhost+1-key.pem"

	rtr := mux.NewRouter()

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("./fonts/"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	rtr.HandleFunc("/", web.Index).Methods("GET")
	rtr.HandleFunc("/register", web.Register).Methods("GET")
	rtr.HandleFunc("/register_user", web.RegisterUser).Methods("POST")
	rtr.HandleFunc("/login", web.Login).Methods("GET")
	rtr.HandleFunc("/acc/login", web.AccLogin).Methods("POST")
	rtr.HandleFunc("/acc/logout", web.AccLogout).Methods("GET")
	rtr.HandleFunc("/acc/profile/{id:[0-9]+}", web.ShowProfile).Methods("GET")
	rtr.HandleFunc("/acc/profile/{id:[0-9]+}/all-posts", web.ShowUserAllPosts).Methods("GET")
	rtr.HandleFunc("/myprofile/{id:[0-9]+}", web.ShowMyProfile).Methods("GET")
	rtr.HandleFunc("/acc/change-pass", web.ModifyPassword).Methods("post")
	rtr.HandleFunc("/forum/{id:[0-9]+}", web.ShowMainThread).Methods("GET")
	rtr.HandleFunc("/thread/{id:[0-9]+}", web.ShowThread).Methods("GET")
	rtr.HandleFunc("/thread/{id:[0-9]+}/del", web.DeleteThread).Methods("POST")
	rtr.HandleFunc("/thread/create", web.CreateThreadPage).Methods("GET")
	rtr.HandleFunc("/mainforum/create", web.CreateMainForum).Methods("GET")
	rtr.HandleFunc("/mainforum/{id:[0-9]+}/modify", web.ModifyMThreadPage).Methods("GET")
	rtr.HandleFunc("/mainforum/{id:[0-9]+}/del", web.DeleteMThread).Methods("POST")
	rtr.HandleFunc("/create-thread", web.CreateThread).Methods("POST")
	rtr.HandleFunc("/mod-thread", web.ModMThread).Methods("POST")
	rtr.HandleFunc("/create-mainforum", web.CreateMForum).Methods("POST")
	rtr.HandleFunc("/thread/{id:[0-9]+}/edit", web.EditTopicPage).Methods("GET")
	rtr.HandleFunc("/edit-topic", web.EditTopic).Methods("POST")
	rtr.HandleFunc("/post", web.AddPost).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}/edit", web.ModifyPost).Methods("GET")
	rtr.HandleFunc("/post/edit", web.ModifyPostButton).Methods("POST")
	rtr.HandleFunc("/post/{{.Id}}/delete", web.DeletePost).Methods("POST")
	rtr.HandleFunc("/secure", web.SecureHandler).Methods("get")
	rtr.HandleFunc("/like/{id:[0-9]+}", web.LikeIt).Methods("post")
	rtr.HandleFunc("/removelike/{id:[0-9]+}", web.RemoveLike).Methods("post")

	rtr.HandleFunc("/admin", web.Admin).Methods("get")
	rtr.HandleFunc("/admin/change-user-group", web.ChangeUserGroup).Methods("post")
	rtr.HandleFunc("/admin/search", web.SearchUser).Methods("get")

	log.Println("Server is online")
	log.Println("Visit https://localhost:8080/")
	log.Println("Press CTRL+C to shutdown the server")

	http.Handle("/", rtr)
	http.ListenAndServeTLS(":8080", certFilePath, keyFilePath, nil)
}
