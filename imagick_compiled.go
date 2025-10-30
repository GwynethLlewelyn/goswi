//go:build cgo && imagick && !spawn

// This is the variant that uses the ImageMagick wrapper around the CGo-compiled library.
// To use this, make sure your environment has a working CGo installation, and then build using:
// `go build -tags imagick`

package main

import (
	"gopkg.in/gographics/imagick.v3/imagick" // we might have alternatives for 6.9 and 7+
)

// Initialise ImageMagick, which we use to convert JPEG2000 to PNG
func init() {
	imagick.Initialize()
	defer imagick.Terminate()
}

// ImageConvert will take sequence of bytes of an image and convert it into another image with minimal compression, possibly resizing it.
// Parameters are []byte of original image, height, width, compression quality
// Returns []byte of converted image
// See https://golangcode.com/convert-pdf-to-jpg/ (gwyneth 20200726)
func ImageConvert(aImage []byte, height, width, compression uint) ([]byte, []byte, error) {
	// some minor error checking on params
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
	// Now that we have checked all parameters, it's time to setup ImageMagick:
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Load the image into imagemagick
	if err := mw.ReadImageBlob(aImage); err != nil {
		return nil, nil, err
	}

	if *config["ginMode"] == "debug" {
		filename := mw.GetFilename()
		format := mw.GetFormat()
		resX, resY, _ := mw.GetResolution()
		x, y, _ := mw.GetSize()
		imageProfile := mw.GetImageProfile("generic")
		length, _ := mw.GetImageLength()
		config.LogDebugf("ImageConvert now attempting to convert image with filename %q and format %q and size %d (%.f ppi), %d (%.f ppi), Generic profile: %q, size in bytes: %d\n", filename, format, x, resX, y, resY, imageProfile, length)
	}

	if err := mw.ResizeImage(height, width, imagick.FILTER_LANCZOS2_SHARP); err != nil {
		return nil, nil, err
	}

	// Must be *after* ReadImage
	// Flatten image and remove alpha channel, to prevent alpha turning black in jpg
	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_OFF); err != nil {
		return nil, nil, err
	}

	// Set any compression (100 = max quality)
	if err := mw.SetCompressionQuality(compression); err != nil {
		return nil, nil, err
	}

	// Move to first image
	mw.SetIteratorIndex(0)

	// Convert into PNG
	var formatType = *config["convertExt"]
	config.LogDebug("Setting format type to", formatType[1:])
	if err := mw.SetFormat(formatType[1:]); err != nil {
		return nil, nil, err
	}

	var err error // We need to define this here because of stupid scoping issues...

	// Return []byte for this image.
	blob, err := mw.GetImageBlob()
	if err != nil {
		return nil, nil, err
	}

	// now do the same for the Retina size
	if err := mw.ResizeImage(height*2, width*2, imagick.FILTER_LANCZOS_SHARP); err != nil {
		// this probably doesn't make sense, but we return the valid image, while
		// sending back nil for the Retina image *and* the error about why this didn't work.
		// (gwyneth 20240620)
		return blob, nil, err
	}

	blobRetina, err := mw.GetImageBlob()
	if err != nil {
		return nil, nil, err
	}

	return blob, blobRetina, nil
}
