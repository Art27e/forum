package main

import (
	"database/sql"
	"forum/models"
	"forum/web"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
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

	rtr.HandleFunc("/", web.Index).Methods("GET")
	rtr.HandleFunc("/register", web.Register).Methods("GET")
	rtr.HandleFunc("/register_user", web.RegisterUser).Methods("POST")
	rtr.HandleFunc("/login", web.Login).Methods("GET")
	rtr.HandleFunc("/acc/login", web.AccLogin).Methods("POST")
	rtr.HandleFunc("/acc/logout", web.AccLogout).Methods("GET")
	rtr.HandleFunc("/acc/profile/{id:[0-9]+}", web.ShowProfile).Methods("GET")
	rtr.HandleFunc("/acc/profile/{id:[0-9]+}/all-posts", web.ShowUserAllPosts).Methods("GET")
	rtr.HandleFunc("/myprofile/{id:[0-9]+}", web.ShowMyProfile).Methods("GET")
	rtr.HandleFunc("/forum/{id:[0-9]+}", web.ShowMainThread).Methods("GET")
	rtr.HandleFunc("/thread/{id:[0-9]+}", web.ShowThread).Methods("GET")
	rtr.HandleFunc("/thread/create", web.CreateThreadPage).Methods("GET")
	rtr.HandleFunc("/create-thread", web.CreateThread).Methods("POST")
	rtr.HandleFunc("/post", web.AddPost).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}/edit", web.ModifyPost).Methods("GET")
	rtr.HandleFunc("/post/edit", web.ModifyPostButton).Methods("POST")
	rtr.HandleFunc("/secure", web.SecureHandler).Methods("get")
	rtr.HandleFunc("/like/{id:[0-9]+}", web.LikeIt).Methods("post")
	rtr.HandleFunc("/removelike/{id:[0-9]+}", web.RemoveLike).Methods("post")

	log.Println("Server is online")
	log.Println("Visit https://localhost:8080/")
	log.Println("Press CTRL+C to shutdown the server")

	http.Handle("/", rtr)
	http.ListenAndServeTLS(":8080",certFilePath,keyFilePath, nil)
}
