package main

import (
	"database/sql"
//	 "encoding/json"
	"fmt"
//	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
	"net/http"
//	"strings"
	"time"
)

type UserProfile struct {
	UserUUID string `form:"useruuid" json:"useruuid"`
	ProfilePartner string
	ProfileAllowPublish bool
	ProfileMaturePublish string
	ProfileURL string
	ProfileWantToMask int
	ProfileWantToText string
	ProfileSkillsMask int
	ProfileSkillsText string
	ProfileLanguages string
	ProfileImage string
	ProfileAboutText string
	ProfileFirstImage string
	ProfileFirstText string
}

// GetProfile connects to the database, does its magic, and spews out a profile. That's the theory at least.
func GetProfile(c *gin.Context) {
	// open database connection
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var profileData UserProfile
	var uuid, username string

	username	= c.MustGet("Username").(string)
	uuid	 	= c.MustGet("UUID").(string)
	err = db.QueryRow("SELECT * FROM userprofile WHERE useruuid = ?", uuid).Scan(&profileData)
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG]: user %q (%s) has no profile", username, uuid)
		}
	}
	c.HTML(http.StatusOK, "profile.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"needsTables"	: false,
		"needsMap"		: false,
		"author"		: *config["author"],
		"description"	: *config["description"],
		"Debug"			: false,	// we will probably need two versions of 'debug mode'... (gwyneth 20200622)
		"titleCommon"	: *config["titleCommon"] + uuid + " Profile",
		"ProfileData"	: fmt.Sprintf("%+v", profileData),
		"Username"		: username,
		"Libravatar"	: "bAavatar",
	})
}