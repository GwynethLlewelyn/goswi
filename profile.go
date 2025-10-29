package main

import (
	"database/sql"
	"encoding/binary"
	//	 "encoding/json"
	// "errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/peterbourgon/diskv/v3"
	"html/template"
	"io"
	//	jsoniter "github.com/json-iterator/go"
	"net/http"
	// "os"
	// "os/exec"
	"path"
	"path/filepath"
	"strings"
	// "time"
)

type UserProfile struct {
	UserUUID             string   `form:"UserUUID" json:"useruuid"`
	ProfilePartner       string   `form:"ProfilePartner" json:"profilePartner"`
	ProfileAllowPublish  bool     `form:"ProfileAllowPublish" json:"profileAllowPublish"` // inside the database, this is binary(1)
	ProfileMaturePublish bool     `form:"ProfileMaturePublish" json:"profileMaturePublish"`
	ProfilePublish       []string `form:"ProfilePublish" json:"profilePublish"` // seems to be needed; values are Allow and Mature
	ProfileURL           string   `form:"ProfileURL" json:"profileURL"`
	ProfileWantToMask    int      `form:"ProfileWantToMask" json:"profileWantToMask"`
	ProfileWantTo        []string `form:"ProfileWantTo[]"`
	ProfileWantToText    string   `form:"ProfileWantToText" json:"profileWantToText"`
	ProfileSkillsMask    int      `form:"ProfileSkillsMask" json:"profileSkillsMask"`
	ProfileSkills        []string `form:"ProfileSkills[]"`
	ProfileSkillsText    string   `form:"ProfileSkillsText" json:"profileSkillsText"`
	ProfileLanguages     string   `form:"ProfileLanguages" json:"profileLanguages"`
	ProfileImage         string   `form:"ProfileImage" json:"profileImage"`
	ProfileAboutText     string   `form:"ProfileAboutText" json:"profileAboutText"`
	ProfileFirstImage    string   `form:"ProfileFirstImage" json:"profileFirstImage"`
	ProfileFirstText     string   `form:"ProfileFirstText" json:"profileFirstText"`
}

// Flagged by SonarCube as being constanly used, so we'll use a `const` for it now.
// (gwyneth 20251025)
const (
	strImageFromOpenSimulator = "Image retrieved from OpenSimulator"
	strUnitBytes              = "bytes."
	strInTheCacheErrorWas     = "in the cache, error was:"
)

// GetProfile connects to the database, does its magic, and spews out a profile. That's the theory at least.
func GetProfile(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("Username")
	uuid := session.Get("UUID")

	// open database connection
	if *config["dsn"] == "" {
		config.LogFatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var (
		profileData UserProfile
		//		avatarProfileImage string	// constructed URL for the profile image (gwyneth 20200719) Note: not used any longer (gwyneth 20200728)
		// allowPublish, maturePublish []byte // it has to be this way to get around a bug in the mySQL driver which is impossible to fix
		allowPublishInt, maturePublishInt                                                                                 uint64 // intermediary things
		unsafeProfileURL, unsafeProfileWantToText, unsafeProfileLanguages, unsafeProfileAboutText, unsafeProfileFirstText string // user-provided data requiring strict sanitising (gwyneth 20200815)
	)
	// Allegedly, this is the only way to extract a binary(n) type into a variable;
	//  we need these so-called '[]byte buffers' to temporarily store conversion results (gwyneth 20210118)
	allowPublish := make([]byte, binary.MaxVarintLen64)
	maturePublish := make([]byte, binary.MaxVarintLen64)

	err = db.QueryRow("SELECT useruuid, profilePartner, profileAllowPublish, profileMaturePublish, profileURL, profileWantToMask, profileWantToText, profileSkillsMask, profileSkillsText, profileLanguages, profileImage, profileAboutText, profileFirstImage, profileFirstText FROM userprofile WHERE useruuid = ?", uuid).Scan(
		&profileData.UserUUID,
		&profileData.ProfilePartner,
		&allowPublish,  // &profileData.ProfileAllowPublish,
		&maturePublish, // &profileData.ProfileMaturePublish,
		&unsafeProfileURL,
		&profileData.ProfileWantToMask,
		&unsafeProfileWantToText,
		&profileData.ProfileSkillsMask,
		&profileData.ProfileSkillsText,
		&unsafeProfileLanguages,
		&profileData.ProfileImage,
		&unsafeProfileAboutText,
		&profileData.ProfileFirstImage,
		&unsafeProfileFirstText,
	)

	profileData.ProfileURL = bluemondaySafeHTML.Sanitize(unsafeProfileURL)
	profileData.ProfileWantToText = bluemondaySafeHTML.Sanitize(unsafeProfileWantToText)
	profileData.ProfileLanguages = bluemondaySafeHTML.Sanitize(unsafeProfileLanguages)
	profileData.ProfileAboutText = bluemondaySafeHTML.Sanitize(unsafeProfileAboutText)
	profileData.ProfileFirstText = bluemondaySafeHTML.Sanitize(unsafeProfileFirstText)
	//		Since we get a int(1) = byte, I'll stick with a byte... (gwyneth 20210117)
	//		That approach doesn't work, so we'll try using encode/binary instead
	allowPublishInt, _ = binary.Uvarint(allowPublish)
	profileData.ProfileAllowPublish = allowPublishInt > 0 // hm!
	maturePublishInt, _ = binary.Uvarint(maturePublish)
	profileData.ProfileMaturePublish = maturePublishInt > 0

	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		config.LogDebugf("retrieving profile from user %q (%s) failed; database error was %v\n", username, uuid, err)
	} else {
		config.LogDebugf("while retrieving profile, allowPublish is %v while maturePublish is %v\n", allowPublish, maturePublish)
		config.LogDebugf("after some magic, allowPublishInt is %d while maturePublishInt is %d\n", allowPublishInt, maturePublishInt)
		config.LogDebugf("retrieving profile from user %q (%s): %+v\n", username, uuid, profileData)
	}

	// see if we have this image already
	// Note: in the future, we might simplify the call by just using the UUID + file extension... (gwyneth 20200727)
	// Note 2: We *also* retrieve a Retina image and store it in the cache, but we do not specifically check for it. The idea is that proper HTML will deal with selecting the 'correct' image, we only need to check for one of them. Also, if the conversion to Retina fails for some reason, that's not a problem, we'll fall back to whatever has been downloaded so far...
	profileImage := filepath.Join( /* PathToStaticFiles, */ *config["cache"], profileData.ProfileImage+*config["convertExt"])
	profileRetinaImage := filepath.Join( /* PathToStaticFiles, */ *config["cache"], profileData.ProfileImage+"@2x"+*config["convertExt"]) // we need the path, but we won't check for it directly
	/*
		if profileImage[0] != '/' {
			profileImage = "/" + profileImage
		}
	*/
	// either this URL exists and is in the cache, or not, and we need to get the image from
	//  OpenSimulator and attempt to convert it... we won't change the URL in the process.
	// Note: Other usages of the diskv cache might not be so obvious... or maybe they all are? (gwyneth 20200727)
	if !imageCache.Has(profileImage) { // this URL is not in the cache yet!
		config.LogDebug("Cache miss on profileImage:", profileImage, " - attempting to download it...")
		// get it!
		profileImageAssetURL := *config["assetServer"] + path.Join("/assets/", profileData.ProfileImage, "/data")
		resp, err := http.Get(profileImageAssetURL)
		if err != nil {
			// handle error
			config.LogError("Oops — OpenSimulator cannot find", profileImageAssetURL, "error was:", err)
		}
		defer resp.Body.Close()
		newImage, err := io.ReadAll(resp.Body)
		if err != nil {
			config.LogError("Oops — could not get contents of", profileImageAssetURL, "from OpenSimulator, error was:", err)
		}
		if len(newImage) == 0 {
			config.LogError(strImageFromOpenSimulator, profileImageAssetURL, "has zero bytes.")
			// we might have to get out of here
		} else {
			config.LogDebug(strImageFromOpenSimulator, profileImageAssetURL, "has", len(newImage), strUnitBytes)
		}
		// Now use ImageMagick to convert this image!
		// Note: I've avoided using ImageMagick because it's compiled with Cgo, but I can't do better
		//  than this. See also https://stackoverflow.com/questions/38950909/c-style-conditional-compilation-in-golang for a way to prevent ImageMagick to be used.
		convertedImage, retinaImage, err := ImageConvert(newImage, 256, 256, 100)
		if err != nil {
			config.LogError("Could not convert", profileImageAssetURL, " - error was:", err)
		}
		if /* convertedImage == nil || */ len(convertedImage) == 0 {
			config.LogError("Converted image is empty")
		}
		if /* retinaImage == nil || */ len(retinaImage) == 0 {
			config.LogError("Converted Retina image is empty")
		}
		config.LogDebug("Regular image from", profileImageAssetURL, "has", len(convertedImage), "bytes; retina image has", len(retinaImage), strUnitBytes)
		// put it into KV cache:
		if err := imageCache.Write(profileImage, convertedImage); err != nil {
			config.LogError("Could not store converted", profileImage, strInTheCacheErrorWas, err)
		}
		// put Retina image into KV cache as well:
		if err := imageCache.Write(profileRetinaImage, retinaImage); err != nil {
			config.LogError("Could not store retina image", string(retinaImage), strInTheCacheErrorWas, err)
		}
	}
	// note that the code will now assume that profileImage does, indeed, have a valid
	//  image URL, and will fail with a broken image (404 error on browser) if it doesn't; thus:
	// TODO(gwyneth): get some sort of default image for when all of the above fails
	// An idea would be just to get a Libravatar! We have it, after all...

	// Do the same for the profile image for First (=Real) Life. Comments as above!
	profileFirstImage := filepath.Join( /*PathToStaticFiles, */ *config["cache"], profileData.ProfileFirstImage+*config["convertExt"])
	profileRetinaFirstImage := filepath.Join( /* PathToStaticFiles, */ *config["cache"], profileData.ProfileFirstImage+"@2x"+*config["convertExt"])

	if !imageCache.Has(profileFirstImage) {
		config.LogDebug("Cache miss on profileFirstImage:", profileFirstImage, " - attempting to download it...")
		profileFirstImageAssetURL := *config["assetServer"] + path.Join("/assets/", profileData.ProfileFirstImage, "/data")
		resp, err := http.Get(profileFirstImageAssetURL)
		if err != nil {
			config.LogError("Oops — OpenSimulator cannot find", profileFirstImageAssetURL, "error was:", err)
		}
		defer resp.Body.Close()
		newImage, err := io.ReadAll(resp.Body)
		if err != nil {
			config.LogError("Oops — could not get contents of", profileFirstImageAssetURL, "from OpenSimulator, error was:", err)
		}
		if len(newImage) == 0 {
			config.LogError(strImageFromOpenSimulator, profileFirstImageAssetURL, "has zero bytes.")
		} else {
			config.LogDebug(strImageFromOpenSimulator, profileFirstImageAssetURL, "has", len(newImage), strUnitBytes)
		}
		convertedImage, retinaImage, err := ImageConvert(newImage, 128, 128, 100)
		if err != nil {
			config.LogError("Could not convert", profileFirstImageAssetURL, " - error was:", err)
		}
		if /* convertedImage == nil || */ len(convertedImage) == 0 {
			config.LogError("Converted image is empty")
		}
		if /* retinaImage == nil || */ len(retinaImage) == 0 {
			config.LogError("Converted Retina image is empty")
		}
		config.LogDebug("Image from", profileFirstImageAssetURL, "has", len(convertedImage), "bytes; retina image has", len(retinaImage), strUnitBytes)
		if err := imageCache.Write(profileFirstImage, convertedImage); err != nil {
			config.LogError("Could not store converted", profileFirstImage, strInTheCacheErrorWas, err)
		}
		// put Retina image into KV cache as well:
		if err := imageCache.Write(profileRetinaFirstImage, retinaImage); err != nil {
			config.LogError("Could not store retina image", string(retinaImage), strInTheCacheErrorWas, err)
		}
	}

	c.HTML(http.StatusOK, "profile.tpl", environment(c, gin.H{
		"needsTables":             false,
		"needsMap":                false,
		"moreValidation":          true,
		"Debug":                   false, // we will probably need two versions of 'debug mode'... (gwyneth 20200622)
		"titleCommon":             *config["titleCommon"] + profileData.UserUUID + " Profile",
		"ProfileData":             fmt.Sprintf("%+v", profileData),
		"ProfileURL":              template.HTML(profileData.ProfileURL),
		"UserUUID":                profileData.UserUUID,
		"ProfilePartner":          profileData.ProfilePartner,
		"ProfileAllowPublish":     profileData.ProfileAllowPublish,
		"ProfileMaturePublish":    profileData.ProfileMaturePublish,
		"ProfileWantToMask":       profileData.ProfileWantToMask,
		"ProfileWantToText":       template.HTML(profileData.ProfileWantToText),
		"ProfileSkillsMask":       profileData.ProfileSkillsMask,
		"ProfileSkillsText":       template.HTML(profileData.ProfileSkillsText),
		"ProfileLanguages":        template.HTML(profileData.ProfileLanguages),
		"ProfileImage":            profileImage,       // OpenSimulator/Second Life profile image
		"ProfileRetinaImage":      profileRetinaImage, // Generated Retina image
		"ProfileAboutText":        template.HTML(profileData.ProfileAboutText),
		"ProfileFirstImage":       profileFirstImage,       // Real life, i.e. 'First Life' image
		"ProfileRetinaFirstImage": profileRetinaFirstImage, // Another generated Retina image
		"ProfileFirstText":        template.HTML(profileData.ProfileFirstText),
	}))
}

// saveProfile is what gets called when someone saves the profile.
func saveProfile(c *gin.Context) {
	var oneProfile UserProfile

	session := sessions.Default(c)
	thisUUID := session.Get("UUID")

	if c.Bind(&oneProfile) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "404.tpl", environment(c, gin.H{
			"errorcode":   http.StatusBadRequest,
			"errortext":   "Saving profile failed",
			"errorbody":   "No form data posted",
			"titleCommon": *config["titleCommon"] + " - Profile",
		}))
		config.LogError("No form data posted for saving profile")

		return
	}

	config.LogDebugf("oneProfile is now %+v\n", oneProfile)

	// check if we really are who we claim to be
	if thisUUID != oneProfile.UserUUID {
		c.HTML(http.StatusUnauthorized, "404.tpl", environment(c, gin.H{
			"errorcode":   http.StatusUnauthorized,
			"errortext":   "No permission",
			"errorbody":   fmt.Sprintf("You have no permission to change the profile for %q", session.Get("Username")),
			"titleCommon": *config["titleCommon"] + " - Profile",
		}))
		config.LogErrorf("Session UUID %q is not the same as Profile UUID %q - profile data change for %q not allowed\n",
			thisUUID, oneProfile.UserUUID, session.Get("Username"))

		return
	}

	// Allegedly we have successfully bound to the form data, so we can proceed to write it to the database.
	if *config["dsn"] == "" {
		config.LogFatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	// Calculate the masks

	var wantToMask, skillsMask int

	for _, bitfield := range oneProfile.ProfileWantTo {
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

	for _, bitfield := range oneProfile.ProfileSkills {
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

	config.LogDebugf("oneProfile.ProfileWantTo is %v, wantToMask is %d, oneProfile.ProfileSkills is %v, skillsMask is %d\n", oneProfile.ProfileWantTo, wantToMask, oneProfile.ProfileSkills, skillsMask)

	config.LogDebugf("oneProfile.ProfileAllowPublish is %+v, oneProfile.ProfileMaturePublish is %+v\n", oneProfile.ProfileAllowPublish, oneProfile.ProfileMaturePublish)

	allowPublish := make([]byte, binary.MaxVarintLen64)
	maturePublish := make([]byte, binary.MaxVarintLen64)

	_ = binary.PutUvarint(allowPublish, 0)
	_ = binary.PutUvarint(maturePublish, 0)

	// we always seem to get checkboxes as a group inside an array, so we do something similar as above with the bitmasks
	//  however
	for _, publish := range oneProfile.ProfilePublish {
		switch publish {
		case "Allow":
			//				allowPublish = append(allowPublish, 1)
			//				allowPublish = 1
			_ = binary.PutUvarint(allowPublish, 1)
		case "Mature":
			//				maturePublish = append(maturePublish, 1)
			//				maturePublish = 1
			_ = binary.PutUvarint(maturePublish, 1)
		}
	}
	// if len(allowPublish) == 0 {
	// 	allowPublish = append(allowPublish, 0)
	// }
	// if len(maturePublish) == 0 {
	// 	maturePublish = append(maturePublish, 0)
	// }

	config.LogDebugf("oneProfile.ProfilePublish is %+v, allowPublish is %+v, maturePublish is %+v\n", oneProfile.ProfilePublish, allowPublish, maturePublish)

	// Update it on database
	result, err := db.Exec("UPDATE userprofile SET profileAllowPublish = ?, profileMaturePublish = ?, profileURL = ?, profileWantToMask = ?, profileWantToText = ?, profileSkillsMask = ?, profileSkillsText = ?, profileLanguages = ?, profileAboutText = ?, profileFirstText = ? WHERE useruuid = ?",
		// oneProfile.ProfilePartner,
		allowPublish,
		maturePublish,
		bluemondaySafeHTML.Sanitize(oneProfile.ProfileURL),
		wantToMask, // oneProfile.ProfileWantToMask,	// images are read-only!
		bluemondaySafeHTML.Sanitize(oneProfile.ProfileWantToText),
		skillsMask, // oneProfile.ProfileSkillsMask,
		bluemondaySafeHTML.Sanitize(oneProfile.ProfileSkillsText),
		bluemondaySafeHTML.Sanitize(oneProfile.ProfileLanguages),
		bluemondaySafeHTML.Sanitize(oneProfile.ProfileAboutText),
		bluemondaySafeHTML.Sanitize(oneProfile.ProfileFirstText),
		oneProfile.UserUUID,
	)

	checkErr(err)

	if numRowsAffected, err := result.RowsAffected(); err != nil {
		c.HTML(http.StatusOK, "404.tpl", environment(c, gin.H{
			"errorcode":   http.StatusInternalServerError,
			"errortext":   "Saving profile failed",
			"errorbody":   fmt.Sprintf("Database error was: %q [%d row(s) affected]", err, numRowsAffected),
			"titleCommon": *config["titleCommon"] + " - Profile",
		}))

		config.LogErrorf("Updating database with new profile for %q failed, error was %s\n", thisUUID, err)
		// TODO(gwyneth): we
		return
	} else {
		config.LogDebugf("Success updating database with new profile for %q, number of rows affected: %d\n", thisUUID, numRowsAffected)
		c.Redirect(http.StatusSeeOther, "/user/profile")
	}
}

// Transformation functions
// These will probably be moved to cache.go or something similar (gwyneth 20200724)
// TODO(gwyneth): Probably split it further in subdirectories
func imageCacheTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	config.LogDebugf("imageCacheTransform: got from KV store key %q transformed into path %v and filename %q\n",
		key, path, path[last])

	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}

func imageCacheInverseTransform(pathKey *diskv.PathKey) string {
	config.LogDebugf("imageCacheInverseTransform: pathKey %v which will be returned as %q\n",
		pathKey, strings.Join(pathKey.Path, "/")+pathKey.FileName) // inefficient but we're just debugging... (g20200727)
	return strings.Join(pathKey.Path, "/") + pathKey.FileName
}

// NOTE: ImageConvert() has been moved to imagick_compiled.go instead.
// imagick_spawn.go uses a fork to spawn an external process, etc.
