//go:build !cgo || !imagick

// This is the new default, that launches an external process called `magick`.
// The version is irrelevant, so long as it understands the same parameters of IM 6.9/7+
// It's up to you to put them in the correct path.
// `go build` will automatically select this (or if your CGo environment is broken)
// TODO(gwyneth): add a configuration option for setting the path for ImageMagick.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"gitlab.com/StellarpowerGroupedProjects/tidbits/go"
	tidbits "gitlab.com/StellarpowerGroupedProjects/tidbits/go"
)

// ImageConvert will take sequence of bytes of an image and convert it into another image with minimal compression, possibly resizing it.
// Parameters are []byte of original image, height, width, compression quality
// Returns []byte of converted image
// This spawns an external ImageMagick process and reads the results — twice, once for the normalSize, the second time
// for the Retina size.
func ImageConvert(aImage []byte, height, width, compression uint) (normalSize []byte, retinaSize []byte, err error) {
	// some minor error checking on params.
	if height == 0 {
		height = 256
	}
	if width == 0 {
		width = height
	}
	if compression == 0 {
		compression = 75
	}
	if len(aImage) == 0 {
		return nil, nil, errors.New("empty image passed to ImageConvert")
	}

	// Now spawn one process for the original size.
	normalSize, err = spawnImageMagick(aImage, height, width, compression)
	if err != nil {
		return nil, nil, err
	}

	// If all went well, do the same for the Retina version. Note that we attempt to return at least
	// the original size, if it was valid. Otherwise, we just send back nil.
	retinaSize, err = spawnImageMagick(aImage, height*2, width*2, compression)
	if err != nil {
		return normalSize, nil, err // returning normalSize makes no difference.
	}

	return
}

// Either the full path was set via `config.ini` or the CLI flags, or we 'assume' this is in the path.
var imagickCommand = "magick"

func init() {
	// Do we have set up `agick` from an absolute path, or simply fall back to the system $PATH?
	if len(*config["ImageMagickCommand"]) != 0 {
		// Check if the absoliute path of this command exists and is properly set to executable.
		if err := tidbits.CheckFileExecutable(*config["ImageMagickCommand"], false); err != nil {
			config.LogErrorf("ImageMagick executable not found at %q; please check the path (or set `ImageMagickCommand` to blank), otherwise images won't work", *config["ImageMagickCommand"])
			return
		}
	}
	// Right, we fall back to using th pah...
	if err := tidbits.CheckFileExecutable(imagickCommand, true); err != nil {
		config.LogError("ImageMagick `imagick`executable not found in path; please check if it's in the path, otherwise images won't work")
	}

	return
}

// Internal function to spawn an ImageMagick process and feed everything to it.
// Since the image resizing operation will be called *twice*, this means
func spawnImageMagick(aImage []byte, height, width, compression uint) ([]byte, error) {
	// Call 'expensive' Sprintf() to create a `widthxheight` string only once.
	var dimensions = fmt.Sprintf("%dx%d", width, height)

	// Format to convert to, based on the extension given, minus the dot at the beginning.
	// The default will be .png, but we'll read it from the configuration, *if* available.
	var formatType string = ".png"

	// Note that this is required, since some tests will **not** properly instantiate *config[].
	if len(config) != 0 && config["convertExt"] != nil && len(*config["convertExt"]) >= 2 {
		formatType = *config["convertExt"]
	}

	config.LogTrace("spawnImageMagick called with `aImage` length ", len(aImage), "resize to:", dimensions, "compression quality:", compression)
	config.LogDebug("Setting format type to", formatType[1:])

	cmd := exec.Command("magick", "-", "-filter", "Lanczos2Sharp", "-resize", dimensions,
		"-quality", fmt.Sprintf("%d", compression),
		"-alpha", "off", "-format", formatType[1:], "-")

	var stdinbuf /*, stdoutbuf, stderrbuf*/ bytes.Buffer
	cmd.Stdin = &stdinbuf
	// cmd.Stdout = &stdoutbuf
	// cmd.Stderr = &stderrbuf

	bytesWritten, copyErr := stdinbuf.Write(aImage)
	if copyErr != nil {
		config.LogFatal("could not pipe image with", dimensions, "to spawned `magick` process, error was", copyErr)
	}
	config.LogDebug("image with", dimensions, "successfully sent to spawned `magick` process, wrote", bytesWritten, "bytes")
	/* 	if err := cmd.Run(); err != nil {
	   		return nil, err
	   	}
	   	if stderrbuf.Len() != 0 {
	   		config.LogDebug("`ìmagick` returned:", stderrbuf.String())
	   	}
	   	return stdoutbuf.Bytes(), nil */

	// outImage is the return value from the command, if anything.
	outImage, err := cmd.Output()
	if err != nil {
		config.LogDebug("`ìmagick` returned error:", err)
		return nil, err
	}
	if len(outImage) == 0 {
		config.LogDebug("`ìmagick` returned image with zero bytes")
		return nil, errors.New("empty image received")
	}
	config.LogTracef("`ìmagick` returned image with %d bytes", len(outImage))

	return outImage, nil
}
