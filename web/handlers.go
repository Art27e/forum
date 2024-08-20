package web

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"forum/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var MyLogin bool
var SecureCheck2 bool

func Index(w http.ResponseWriter, r *http.Request) { // homepage
	t, err := template.ParseFiles("templates/header.html", "templates/index.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	// Reset any warning message
	models.WarningMsg = false

	// Here we create table for our main forum threads, we manually write them later. With moderation functions, it will be added in the website
	statement, err := models.Db.Prepare("CREATE TABLE IF NOT EXISTS main_threads (id INTEGER PRIMARY KEY, title TEXT, description TEXT)")
	if err != nil {
		log.Println("Error executing statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	statement.Exec()

	// Here we create table for likes at our website
	likeStatement, err := models.Db.Prepare("CREATE TABLE IF NOT EXISTS likes (id INTEGER PRIMARY KEY, user_id INTEGER REFERENCES users(id), post_id INTEGER REFERENCES posts(id))")
	if err != nil {
		log.Println("Error executing statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	likeStatement.Exec()

	// As we dont have admin functions atm, we gonna write our threads manually here
	mainThread := models.MainThread{
		{
			Title:       "Forum Development",
			Description: "Discuss bugs, updates, planned works and improvements",
		},
		{
			Title:       "Coding",
			Description: "Learn coding",
		},
		{
			Title:       "PC & Consoles and Video-games",
			Description: "Talk about PC, consoles, games",
		},
		{
			Title:       "Sports",
			Description: "Discuss sports here",
		},
	}

	// count here to check, if data already exists there. If no, then we insert values
	count := 0
	err = models.Db.QueryRow("SELECT COUNT(*) FROM main_threads").Scan(&count)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		insert, err := models.Db.Prepare("INSERT INTO main_threads (title, description) VALUES (?, ?)")
		if err != nil {
			log.Println("Error preparing insert statement:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		for _, q := range mainThread {
			_, err := insert.Exec(q.Title, q.Description)
			if err != nil {
				log.Println("Error executing insert statement:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		defer insert.Close()
	}
	// here we going to store main threads data
	mThreads := []models.Thread{}

	// We search results in our table
	res, err := models.Db.Query("SELECT * FROM main_threads")
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for res.Next() {
		var mainThreadf models.Thread
		err = res.Scan(&mainThreadf.Id, &mainThreadf.Title, &mainThreadf.Description)
		if err != nil {
			log.Println("Error scanning result:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Topic amount in certain main thread
		mainThreadf.TopicCount = mainThreadf.NumTopics()
		// Count posts amount in certain main thread
		mainThreadf.PostsCount = mainThreadf.NumReplies()
		// Finding thread id where was a last post
		lastpoststopicId := 0
		err = models.Db.QueryRow("SELECT thread_id FROM posts WHERE mainthread_id = $1 ORDER BY created_at DESC LIMIT 1", mainThreadf.Id).Scan(&lastpoststopicId)
		if err != nil {
			lastpoststopicId = -1 // if none posts, then we set variable to -1
		}
		mainThreadf.LastPost = lastpoststopicId

		var lastpostsTime time.Time
		var lastpostsUserId uint16
		var lastTopicIds uint16
		topicNameForLast := ""
		lastpostsUser := ""
		if lastpoststopicId != -1 {
			err = models.Db.QueryRow("SELECT title FROM threads WHERE id = ?", lastpoststopicId).Scan(&topicNameForLast)
			if err != nil {
				log.Println("Error querying database:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			err = models.Db.QueryRow("SELECT created_at FROM posts WHERE mainthread_id = $1 ORDER BY created_at DESC LIMIT 1", mainThreadf.Id).Scan(&lastpostsTime)
			if err != nil {
				lastpoststopicId = -1
			}
			err = models.Db.QueryRow("SELECT user_id FROM posts WHERE thread_id = ? ORDER BY created_at DESC LIMIT 1", lastpoststopicId).Scan(&lastpostsUserId)
			if err != nil {
				lastpoststopicId = -1
			}
			err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", lastpostsUserId).Scan(&lastpostsUser)
			if err != nil {
				lastpoststopicId = -1
			}
			err = models.Db.QueryRow("SELECT id FROM threads WHERE title = ?", topicNameForLast).Scan(&lastTopicIds)
			if err != nil {
				log.Println("Error querying database:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		mainThreadf.LastPostTime = lastpostsTime
		mainThreadf.LastTopicTitle = topicNameForLast
		mainThreadf.LastTopicUser = lastpostsUser
		mainThreadf.LastTopicUserId = lastpostsUserId
		mainThreadf.LastTopicId = lastTopicIds

		mThreads = append(mThreads, mainThreadf)
	}

	// Check if session exists
	_, err = r.Cookie("session" + models.LoggedUser)
	if err != nil {
		models.IsLoggedIn = false
		models.LoginCheck = false
	} else {
		models.IsLoggedIn = true
		models.LoginCheck = true
	}

	if models.LoggedUser != "" {
		models.LoginCheck = true
	} else {
		models.LoginCheck = false
	}

	ourData := models.Data{
		IsLoggedIn:  models.IsLoggedIn,
		ProfileShow: models.LoginCheck,
		Username:    models.LoggedUser,
		UserId:      models.UserId,
		Threads:     mThreads,
	}

	err = t.ExecuteTemplate(w, "index", ourData)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CreateThreadPage(w http.ResponseWriter, r *http.Request) { // create topic html page handler
	t, err := template.ParseFiles("templates/createthread.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	dataTransfer := models.ForumData{}
	dataTransfer.IsLoggedIn = models.IsLoggedIn

	if dataTransfer.IsLoggedIn {
		dataTransfer.UserId = models.UserId
	}

	err = t.ExecuteTemplate(w, "create_thread", dataTransfer)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CreateThread(w http.ResponseWriter, r *http.Request) { // post action
	followedThreadId := models.FollowMainThreadId
	topicName := r.FormValue("topic")
	topicMsg := r.FormValue("msg-create")

	createdAt, err := GetCurrentTime()
	if err != nil {
		log.Println("Error converting time", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Who creates new topic
	err = models.Db.QueryRow("SELECT id FROM users WHERE username = ?", models.LoggedUser).Scan(&models.Creator)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username", http.StatusUnauthorized)
			return
		}
	}
	// Install new data
	// topic
	insert, err := models.Db.Prepare("INSERT INTO threads (title, user_id, mainthread_id, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Println("Error preparing insert statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	res, err := insert.Exec(topicName, models.Creator, followedThreadId, createdAt)
	if err != nil {
		log.Println("Error executing insert statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	threadIdeshka, err := res.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert ID:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer insert.Close()

	// post
	insertPost, err := models.Db.Prepare("INSERT INTO posts (body,user_id,thread_id,mainthread_id, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Error preparing insertPost statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// for redirect we find right main thread id
	MainThreadId := ""
	err = models.Db.QueryRow("SELECT mainthread_id FROM threads WHERE id = ?", threadIdeshka).Scan(&MainThreadId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid thread id", http.StatusUnauthorized)
			return
		}
	}
	insertPost.Exec(topicMsg, models.Creator, threadIdeshka, MainThreadId, createdAt)
	defer insertPost.Close()

	// redirect after creating new topic
	redirectLocation := "/thread/" + strconv.Itoa(int(threadIdeshka))
	http.Redirect(w, r, redirectLocation, http.StatusSeeOther)
}

func ShowThread(w http.ResponseWriter, r *http.Request) { // show topic with all posts
	t, err := template.ParseFiles("templates/thread.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	var followThreadId uint16
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	varsToStr, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("problem related to convertation", err)
	}

	followThreadId = uint16(varsToStr)
	models.FollowThreadId = followThreadId

	dataTransfer := models.ForumData{}   // the main data
	postCollection := []models.PostMsg{} // data stores posts

	dataTransfer.IsLoggedIn = models.IsLoggedIn

	// if login is active, then UserId = is me
	if dataTransfer.IsLoggedIn {
		dataTransfer.UserId = models.UserId
	}

	// SQL command to select a specific thread by ID
	query := "SELECT id, body, user_id, thread_id, mainthread_id, created_at FROM posts WHERE thread_id = ? ORDER BY created_at ASC;"

	// Execute the query with thread ID as a parameter
	rows, err := models.Db.Query(query, followThreadId)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	mainThreadIdeshka := 0

	for rows.Next() {
		var messageFromUser models.PostMsg
		var authorID int
		var threadID int
		var mainthreadID int
		var timeTime time.Time

		err = rows.Scan(&messageFromUser.Id, &messageFromUser.Body, &authorID, &threadID, &mainthreadID, &timeTime)
		if err != nil {
			log.Println("error 0", err)
			return
		}

		// last posts time
		var lastpostsTime time.Time
		err = models.Db.QueryRow("SELECT created_at FROM posts WHERE body = $1 ORDER BY created_at DESC LIMIT 1", messageFromUser.Body).Scan(&lastpostsTime)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		messageFromUser.LastPostTime = lastpostsTime

		// Get post's author
		var authorUsername string
		err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", authorID).Scan(&authorUsername)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		messageFromUser.Author = authorUsername
		// Get topic name
		var topicName string
		err = models.Db.QueryRow("SELECT title FROM threads WHERE id = ?", threadID).Scan(&topicName)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Get topic id
		var postId uint16
		err = models.Db.QueryRow("SELECT id FROM posts WHERE body = ?", messageFromUser.Body).Scan(&postId)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Total likes for certain post
		var likesTotal uint16
		err = models.Db.QueryRow("SELECT count(*) FROM likes where post_id = $1", postId).Scan(&likesTotal)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Creation date
		var findTime time.Time
		err = models.Db.QueryRow("SELECT created_at FROM threads WHERE id = ?", threadID).Scan(&findTime)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Here we replace newlines in our messages with <br> for correct display in html
		messageFromUser.Body = strings.ReplaceAll(messageFromUser.Body, "\r\n", "<br>")

		// write to data structure
		messageFromUser.ThreadTitle = topicName
		messageFromUser.CreatedAt = timeTime
		messageFromUser.ThreadId = uint16(threadID)
		messageFromUser.AuthorID = uint16(authorID)
		messageFromUser.Likes = likesTotal

		// open possibility to edit post only for 5 min since post was sent, check time range and author
		var checkPostHost bool
		if models.LoginCheck && messageFromUser.AuthorID == models.UserId && time.Since(messageFromUser.CreatedAt) <= 5*time.Minute {
			checkPostHost = true
		}
		messageFromUser.CheckHost = checkPostHost

		mainThreadIdeshka = mainthreadID
		// Posts are stored in slice of data structures
		postCollection = append(postCollection, messageFromUser)
	}
	dataTransfer.MainThreadId = uint16(mainThreadIdeshka)
	dataTransfer.Messages = append(dataTransfer.Messages, postCollection...)

	err = t.ExecuteTemplate(w, "thread", dataTransfer)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func ShowUserAllPosts(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/thread.html", "templates/header.html", "templates/footer.html", "templates/showuserposts.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)

	statsData := models.Stats{}
	postCollection := []models.PostMsg{} // data stores posts

	// SQL command to select posts by user
	query := "SELECT id, body, user_id, thread_id, mainthread_id, created_at FROM posts WHERE user_id = ? ORDER BY created_at ASC;"

	// Execute the query with thread ID as a parameter
	rows, err := models.Db.Query(query, vars["id"])
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var messageFromUser models.PostMsg
		var authorID uint16
		var threadID uint16
		var mainthreadID uint16
		var timeTime time.Time
		err := rows.Scan(&messageFromUser.Id, &messageFromUser.Body, &authorID, &threadID, &mainthreadID, &timeTime)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Get topic name
		var topicName string
		err = models.Db.QueryRow("SELECT title FROM threads WHERE id = ?", threadID).Scan(&topicName)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Here we replace newlines in our messages with <br> for correct display in html
		messageFromUser.ThreadId = threadID
		messageFromUser.ThreadTitle = topicName
		messageFromUser.Body = strings.ReplaceAll(messageFromUser.Body, "\r\n", "<br>")
		postCollection = append(postCollection, messageFromUser)

	}

	var userName string
	err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", vars["id"]).Scan(&userName)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	statsData.Messages = postCollection
	statsData.UserId = models.UserId
	statsData.UserNickname = userName
	statsData.IsLoggedIn = models.IsLoggedIn

	err = t.ExecuteTemplate(w, "show-user-all-posts", statsData)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func LikeIt(w http.ResponseWriter, r *http.Request) { // Like mechanism
	// After like complete, redirect us to the topic with our post
	threadIdeshka := models.FollowThreadId
	toStrIdOfpage := strconv.Itoa(int(threadIdeshka))
	redirectLocation := "/thread/" + toStrIdOfpage
	http.Redirect(w, r, redirectLocation, http.StatusSeeOther)

	var followPostId uint16
	vars := mux.Vars(r)

	varsToStr, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("problem related to convertation", err)
	}

	followPostId = uint16(varsToStr)
	models.FollowPostId = followPostId

	// Knowing LoggedUser, we find his ID
	err = models.Db.QueryRow("SELECT id FROM users WHERE username = ?", models.LoggedUser).Scan(&models.Creator)
	if err != nil {
		if err == sql.ErrNoRows {
			//http.Error(w, "Invalid username", http.StatusUnauthorized)
			return
		}
	}

	// Check if we didnt put like earlier
	var checkLike bool
	err = models.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)", models.FollowPostId, models.Creator).Scan(&checkLike)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !checkLike {
		// We add like to the table, connecting it with post we like, previously checking if it doesn't exist
		insertPost, err := models.Db.Prepare("INSERT INTO likes (user_id, post_id) VALUES (?, ?)")
		if err != nil {
			log.Println("Error preparing insertPost statement:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		insertPost.Exec(models.Creator, models.FollowPostId)
		defer insertPost.Close()
	}
}

func RemoveLike(w http.ResponseWriter, r *http.Request) { // Remove like
	// After like complete, redirect us to the topic with our post
	threadIdeshka := models.FollowThreadId
	toStrIdOfpage := strconv.Itoa(int(threadIdeshka))
	redirectLocation := "/thread/" + toStrIdOfpage
	http.Redirect(w, r, redirectLocation, http.StatusSeeOther)

	var followPostId uint16
	vars := mux.Vars(r)

	varsToStr, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("problem related to convertation", err)
	}

	followPostId = uint16(varsToStr)
	models.FollowPostId = followPostId

	// Knowing LoggedUser, we find his ID
	err = models.Db.QueryRow("SELECT id FROM users WHERE username = ?", models.LoggedUser).Scan(&models.Creator)
	if err != nil {
		if err == sql.ErrNoRows {
			//http.Error(w, "Invalid username", http.StatusUnauthorized)
			return
		}
	}

	// We remove our like
	removeLike, err := models.Db.Prepare("DELETE FROM likes WHERE post_id = ? AND user_id = ?")
	if err != nil {
		log.Println("Error preparing removeLike statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	removeLike.Exec(models.FollowPostId, models.Creator)
	defer removeLike.Close()
}

func ShowMainThread(w http.ResponseWriter, r *http.Request) { // Show list of topics
	t, err := template.ParseFiles("templates/threads.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	varsToInt, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("Error converting str -> int", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	models.FollowMainThreadId = uint16(varsToInt)

	// The main data
	dataTransfer := models.ForumData{}
	dataTransfer.IsLoggedIn = models.IsLoggedIn

	if dataTransfer.IsLoggedIn {
		dataTransfer.UserId = models.UserId
	}

	// Create table for threads
	statThreads, err := models.Db.Prepare("CREATE TABLE IF NOT EXISTS threads (id INTEGER PRIMARY KEY, title TEXT, user_id INTEGER REFERENCES users(id), mainthread_id INTEGER REFERENCES main_threads(id), created_at TIMESTAMP NOT NULL)")
	if err != nil {
		log.Println("Error preparing statThreads statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	statThreads.Exec()

	// Create table for posts
	statPosts, err := models.Db.Prepare("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body TEXT, user_id INTEGER REFERENCES users(id), thread_id INTEGER REFERENCES threads(id), mainthread_id INTEGER REFERENCES main_threads(id), created_at TIMESTAMP NOT NULL)")
	if err != nil {
		log.Println("Error preparing statPosts statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	statPosts.Exec()

	sThreads := []models.SubThread{}

	// This query for listing all topics ordered by last post
	query := `SELECT t.id, t.title, t.user_id, t.mainthread_id, t.created_at
	FROM threads t
	INNER JOIN (
		SELECT thread_id, MAX(created_at) AS max_created_at
		FROM posts
		GROUP BY thread_id
	) p ON t.id = p.thread_id
	WHERE t.mainthread_id = ?
	ORDER BY p.max_created_at DESC`

	// Execute the query with the main thread ID as a parameter
	rows, err := models.Db.Query(query, models.FollowMainThreadId)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var subThreadf models.SubThread
		var authorID int
		var creationDate time.Time

		err = rows.Scan(&subThreadf.Id, &subThreadf.Title, &authorID, &subThreadf.MainThreadId, &creationDate)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get username knowing id
		authorUsername := ""
		err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", authorID).Scan(&authorUsername)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		subThreadf.Author = authorUsername
		subThreadf.UserId = uint16(authorID)

		// We count total posts in a certain thread
		count := 0
		err = models.Db.QueryRow("SELECT count(*) FROM posts where thread_id = $1", subThreadf.Id).Scan(&count)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		subThreadf.NumReplies = count

		// Having id we get date of creation
		var findTime time.Time
		err = models.Db.QueryRow("SELECT created_at FROM threads WHERE id = ?", subThreadf.Id).Scan(&findTime)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		subThreadf.CreatedAt = findTime

		// Last post author id
		var lastpostauthorideshka uint16
		err = models.Db.QueryRow("SELECT user_id FROM posts WHERE thread_id = $1 ORDER BY created_at DESC LIMIT 1", subThreadf.Id).Scan(&lastpostauthorideshka)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		subThreadf.LastPostAuthorId = lastpostauthorideshka

		// last post author nickname
		lastpostsUser := ""
		err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", lastpostauthorideshka).Scan(&lastpostsUser)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		subThreadf.LastPostAuthor = lastpostsUser

		// last posts time
		var lastpostsTime time.Time
		err = models.Db.QueryRow("SELECT created_at FROM posts WHERE thread_id = $1 ORDER BY created_at DESC LIMIT 1", subThreadf.Id).Scan(&lastpostsTime)
		if err != nil {
			log.Println("Error querying database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		subThreadf.LastPostTime = lastpostsTime

		sThreads = append(sThreads, subThreadf)
	}
	// Get topic title
	mainthreadName := ""
	err = models.Db.QueryRow("SELECT title FROM main_threads WHERE id = ?", models.FollowMainThreadId).Scan(&mainthreadName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	dataTransfer.MainThreadName = mainthreadName

	// Here we save all our created topics
	dataTransfer.Topics = append(dataTransfer.Topics, sThreads...)

	err = t.ExecuteTemplate(w, "show_threads", dataTransfer)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func Register(w http.ResponseWriter, r *http.Request) { // handler. html page for registration

	t, err := template.ParseFiles("templates/register.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// our data includes login status check and booleans to activate registration failure error messages
	data := models.ErrorMessage{
		WarningMessage: models.WarningMsg,
		PassCheckMsg:   models.WarningMsg2,
		IsLoggedIn:     models.IsLoggedIn,
		UserExists:     models.IfUserExists,
	}

	// reset all errors and warning messages after page refresh
	models.IfUserExists = false
	models.WarningMsg = false
	models.WarningMsg2 = false

	err = t.ExecuteTemplate(w, "registration", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func RegisterUser(w http.ResponseWriter, r *http.Request) { // post action for register

	// If we dont have any table to store our users, we create it
	statement, err := models.Db.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, password TEXT, email TEXT)")
	if err != nil {
		log.Println("Error preparing statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	statement.Exec()

	// We check if nickname is already taken
	var existsUser bool
	err = models.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", r.FormValue("username")).Scan(&existsUser)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// check if email is already taken
	var existsEmail bool
	err = models.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", r.FormValue("email")).Scan(&existsEmail)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Email or nickname is taken, try again
	if existsUser || existsEmail {
		models.IfUserExists = true
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Allowed symbols check
	for _, letter := range r.FormValue("username") {
		// we going to disallow any symbols in nickname, but allow "_"
		if letter >= 30 && letter < 48 || letter >= 58 && letter < 65 || letter >= 91 && letter < 95 || letter == 96 || letter == 126 {
			models.WarningMsg = true // if disallowed symbol is found, then activate warning message on the website and redirect to register page to try again
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
	}

	// Password reqs check
	passWord := r.FormValue("password")

	checkUpper := false
	checkLower := false
	checkNum := false
	checkSymb := false
	checkWrongSymb := true

	for _, letter := range r.FormValue("password") {
		switch {
		case unicode.IsUpper(letter):
			checkUpper = true
		case unicode.IsLower(letter):
			checkLower = true
		case unicode.IsDigit(letter):
			checkNum = true
		case unicode.IsPunct(letter):
			checkSymb = true
		case unicode.IsSymbol(letter):
			checkWrongSymb = false
		}
	}

	finalCheck := checkLower && checkUpper && checkNum && checkSymb && checkWrongSymb
	if !finalCheck {
		models.WarningMsg2 = true // if reqs not met, redirect and try again
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Encryption
	h := sha256.Sum256([]byte(passWord))
	salty := "ora4vng3"
	salt := sha256.Sum256([]byte(salty))
	newPass := fmt.Sprintf("%x", h) + fmt.Sprintf("%x", salt)

	userName := r.FormValue("username")
	eMail := r.FormValue("email")

	// We insert our nickname, pass and email from inputs to sql table named users
	insert, err := models.Db.Prepare("INSERT INTO users (username, password, email) VALUES (?, ?, ?)")
	if err != nil {
		log.Println("Error preparing insert statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	insert.Exec(userName, newPass, eMail)
	defer insert.Close()

	// User data check, if registration went wrong somehow, it should return 500 error code webpage
	row, err := models.Db.Query("SELECT id, username, password, email FROM users")
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for row.Next() {
		row.Scan(&userName, newPass, eMail)
	}

	// After registration form, we are redirected to the main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Login(w http.ResponseWriter, r *http.Request) { // handler for html page login

	t, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// our data includes text messages for errors/warnings, boolean for login failures and login status checks
	data := models.ErrorMessage{
		IsLoggedIn:   models.IsLoggedIn,
		WarningText:  models.WarningMsgText,
		FailureCheck: models.CheckLoginFail,
	}

	// reset all errors and warning messages after page refresh
	models.CheckLoginFail = false
	models.WarningMsg = false

	err = t.ExecuteTemplate(w, "login", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func AccLogin(w http.ResponseWriter, r *http.Request) { // post action for login process

	// Data from our website input areas
	userName := r.FormValue("username")
	passWord := r.FormValue("password")

	// Getting ID from the username for our further operations, if not found then write messages on the website and try again
	var profileId uint16
	err := models.Db.QueryRow("SELECT id FROM users WHERE username = ?", userName).Scan(&profileId)
	if err != nil {
		defer http.Redirect(w, r, "/login", http.StatusSeeOther)
		models.CheckLoginFail = true
		models.WarningMsgText = "User not exist, wrong user or password"
		return
	}

	var profilek string
	err = models.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", userName).Scan(&profilek)
	if err != nil {
		defer http.Redirect(w, r, "/login", http.StatusSeeOther)
		models.CheckLoginFail = true
		models.WarningMsgText = "User not exist, wrong user or password"
		return
	}

	// Query for the password associated with the provided username, if not found then write messages on the website and try again
	passInDb := ""
	err = models.Db.QueryRow("SELECT password FROM users WHERE username = ?", userName).Scan(&passInDb)
	if err != nil {
		if err == sql.ErrNoRows { // было err
			defer http.Redirect(w, r, "/login", http.StatusSeeOther)
			models.CheckLoginFail = true
			models.WarningMsgText = "User not exist, wrong user or password"
			return
		}
	}

	h2 := sha256.Sum256([]byte(passWord))
	salty2 := "ora4vng3"
	salt2 := sha256.Sum256([]byte(salty2))
	newPass2 := fmt.Sprintf("%x", h2) + fmt.Sprintf("%x", salt2)

	// Logics, in case of successful login
	if passInDb == newPass2 {
		MyLogin = true
		models.LoggedUser = userName // track logged in nickname
		models.UserId = profileId    // track his profile id
		models.CheckLoginFail = false
		models.WarningMsgText = "You're already logged in"
		CreateSession(w, true) // activating the session
		http.Redirect(w, r, "/secure", http.StatusSeeOther)
	} else {
		models.CheckLoginFail = true
		models.WarningMsgText = "User not exist, wrong user or password"
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func SecureHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateSession(write http.ResponseWriter, login bool) {
	// We create cookie session for user here
	var err error

	models.SessionToken, err = GenerateSessionToken()
    if err != nil {
        log.Println("Error generating session token:", err)
        return
    }

	Сookie := &http.Cookie{
		Name:  "session" + models.LoggedUser,
		Value: models.SessionToken,
		Path:  "/",
		HttpOnly: true,
		Secure: true,
	}

	if login {
		Сookie.Expires = time.Now().Add(30 * time.Minute) // keep logged in for 30 min
	} else {
		Сookie.MaxAge = -1 // for logging out
	}

	http.SetCookie(write, Сookie)
}

func AccLogout(w http.ResponseWriter, r *http.Request) {

	CreateSession(w, false) // cookie session over
	models.IsLoggedIn = false
	models.UserId = 0      // reset profile id
	models.LoggedUser = "" // reset logged in user's nickname

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func ShowProfile(w http.ResponseWriter, r *http.Request) {
	// Profile pages are under development. Codes in showProfile and showMyProfile may be similar to each other.
	t, err := template.ParseFiles("templates/profile.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	// Logics is simple. We see some number in URL, it's actually our user id. We scan this ID to get nickname
	user := ""
	err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", vars["id"]).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username", http.StatusUnauthorized)
			return
		}
	}
	// total posts
	count := 0
	err = models.Db.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", vars["id"]).Scan(&count)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	toInt, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("Error converting str -> int", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if models.LoggedUser == user {
		models.LoginCheck = true
	} else {
		models.LoginCheck = false
	}

	ourData := models.Data{
		IsLoggedIn:  models.IsLoggedIn,
		ProfileShow: models.LoginCheck,
		Username:    user,
		UserId:      uint16(toInt),
		TotalPosts:  uint16(count),
	}

	vars["id"] = strconv.Itoa(int(ourData.UserId))

	err = t.ExecuteTemplate(w, "profile", ourData)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func ShowMyProfile(w http.ResponseWriter, r *http.Request) {
	// Profile pages are under development. Codes in showProfile and showMyProfile may be similar to each other.

	t, err := template.ParseFiles("templates/myprofile.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	// Who owns this ID
	user := ""
	err = models.Db.QueryRow("SELECT username FROM users WHERE id = ?", vars["id"]).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username", http.StatusUnauthorized)
			return
		}
	}

	toInt, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("Error converting str -> int", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if models.LoggedUser == user {
		models.LoginCheck = true
	} else {
		models.LoginCheck = false
	}

	ourData := models.Data{
		IsLoggedIn:  models.IsLoggedIn,
		ProfileShow: models.LoginCheck,
		Username:    user,
		UserId:      uint16(toInt),
	}

	vars["id"] = strconv.Itoa(int(ourData.UserId))

	err = t.ExecuteTemplate(w, "myprofile", ourData)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func ModifyPost(w http.ResponseWriter, r *http.Request) {
	// Here is a page for editing our post, if all conditions have been respected

	t, err := template.ParseFiles("templates/edit.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		log.Println("Error parsing template files:", err)
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	// if some case occurs, when we are on modify post page somehow without login. Probability of it is low, anyway we have to be redirected to the main page in that case
	if !models.IsLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// vars will be our message id from sql table in the further
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	// Data transmit, we use EditMsg structure
	var data models.EditMsg
	data.IsLoggedIn = models.IsLoggedIn
	data.UserId = models.UserId

	models.SaveVars = vars["id"] // to remember post id

	// Getting message body having ID from url. We should see message body in the textarea on our website, so we can easily edit the msg.
	msg := ""
	err = models.Db.QueryRow("SELECT body FROM posts WHERE id = ?", vars["id"]).Scan(&msg)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Body = msg

	err = t.ExecuteTemplate(w, "edit", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Oops, something went wrong", 404)
		return
	}
}

func ModifyPostButton(w http.ResponseWriter, r *http.Request) {
	// Here is POST action for editing our message
	if !models.IsLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// Modified post changes
	createdAt, err := GetCurrentTime()
	if err != nil {
		log.Println("Error converting time", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	bodyText := r.FormValue("edit-post-textarea")
	// we edit text body and date, leaving old post id
	modPost, err := models.Db.Prepare("UPDATE posts SET body = ?, created_at = ? WHERE id = ?")
	if err != nil {
		log.Println("Error preparing modPost statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	modPost.Exec(bodyText, createdAt, models.SaveVars)
	defer modPost.Close()

	// We have presaved SaveVars with our last visited forum id. After posting modified post, we will be redirected to the right topic
	var threadIdForRedirect uint16
	err = models.Db.QueryRow("SELECT thread_id FROM posts WHERE id = ?", models.SaveVars).Scan(&threadIdForRedirect)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/thread/"+strconv.Itoa(int(threadIdForRedirect)), http.StatusSeeOther)
}

func AddPost(w http.ResponseWriter, r *http.Request) {
	// In case if we want to add post being logged out - redirect to the main page
	if !models.IsLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	createdAt, err := GetCurrentTime()
	if err != nil {
		log.Println("Error converting time", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	threadIdeshka := models.FollowThreadId

	// Knowing LoggedUser, we find his ID
	err = models.Db.QueryRow("SELECT id FROM users WHERE username = ?", models.LoggedUser).Scan(&models.Creator)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username", http.StatusUnauthorized)
			return
		}
	}

	// Knowing thread_id (from FollowThreadId variable) we can find our mainthread id
	mainThreadId := ""
	err = models.Db.QueryRow("SELECT mainthread_id FROM threads WHERE id = ?", threadIdeshka).Scan(&mainThreadId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid thread id", http.StatusUnauthorized)
			return
		}
	}

	// Add Post. Insert new data

	bodyText := r.FormValue("msg")

	// We create a table with 6 columns (id, body, user_id, thread_id, mainthread_id, created_at)
	insertPost, err := models.Db.Prepare("INSERT INTO posts (body,user_id,thread_id,mainthread_id, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Error preparing insertPost statement:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Text is taken from our textarea "msg" and Creator is logged in user.
	insertPost.Exec(bodyText, models.Creator, threadIdeshka, mainThreadId, createdAt)
	defer insertPost.Close()

	// After complete redirect us to the topic with our post
	toStrIdOfpage := strconv.Itoa(int(threadIdeshka))
	redirectLocation := "/thread/" + toStrIdOfpage
	http.Redirect(w, r, redirectLocation, http.StatusSeeOther)
}
