package main

import (
	"database/sql"
	"encoding/gob"
//	"fmt"
	"log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"strconv"
)

type OfflineIM struct {
	ID string			`json:"ID"`
	PrincipalID string	`json:"PrincipalID"`
	Username string		`json:"Username"`	// will be constructed by getting it from the UserAccounts table
	Libravatar string	`json:"Libravatar"`
	FromID string		`json:"FromID"`
	Message string		`json:"Message"`
	TMStamp string		`json:"TMStamp"`
}

const MaxNumberMessages int = 5	// maximum number of messages to retrieve

// For some very, very, very stupid reason, we need to register our message type (and probably others) when starting...
func init() {
	gob.RegisterName("listOfOfflineIMs", []OfflineIM{})
}

// GetTopOfflineMessages will retrieve the top first 5 messages and put it on the session, to avoid constant reloading
func GetTopOfflineMessages(c *gin.Context) {
	session		:= sessions.Default(c)
	username	:= session.Get("Username")
	uuid		:= session.Get("UUID")

	if uuid == "" {
		log.Println("[WARN]: GetTopOfflineMessages(): No UUID stored; messages for this user cannot get retrieved")
	}

	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	// first count how many messages we have, we will need this later.
	// According to the Internet, current versions of MariaDB/MySQL are actually much faster doing _two_ queries, one just for counting rows, since it's allegedly optimised; in this case, we can simplify the whole query as well.

	var numberMessages int

	err = db.QueryRow("SELECT COUNT(*) FROM im_offline WHERE im_offline.PrincipalID = ?", uuid).Scan(&numberMessages)
	checkErr(err)

	if numberMessages > 0 {
		rows, err := db.Query("SELECT ID, im_offline.PrincipalID, FromID, Message, TMStamp, FirstName, LastName, Email FROM im_offline, UserAccounts WHERE im_offline.PrincipalID = ? AND UserAccounts.PrincipalID = im_offline.FromID ORDER BY TMStamp ASC LIMIT ?", uuid, strconv.Itoa(MaxNumberMessages))
		checkErr(err)

		defer rows.Close()

		var (
			oneMessage OfflineIM
			messages []OfflineIM
			firstName, lastName, email string
		)

		for i := 1; rows.Next(); i++ {
			err = rows.Scan(
				&oneMessage.ID,
				&oneMessage.PrincipalID,
				&oneMessage.FromID,
				&oneMessage.Message,
				&oneMessage.TMStamp,
				&firstName,
				&lastName,
				&email,
			)
			oneMessage.Username = firstName + " " + lastName
			oneMessage.Libravatar = getAvatar(email, oneMessage.Username, 60)

			// if *config["ginMode"] == "debug" {
			// 	log.Printf("[DEBUG]: message # %d from user %q <%s> to %q is: %q\n", i, oneMessage.Username, email, username, oneMessage.Message)
			// }
			messages = append(messages, oneMessage)
		}
		checkErr(err)

		// if *config["ginMode"] == "debug" {
		// 	log.Printf("[DEBUG]: GetTopOfflineMessages(): All messages for user %q: %+v\n", username, messages)
		// }

		session.Set("Messages", messages)
		session.Set("numberMessages", numberMessages)
	} else {	// no messages for this user
		session.Set("Messages", nil)
		session.Set("numberMessages", numberMessages)
	}
	if err := session.Save(); err != nil {
		log.Printf("[WARN]: GetTopOfflineMessages(): Could not save messages to user %q on the session, error was: %q\n", username, err)
	}
}