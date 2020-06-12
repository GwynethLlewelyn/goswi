// First attempt to get some data from OpenSim

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	jsoniter "github.com/json-iterator/go"
	"log"
	"net/http"
	"strings"
	"time"
)

// We need to pass JSON to templates, because it won't work otherwise.
// var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SimpleRegion is a very simple struct just to get a region's name and location.
// In the future, it might have extra fields for linking to the grid map.
type SimpleRegion struct {
	regionName string	`json:"regionName"`	// we'll JSONify this later
	locX int			`json:"locX"`
	locY int			`json:"locY"`
}

// GetStats will be used on the Welcome template (and possibly elsewhere) to display some in-world stats.
func GetStats(c *gin.Context) {
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

	var simpleRegion SimpleRegion
	
	regionsTable := `"data": [`
	
	for rows.Next() {
			err = rows.Scan(
				&simpleRegion.regionName,
				&simpleRegion.locX,
				&simpleRegion.locY,

			)
		// Log.Debug("Row extracted:", Object)
		regionsTable += fmt.Sprintf(`{ "Region" : "%s", "locX" : "%d", "locY" : "%d"} ,`, 
								simpleRegion.regionName, simpleRegion.locX, simpleRegion.locY)
	}
	checkErr(err)
	defer rows.Close()
	regionsTable = strings.TrimSuffix(regionsTable, ",")
	regionsTable += "]"
	
	
	// Online users is TBD.
	usersOnline := `"data": [ { "Avatar Name": "--(not implemented yet)--" } ]`
	
	c.HTML(http.StatusOK, "welcome.tpl", gin.H{
			"now": formatAsYear(time.Now()),
			"needsTables"	: true,
			"author"		: author,
			"description"	: description,
			"jsCallDataTable"	: `$('#regionsTable').DataTable(` + regionsTable +
`	);
	$('#usersOnline').DataTable(` + usersOnline +
`	);`,
	})
}