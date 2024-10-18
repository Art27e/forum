package models

import (
	"database/sql"
	"time"
)

type User struct {
	Id       uint16
	Username string
	Password string
	Email    string
}

type Data struct {
	IsLoggedIn  bool
	ProfileShow bool
	Username    string
	UserId    uint16
	OthersId      uint16
	Group       string
	PermRights  int
	Admin       bool
	Threads     []Thread
	TotalPosts  uint16
	WarningMsg  bool
	WarningText string
	Users      []User
}

type ForumData struct {
	IsLoggedIn     bool
	Admin          bool
	UserId         uint16
	Messages       []PostMsg
	Topics         []SubThread
	MainThreadName string
	MainThreadId   uint16
	UserData       Data // нужно будет все исправить, первые два должны будут быть удалены и включены в структуру и все поменять в обработчиках
}

type Thread struct {
	Id                 uint16
	Title, Description string
	TopicCount         int
	PostsCount         int
	LastPost           int
	LastTopicTitle     string
	LastPostTime       time.Time
	LastTopicUser      string
	LastTopicUserId    uint16
	LastTopicId        uint16
	Admin              bool
	IsLoggedIn             bool
	UserId             uint16
}

type MainThread []struct {
	Id          uint16
	Title       string
	Description string
}

type SubThread struct {
	Id               uint16
	Title            string
	Author           string
	UserId           uint16
	MainThreadId     uint16
	NumReplies       int
	CreatedAt        time.Time
	LastPostAuthorId uint16
	LastPostAuthor   string
	LastPostTime     time.Time
}

type PostMsg struct {
	Id           uint16
	Author       string
	AuthorID     uint16
	Body         string
	ThreadId     uint16
	ThreadTitle  string
	CreatedAt    time.Time
	LastPostTime time.Time
	Admin        bool
	CheckHost    bool
	MainThreadId uint16
	Likes        uint16
}

type Stats struct {
	IsLoggedIn   bool
	Admin        bool
	UserId       uint16
	UserNickname string
	Messages     []PostMsg
}

type PostForStats struct {
	Id         uint16
	Body       string
	Author     string
	TopicId    uint16
	TopicTitle string
}

type EditMsg struct {
	IsLoggedIn bool
	Id         uint16
	Body       string
	UserId     uint16
	Admin      bool
	Errors     ErrorMessage
}

type ErrorMessage struct {
	WarningMessage bool
	PassCheckMsg   bool
	IsLoggedIn     bool
	WarningText    string
	FailureCheck   bool
	UserId         uint16
	UserExists     bool
}

// Methods, in further going to use more of them. Atm only to show amount of replies and topics for main threads
func (mainThreadf *Thread) NumTopics() (count int) {
	err := Db.QueryRow("SELECT count(*) FROM threads where mainthread_id = $1", mainThreadf.Id).Scan(&count)
	if err != nil {
		return
	}
	return
}

func (mainThreadf *Thread) NumReplies() (count int) {
	err := Db.QueryRow("SELECT count(*) FROM posts WHERE mainthread_id = $1", mainThreadf.Id).Scan(&count)
	if err != nil {
		return
	}
	return
}

var Db *sql.DB                // database we use
var IsLoggedIn bool           // shows, if we are logged in or logged out
var UserId uint16             // logged user ID, in new updates going to unite with Creator variable
var UserGroup string          // logged user group relation
var CheckAdminRights bool     // if user has admin rights
var Creator uint16            // logged user ID, in new updates going to unite with UserId variable
var LoggedUser string         // nickname of logged in user
var SessionToken string       // our session token
var LoginCheck bool           // just another login check boolean, probably will be removed in new versions
var CheckLoginFail bool       // for errors, warning messages related to login
var WarningMsg bool           // for errors, warning messages related to login
var WarningMsg2 bool          // for password check warnings
var SaveVars string           // to track post id, when we wanna edit it
var WarningMsgText string     // contains some text for error, probably will be replaced only with html template errors later
var FollowMainThreadId uint16 // to track main thread location, in new updates logics probably will be modified
var FollowThreadId uint16     // to track topic location, in new updates logics probably will be modified
var FollowPostId uint16       // to track post id for likes
var IfUserExists bool         // for registration, if user already exists
var WarningTxt string         // for password modify success text
