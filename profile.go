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
//	"time"
)

type UserProfile struct {
	UserUUID string 			`form:"UserUUID" json:"useruuid"`
	ProfilePartner string		`form:"ProfilePartner" json:"profilePartner"`
	ProfileAllowPublish bool	`form:"ProfileAllowPublish" json:"profileAllowPublish"`	// inside the database, this is binary(1)
	ProfileMaturePublish bool	`form:"ProfileMaturePublish" json:"profileMaturePublish"`
	ProfilePublish []string		`form:"ProfilePublish" json:"profilePublish"`	// seems to be needed; values are Allow and Mature
	ProfileURL string			`form:"ProfileURL" json:"profileURL"`
	ProfileWantToMask int		`form:"ProfileWantToMask" json:"profileWantToMask"`
	ProfileWantTo []string		`form:"ProfileWantTo[]"`
	ProfileWantToText string	`form:"ProfileWantToText" json:"profileWantToText"`
	ProfileSkillsMask int		`form:"ProfileSkillsMask" json:"profileSkillsMask"`
	ProfileSkills []string		`form:"ProfileSkills[]"`
	ProfileSkillsText string	`form:"ProfileSkillsText" json:"profileSkillsText"`
	ProfileLanguages string		`form:"ProfileLanguages" json:"profileLanguages"`
	ProfileImage string			`form:"ProfileImage" json:"profileImage"`
	ProfileAboutText string		`form:"ProfileAboutText" json:"profileAboutText"`
	ProfileFirstImage string	`form:"ProfileFirstImage" json:"profileFirstImage"`
	ProfileFirstText string		`form:"ProfileFirstText" json:"profileFirstText"`
}

// GetProfile connects to the database, does its magic, and spews out a profile. That's the theory at least.
func GetProfile(c *gin.Context) {
	session		:= sessions.Default(c)
	username	:= session.Get("Username")
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
		allowPublish, maturePublish []byte // it has to be this way to get around a bug in the mySQL driver which is impossible to fix
	)
	err = db.QueryRow("SELECT useruuid, profilePartner, profileAllowPublish, profileMaturePublish, profileURL, profileWantToMask, profileWantToText, profileSkillsMask, profileSkillsText, profileLanguages, profileImage, profileAboutText, profileFirstImage, profileFirstText FROM userprofile WHERE useruuid = ?", uuid).Scan(
			&profileData.UserUUID,
			&profileData.ProfilePartner,
			&allowPublish,	// &profileData.ProfileAllowPublish,
			&maturePublish,	// &profileData.ProfileMaturePublish,
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
		profileData.ProfileAllowPublish		= (allowPublish[0] != 0)
		profileData.ProfileMaturePublish	= (maturePublish[0] != 0)

	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[ERROR]: retrieving profile from user %q (%s) failed; database error was %v\n", username, uuid, err)
		}
	} else {
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG]: while retrieving profile, allowPublish is %v while maturePublish is %v\n", allowPublish, maturePublish)
			log.Printf("[DEBUG]: retrieving profile from user %q (%s): %+v\n", username, uuid, profileData)
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

	c.HTML(http.StatusOK, "profile.tpl", environment(c, gin.H{
		"needsTables"	: false,
		"needsMap"		: false,
		"moreValidation" : true,
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
	}))
}

// saveProfile is what gets called when someone saves the profile.
func saveProfile(c *gin.Context) {
	var oneProfile UserProfile

	session := sessions.Default(c)
	thisUUID := session.Get("UUID")

	if c.Bind(&oneProfile) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "404.tpl", environment(c, gin.H{
			"errorcode"		: http.StatusBadRequest,
			"errortext"		: "Saving profile failed",
			"errorbody"		: "No form data posted",
			"titleCommon"	: *config["titleCommon"] + " - Profile",
		}))
		log.Println("[ERROR] No form data posted for saving profile")

		return
	}

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] oneProfile is now %+v\n", oneProfile)
	}

	// check if we really are who we claim to be
	if thisUUID != oneProfile.UserUUID {
		c.HTML(http.StatusUnauthorized, "404.tpl", environment(c, gin.H{
			"errorcode"		: http.StatusUnauthorized,
			"errortext"		: "No permission",
			"errorbody"		: fmt.Sprintf("You have no permission to change the profile for %q", session.Get("Username")),
			"titleCommon"	: *config["titleCommon"] + " - Profile",
		}))
		log.Printf("[ERROR] Session UUID %q is not the same as Profile UUID %q - profile data change for %q not allowed\n",
		thisUUID, oneProfile.UserUUID, session.Get("Username"))

		return
	}

	// Allegedly we have successfully bound to the form data, so we can proceed to write it to the database.
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	// Calculate the masks

	var wantToMask, skillsMask int

	for _, bitfield := range(oneProfile.ProfileWantTo) {
		switch bitfield {
			case "Build":
				wantToMask += 1
			case "Meet":
				wantToMask += 4
			case "Group":
				wantToMask += 8
			case "Sell":
				wantToMask += 32
			case "Explore":
				wantToMask += 2
			case "BeHired":
				wantToMask += 64
			case "Buy":
				wantToMask += 16
			case "Hire":
				wantToMask += 128
		}
	}

	for _, bitfield := range(oneProfile.ProfileSkills) {
		switch bitfield {
			case "Textures":
				skillsMask += 1
			case "Modeling":
				skillsMask += 8
			case "Scripting":
				skillsMask += 16
			case "Architecture":
				skillsMask += 2
			case "EventPlanning":
				skillsMask += 4
			case "CustomCharacters":
				skillsMask += 32
		}
	}

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] oneProfile.ProfileWantTo is %v, wantToMask is %d, oneProfile.ProfileSkills is %v, skillsMask is %d\n", oneProfile.ProfileWantTo, wantToMask, oneProfile.ProfileSkills, skillsMask)
	}

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] oneProfile.ProfileAllowPublish is %+v, oneProfile.ProfileMaturePublish is %+v\n", oneProfile.ProfileAllowPublish, oneProfile.ProfileMaturePublish)
	}

	var allowPublish, maturePublish []byte // see comment under GetProfile

	// we always seem to get checkboxes as a group inside an array, so we do something similar as above with the bitmasks
	//  however
	for _, publish := range(oneProfile.ProfilePublish) {
		switch publish {
			case "Allow":
				allowPublish = append(allowPublish, 1)
			case "Mature":
				maturePublish = append(maturePublish, 1)
		}
	}
	if len(allowPublish) == 0 {
		allowPublish = append(allowPublish, 0)
	}
	if len(maturePublish) == 0 {
		maturePublish = append(maturePublish, 0)
	}

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] oneProfile.ProfilePublish is %+v, allowPublish is %+v, maturePublish is %+v\n", oneProfile.ProfilePublish, allowPublish, maturePublish)
	}

	// Update it on database
	result, err := db.Exec("UPDATE userprofile SET profileAllowPublish = ?, profileMaturePublish = ?, profileURL = ?, profileWantToMask = ?, profileWantToText = ?, profileSkillsMask = ?, profileSkillsText = ?, profileLanguages = ?, profileAboutText = ?, profileFirstText = ? WHERE useruuid = ?",
		// oneProfile.ProfilePartner,
		allowPublish,
		maturePublish,
		oneProfile.ProfileURL,
		wantToMask,						// oneProfile.ProfileWantToMask,	// images are read-only!
		oneProfile.ProfileWantToText,
		skillsMask,						// oneProfile.ProfileSkillsMask,
		oneProfile.ProfileSkillsText,
		oneProfile.ProfileLanguages,
		oneProfile.ProfileAboutText,
		oneProfile.ProfileFirstText,
		oneProfile.UserUUID,
	)

	checkErr(err)

	if numRowsAffected, err := result.RowsAffected(); err != nil {
		c.HTML(http.StatusOK, "404.tpl", environment(c, gin.H{
			"errorcode"		: http.StatusInternalServerError,
			"errortext"		: "Saving profile failed",
			"errorbody"		: fmt.Sprintf("Database error was: %q [%d row(s) affected]", err, numRowsAffected),
			"titleCommon"	: *config["titleCommon"] + " - Profile",
		}))

		log.Printf("[ERROR] Updating database with new profile for %q failed, error was %s\n", thisUUID, err)
		// TODO(gwyneth): we
		return
	} else {
		if *config["ginMode"] == "debug" {
			log.Printf("[INFO] Success updating database with new profile for %q, number of rows affected: %d\n", thisUUID, numRowsAffected)
		}
		c.Redirect(http.StatusSeeOther, "/user/profile")
	}
}

// Transformation functions
// These will probably be moved to cache.go or something similar (gwyneth 20200724)
// TODO(gwyneth): Probably split it further in subdirectories
func imageCacheTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] imageCacheTransform: got from KV store key %q transformed into path %v and filename %q\n",
			key, path, path[last])
	}
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}

func imageCacheInverseTransform(pathKey *diskv.PathKey) string {
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] imageCacheInverseTransform: pathKey %v which will be returned as %q\n",
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