// cache implements some functionality to deal with cache
package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Cache is mostly a placeholder struct for creating Cache-specific methods.
type Cache struct {
	// it makes sense to store things here as well
	directory string	// path to where the cache items are stored
	url string			// URL pointing to the cache directory
}

// ConversionFunction accepts one parameter: what to call before saving the item to cache.
type ConversionFunction func (string) string

// New returns a new Cache object, initialising some things.
func (c Cache) New(dir string, u string) (Cache) {
	c.directory = dir
	c.url = u
	// that's it for now!
	return
}

// Store returns the URL for the cached item.

// ConvertAndStore returns an URL for a cached item.
func (c Cache) ConvertAndStore(itemFileName string, itemURL string, conversionFunc ConversionFunction) string {
	fd, ferr := os.Open(itemURL)
	defer fd.Close()
	if ferr != nil {
		// file is not in the cache yet, so grab it and save it
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG] File", itemURL, "not in cache yet - trying to load it from", itemFileName)
		}
		err := downloadFile(itemFileName, itemURL)
		if err != nil {
			log.Println("[WARN] Item", itemFileName, "didn't get saved to cache")
		}
	}
	fd.Close()	// we don't need it any more
	}
}

// General purpose functions
// downloadFile will create a file with the content of an URL.
// See https://stackoverflow.com/a/33845771/1035977 by Pablo Jomer
func downloadFile(filename string, url string) (err error) {
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