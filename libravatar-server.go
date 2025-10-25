// This implements a minimalistic Libravatar-compatible, federated server, which will run as part of gOSWI.
// It returns Profile images from OpenSim to be used as Libravatars.
// For this to work, DNS needs to be properly setup with:
//  (... mumble mumble needs further explanation... )
//
// Partially based on Surrogator, written in PHP by Christian Weiske (cweiske@cweiske.de) and licensed under the AGPL v3

/* Lots of things to do here */
package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	//	"gopkg.in/gographics/imagick.v3/imagick"	// not needed since we call the conversion function (gwyneth 20200910)
)

// Database fields that we will retrieve later for checking email, names, profile image, etc.
type LibravatarProfile struct {
	ProfileID, FirstName, LastName, Email, ProfileImage string
}

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
	//var size uint

	size, err := strconv.ParseUint(c.DefaultQuery("s", "80"), 10, 0)
	if size == 80 {
		size, err = strconv.ParseUint(c.DefaultQuery("size", "80"), 10, 0)
	}
	if err != nil {
		config.LogWarn("Libravatar: size is not an integer, 80 assumed")
		size = 80
	}
	var defaultParam = c.DefaultQuery("d", "")
	if defaultParam == "" {
		defaultParam = c.DefaultQuery("default", "")
	}

	// Create filename. (it's horrible, but that's how both Gravatar + Libravatar work) (gwyneth 20200908)
	// Note that we get this through Blue Monday's sanitiser, to make sure the requests are valid.
	profileImageFilename := bluemondaySafeHTML.Sanitize(strings.TrimPrefix(c.Request.URL.RequestURI(), "/avatar/"))

	config.LogDebug("PathToStaticFiles is", PathToStaticFiles, "and profileImageFilename is now", profileImageFilename)
	// check if image exists on the diskv cache; code shares similarities with profile.go (gwyneth 20200908)
	profileImage := filepath.Join( /* PathToStaticFiles, */ *config["cache"], profileImageFilename)

	if imageCache.Has(profileImage) {
		config.LogDebug("Libravatar: returning file", profileImage)
		// c.Header("Content-Transfer-Encoding", "binary")
		// c.Header("Content-Type", "image/png")
		// c.File(profileImage)

		// assemble path to static file on disk, because, path complications (gwyneth 20200908).
		pathToProfileImage := filepath.Clean(filepath.Join(PathToStaticFiles, profileImage))
		config.LogDebugf("Libravatar: pathToProfileImage is now %q\n", pathToProfileImage)

		if fileContent, err := os.ReadFile(pathToProfileImage); err == nil {
			mime := mimetype.Detect(fileContent)
			config.LogDebugf("Libravatar: file %q for profileImage %q is about to be returned, MIME type is %q, file size is %d\n", pathToProfileImage, profileImage, mime.String(), len(fileContent))
			c.Data(http.StatusOK, mime.String(), fileContent) // note: mime.String() will return "application/octet-stream" if the image type was not detected
			return
		} else {
			c.String(http.StatusNotFound, fmt.Sprintf("Libravatar: file not found for received hash: %q; desired size is: %d and default param is %q\n", params.Hash, size, defaultParam))
			config.LogErrorf("Libravatar: imageCache error; file %q is in hash table but %q is not on filesystem! Error was: %v\n",
				profileImage, pathToProfileImage, err)
			// this probably means that the imageCache is corrupted, e.g. it has keys for non-existing files
			return
		}
	} else {
		/*
			Image not found in cache, let's get it from OpenSimulator!
			Again, this is very similar to profile.go (gwyneth 20200908).
			The difference is that we need to do the following:
			- Check entries on the database
			- See if we get a match on the hash for the email address stored on the database (try MD5 or SHA256 depending on key size)
			- If not, check for a hash of (email) AvatarFirstName.AvatarLastName@<hostname> and/or (OpenID) <hostname>/AvatarFirstName.AvatarLastName
		*/
		var (
			hashType             = "MD5" // try this first
			oneLibravatarProfile LibravatarProfile
			//			username string				// not needed yet!
		)
		if len(params.Hash) > 32 {
			hashType = "SHA256"
		}

		// open database connection
		if *config["dsn"] == "" {
			log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
		}
		db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
		checkErrFatal(err)

		defer db.Close()

		err = db.QueryRow("SELECT PrincipalID, FirstName, LastName, Email, profileImage FROM UserAccounts, userprofile WHERE "+hashType+" (LOWER(Email)) = ? AND useruuid = PrincipalID AND profileImage <> '00000000-0000-0000-0000-000000000000'", params.Hash).Scan(
			&oneLibravatarProfile.ProfileID,
			&oneLibravatarProfile.FirstName,
			&oneLibravatarProfile.LastName,
			&oneLibravatarProfile.Email,
			&oneLibravatarProfile.ProfileImage,
		)

		if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
			config.LogDebugf("Libravatar: retrieving profile for hash %q failed; database error was %v\n", params.Hash, err)
			// no rows found, so we can assume that either the email is NULL or possibly there isn't a profileImage
			// First we will attempt to do some hashing on the 'fake' email:

		} else {
			// match found for email on database!
			config.LogDebugf("Libravatar: retrieving profile for hash %q: %+v\n", params.Hash, oneLibravatarProfile)

			// get the image from OpenSimulator!
			profileImageAssetURL := *config["assetServer"] + path.Join("/assets/", oneLibravatarProfile.ProfileImage, "/data")
			resp, err := http.Get(profileImageAssetURL)
			if err != nil {
				// handle error
				config.LogError("Libravatar: Oops — OpenSimulator cannot find", profileImageAssetURL, "error was:", err)
			}
			defer resp.Body.Close()
			newImage, err := io.ReadAll(resp.Body)
			if err != nil {
				config.LogError("Libravatar: Oops — could not get contents of", profileImageAssetURL, "from OpenSimulator, error was:", err)
			}
			if len(newImage) == 0 {
				config.LogError("Libravatar: Image retrieved from OpenSimulator", profileImageAssetURL, "has zero bytes.")
				// we might have to get out of here
			} else {
				config.LogDebug("Libravatar: Image retrieved from OpenSimulator", profileImageAssetURL, "has", len(newImage), "bytes.")
			}
			// Now use ImageMagick to convert this image!
			// Unlike what happened on GetProfile(), here we're ignoring the Retina version (gwyneth 20200910)
			convertedImage, _, err := ImageConvert(newImage, uint(size), uint(size), 100)
			if err != nil {
				config.LogError("Libravatar: Could not convert", profileImageAssetURL, " - error was:", err)
			}
			if len(convertedImage) == 0 {
				config.LogError("Libravatar: Converted image is empty")
			}
			config.LogDebug("Libravatar: Regular image from", profileImageAssetURL, "has", len(convertedImage), "bytes.")
			// put it into KV cache:
			if err := imageCache.Write(profileImage, convertedImage); err != nil {
				config.LogError("Libravatar: Could not store converted", profileImage, "in the cache, error was:", err)
			}

			mime := mimetype.Detect(convertedImage)
			config.LogDebugf("Libravatar: file for profileImage %q is about to be returned, MIME type is %q, file size is %d\n", profileImage, mime.String(), len(convertedImage))
			c.Data(http.StatusOK, mime.String(), convertedImage)
			return
		}
	}
	c.String(http.StatusNotFound, fmt.Sprintf("Libravatar: File not in image cache but could not retrieve it anyway; received hash was %q; desired size was: %d and default param was %q\n", params.Hash, size, defaultParam))
}
