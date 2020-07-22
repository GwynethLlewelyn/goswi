// cache implements some functionality to deal with caching.
// TODO(gwyneth): Future: this will be its own standalone package
package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
//	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// ConversionFunction accepts one parameter: what to call before saving the item to cache.
// This function is just needed if the downloaded file is _not_ browser-viewable.
// A function performing conversion will receive the path to the file requiring conversion, and
// returns an error. Actual construction of the URL path occurs inside
type ConversionFunction func (string) (string, error)

// Cache is mostly a placeholder struct for creating Cache-specific methods.
type Cache struct {
	convFunc ConversionFunction	// the function to call when converting from an invalid type to a valid one
	cacheDir string	// path to where the cache items are stored
	baseURL string	// URL pointing to the cache directory
	postfix string	// if file has an invalid type, we temporarily assign this
	convertedExtension string	// the extension we're going to convert to
}

// Download checks if the file is not in the cache yet, and downloads it if it isn't, returning URL to cached file (and error).
// As a side-effect, it calls a conversion function if the file extension is unknown.
func (c Cache) Download(itemURL string) (string, error) {
	if itemURL == "" {
		log.Println("[ERROR] Download: Empty itemURL!")
		return "", errors.New("Empty itemURL")
	}
	// generate itemFileName from itemURL; we need to do some slicing & dicing!
//	var queryString string
	var cleanURL string
	// split by eventual "?"
	i := strings.LastIndex(itemURL, "?")
	if i < 0 {	// are we clean?
//		queryString = ""
		cleanURL = itemURL
	} else {
//		queryString = itemURL[i:]
		cleanURL = itemURL[:i]
	}
	_, file := path.Split(cleanURL)	// now split things between the 'main' URL and the filename at the end
	ext := path.Ext(file)				// check if we have an extension or if the extension is invalid
	if ext == "" || !isValidExtension(ext) {
		ext = c.postfix					// invalid extension, so we add 'our' postfix for this
	}
	// construct the filename that will be placed on cache
	itemFileName := filepath.Join(c.cacheDir, file + ext)
	fd, ferr := os.Open(itemFileName)	// Check if this file already exists in the cache directory
	defer fd.Close()
	if ferr != nil {
		// file is not in the cache yet, so grab it and save it
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG] Download: File", itemURL, "not in cache yet - trying to load it from", itemFileName)
		}
		err := downloadFile(itemFileName, itemURL)
		if err != nil {
			log.Println("[WARN] Download: Item", itemFileName, "didn't get saved to cache")
		} else {
			return "", err
		}
	}
	// Assemble the cache URL
	cacheURL := path.Join(c.baseURL, file + ext)

	// Now the file is in the cache for sure, let's check if it is the 'right' kind of file!
	if !isValidExtension(ext) { // nope — this file needs to be converted first
		if c.convFunc == nil {
			log.Println("[WARN] Download: No conversion function supplied")
			return cacheURL, errors.New("No conversion function supplied")
		} else {
			// call conversion function on the downloaded file
			output, err := c.convFunc(itemFileName)
			if err != nil {
				log.Println("[ERROR] Download: Couldn't launch conversion command", err)
				return cacheURL, err
			} else {
				if *config["ginMode"] == "debug" {
					log.Printf("[DEBUG] Download: Output from converting command was %q\n", output)
				}
				cacheURL = path.Join(c.baseURL, file + c.convertedExtension)
			}
		}
	}
	return cacheURL, nil
}

// GC runs the Garbage Collector, which basically starts every hour and will delete anything with more than 4 hours or so.
// TODO(gwyneth): Figure out a way to deal with different timers (e.g. run every minute, delete everything every X days, etc.)
func (c Cache) GC() {
	// we'll have some code here, a goroutine, and a channel so that we can kill the goroutine
}

// General purpose functions
// downloadFile will create a file with the content of an URL.
// See https://stackoverflow.com/a/33845771/1035977 by Pablo Jomer
func downloadFile(filename string, url string) error {
	// Create the file
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
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