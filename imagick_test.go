//go:build !cgo || !imagick

// Test battery for ImageMagick's spawn version.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	inputFolder  = "./assets/images/"
	outputFolder = "./testdata/output"
)

func TestMain(m *testing.M) {
	fmt.Println("Entering test battery main configuration...")

	// deal with -removeall flag
	removeAll := false
	for _, flag := range os.Args {
		if flag == "--remove-all" {
			removeAll = true
		}
	}

	if len(config) == 0 {
		fmt.Println("`config` is essentially empty")
		config = make(map[string]*string)
	}

	if config["convertExt"] == nil {
		fmt.Println("`config` exists, but `convertExt` is not one of its members")
		config["convertExt"] = new(string)
	}

	if len(*config["convertExt"]) < 2 {
		fmt.Println("`*config[\"convertExt\"]` exists, but it's too small to be a valid extension; setting to .png")
		*config["convertExt"] = ".png"
	}
	// we also need the debugging value
	if config["ginMode"] == nil {
		fmt.Println("`config` exists, but `ginMode` is not one of its members")
		config["ginMode"] = new(string)
	}

	if len(*config["ginMode"]) < 2 {
		fmt.Printf("`*config[\"ginMode\"]` exists, but it's too small to be valid: %q\n", *config["ginMode"])
		*config["ginMode"] = LevelTraceValue
	}
	// and having the path to ImageMagick
	if config["ImageMagickCommand"] == nil {
		fmt.Println("`config` exists, but `ImageMagickCommand` is not one of its members")
		config["ImageMagickCommand"] = new(string)
		*config["ImageMagickCommand"] = ""
	}

	fmt.Println("Configuration ok! ✅")

	// Make sure that the output dirs are available.
	// This is mostly because they will *not* be created automagically... unless we create them here!
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		fmt.Printf("Cannot create output directories under %s! Please check what's wrong. Error was: %q", outputFolder, err)
	}

	fmt.Println("Running main tests...")

	code := m.Run()

	// cleanup as necessary
	if removeAll {
		cleanup := filepath.Join(outputFolder, "*")
		fmt.Printf("Removing test images from %q...", cleanup)
		if err := removeGlob(cleanup); err != nil {
			fmt.Println("failed! Remove them manually from", outputFolder)
			if code == 0 { // use a different exit code just for this.
				code = 2
			}
		}
	}

	os.Exit(code)
}

// Tests if our logging software is working at all.
func TestLog(t *testing.T) {
	// This may crash but it shouldn't.
	t.Logf("Testing for logging level: %q", *config["ginMode"])

	// if any of them fails, it will be probably the very first.
	config.LogTrace("✅")
	config.LogDebug("✅")
	config.LogInfo("✅")
	config.LogWarn("✅")
	config.LogError("✅")

	// TODO(gwyneth):These need to be caught with `defer` tricks, or else they will abort the tests.
	// See my 'other code' as a reference n how to do it properly...
	// config.LogFatal("✅")
	// config.LogPanic("✅")
}

// Simple test to see if we can make thumbnails out of the original images.
// Note: The test will not output anything visible, just check if the files were written out correctly as expected.
// A visual confirmation might be required.
// For a reference, see https://eli.thegreenplace.net/2022/file-driven-testing-in-go/
func TestImageConvert_Spawn(t *testing.T) {
	if config["convertExt"] == nil || len(*config["convertExt"]) < 2 {
		t.Fatal("could not find valid configuration for the conversion file extension")
	}

	// Find the paths of all input files in the data directory.
	paths, err := filepath.Glob(filepath.Join(inputFolder, "*"+*config["convertExt"]))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		_, filename := filepath.Split(path)
		testname := filename[:len(filename)-len(filepath.Ext(path))]

		// Each path turns into a test: the test name is the filename without the
		// extension.
		t.Run(testname, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal("error reading source file:", err)
			}

			t.Logf("Opening %q...", path)
			// Send the images for convertion:
			normalSize, retinaSize, err := ImageConvert(source, 256, 256, 75)
			if err != nil {
				t.Fatal("error calling ImageConvert:", err)
			}

			// now write them out:
			if err := os.WriteFile(filepath.Join(outputFolder, testname+"-256x256.png"), normalSize, os.FileMode(int(0644))); err != nil {
				t.Fatal("error saving 256x256 (normal) version:", err)
			}
			t.Log(testname + "(normal): ✅")
			if err := os.WriteFile(filepath.Join(outputFolder, testname+"-512x512.png"), retinaSize, os.FileMode(int(0644))); err != nil {
				t.Fatal("error saving 512x512 (Retina) version:", err)
			}
			t.Log(testname + "(Retina): ✅")
		})
	}
}

// Stupid function to do what Go doesn't do: remove everything *inside* a path.
func removeGlob(path string) (err error) {
	contents, err := filepath.Glob(path)
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			return
		}
	}
	return
}
