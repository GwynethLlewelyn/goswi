//go:build !cgo || !imagick || spawn

// This is the new default, that launches an external process called `magick`.
// The version is irrelevant, so long as it understands the same parameters of IM 6.9/7+
// It's up to you to put them in the correct path.
// `go build` will automatically select this (or if your CGo environment is broken)
// TODO(gwyneth): add a configuration option for setting the path for ImageMagick.
package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os/exec"
	"text/template"

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

// Global: either the full path was set via `config.ini` or the CLI flags, or we 'assume' this is in the path.
var imagickCommand = "magick"

/* string of parameters to be passed to spawned ImageMagick. This is a (reasonable) default.
 * Note that the placeholders are Go text templates:
 *   {{.width}} — Image width;
 *   {{.height}} - Image height;
 *   {{.compression}} - Compression level (usually 0–100, 75–80 recommended);
 *   {{.fileFormat}} - Image file format type, e.g. "webp", "png", etc.
 *
 * The assumption is that the values will be replaced dynamically at runtime.
 */
const imagickParamsDefault = `- -filter Lanczos2Sharp -resize {{- if .width ne 0 -}}{{- .width -}}{{- end -}}x{{- if .height ne 0 -}}{{- .height -}}{{- end -}} -quality {{.compression}} -alpha off -format {{.fileFormat}} -`

// Global: parameters sent to `imagickCommand`, set to default for good measure.
// Note that this is a *template* (string), *not* an array of parameters to pass,
// which *must* be constructed firs by our code! (gwyneth 20251031)
var imagickParams string = imagickParamsDefault

// struct to be passed to the text templating engine, because Go developers
// *love* structs! (gwyneth )20251030
type paramsType struct {
	width, height, compression uint
	fileFormat                 string // e.g. "webp", "png", etc.
}

// Initialisation of this submodule.
func init() {
	// Do we have set up `magick` from an absolute path, or simply fall back to the system $PATH?
	if config["ImageMagickCommand"] != nil && len(*config["ImageMagickCommand"]) != 0 {
		// Check if the absoliute path of this command exists and is properly set to executable.
		if err := tidbits.CheckFileExecutable(*config["ImageMagickCommand"], false); err != nil {
			config.LogErrorf("ImageMagick `imagick` executable not found in %q; please check the path (or set `ImageMagickCommand` to blank), otherwise images won't work", *config["ImageMagickCommand"])
			return
		}
		imagickCommand = *config["ImageMagickCommand"]
	}
	// Right, we fall back to using the executable in the path...
	if err := tidbits.CheckFileExecutable(imagickCommand, true); err != nil {
		config.LogError("ImageMagick `imagick` executable not found in path; please check if it's in the path, otherwise images won't work")
	}

	if config["ImageMagickParams"] != nil && len(*config["ImageMagickParams"]) != 0 {
		// If ImageMagickParams is configured, then assign it instead (otherwise, keep the defaults). (gwyneth 20251030)
		imagickParams = *config["ImageMagickParams"]
	}

	return
}

// Returns all parsed ImageMagick parameters as an array of strings.
func parseParams(paramList string, height, width, compression uint, fileFormat string) ([]string, error) {
	// Some trivial initial checks.
	if len(paramList) == 0 {
		// empty string? use the default!
		paramList = imagickParamsDefault
		config.LogWarn("empty `paramList`, falling back to default")
	}
	if len(fileFormat) < 2 {
		// empty string or too short? use the default, "png"
		fileFormat = "png"
		config.LogWarn("empty or too short `fileFormat`, falling back to default, 'png'")
	}
	// If the fileFormat begins with a dot (because it's derived from the file extension),
	// then strip the dot, and just retain the rest.
	if fileFormat[0] == '.' {
		fileFormat = fileFormat[1:]
	}
	// TODO(gwyneth): check if the file type is a valid file type.
	// This is not trivial, as it requires asking ImageMagick what formats it supports,
	// which, in turn, depends on the compiled-in options. (gwyneth 20251031)

	var err error        // dealing with scope issues.
	var buf bytes.Buffer // result of applying template to parameters.

	// Instanciate a new template, giving it a name.
	t := template.New("params")
	t, err = t.Parse(paramList)
	if err != nil {
		// something went wrong when parsing the template: try again, this time with defaults.
		config.LogErrorf("parseParams(): could not parse parameters with %q, falling back to default parameters template; error was %q", paramList, err)
		t, err = t.Parse(imagickParamsDefault)
		if err != nil {
			config.LogErrorf("parseParams(): could not parse default parameters, aborting; error was %q", err)
			return nil, err
		}
	}

	// To parse the CLI parameters, we're using Go's native templating system, which,
	// for text, requires all components to be applied by the template to come from
	// either a map or a struct.
	var params = paramsType{
		width:       width,
		height:      height,
		compression: compression,
		fileFormat:  fileFormat,
	}

	// Now execute the parsed template with the params we pushed into the struct above.
	err = t.Execute(&buf, params)
	if err != nil {
		// something went wrong when executing the template: abort!
		return nil, fmt.Errorf("parseParams(): could not execute %q, wrong parameters; error was %q", buf, err)
	}
	// Slice result, tokenizing it by spaces. We use the CSV package because it neatly eats up unnecessary
	// quotes (required for preserving spaces inside a parameter).
	r := csv.NewReader(&buf)
	r.Comma = ' '
	r.TrimLeadingSpace = true // exec.Command() wants neatly trimmed parameters without extra space.
	return r.Read()           // outputs a []string and an error, exactly what we need.
}

// Internal function to spawn an ImageMagick process and feed everything to it.
// This image resizing operation will be called *twice*, once for normal size, another for Retina size.
func spawnImageMagick(aImage []byte, height, width, compression uint) ([]byte, error) {
	// Auxiliary result for avoiding all 'expensive' Sprintf(), by creating a `widthxheight` string only once.
	var dimensions = fmt.Sprintf("%dx%d", width, height)

	// Format to convert to, based on the extension given, minus the dot at the beginning.
	// The default                                                                                                                           will be .png, but we'll read it from the configuration, *if* available.
	var formatType string = ".png"

	// Note that this is required, since some tests may **not** properly instantiate *config[].
	if len(config) != 0 && config["convertExt"] != nil && len(*config["convertExt"]) >= 2 {
		formatType = *config["convertExt"]
	}

	config.LogTrace("spawnImageMagick called with `aImage` length ", len(aImage), "resize to:", dimensions, "compression quality:", compression)
	config.LogDebug("Setting format type to", formatType[1:])

	// Construct the list of parameters to pass to èxec.Command().
	// Default params are: "-", "-filter", "Lanczos2Sharp", "-resize", dimensions,
	// "-quality", fmt.Sprintf("%d", compression),
	// "-alpha", "off", "-format", formatType[1:], "-"
	params, err := parseParams(imagickParams, height, width, compression, formatType)
	if err != nil {
		config.LogErrorf("Could not parse command parameters as an array of strings, error was: %q, reverting to default parameters", err)
		// In this case, we supercede the automatic parsing, and just do it manually.
		// Of course, this works only fpr `imagick`. (gwyneth 20251031)
		params = []string{"-", "-filter", "Lanczos2Sharp", "-resize", dimensions,
			"-quality", fmt.Sprintf("%d", compression),
			"-alpha", "off", "-format", formatType[1:], "-"}
	}

	// We got the command path to execute and have all the parameters correcly parsed,
	// so let the games begin!
	cmd := exec.Command(imagickCommand, params...)

	// I'm not 100% sure if things need to be done this way.
	// var stdinbuf bytes.Buffer
	// cmd.Stdin = &stdinbuf

	// bytesWritten, copyErr := stdinbuf.Write(aImage)
	// bytesWritten, copyErr := cmd.Stdin.Read(aImage)
	// if copyErr != nil {
	// 	config.LogFatal("could not pipe image with", dimensions, "to spawned `magick` process, error was", copyErr, "bytes written:", bytesWritten)
	// }

	// Attempt #2: Pipe aImage ([]byte) directly into cmd's stdin.

	var bytesWritten int // number of bytes actually written to the spawned process.
	var writeErr error   // error returned by the pipe to the spawned process.

	// start by creating a pipe:
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not open communications with %q, error was %q", imagickCommand, err)
	}

	// Now go into the background and write our nice image to the pipe.
	// Buffering, etc., should be working automagically.
	go func() {
		defer stdin.Close()
		bytesWritten, writeErr = stdin.Write(aImage)
	}()
	// What happens at this stage is a bit muddy.
	// The Go 'official' examples never check for errors or count the bytes sent.
	// So, probably we'll only see the result when calling cmd.Output(), which allegedly handles
	// all waiting and goroutine synchronisation for us:
	outImage, outErr := cmd.Output()
	// At this stage, the goroutine ought to be synced, and we
	// *might* be able to read the errors from the *writing* operation on the pipe:
	if writeErr != nil {
		return nil, fmt.Errorf("could not write image file to convert to %q, error was %q, bytes written %d", imagickCommand, writeErr, bytesWritten)
	}
	// Note: a value of zero *may* be possible, under race conditions, and might not be an actual error.
	// TODO(gwyneth): check if that's the case using the test battery. (gwyneth 20251031)
	if bytesWritten == 0 {
		return nil, fmt.Errorf("could not write image file to convert to %q, error was %q, bytes written %d", imagickCommand, writeErr, bytesWritten)
	}

	// We should also be able to have an idea of how many bytes were originally sent to the
	// pipe, purely for debugging purposes.
	config.LogDebugf("image with size %s was successfully sent to spawned %q process, wrote %d bytes", dimensions, imagickCommand, bytesWritten)

	if outErr != nil {
		config.LogDebugf("%q returned error %q", imagickCommand, outErr)
		return nil, outErr
	}
	if len(outImage) == 0 {
		config.LogDebugf("%q returned image with zero bytes", imagickCommand)
		return nil, fmt.Errorf("empty image received from call to %q, no OS errors were returned", imagickCommand)
	}
	config.LogTracef("%q returned converted image with %d bytes", imagickCommand, len(outImage))

	return outImage, nil
}
