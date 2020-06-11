// First attempt to get some data from OpenSim

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"time"
)

// SimpleRegion is a very simple struct just to get a region's name and location.
// In the future, it might have extra fields for linking to the grid map.
type SimpleRegion struct {
	regionName string	`json:"regionName"`	// we'll JSONify this later
	locX int			`json:"locX"`
	locY int			`json:"locY"`
}

func testFunc(s string) string {
	return fmt.Sprintln("This is a string:", s)
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

	var (
		rowArr []interface{}
		simpleRegion SimpleRegion
	)
	
	for rows.Next() {
			err = rows.Scan(
				&simpleRegion.regionName,
				&simpleRegion.locX,
				&simpleRegion.locY,

			)
		// Log.Debug("Row extracted:", Object)
		rowArr = append(rowArr, simpleRegion)
	}
	checkErr(err)
	defer rows.Close()
	
	// Prepare a simple table. For now, we'll just print stuff out. This will probably get JSONified at some point and loaded in a 'real' table.
	regionsTable := fmt.Sprintf("<pre>%#v</pre>\n", rowArr)
	
	// Online users is TBD.
	usersOnline := "<pre>(nothing yet)</pre>\n"
	
	c.HTML(http.StatusOK, "welcome.tpl", gin.H{
			"now": formatAsYear(time.Now()),
			"regionsTable"	: regionsTable,
			"usersOnline"	: usersOnline,
	})
}