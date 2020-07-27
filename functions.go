package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
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

// formatAsDate is a function for the templating system, which will be registered below.
func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

// formatAsYear is another function for the templating system, which will be registered below.
func formatAsYear(t time.Time) string {
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
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