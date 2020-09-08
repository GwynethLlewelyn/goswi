// This implements a minimalistic Libravatar-compatible, federated server, which will run as part of gOSWI.
// It returns Profile images from OpenSim to be used as Libravatars.
// For this to work, DNS needs to be properly setup with:
//
//
// Partially based on Surrogator, written in PHP by Christian Weiske (cweiske@cweiske.de) and licensed under the AGPL v3

/* Lots of things to do here */
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
//	"strings"
)

// LibravatarParams is required by the ShouldBindURI() call, it cannot be a simple string for some reason...
type LibravatarParams struct {
	Hash string `uri:"hash"`
}

// Libravatar is the handle, being called with ../avatar/<hash>?s=<size>&d=<default> .
func Libravatar(c *gin.Context) {
	// start by parsing what we can
	var params LibravatarParams // see what parameters we have here; at least we ought to get the hash...

	if err := c.ShouldBindUri(&params); err != nil {
		// this is fatal!
		c.String(http.StatusInternalServerError, "Libravatar: Cannot bind to hash parameters")
		return // not necessary, I think
	}
	size, err := strconv.Atoi(c.DefaultQuery("s", "80"))
	if size == 80 {
		size, err = strconv.Atoi(c.DefaultQuery("size", "80"))
	}
	if err != nil {
		log.Println("[WARN] Libravatar: size is not an integer, 80 assumed")
	}
	var defaultParam string = c.DefaultQuery("d", "")
	if (defaultParam == "") {
		defaultParam = c.DefaultQuery("default", "")
	}
	c.String(http.StatusNotImplemented, fmt.Sprintf("Libravatar: Not implemented, but received hash: %q; desired size is: %d and default param is %q\n", params.Hash, size, defaultParam))
}
