// First attempt to get some data from OpenSim

package main

import (
	"database/sql"
//	"encoding/json"
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
//var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SimpleRegion is a very simple struct just to get a region's name and location.
// In the future, it might have extra fields for linking to the grid map.
type SimpleRegion struct {
	regionName string	`json:"regionName"`	// we'll JSONify this later
	locX int			`json:"locX"`
	locY int			`json:"locY"`
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
	avatarName	string `json:"Avatar Name"`
}


// GetStats will be used on the Welcome template (and possibly elsewhere) to display some in-world stats.
func GetStats(c *gin.Context) {
	// Declare some variables used to JSONify everything. (gwyneth 20200612)
	var (
		oneViewer Viewer
		viewerTable []Viewer
		simpleRegion SimpleRegion
		regionsTable []SimpleRegion
		oneOnlineUser SimpleUser
		userTable []SimpleUser
		err error
//		viewerDataJSON, usersOnlineJSON, regionsTableJSON []byte
	)
		
	// TODO(gwyneth): deal with channel=Firestorm-Releasex64&grid=btgrid&lang=en&login_content_version=2&os=Mac%20OS%20X%2010.15.6&sourceid=&version=6.3.9%20%2858205%29"
	if c.Bind(&oneViewer) == nil { // nil means no errors
//		if viewerDataJSON, err = jsoniter.Marshal(viewer); err != nil {
//			checkErr(err)
//		}
		viewerTable = append(viewerTable, oneViewer)
	} else {
		checkErr(err)
	}
	log.Println("[DEBUG] Data from viewer:", viewerTable)
	
	// open database connection
	if *DSN == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *DSN) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	rows, err := db.Query("SELECT regionName, locX, locY FROM regions ORDER BY regionName ASC LIMIT 50")
	checkErr(err)

	defer rows.Close()
		
	for rows.Next() {
			err = rows.Scan(
				&simpleRegion.regionName,
				&simpleRegion.locX,
				&simpleRegion.locY,

			)
//		log.Println("[DEBUG] Row extracted:", simpleRegion)
		simpleRegion.locX /= 256
		simpleRegion.locY /= 256
		regionsTable = append(regionsTable, simpleRegion)
	}
	checkErr(err)
	defer rows.Close()

/*
	if regionsTableJSON, err = jsoniter.Marshal(regionsTable); err != nil {
		checkErr(err)
	}
	log.Printf("[DEBUG] Original data for regionsTable: >>%v<<\n", regionsTable)
*/
	log.Println("[DEBUG] Data from regionsTable:", regionsTable)

	
	// Online users is TBD.
//	usersOnline := [ ("Avatar Name"), ("Nobody IsOnline") ], [("Avatar Name"), ("Me Neither")] ]
//	var oneUserOnline = SimpleUser{avatarName: "Nobody IsOnline"}
/*
	if usersOnlineJSON, err = jsoniter.Marshal(oneUserOnline); err != nil {
		checkErr(err)
	}
*/


/*
	if usersOnlineJSON, err = json.Marshal(usersOnline); err != nil {
		checkErr(err)
	}
*/

	oneOnlineUser = SimpleUser{avatarName: "Nobody IsOnline"}
	userTable = append(userTable, oneOnlineUser)

	log.Println("[DEBUG] One online user:", oneOnlineUser, "; table for usersOnline:", userTable)
	
	c.HTML(http.StatusOK, "welcome.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"needsTables"	: true,
			"author"		: author,
			"description"	: description,
			"viewerData"	: viewerTable,
			"regionsTable"	: regionsTable,
			"usersOnline"	: userTable,
	})
}