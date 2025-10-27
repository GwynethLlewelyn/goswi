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
	"io"
	"os/exec"
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

// Internal function to spawn an ImageMagick process and feed everything to it.
// Since the image resizing operation will be called *twice*, this means
func spawnImageMagick(aImage []byte, height, width, compression uint) ([]byte, error) {
	// Call 'expensive' Sprintf() to create a `widthxheight` string only once.
	var dimensions = fmt.Sprintf("%dx%d", width, height)

	// Format to convert to, based on the extension given, minus the dot at the beginning.
	var formatType = *config["convertExt"]
	config.LogDebug("Setting format type to", formatType[1:])

	// spawn imagick process, preparing it to accept from stdin and write to stdout.
	cmd := exec.Command("magick", "-", "-filter", "Lanczos2Sharp", "-resize", dimensions,
		"-quality", fmt.Sprintf("%d", compression),
		"-alpha", "off", "-format", formatType[1:], "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdinImage := bytes.NewBuffer(aImage)
	if stdinImage == nil {
		return nil, errors.New("failed to allocate memory for image buffer with size: " + dimensions)
	}

	// TODO: I have to think a bit to see if it makes sense running a gooutine here, or not.
	go func() {
		bytesWritten, copyErr := io.Copy(stdin, stdinImage)
		if copyErr != nil {
			config.LogFatalf("could not pipe image with %s to spawned `magick` process, error was %q", dimensions, copyErr)
		}
		config.LogDebugf("image with %s successfully sent to spawned `magick` process, wrote %d bytes", dimensions, bytesWritten)
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	stdoutImage, outputErr := io.ReadAll(stdout)
	if outputErr != nil {
		config.LogErrorf("could not capture result of converting image to %s, error was %q", dimensions, outputErr)
		stdoutImage = nil // should be nil by default...
	}

	// clean exit, even if the saving to a buffer failed. If that's the case,
	// we'll return the error for outputErr; otherwise, it will be nil.
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return stdoutImage, outputErr
}
