package web

import (
	"fmt"
	"forum/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

func Admin(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/admin.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	if !models.CheckAdminRights {
		http.NotFound(w, r)
		return
	}

	data := models.Data{}

	data.Admin = models.CheckAdminRights
	data.Group = models.UserGroup
	data.IsLoggedIn = models.LoginCheck

	// Fetch search query if provided
	query := r.URL.Query().Get("query")
	
	// Create the SQL pattern for searching
	searchPattern := "%" + query + "%"

	// Fetch matching users from the database
	rows, err := models.Db.Query(`SELECT id, username FROM users WHERE username LIKE ?`, searchPattern)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var id int
		var username string
		if err := rows.Scan(&id, &username); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		users = append(users, models.User{Id: uint16(id), Username: username})
	}
	data.Users = users

	err = t.ExecuteTemplate(w, "admin_panel", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CreateMainForum(w http.ResponseWriter, r *http.Request) { // main thread create for admin rights
	t, err := template.ParseFiles("templates/createmainthread.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	if !models.CheckAdminRights {
		http.NotFound(w, r)
		return
	}

	dataTransfer := models.ForumData{}
	dataTransfer.IsLoggedIn = models.IsLoggedIn
	dataTransfer.Admin = models.CheckAdminRights

	if dataTransfer.IsLoggedIn {
		dataTransfer.UserId = models.UserId
	}

	err = t.ExecuteTemplate(w, "create_mainthread", dataTransfer)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CreateMForum(w http.ResponseWriter, r *http.Request) { // main thread create for admin rights
	name := r.FormValue("forum")
	description := r.FormValue("description")

	insert, err := models.Db.Prepare("INSERT INTO main_threads (title, description) VALUES (?, ?)")
	if err != nil {
		log.Println("Error preparing insert statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	res, err := insert.Exec(name, description)
	if err != nil {
		log.Println("Error executing insert statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	forumID, err := res.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert ID:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer insert.Close()

	// redirect after creating new forum
	redirectTo := "/forum/" + strconv.Itoa(int(forumID))
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func ModifyMThreadPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/modifymainthread.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	urlId := vars["id"]
	models.SaveVars = urlId

	if !models.CheckAdminRights {
		http.NotFound(w, r)
		return
	}

	description := ""
	err = models.Db.QueryRow("SELECT description FROM main_threads WHERE id = ?", urlId).Scan(&description)
	if err != nil {
		fmt.Println("Позже сделать ошибку")
	}

	title := ""
	err = models.Db.QueryRow("SELECT title FROM main_threads WHERE id = ?", urlId).Scan(&title)
	if err != nil {
		fmt.Println("Позже сделать ошибку")
	}

	data := models.Thread{}

	data.Admin = models.CheckAdminRights
	data.IsLoggedIn = models.IsLoggedIn
	data.Title = title
	data.Description = description
	data.UserId = models.UserId

	err = t.ExecuteTemplate(w, "modify_mainthread", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ModMThread(w http.ResponseWriter, r *http.Request) {

	if !models.IsLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	modThread, err := models.Db.Prepare("UPDATE main_threads SET title = ?, description = ? WHERE id = ?")
	if err != nil {
		log.Println("Error preparing modPost statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	modThread.Exec(title, description, models.SaveVars)
	defer modThread.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	if !models.IsLoggedIn {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !models.CheckAdminRights {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	postId := r.FormValue("post-id")

	query := "DELETE FROM posts WHERE id = ?"
	_, err := models.Db.Exec(query, postId)
	if err != nil {
		fmt.Println("Error executing DELETE query:", err)
		return
	}

	http.Redirect(w, r, "/thread/"+strconv.Itoa(int(models.FollowThreadId)), http.StatusSeeOther)
}

func EditTopicPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/modifytopic.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	if !models.IsLoggedIn {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !models.CheckAdminRights {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	urlId := vars["id"]

	data := models.Thread{}

	topicTitle := ""
	err = models.Db.QueryRow("SELECT title FROM threads WHERE id = ?", urlId).Scan(&topicTitle)
	if err != nil {
		fmt.Println("Позже сделать ошибку")
	}

	data.Title = topicTitle
	models.SaveVars = urlId

	err = t.ExecuteTemplate(w, "modify_topic", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func EditTopic(w http.ResponseWriter, r *http.Request) {
	if !models.IsLoggedIn {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !models.CheckAdminRights {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("topic")

	update, err := models.Db.Prepare("UPDATE threads SET title = ? WHERE id = ?")
	if err != nil {
		log.Println("Error preparing modPost statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	update.Exec(title, models.SaveVars)
	defer update.Close()

	var threadIdForRedirect uint16
	err = models.Db.QueryRow("SELECT mainthread_id FROM threads WHERE id = ?", models.SaveVars).Scan(&threadIdForRedirect)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/forum/"+strconv.Itoa(int(threadIdForRedirect)), http.StatusSeeOther)
}

func SearchUser(w http.ResponseWriter, r *http.Request) {
	if !models.CheckAdminRights {
		http.NotFound(w, r)
	}

	// Get the selected username from the form
	username := r.FormValue("query")

	// Check if the username is not empty
	if username == "" {
		log.Println("No username provided")
		http.Error(w, "No username provided", http.StatusBadRequest)
		return
	}

	// Query the database for the user ID based on the username
	var id int
	err := models.Db.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	if err != nil {
		log.Println("Error fetching user ID:", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Redirect to the user's profile page
	http.Redirect(w, r, fmt.Sprintf("/acc/profile/%d", id), http.StatusFound)
}

func ChangeUserGroup(w http.ResponseWriter, r *http.Request) {
	if !models.CheckAdminRights {
		http.NotFound(w, r)
	}

	selectedGroup := r.FormValue("category")
	userId := r.FormValue("user_id")

	switch selectedGroup {
	case "admins":
		query := fmt.Sprintf(`UPDATE 'users' SET 'group' = 'admins' WHERE id IS (%s);`, userId)
		_, err := models.Db.Exec(query)
		if err != nil {
			panic(err)
		}
	case "vip":
		query := fmt.Sprintf(`UPDATE 'users' SET 'group' = 'V.I.P' WHERE id IS (%s);`, userId)
		_, err := models.Db.Exec(query)
		if err != nil {
			panic(err)
		}
	case "users":
		query := fmt.Sprintf(`UPDATE 'users' SET 'group' = 'users' WHERE id IS (%s);`, userId)
		_, err := models.Db.Exec(query)
		if err != nil {
			panic(err)
		}
	case "banned":
		query := fmt.Sprintf(`UPDATE 'users' SET 'group' = 'banned' WHERE id IS (%s);`, userId)
		_, err := models.Db.Exec(query)
		if err != nil {
			panic(err)
		}
	}

	http.Redirect(w, r, "/acc/profile/"+userId, http.StatusFound)
}

func DeleteMThread(w http.ResponseWriter, r *http.Request) {
	if !models.CheckAdminRights {
		http.NotFound(w, r)
	}

	forumId := r.FormValue("forum-id")

	query := "DELETE FROM main_threads WHERE id = ?"
	_, err := models.Db.Exec(query, forumId)
	if err != nil {
		fmt.Println("Error executing DELETE query:", err)
		return
	}

	query2 := "DELETE FROM threads WHERE mainthread_id = ?"
	_, err = models.Db.Exec(query2, forumId)
	if err != nil {
		fmt.Println("Error executing DELETE query:", err)
		return
	}

	query3 := "DELETE FROM posts WHERE mainthread_id = ?"
	_, err = models.Db.Exec(query3, forumId)
	if err != nil {
		fmt.Println("Error executing DELETE query:", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DeleteThread(w http.ResponseWriter, r *http.Request) {
	if !models.CheckAdminRights {
		http.NotFound(w, r)
	}

	vars := mux.Vars(r)
	urlId := vars["id"]
	models.SaveVars = urlId

	var threadIdForRedirect uint16
	err := models.Db.QueryRow("SELECT mainthread_id FROM threads WHERE id = ?", models.SaveVars).Scan(&threadIdForRedirect)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	topicId := r.FormValue("thread-id")

	query := "DELETE FROM threads WHERE id = ?"
	_, err = models.Db.Exec(query, topicId)
	if err != nil {
		fmt.Println("Error executing DELETE query:", err)
		return
	}

	query2 := "DELETE FROM posts WHERE thread_id = ?"
	_, err = models.Db.Exec(query2, topicId)
	if err != nil {
		fmt.Println("Error executing DELETE query:", err)
		return
	}

	http.Redirect(w, r, "/forum/"+strconv.Itoa(int(threadIdForRedirect)), http.StatusSeeOther)
}
