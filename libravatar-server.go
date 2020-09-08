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
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("s", "80"))
	if size == 80 {
		size, err = strconv.Atoi(c.DefaultQuery("size", "80"))
	}
	if err != nil {
		log.Println("[WARN] Libravatar: size is not an integer, 80 assumed")
		size = 80
	}
	var defaultParam string = c.DefaultQuery("d", "")
	if (defaultParam == "") {
		defaultParam = c.DefaultQuery("default", "")
	}
	// create filename (it's horrible, but that's how both Gravatar + Libravatar work) (gwyneth 20200908)

	profileImageFilename := strings.TrimPrefix(c.Request.URL.RequestURI(), "/avatar/")

	if *config["ginMode"] == "debug" {
		log.Println("[DEBUG] PathToStaticFiles is", PathToStaticFiles, "and profileImageFilename is now", profileImageFilename)
	}
	// check if image exists on the diskv cache; code shares similarities with profile.go (gwyneth 20200908)
	profileImage := filepath.Join(/* PathToStaticFiles, */ *config["cache"], profileImageFilename)

	if imageCache.Has(profileImage) {
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG] Libravatar: returning file", profileImage)
		}
		// c.Header("Content-Transfer-Encoding", "binary")
		// c.Header("Content-Type", "image/png")
		// c.File(profileImage)

		// assemble path to static file on disk, because, path complications (gwyneth 20200908)
		pathToProfileImage := filepath.Join(PathToStaticFiles, profileImage)
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG] Libravatar: pathToProfileImage is now %q\n", pathToProfileImage)
		}

		if fileContent, err := ioutil.ReadFile(pathToProfileImage); err != nil {
			mime := mimetype.Detect(fileContent)
			if *config["ginMode"] == "debug" {
				log.Printf("[DEBUG] Libravatar: file %q for profileImage %q is about to be returned, MIME type is %q, file size is %d\n", pathToProfileImage, profileImage, mime.String(), len(fileContent))
			}
			c.Data(http.StatusOK, mime.String(), fileContent)	// note: mime.String() will return "application/octet-stream" if the image type was not detected
			return
		} else {
			c.String(http.StatusNotFound, fmt.Sprintf("Libravatar: File not found for received hash: %q; desired size is: %d and default param is %q\n", params.Hash, size, defaultParam))
			log.Printf("[ERROR] Libravatar: imageCache error; file %q is in hash table but %q is not on filesystem! Error was: %v\n",
				profileImage, pathToProfileImage, err)
			// this probably means that the imageCache is corrupted, e.g. it has keys for non-existing files
			return
		}
	} else {
		// Image not found in cache, let's get it from OpenSimulator!
		// Again, this is very similar to profile.go (gwyneth 20200908).
	}
	// If all else fails:

	c.String(http.StatusNotFound, fmt.Sprintf("Libravatar: File not in image cache but could not retrieve it anyway; received hash was %q; desired size was: %d and default param was %q\n", params.Hash, size, defaultParam))

}
