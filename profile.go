package main

import (
	"database/sql"
//	 "encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/peterbourgon/diskv/v3"
	"gopkg.in/gographics/imagick.v3/imagick"
	"io/ioutil"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
	"net/http"
	// "os"
	// "os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type UserProfile struct {
	UserUUID string 			`form:"useruuid" json:"useruuid"`
	ProfilePartner string		`form:"profilePartner" json:"profilePartner"`
	ProfileAllowPublish int		`form:"profileAllowPublish" json:"profileAllowPublish"`
	ProfileMaturePublish int	`form:"profileMaturePublish" json:"profileMaturePublish"`
	ProfileURL string			`form:"profileURL" json:"profileURL"`
	ProfileWantToMask int		`form:"profileWantToMask" json:"profileWantToMask"`
	ProfileWantToText string	`form:"profileWantToText" json:"profileWantToText"`
	ProfileSkillsMask int		`form:"profileSkillsMask" json:"profileSkillsMask"`
	ProfileSkillsText string	`form:"profileSkillsText" json:"profileSkillsText"`
	ProfileLanguages string		`form:"profileLanguages" json:"profileLanguages"`
	ProfileImage string			`form:"profileImage" json:"profileImage"`
	ProfileAboutText string		`form:"profileAboutText" json:"profileAboutText"`
	ProfileFirstImage string	`form:"profileFirstImage" json:"profileFirstImage"`
	ProfileFirstText string		`form:"profileFirstText" json:"profileFirstText"`
}

// GetProfile connects to the database, does its magic, and spews out a profile. That's the theory at least.
func GetProfile(c *gin.Context) {
	session		:= sessions.Default(c)
	username	:= session.Get("Username")
	libravatar	:= session.Get("Libravatar")
	uuid		:= session.Get("UUID")

	// open database connection
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var (
		profileData UserProfile
//		avatarProfileImage string	// constructed URL for the profile image (gwyneth 20200719) Note: not used any longer (gwyneth 20200728)
//		allowPublish, maturePublish string // it has to be this way to get around a bug in the mySQL driver which is impossible to fix
	)
	err = db.QueryRow("SELECT useruuid, profilePartner, profileAllowPublish, profileMaturePublish, profileURL, profileWantToMask, profileWantToText, profileSkillsMask, profileSkillsText, profileLanguages, profileImage, profileAboutText, profileFirstImage, profileFirstText FROM userprofile WHERE useruuid = ?", uuid).Scan(
			&profileData.UserUUID,
			&profileData.ProfilePartner,
			&profileData.ProfileAllowPublish,
			&profileData.ProfileMaturePublish,
			&profileData.ProfileURL,
			&profileData.ProfileWantToMask,
			&profileData.ProfileWantToText,
			&profileData.ProfileSkillsMask,
			&profileData.ProfileSkillsText,
			&profileData.ProfileLanguages,
			&profileData.ProfileImage,
			&profileData.ProfileAboutText,
			&profileData.ProfileFirstImage,
			&profileData.ProfileFirstText,
		)
		// profileData.ProfileAllowPublish		= (allowPublish != "")
		// profileData.ProfileMaturePublish	= (maturePublish != "")
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG]: retrieving profile from user %q (%s) failed; database error was %v", username, uuid, err)
		}
	}

	// see if we have this image already
	// Note: in the future, we might simplify the call by just using the UUID + file extension... (gwyneth 20200727)
	// Note 2: We *also* retrieve a Retina image and store it in the cache, but we do not specifically check for it. The idea is that proper HTML will deal with selecting the 'correct' image, we only need to check for one of them. Also, if the conversion to Retina fails for some reason, that's not a problem, we'll fall back to whatever has been downloaded so far...
	profileImage := filepath.Join(PathToStaticFiles, "/", *config["cache"], profileData.ProfileImage + *config["convertExt"])
	profileRetinaImage := filepath.Join(PathToStaticFiles, "/", *config["cache"], profileData.ProfileImage + "@2x" + *config["convertExt"]) // we need the path, but we won't check for it directly
	/*
	if profileImage[0] != '/' {
		profileImage = "/" + profileImage
	}
	*/
	// either this URL exists and is in the cache, or not, and we need to get the image from
	//  OpenSimulator and attempt to convert it... we won't change the URL in the process.
	// Note: Other usages of the diskv cache might not be so obvious... or maybe they all are? (gwyneth 20200727)
	if !imageCache.Has(profileImage) { // this URL is not in the cache yet!
		if *config["ginMode"] == "debug" {
			log.Println("[INFO] Cache miss on profileImage:", profileImage, " - attempting to download it...")
		}
		// get it!
		profileImageAssetURL := *config["assetServer"] + path.Join("/assets/", profileData.ProfileImage, "/data")
		resp, err := http.Get(profileImageAssetURL)
		defer resp.Body.Close()
		if err != nil {
			// handle error
			log.Println("[ERROR] Oops — OpenSimulator cannot find", profileImageAssetURL)
		}
		newImage, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("[ERROR] Oops — could not get contents of", profileImageAssetURL, "from OpenSimulator")
		}
		if len(newImage) == 0 {
			log.Println("[ERROR] Image retrieved from OpenSimulator", profileImageAssetURL, "has zero bytes.")
			// we might have to get out of here
		} else {
			if *config["ginMode"] == "debug" {
				log.Println("[INFO] Image retrieved from OpenSimulator", profileImageAssetURL, "has", len(newImage), "bytes.")
			}
		}
		// Now use ImageMagick to convert this image!
		// Note: I've avoided using ImageMagick because it's compiled with CGo, but I can't do better
		//  than this. See also https://stackoverflow.com/questions/38950909/c-style-conditional-compilation-in-golang for a way to prevent ImageMagick to be used.
		convertedImage, retinaImage, err := ImageConvert(newImage, 256, 256, 100)
		if err != nil {
			log.Println("[ERROR] Could not convert", profileImageAssetURL, " - error was:", err)
		}
		if convertedImage == nil || len(convertedImage) == 0 {
			log.Println("[ERROR] Converted image is empty")
		}
		if retinaImage == nil || len(retinaImage) == 0 {
			log.Println("[ERROR] Converted Retina image is empty")
		}
		if *config["ginMode"] == "debug" {
			log.Println("[INFO] Regular image from", profileImageAssetURL, "has", len(convertedImage), "bytes; retina image has", len(retinaImage), "bytes.")
		}
		// put it into KV cache:
		if err := imageCache.Write(profileImage, convertedImage); err != nil {
			log.Println("[ERROR] Could not store converted", profileImage, "in the cache, error was:", err)
		}
		// put Retina image into KV cache as well:
		if err := imageCache.Write(profileRetinaImage, retinaImage); err != nil {
			log.Println("[ERROR] Could not store retina image", retinaImage, "in the cache, error was:", err)
		}
	}
	// note that the code will now assume that profileImage does, indeed, have a valid
	//  image URL, and will fail with a broken image (404 error on browser) if it doesn't; thus:
	// TODO(gwyneth): get some sort of default image for when all of the above fails

	// Do the same for the profile image for First (=Real) Life. Comments as above!
	profileFirstImage := filepath.Join(PathToStaticFiles, "/", *config["cache"], profileData.ProfileFirstImage + *config["convertExt"])
	profileRetinaFirstImage := filepath.Join(PathToStaticFiles, "/", *config["cache"], profileData.ProfileFirstImage + "@2x" + *config["convertExt"])

	if !imageCache.Has(profileFirstImage) {
		if *config["ginMode"] == "debug" {
			log.Println("[INFO] Cache miss on profileFirstImage:", profileFirstImage, " - attempting to download it...")
		}
		profileFirstImageAssetURL := *config["assetServer"] + path.Join("/assets/", profileData.ProfileFirstImage, "/data")
		resp, err := http.Get(profileFirstImageAssetURL)
		defer resp.Body.Close()
		if err != nil {
			log.Println("[ERROR] Oops — OpenSimulator cannot find", profileFirstImageAssetURL)
		}
		newImage, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("[ERROR] Oops — could not get contents of", profileFirstImageAssetURL, "from OpenSimulator")
		}
		if len(newImage) == 0 {
			log.Println("[ERROR] Image retrieved from OpenSimulator", profileFirstImageAssetURL, "has zero bytes.")
		} else {
			if *config["ginMode"] == "debug" {
				log.Println("[INFO] Image retrieved from OpenSimulator", profileFirstImageAssetURL, "has", len(newImage), "bytes.")
			}
		}
		convertedImage, retinaImage, err := ImageConvert(newImage, 128, 128, 100)
		if err != nil {
			log.Println("[ERROR] Could not convert", profileFirstImageAssetURL, " - error was:", err)
		}
		if convertedImage == nil || len(convertedImage) == 0 {
			log.Println("[ERROR] Converted image is empty")
		}
		if retinaImage == nil || len(retinaImage) == 0 {
			log.Println("[ERROR] Converted Retina image is empty")
		}
		if *config["ginMode"] == "debug" {
			log.Println("[INFO] Image from", profileFirstImageAssetURL, "has", len(convertedImage), "bytes; retina image has", len(retinaImage), "bytes.")
		}
		if err := imageCache.Write(profileFirstImage, convertedImage); err != nil {
			log.Println("[ERROR] Could not store converted", profileFirstImage, "in the cache, error was:", err)
		}
		// put Retina image into KV cache as well:
		if err := imageCache.Write(profileRetinaFirstImage, retinaImage); err != nil {
			log.Println("[ERROR] Could not store retina image", retinaImage, "in the cache, error was:", err)
		}
	}

	c.HTML(http.StatusOK, "profile.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"needsTables"	: false,
		"needsMap"		: false,
		"moreValidation" : true,
		"author"		: *config["author"],
		"description"	: *config["description"],
		"logo"			: *config["logo"],
		"logoTitle"		: *config["logoTitle"],
		"sidebarCollapsed" : *config["sidebarCollapsed"],
		"Debug"			: false,	// we will probably need two versions of 'debug mode'... (gwyneth 20200622)
		"titleCommon"	: *config["titleCommon"] + profileData.UserUUID + " Profile",
		"ProfileData"	: fmt.Sprintf("%+v", profileData),
		"ProfileURL"	: profileData.ProfileURL, // TODO(gwyneth): This ought to be sanitized!!
		"UserUUID"				: profileData.UserUUID,
		"ProfilePartner"		: profileData.ProfilePartner,
		"ProfileAllowPublish"	: profileData.ProfileAllowPublish,
		"ProfileMaturePublish"	: profileData.ProfileMaturePublish,
		"ProfileWantToMask"	: profileData.ProfileWantToMask,
		"ProfileWantToText"	: profileData.ProfileWantToText,
		"ProfileSkillsMask"	: profileData.ProfileSkillsMask,
		"ProfileSkillsText"	: profileData.ProfileSkillsText,
		"ProfileLanguages"	: profileData.ProfileLanguages,
		"ProfileImage"		: profileImage,						// OpenSimulator/Second Life profile image
		"ProfileRetinaImage"	: profileRetinaImage,			// Generated Retina image
		"ProfileAboutText"	: profileData.ProfileAboutText,
		"ProfileFirstImage"	: profileFirstImage,				// Real life, i.e. 'First Life' image
		"ProfileRetinaFirstImage"	: profileRetinaFirstImage,	// Another generated Retina image
		"ProfileFirstText"	: profileData.ProfileFirstText,
		"Username"			: username,
		"Libravatar"		: libravatar,
	})
}

// saveProfile is what gets called when someone saves the profile.
func saveProfile(c *gin.Context) {
	var oneProfile UserProfile

	session := sessions.Default(c)

	if c.Bind(&oneProfile) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "404.tpl", gin.H{
			"errorcode"		: http.StatusBadRequest,
			"errortext"		: "Saving profile failed",
			"errorbody"		: "No form data posted",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"titleCommon"	: *config["titleCommon"] + " - Profile",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
		log.Println("[ERROR] No form data posted for saving profile")

		return
	}
	c.HTML(http.StatusOK, "404.tpl", gin.H{
		"errorcode"		: http.StatusOK,
		"errortext"		: "Saving profile succeeded",
		"errorbody"		: "But... we still haven't done the coding!... So nothing actually happened",
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"logo"			: *config["logo"],
		"logoTitle"		: *config["logoTitle"],
		"sidebarCollapsed" : *config["sidebarCollapsed"],
		"titleCommon"	: *config["titleCommon"] + " - Profile",
		"Username"		: session.Get("Username"),
		"Libravatar"	: session.Get("Libravatar"),
	})
	log.Println("[INFO] Got form data for profile but code isn't implemented yet")

	return
}

// Transformation functions
// These will probably be moved to cache.go or something similar (gwyneth 20200724)
// TODO(gwyneth): Probably split it further in subdirectories
func imageCacheTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] imageCacheTransform: got key %q transformed into path %v and filename %q\n",
			key, path, path[last])
	}
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}

func imageCacheInverseTransform(pathKey *diskv.PathKey) string {
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] imageCacheInverseTransform: got pathKey %v which will be returned as %q\n",
			pathKey, strings.Join(pathKey.Path, "/") + pathKey.FileName) // inefficient but we're just debugging... (gwyneth 20200727)
	}
	return strings.Join(pathKey.Path, "/") + pathKey.FileName
}

// ImageConvert will take sequence of bytes of an image and convert it into another image with minimal compression, possibly resizing it.
// Parameters are []byte of original image, height, width, compression quality
// Returns []byte of converted image
// See https://golangcode.com/convert-pdf-to-jpg/ (gwyneth 20200726)
// TODO(gwyneth): We might also generate a Retina image; how will it be saved? Through the KV store?
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
	if aImage == nil || len(aImage) == 0 {
		return nil, nil, errors.New("Empty image passed to ImageConvert")
	}
	// Now that we have checked all parameters, it's time to setup ImageMagick:
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

    // Load the image into imagemagick
    if err := mw.ReadImageBlob(aImage); err != nil {
		return nil, nil, err
	}

	if *config["ginMode"] == "debug" {
		filename		:= mw.GetFilename()
		format			:= mw.GetFormat()
		resX, resY, _	:= mw.GetResolution()
		x, y, _			:= mw.GetSize()
		imageProfile	:= mw.GetImageProfile("generic")
		length,	_		:= mw.GetImageLength()
		log.Printf("[DEBUG] ImageConvert now attempting to convert image with filename %q and format %q and size %d (%.f ppi), %d (%.f ppi), Generic profile: %q, size in bytes: %d\n", filename, format, x, resX, y, resY, imageProfile, length)
	}

	if err := mw.ResizeImage(height, width, imagick.FILTER_LANCZOS_SHARP); err != nil {
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
	var formatType string = *config["convertExt"]
	if *config["ginMode"] == "debug" {
		log.Println("[DEBUG] Setting format type to", formatType[1:])
	}
    if err := mw.SetFormat(formatType[1:]); err != nil {
        return nil, nil, err
    }

    // Return []byte for this image
	blob := mw.GetImageBlob()

	// now do the same for the Retina size
	if err := mw.ResizeImage(height * 2, width * 2, imagick.FILTER_LANCZOS_SHARP); err != nil {
		return blob, nil, err
	}

	blobRetina := mw.GetImageBlob()

    return blob, blobRetina, nil
}