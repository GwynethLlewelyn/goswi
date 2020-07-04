package main

import (
	"database/sql"
//	 "encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
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
	UserUUID string 			`form:"useruuid" json:"useruuid"`
	ProfilePartner string		`form:"profilePartner" json:"profilePartner"`
	ProfileAllowPublish bool	`form:"profileAllowPublish" json:"profileAllowPublish"`
	ProfileMaturePublish bool	`form:"profileMaturePublish" json:"profileMaturePublish"`
	ProfileURL string			`form:"profileURL" json:"profileURL"`
	ProfileWantToMask int		`form:"profileWantToMask" json:"profileWantToMask"`
	ProfileWantToText string	`form:"profileWantToText" json:"profileWantToText"`
	ProfileSkillsMask int		`form:"profileSkillsMask" json:"profileSkillsMask"`
	ProfileSkillsText string	`form:"profileSkillsText" json:"profileSkillsText"`
	ProfileLanguages string		`form:"profileLanguages" json:"profileLanguages"`
	ProfileImage string			`form:"profileImage" json:"profileImage"`
	ProfileAboutText string		`form:"profileAboutText" json:"profileAboutText"`
	ProfileFirstImage string	`form:"profileFirstImage" json:"profileFirstImage"`
	ProfileFirstText string		`form:"profileFirstText" json:"profileFirstText"`
}

// GetProfile connects to the database, does its magic, and spews out a profile. That's the theory at least.
func GetProfile(c *gin.Context) {
	session		:= sessions.Default(c)
	username	:= session.Get("Username")
	libravatar	:= session.Get("Libravatar")
	uuid		:= session.Get("UUID")

	// open database connection
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var (
		profileData UserProfile
		allowPublish, maturePublish string // it has to be this way to get around a bug in the mySQL driver which is impossible to fix
	)
	err = db.QueryRow("SELECT useruuid, profilePartner, profileAllowPublish, profileMaturePublish, profileURL, profileWantToMask, profileWantToText, profileSkillsMask, profileSkillsText, profileLanguages, profileImage, profileAboutText, profileFirstImage, profileFirstText FROM userprofile WHERE useruuid = ?", uuid).Scan(
			&profileData.UserUUID,
			&profileData.ProfilePartner,
			&allowPublish,
			&maturePublish,
			&profileData.ProfileURL,
			&profileData.ProfileWantToMask,
			&profileData.ProfileWantToText,
			&profileData.ProfileSkillsMask,
			&profileData.ProfileSkillsText,
			&profileData.ProfileLanguages,
			&profileData.ProfileImage,
			&profileData.ProfileAboutText,
			&profileData.ProfileFirstImage,
			&profileData.ProfileFirstText,
		)
		profileData.ProfileAllowPublish		= (allowPublish != "")
		profileData.ProfileMaturePublish	= (maturePublish != "")
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG]: user %q (%s) has no profile; database error was %v", username, uuid, err)
		}
	}
	c.HTML(http.StatusOK, "profile.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"needsTables"	: false,
		"needsMap"		: false,
		"author"		: *config["author"],
		"description"	: *config["description"],
		"Debug"			: false,	// we will probably need two versions of 'debug mode'... (gwyneth 20200622)
		"titleCommon"	: *config["titleCommon"] + profileData.UserUUID + " Profile",
		"ProfileData"	: fmt.Sprintf("%+v", profileData),
		"ProfileURL"	: profileData.ProfileURL, // TODO(gwyneth): This ought to be sanitized!!
		"ProfileImage"	: profileData.ProfileImage,
		"Username"		: username,
		"Libravatar"	: libravatar,
	})
}