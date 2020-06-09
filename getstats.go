// First attempt to get some data from OpenSim

package main

import (
	"database/sql"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

type SimpleRegion struct { 
	regionName string
	locX int
	locY int
}

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
	
	for rows.Next() {
			err = rows.Scan(
				&SimpleRegion.regionName,
				&SimpleRegion.locX,
				&SimpleRegion.locY,
				
			)
		// Log.Debug("Row extracted:", Object)
		rowArr = append(rowArr, SimpleRegion)
	}
	checkErr(err)
	defer rows.Close()	
	}
}