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
	"net/http"
//	"strings"
)

// LibravatarParams is required by the ShouldBindURI() call, it cannot be a simple string for some reason...
type LibravatarParams struct {
	Hash string `uri:"hash"`
}

// Libravatar is the handle, being called with ../avatar/<hash>?s=<size>&otherparameters .
func Libravatar(c *gin.Context) {
	var params LibravatarParams // see what parameters we have here; at least we ought to get the hash...

	if err := c.ShouldBindUri(&params); err != nil {
		c.String(http.StatusInternalServerError, "Libravatar: Cannot bind to hash parameters")
	}
	c.String(http.StatusNotImplemented, fmt.Sprintf("Libravatar: Not implemented, but received hash: %q\n", params.Hash))
}
