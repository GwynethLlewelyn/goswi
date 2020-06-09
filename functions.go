package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
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

    usr, err := user.Current()
    if err != nil {
        return "", err
    }
    return filepath.Join(usr.HomeDir, path[1:]), nil
}