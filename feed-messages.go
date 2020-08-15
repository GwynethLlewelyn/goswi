// These functions deal with the 'feed' table, which I have no idea if it's something standard or not.
// But I'm using it nevertheless, since the code is the same as FeedMessages, and at least I'll give some use to the
// notification area. (gwyneth 20200815)
package main

import (
	"database/sql"
	"encoding/gob"
//	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

type FeedMessage struct {
	PostParentID string	`json:"PostParentID"`
	PosterID string		`json:"PosterID"`	// UUID of poster. Feed messages are seen by everyone.
	PostID string		`json:"PostID"`		// primary key
	Username string		`json:"Username"`	// will be constructed by getting it from the UserAccounts table
	Libravatar string	`json:"Libravatar"`
	PostMarkup string	`json:"PostMarkup"`	// actual message. May contain HTML.
	Chronostamp string	`json:"Chronostamp"`
	Visibility int		`json:"Visibility"`	// Ignored on this implementation
	Comment int			`json:"Comment"`	// Ignored on this implementation
	Commentlock string	`json:"Commentlock"`	// possibly the UUID of the avatar locking this thread for commenting
	Editlock string		`json:"Editlock"`	// possibly the UUID of the avatar locking this message for editing
	Feedgroup string	`json:"Feedgroup"`
}

const MaxNumberFeedMessages int = 5	// maximum number of feed messages to retrieve

// For some very, very, very stupid reason, we need to register our message type (and probably others) when starting...
func init() {
	gob.RegisterName("listOfFeedMessages", []FeedMessage{})
}

// GetTopFeedMessages will retrieve the top first 5 feed messages and put it on the session, to avoid constant reloading
func GetTopFeedMessages(c *gin.Context) {
	session		:= sessions.Default(c)
	username	:= session.Get("Username")
	uuid		:= session.Get("UUID")

	if uuid == "" {
		log.Println("[WARN]: GetTopFeedMessages(): No UUID stored; messages for this user cannot get retrieved")
	}

	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"] + "?parseTime=true") // this will allow parsing MySQL timestamps into Time vars; see https://stackoverflow.com/a/46613451/1035977
	checkErrFatal(err)

	defer db.Close()

	// first count how many messages we have, we will need this later.
	// According to the Internet, current versions of MariaDB/MySQL are actually much faster doing _two_ queries, one just for counting rows, since it's allegedly optimised; in this case, we can simplify the whole query as well.

	var numberMessages int

	err = db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&numberMessages)
	checkErr(err)

	if numberMessages > 0 {
		rows, err := db.Query("SELECT PostParentID, PosterID, PostID, PostMarkup, Chronostamp, Visibility, Comment, Commentlock, Editlock, Feedgroup, FirstName, LastName, Email FROM feeds, UserAccounts WHERE UserAccounts.PrincipalID = PosterID ORDER BY Chronostamp ASC LIMIT ?", strconv.Itoa(MaxNumberFeedMessages))
		checkErr(err)

		defer rows.Close()

		var (
			oneMessage FeedMessage
			messages []FeedMessage
			firstName, lastName, email string
			messageTimeStamp sql.NullTime // sql.NullTime will match timestamps with NULLs without crashing; see https://stackoverflow.com/a/60293251/1035977
		)

		for /* i := 1; */ rows.Next() /* ; i++ */ {		// uncomment for special
			err = rows.Scan(
				&oneMessage.PostParentID,
				&oneMessage.PosterID,
				&oneMessage.PostID,
				&oneMessage.PostMarkup,
				&messageTimeStamp,
				&oneMessage.Visibility,
				&oneMessage.Comment,
				&oneMessage.Commentlock,
				&oneMessage.Editlock,
				&oneMessage.Feedgroup,
				&firstName,
				&lastName,
				&email,
			)
			oneMessage.Username = firstName + " " + lastName
			oneMessage.Libravatar = getLibravatar(email, oneMessage.Username, 60)
			// do something to the time
			if messageTimeStamp.Valid {
				oneMessage.Chronostamp = humanize.Time(messageTimeStamp.Time)
			} else {
				oneMessage.Chronostamp = ""
			}

			// if *config["ginMode"] == "debug" {
			// 	log.Printf("[DEBUG]: message # %d from user %q <%s> to %q is: %q\n", i, oneMessage.Username, email, username, oneMessage.Message)
			// }
			messages = append(messages, oneMessage)
		}
		checkErr(err)

		// if *config["ginMode"] == "debug" {
		// 	log.Printf("[DEBUG]: GetTopFeedMessages(): All messages for user %q: %+v\n", username, messages)
		// }

		session.Set("FeedMessages", messages)
		session.Set("numberFeedMessages", numberMessages)
	} else {	// no messages for this user
		session.Set("FeedMessages", nil)
		session.Set("numberFeedMessages", numberMessages)
	}
	if err := session.Save(); err != nil {
		log.Printf("[WARN]: GetTopFeedMessages(): Could not save messages to user %q on the session, error was: %q\n", username, err)
	}
}