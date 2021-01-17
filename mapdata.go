package main

import (
	"database/sql"
//	"fmt"
	"github.com/gin-gonic/gin"
//	. "github.com/siongui/godom"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

type Map struct {
	Grid []aRegion		`xml:"Grid"`
}

type aRegion struct {
	Uuid string			`xml:"Uuid"`
	RegionName string	`xml:"RegionName"`
	LocX int			`xml:"LocX"`
	LocY int			`xml:"LocY"`
	SizeX int			`xml:"SizeX"`
	SizeY int			`xml:"SizeY"`
}

// GetMapaData is the Go equivalent of data/map.php from https://github.com/hawddamor/opensimmaps.
func GetMapData(c *gin.Context) {
/*
	// Original PHP Code is here
	<?
	include("../../../settings/config.php");
	include("../../../settings/mysql.php");

	//Creates XML string and XML document using the DOM
	$dom = new DomDocument('1.0', "UTF-8");

	$map = $dom->appendChild($dom->createElement('Map'));

	$DbLink = new DB;
	$DbLink->query("SELECT uuid,regionName,locX,locY,sizeX,sizeY FROM ".C_REGIONS_TBL);
		while(list($UUID,$regionName,$locX,$locY,$dbsizeX,$dbsizeY) = $DbLink->next_record())
		{
			$grid = $map->appendChild($dom->createElement('Grid'));

			$uuid = $grid->appendChild($dom->createElement('Uuid'));
			$uuid->appendChild($dom->createTextNode($UUID));

			$region = $grid->appendChild($dom->createElement('RegionName'));
			$region->appendChild($dom->createTextNode($regionName));

			$locationX = $grid->appendChild($dom->createElement('LocX'));
			$locationX->appendChild($dom->createTextNode($locX/256));

			$locationY = $grid->appendChild($dom->createElement('LocY'));
			$locationY->appendChild($dom->createTextNode($locY/256));

	        $sizeX = $grid->appendChild($dom->createElement('SizeX'));
	        $sizeX->appendChild($dom->createTextNode($dbsizeX));

	        $sizeY = $grid->appendChild($dom->createElement('SizeY'));
	        $sizeY->appendChild($dom->createTextNode($dbsizeY));
		}

	$dom->formatOutput = true; // set the formatOutput attribute of
	                            // domDocument to true
	// save XML as string or file
	$test1 = $dom->saveXML(); // put string in test1
	//echo $test1;
	header("Content-type: text/xml");
	echo $test1;

	?>
*/
	// open database connection
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var (
		oneMap Map
		oneRegion aRegion
	)

	rows, err := db.Query("SELECT uuid, regionName, locX, locY, sizeX, sizeY FROM regions")
	checkErr(err)

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&oneRegion.Uuid,
			&oneRegion.RegionName,
			&oneRegion.LocX,
			&oneRegion.LocY,
			&oneRegion.SizeX,
			&oneRegion.SizeY,
		)
		oneRegion.LocX /= 256
		oneRegion.LocY /= 256
		oneMap.Grid = append(oneMap.Grid, oneRegion)
	}
	checkErr(err)
//	log.Println("[DEBUG] XML response from mapdata.go:", oneMap)

	c.Header("Access-Control-Allow-Origin", "*")	// because of CORS
	// TODO(gwyneth): also see https://github.com/gin-contrib/cors as a much better solution.
	c.XML(http.StatusOK, oneMap)
}