package main

import (
	"database/sql"
//	"fmt"
	"log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"strconv"
)

type offlineIM struct {
	ID string			`json:"ID"`
	PrincipalID string	`json:"PrincipalID"`
	Username string		`json:"Username"`	// will be constructed by getting it from the UserAccounts table
	Libravatar string	`json:"Libravatar"`
	FromID string		`json:"FromID"`
	Message string		`json:"Message"`
	TMStamp string		`json:"TMStamp"`
}

const MaxNumberMessages int = 5	// maximum number of messages to retrieve

// GetOfflineMessages will retrieve the top first 5 messages and put it on the session, to avoid constant reloading
func GetOfflineMessages(c *gin.Context) {
	session		:= sessions.Default(c)
	username	:= session.Get("Username")
	uuid		:= session.Get("UUID")

	if uuid == "" {
		log.Println("No UUID stored; messages for this user cannot get retrieved")
	}

	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()
	rows, err := db.Query("SELECT ID, im_offline.PrincipalID, FromID, Message, TMStamp, FirstName, LastName, Email FROM im_offline, UserAccounts WHERE im_offline.PrincipalID = ? AND UserAccounts.PrincipalID = im_offline.FromID ORDER BY TMStamp ASC LIMIT ?", uuid, strconv.Itoa(MaxNumberMessages))
	checkErr(err)

	defer rows.Close()

	var (
		oneMessage offlineIM
		messages []offlineIM
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

		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG]: message # %d from user %q <%s> to %q is: %q\n", i, oneMessage.Username, email, username, oneMessage.Message)
		}
		messages = append(messages, oneMessage)
	}
	checkErr(err)

	// if *config["ginMode"] == "debug" {
	// 	log.Printf("[DEBUG]: All messages for user %q: %+v\n", username, messages)
	// }

	session.Set("Messages", messages)
	session.Save()
}