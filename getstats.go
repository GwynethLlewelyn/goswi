// Getting stats from OpenSimulator — viewer version, online users, regions, and map

package main

import (
	"database/sql"
// 	"encoding/json"
	"fmt"
	"github.com/gin-contrib/location"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	jsoniter "github.com/json-iterator/go"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// We need to pass JSON to templates, because it won't work otherwise.
// var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SimpleRegion is a very simple struct just to get a region's name. location and size.
// In the future, it might have extra fields for linking to the grid map.
// Added sizeX/Y for 'official' statistics (gwyneth 20200816).
type SimpleRegion struct {
	RegionName string	`form:"regionName" json:"regionName"`	// we'll JSONify this later
	LocX int			`form:"locX" json:"locX,string"`
	LocY int			`form:"locY" json:"locY,string"`
	SizeX int			`form:"sizeX" json:"sizeX,string"`		// Note that the current gridmap does assume sizeX/Y == 256 (gwyneth 20200816)
	SizeY int			`form:"sizeY" json:"sizeY,string"`
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

	// Deal with what comes from the SL viewer, e.g. something like channel=Firestorm-Releasex64&grid=btgrid&lang=en&login_content_version=2&os=Mac%20OS%20X%2010.15.6&sourceid=&version=6.3.9%20%2858205%29"
	if c.Bind(&oneViewer) == nil { // nil means no errors
		if oneViewer.ViewerName != "" {	// apparently, it binds even if there is nothing to bind to; so we check this first before appending to the table; it means the table will be nil, and commented out on the template (gwyneth 20200616)
			viewerInfo = append(viewerInfo, oneViewer)
		}
	} else {
		checkErr(err)
	}
	if *config["ginMode"] == "debug" {
		log.Println("[DEBUG] Data from viewer:", viewerInfo)
	}

	// open database connection
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	rows, err := db.Query("SELECT regionName, locX, locY FROM regions ORDER BY regionName ASC")
	checkErr(err)

	defer rows.Close()

	var regionName template.HTML

	for rows.Next() {
		err = rows.Scan(
			&regionName,
			&simpleRegion.LocX,
			&simpleRegion.LocY,
		)
		simpleRegion.LocX /= 256
		simpleRegion.LocY /= 256
		simpleRegion.RegionName = fmt.Sprintf(`<a class="class-link text-secondary" href="secondlife://%s/127/127/24/" onclick="goInWorld('secondlife://%s/127/127/24/');">%s</a>`, regionName, regionName, regionName)
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
	if *config["ginMode"] == "debug" {
		log.Println("[DEBUG] Data from userTable:", userTable)
	}
	session := sessions.Default(c)

	c.HTML(http.StatusOK, "welcome.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"needsTables"	: true,
			"needsMap"		: true,
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"slideshow"		: slideshow,
			"viewerInfo"	: viewerInfo,
			"regionsTable"	: regionsTable,
			"usersOnline"	: userTable,
			"Debug"			: false,	// we will probably need two versions of 'debug mode'... (gwyneth 20200622)
			"titleCommon"	: *config["titleCommon"] + "Welcome!",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
	})
}

// Implementation of OpenSimulator statistics according to https://github.com/BillBlight/OS_Simple_Stats/blob/master/stats.php (gwyneth 20200816)
func OSSimpleStats(c *gin.Context) {
	var gStatus string = "ONLINE"
	var serverBuilder strings.Builder
	var server string
	_, err := serverBuilder.WriteString(*config["ROBUSTserver"])
	if err != nil {
		log.Panicf("[ERROR] OSSimpleStats(): Could not add ROBUSTserver string %q\n", *config["ROBUSTserver"])
	}
	i := strings.Index(serverBuilder.String(), "//")
	if i != -1 {
		server = (serverBuilder.String())[i+2:]
	} else {
		server = serverBuilder.String()
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] OSSimpleStats(): ROBUST server is at %q\n", server)
	}

	conn, err := net.Dial("tcp", server)
	// TODO(gwyneth): I'll probably put a timeout here somewhere (gwyneth 20200817).
	if err != nil {
		log.Printf("[ERROR] OSSimpleStats(): ROBUST server %q unavailable; error was: %q", server, err)
		gStatus = "OFFLINE"
	}
	conn.Close()

	// TODO(gwyneth): for the rest of the things, we will limit this to 1 query every X minutes, or else everything blows up; we might return a cached result (gwyneth 20200817).

	// open database connection
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"] + "?parseTime=true")
	checkErrFatal(err)

	defer db.Close()

	preshguser := 0
	checkErr(db.QueryRow("SELECT COUNT(*) FROM GridUser WHERE UserID LIKE '%htt%' AND BINARY Login > UNIX_TIMESTAMP(NOW()) - 2592000").Scan(&preshguser)) // 2592000 = 1 month

	nowonlinescounter := 0
	checkErr(db.QueryRow("SELECT COUNT(*) FROM Presence").Scan(&nowonlinescounter))

	pastmonth := 0
	checkErr(db.QueryRow("SELECT DISTINCT COUNT(*) FROM GridUser WHERE BINARY Logout > UNIX_TIMESTAMP(NOW()) - 2592000").Scan(&pastmonth))

	totalaccounts := 0
	checkErr(db.QueryRow("SELECT COUNT(*) FROM UserAccounts").Scan(&totalaccounts))

	totalregions := 0
	totalvarregions := 0
	totalsingleregions := 0
	totalsize := 0
	var simpleRegion SimpleRegion

	rows, err := db.Query("SELECT sizeX, sizeY FROM regions")
	checkErr(err)

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&simpleRegion.SizeX,
			&simpleRegion.SizeY,
		)
		totalregions++
		if simpleRegion.SizeX == 256 {
			totalsingleregions++
		} else {
			totalvarregions++
		}
		totalsize += simpleRegion.SizeX * simpleRegion.SizeY / 1000
	}
	checkErr(err)

	// now handle formats by type; e.g. .../stats?format=json replies with JSON
	var format string

	if c.Bind(&format) != nil { // nil means no errors
		checkErr(err)
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] OSSimpleStats(): Format for stats is: %v\n", format)
	}
	url := location.Get(c) // get info about hostname

	// Create object to send to templating system
	var arr = gin.H{
		"GridStatus"				: gStatus,
		"Online_Now"				: nowonlinescounter,
		"HG_Visitors_Last_30_Days"	: preshguser,
		"Local_Users_Last_30_Days"	: pastmonth,
		"Total_Active_Last_30_Days"	: pastmonth + preshguser,
		"Registered_Users"			: totalaccounts,
		"Regions"					: totalregions,
		"Var_Regions"				: totalvarregions,
		"Single_Regions"			: totalsingleregions,
		"Total_LandSize"			: totalsize,
		"Login_URL"					: *config["assetServer"],
		"Website"					: url.Scheme + "://" + url.Host,
		"Login_Screen"				: url.Scheme + "://" + url.Host + "/welcome",
	}

	switch format {
		case "json":
			c.JSON(http.StatusOK, arr)
		case "xml":
			c.XML(http.StatusOK, arr)
		case "yaml":
			c.YAML(http.StatusOK, arr)
		default:
			c.HTML(http.StatusOK, "stats.tpl", environment(c, arr))
	}
}

