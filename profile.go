package main

import (
	"database/sql"
//	 "encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	"io"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type UserProfile struct {
	UserUUID string 			`form:"useruuid" json:"useruuid"`
	ProfilePartner string		`form:"profilePartner" json:"profilePartner"`
	ProfileAllowPublish bool	`form:"profileAllowPublish" json:"profileAllowPublish"`
	ProfileMaturePublish bool	`form:"profileMaturePublish" json:"profileMaturePublish"`
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
		avatarProfileImage string	// constructed URL for the profile image (gwyneth 20200719)
		allowPublish, maturePublish string // it has to be this way to get around a bug in the mySQL driver which is impossible to fix
	)
	err = db.QueryRow("SELECT useruuid, profilePartner, profileAllowPublish, profileMaturePublish, profileURL, profileWantToMask, profileWantToText, profileSkillsMask, profileSkillsText, profileLanguages, profileImage, profileAboutText, profileFirstImage, profileFirstText FROM userprofile WHERE useruuid = ?", uuid).Scan(
			&profileData.UserUUID,
			&profileData.ProfilePartner,
			&allowPublish,
			&maturePublish,
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
		profileData.ProfileAllowPublish		= (allowPublish != "")
		profileData.ProfileMaturePublish	= (maturePublish != "")
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG]: user %q (%s) has no profile; database error was %v", username, uuid, err)
		}
	}

	// For ProfileImage/ProfileFirstImage we need to convert them from JPEG2000 to something else.
	// cache check first!
	imageAssetFileName := filepath.Join(*config["cache"], profileData.ProfileImage + ".jp2")
	imageFileName := filepath.Join(*config["cache"], profileData.ProfileImage + *config["jp2convertExt"]) // name of converted file
	imageAssetURL := *config["assetServer"] + path.Join("/assets/", profileData.ProfileImage, "/data")

	fd, ferr := os.Open(imageAssetFileName)
	defer fd.Close()
	if ferr != nil {
		// file is not in the cache yet, so grab it and save it
		err := downloadFile(imageAssetFileName, imageAssetURL)
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG] File", imageAssetFileName, "not in cache yet - trying to load it from", imageAssetURL)
		}
		if err != nil {
			log.Println("[WARN] Asset", profileData.ProfileImage, "didn't get saved to cache")
		}
	}
	fd.Close()	// we don't need it any more
	// Now we have this asset in the cache, so, we just need to convert it, but first see if we have already converted it (gwyneth 20200718).
	fd, ferr = os.Open(imageFileName)	// this is the final image /path/to/cache/imageUUID.jpeg
	defer fd.Close()
	if ferr != nil {
		// image not in the cache yet, let's convert it! (gwyneth 20200718)
		// We launch an external command simply because there is no native Go library for JPEG2000 and we're trying to avoid using CGo (we could use it with ImageMagick)
		if *config["jp2convert"] == "" {
			// should not happen, unless a stupid user overwrote it
			log.Println("[ERROR] Empty path to conversion utility! Please add a valid entry for 'jp2convert' in your config.ini")
		} else {
			// first let's fill the %s with the actual filenames
			cmdString := fmt.Sprintf(*config["jp2convert"], imageAssetFileName, imageFileName)
			if *config["ginMode"] == "debug" {
				log.Println("[DEBUG] Command string extracted from config.ini and parsed:", cmdString)
			}
			// now turn it into a []string
			// code comes from https://stackoverflow.com/a/49429437/1035977 by @vahdet
			slc := strings.Split(cmdString, " ")
			for i := range slc {
				slc[i] = strings.TrimSpace(slc[i]) // trim whitespace
			}
			cmd := exec.Command(slc[0], slc[1:]...)
			if *config["ginMode"] == "debug" {
				log.Println("[DEBUG] Converting command is", cmd)
			}
			output, err := cmd.CombinedOutput()	// move
			if err != nil {
				log.Println("[ERROR] Couldn't launch conversion command", err)
			} else {
				if *config["ginMode"] == "debug" {
					log.Printf("[DEBUG] Output from converting command was %q\n", output)
				}
			}
			// we're finished for now; the image is now on the cache; we can close everything
			//  and move on!
			fd.Close()
		}
	}
	// ok, now we can set the image URL pointing to the cached file!
	// even if it doesn't exist, or failed to convert, we will always get something, namely, a broken image, but that's ok, it won't crash the application (gwyneth 20200719).
	// TODO(gwyneth): if the image failed to convert, assign it to a default image (maybe a default image asset from OpenSimulator?) (gwyneth 20200719).
	// note that this is the same as imageFileName, but while imageFileName was constructed with filepath (since it points to a file using the filesystem), avatarProfileImage is constructed with path, since it's supposed to be an URL (gwyneth 20200719)
	// TODO(gwyneth): add debugging on the handler for the cache to see if it is actually being called (nginx might be serving it statically); also, there should be a cache garbage collector if the user changed their profile image.
	avatarProfileImage = path.Join(*config["cache"], profileData.ProfileImage + *config["jp2convertExt"])	// note that hopefully the router is set correctly on main()! (gwyneth 20200719)

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] After conversion, ProfileImage is now %q\n", avatarProfileImage)
	}
	// Because of the way path.Join() 'cleans up' links, we might end up without a leading slash; so we just need to check for that.
	if avatarProfileImage[0] != '/' {
		avatarProfileImage = "/" + avatarProfileImage
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] ProfileImage is now %q\n", avatarProfileImage)
	}

	// use the cache mechanism for it
	cache := Cache{func (imageFileName string) (string, error) {
			// We launch an external command simply because there is no native Go library for JPEG2000 and we're trying to avoid using CGo (we could use it with ImageMagick)
			if *config["jp2convert"] == "" {
				// should not happen, unless a stupid user overwrote it
				log.Println("[ERROR] Empty path to conversion utility! Please add a valid entry for 'jp2convert' in your config.ini")
				return "", errors.New("Empty path to conversion utility")
			} else {
				// we know this exists and has an invalid extension; so we need to replace it later
				// first we remove the old extension
				var imageDestFileName string
				i := strings.LastIndex(imageFileName, ".")
				if i < 0 {
					// no extension, so we create a filename with a new one
					imageDestFileName = imageFileName + *config["jp2convertExt"]
				} else {
					imageDestFileName = imageFileName[:i] + *config["jp2convertExt"]
				}
				// then let's fill the %s with the actual filenames
				cmdString := fmt.Sprintf(*config["jp2convert"], imageFileName, imageDestFileName)
				if *config["ginMode"] == "debug" {
					log.Println("[DEBUG] Command string extracted from config.ini and parsed:", cmdString)
				}
				// now turn it into a []string
				// code comes from https://stackoverflow.com/a/49429437/1035977 by @vahdet
				slc := strings.Split(cmdString, " ")
				for i := range slc {
					slc[i] = strings.TrimSpace(slc[i]) // trim whitespace
				}
				cmd := exec.Command(slc[0], slc[1:]...)
				if *config["ginMode"] == "debug" {
					log.Println("[DEBUG] Converting command is", cmd)
				}
				output, err := cmd.CombinedOutput()	// move
				if err != nil {
					log.Println("[ERROR] Couldn't launch conversion command", err)
				} else {
					if *config["ginMode"] == "debug" {
						log.Printf("[DEBUG] Output from converting command was %q\n", output)
					}
				}
				// we're finished for now; the image is now on the cache
				return string(output), err
			}
			return "", nil
		},
		*config["cache"], *config["cache"],		// seems redundant, but this allows different options
		".jp2", *config["jp2convertExt"],
	}

	firstLifeImage, cerr := cache.Download(*config["assetServer"] + path.Join("/assets/", profileData.ProfileFirstImage, "/data"))	// wicked!! (gwyneth 20200722)
	if cerr != nil {
		if *config["ginMode"] == "debug" {
			log.Println("[WARN] Cache download returned", firstLifeImage, "with error:", cerr)
		}		
	}

	c.HTML(http.StatusOK, "profile.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"needsTables"	: false,
		"needsMap"		: false,
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
		"ProfileImage"		: avatarProfileImage,				// OpenSimulator/Second Life profile image
		"ProfileAboutText"	: profileData.ProfileAboutText,
		"ProfileFirstImage"	: firstLifeImage,	// Real life, i.e. 'First Life' image
		"ProfileFirstText"	: profileData.ProfileFirstText,
		"Username"		: username,
		"Libravatar"	: libravatar,
	})
}