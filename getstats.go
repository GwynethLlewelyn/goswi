// First attempt to get some data from OpenSim

package main

import (
	"database/sql"
// 	"encoding/json"
//	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
	"net/http"
//	"strings"
	"time"
)

// We need to pass JSON to templates, because it won't work otherwise.
// var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SimpleRegion is a very simple struct just to get a region's name and location.
// In the future, it might have extra fields for linking to the grid map.
type SimpleRegion struct {
	RegionName string	`form:"regionName" json:"regionName"`	// we'll JSONify this later
	LocX int			`form:"locX" json:"locX,string"`
	LocY int			`form:"locY" json:"locY,string"`
}

// Apparently this is what we get with /welcome — some information from the viewer! (gwyneth 20200612)
type Viewer struct {
	ViewerName	string `form:"channel" json:"channel"`
	Grid		string `form:"grid" json:"grid"`
	Language	string `form:"lang" json:"lang"`
	LoginContentVersion	string `form:"login_content_version" json:"login_content_version"`
	OS			string `form:"os" json:"os"`
	SourceID	string `form:"sourceid" json:"sourceid"`
	Version		string `form:"version" json:"version"`
}

type SimpleUser struct {
	AvatarName	string `form:"Avatar Name" json:"Avatar Name"`
}


// GetStats will be used on the Welcome template (and possibly elsewhere) to display some in-world stats.
func GetStats(c *gin.Context) {
	// Declare some variables used to JSONify everything. (gwyneth 20200612)
	var (
		oneViewer Viewer
		viewerInfo []Viewer
		simpleRegion SimpleRegion
		regionsTable []SimpleRegion
		userTable []SimpleUser
		err error
	)
		
	// TODO(gwyneth): deal with *empty* channel=Firestorm-Releasex64&grid=btgrid&lang=en&login_content_version=2&os=Mac%20OS%20X%2010.15.6&sourceid=&version=6.3.9%20%2858205%29"
	if c.Bind(&oneViewer) == nil { // nil means no errors
		if oneViewer.ViewerName != "" {	// apparently, it binds even if there is nothing to bind to; so we check this first before appending to the table; it means the table will be nil, and commented out on the template (gwyneth 20200616)
			viewerInfo = append(viewerInfo, oneViewer)
		}
	} else {
		checkErr(err)
	}
	log.Println("[DEBUG] Data from viewer:", viewerInfo)
	
	// open database connection
	if *DSN == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *DSN) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	rows, err := db.Query("SELECT regionName, locX, locY FROM regions ORDER BY regionName ASC")
	checkErr(err)

	defer rows.Close()
		
	for rows.Next() {
		err = rows.Scan(
			&simpleRegion.RegionName,
			&simpleRegion.LocX,
			&simpleRegion.LocY,
		)
		simpleRegion.LocX /= 256
		simpleRegion.LocY /= 256
		regionsTable = append(regionsTable, simpleRegion)
	}
	checkErr(err)	
//	log.Println("[DEBUG] Data from regionsTable:", regionsTable)

	rows, err = db.Query("SELECT PrincipalID, FirstName, LastName FROM UserAccounts WHERE PrincipalID IN (SELECT UserID FROM Presence)")
	checkErr(err)

	var principalID, firstName, lastName string	// temporary to get replies from the database
	
	for rows.Next() {
		err = rows.Scan(&principalID, &firstName, &lastName)
		userTable = append(userTable, SimpleUser{AvatarName: firstName + " " + lastName})
	}
	checkErr(err)	
	log.Println("[DEBUG] Data from userTable:", userTable)
	
	c.HTML(http.StatusOK, "welcome.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"needsTables"	: true,
			"needsMap"		: true,
			"author"		: author,
			"description"	: description,
			"viewerInfo"	: viewerInfo,
			"regionsTable"	: regionsTable,
			"usersOnline"	: userTable,
			"Debug"			: false,
	})
}