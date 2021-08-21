package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/base64"
	"fmt"
	"github.com/gin-contrib/sessions"
//	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
//	"html/template"
	"log"
	"math"
	"os"
	osUser "os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// checkErrFatal logs a fatal error and does whatever log.Fatal() is supposed to do.
func checkErrFatal(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		log.Fatal(filepath.Base(file), ":", line, ":", pc, ok, " - panic:", err)
	}
}

// checkErrPanic logs a fatal error and panics.
func checkErrPanic(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		log.Panic(filepath.Base(file), ":", line, ":", pc, ok, " - panic:", err)
	}
}

// checkErr checks if there is an error, and if yes, it logs it out and continues.
//  this is for 'normal' situations when we want to get a log if something goes wrong but do not need to terminate execution.
func checkErr(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		fmt.Fprintln(os.Stderr, filepath.Base(file), ":", line, ":", pc, ok, " - error:", err)
	}
}

// expandPath expands the tilde as the user's home directory.
//  found at http://stackoverflow.com/a/43578461/1035977
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := osUser.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

/**
*	Cryptographic helper functions
**/

// GetMD5Hash calculated the MD5 hash of any string. See aviv's solution on SO: https://stackoverflow.com/a/25286918/1035977
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// generateSessionToken uses the same approach as OpenSimulator, which is to return a newly created UUID.
func generateSessionToken() string {
	return uuid.New().String()
}

// randomBase64String is Steven Soroka's simple solution to generate a cryptographically secure random string with base64 encoding (see https://stackoverflow.com/a/55860599/1035977) (gwyneth 20200706)
func randomBase64String(l int) string {
    buff := make([]byte, int(math.Round(float64(l)/float64(1.33333333333))))
    rand.Read(buff)
    str := base64.RawURLEncoding.EncodeToString(buff)
    return str[:l] // strip 1 extra character we get from odd length results
}

// isValidExtension looks up a file extension and checks if it is valid for using inside HTML <img>.
// It's a switch because it's more efficient: https://stackoverflow.com/a/52710077/1035977
func isValidExtension(lookup string) bool {
	switch strings.ToLower(lookup) {
		// A full list of valid extensions is here: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/img
		// I've added .mp4 for the sake of convenience (gwyneth 20200722)
		case
		".bmp",
		".cur",
		".ico",
		".jfif",
		".pjp",
		".pjpeg",
		".apng",
		".gif",
		".jpeg",
		".jpg",
		".mp4",
		".png",
		".svg",
		".webp":
		return true
	}
	return false
}

// MergeMaps adds lots of map[string]interface{} together, returning the merged map[string]interface{}.
// It overwrites duplicate keys, maps to the right overwriting whatever keys are on the left.
// This allows for setting 'default' arguments later below, which can be overriden.
// See https://play.golang.org/p/8a9cXdSL_o3 as well as https://stackoverflow.com/a/39406305/1035977.
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

/**
 * Functions used inside templates.
 **/

// bitTest applies a mask to a flag and returns true if the bit is set in the mask, false otherwise.
func bitTest(flag int, mask int) bool {
	return (flag & mask) != 0
}

// formatAsDate is a function for the templating system.
func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

// formatAsYear is another function for the templating system.
func formatAsYear(t time.Time) string {
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
}

/**
 * Auxiliary functions for the Gin Gonic environment.
 **/

// environment pushes a lot of stuff into the common environment
func environment(c *gin.Context, env gin.H) gin.H {
	session := sessions.Default(c)
	var sidebarCollapsed string = ""	// false by default
	if session.Get("sidebarCollapsed") == "true" {
		sidebarCollapsed = "true"
	} else if *config["sidebarCollapsed"] == "true" {
		sidebarCollapsed = "true"
	}

	var data = gin.H{
		/* common environment */
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"logo"			: *config["logo"],
		"logoTitle"		: *config["logoTitle"],
		"sidebarCollapsed" : sidebarCollapsed,
		"titleCommon"	: *config["titleCommon"],
		"StatsDir"		: *config["gridstats"],
		/* session data */
		"Username"		: session.Get("Username"),
		"UUID"			: session.Get("UUID"),
		"Libravatar"	: session.Get("Libravatar"),
		"Token"			: session.Get("Token"),
		"Email"			: session.Get("Email"),
		"RememberMe"	: session.Get("RememberMe"),
		"Messages"		: session.Get("Messages"),
		"numberMessages": session.Get("numberMessages"),
		"FeedMessages"		: session.Get("FeedMessages"),
		"numberFeedMessages": session.Get("numberFeedMessages"),
	}

	retMap := MergeMaps(data, env)

	// if *config["ginMode"] == "debug" && retMap["Username"] != nil && retMap["Username"] != "" {
	// 	log.Printf("[DEBUG]: environment(): All messages for user %q: %+v\n", retMap["Username"], retMap["Messages"])
	// }

	return retMap
}