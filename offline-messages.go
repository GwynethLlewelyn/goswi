package main

import (
	"database/sql"
//	"fmt"
	"log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type offlineIM struct {
	ID string			`json:"ID"`
	PrincipalID string	`json:"PrincipalID"`
	FromID string		`json:"FromID"`
	Message string		`json:"Message"`
	TMStamp string		`json:"TMStamp"`
}

const MaxNumberMessages int = 5	// maximum number of messages to retrieve

// GetOfflineMessages will retrieve the top first 5 messages and put it on the session, to avoid constant reloading
func GetOfflineMessages(c *gin.Context) {
	session		:= sessions.Default(c)
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
	return
}