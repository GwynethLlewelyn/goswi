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

/*
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
*/

	// attempting a new method!

	// see if we have this image already
	// Note: in the future, we might simplify the call by just using the UUID + file extension... (gwyneth 20200727)
	profileFirstImage := filepath.Join(PathToStaticFiles, "/", *config["cache"], profileData.ProfileFirstImage + *config["jp2convertExt"])
/*
	if profileFirstImage[0] != '/' {
		profileFirstImage = "/" + profileFirstImage
	}
*/
	// either this URL exists and is in the cache, or not, and we need to get the image from
	//  OpenSimulator and attempt to convert it... we won't change the URL in the process.
	// Note: Other usages of the diskv cache might not be so obvious... or maybe they all are? (gwyneth 20200727)
	if !imageCache.Has(profileFirstImage) { // this URL is not in the cache yet!
		if *config["ginMode"] == "debug" {
			log.Println("[INFO] Cache miss on profileFirstImage:", profileFirstImage, " - attempting to download it...")
		}
		// get it!
		profileFirstImageAssetURL := *config["assetServer"] + path.Join("/assets/", profileData.ProfileFirstImage, "/data")
		resp, err := http.Get(profileFirstImageAssetURL)
		defer resp.Body.Close()
		if err != nil {
			// handle error
			log.Println("[ERROR] Oops — OpenSimulator cannot find", profileFirstImageAssetURL)
		}
		newImage, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("[ERROR] Oops — could not get contents of", profileFirstImageAssetURL, "from OpenSimulator")
		}
		if len(newImage) == 0 {
			log.Println("[ERROR] Image retrieved from OpenSimulator", profileFirstImageAssetURL, "has zero bytes.")
			// we might have to get out of here
		} else {
			if *config["ginMode"] == "debug" {
				log.Println("[INFO] Image retrieved from OpenSimulator", profileFirstImageAssetURL, "has", len(newImage), "bytes.")
			}
		}
		// Now use ImageMagick to convert this image!
		// Note: I've avoided using ImageMagick because it's compiled with CGo, but I can't do better
		//  than this. See also https://stackoverflow.com/questions/38950909/c-style-conditional-compilation-in-golang for a way to prevent ImageMagick to be used.

		convertedImage, err := ImageConvert(newImage, 128, 128, 100)
		if err != nil {
			log.Println("[ERROR] Could not convert", profileFirstImageAssetURL, " - error was:", err)
		}
		if convertedImage == nil || len(convertedImage) == 0 {
			log.Println("[ERROR] Converted image is empty")
		}
		if *config["ginMode"] == "debug" {
			log.Println("[INFO] Image from", profileFirstImageAssetURL, "has", len(convertedImage), "bytes.")
		}

		// put it into KV cache:
		if err := imageCache.Write(profileFirstImage, convertedImage); err != nil {
			log.Println("[ERROR] Could not store converted", profileFirstImage, "in the cache, error was:", err)
		}
	}
		// note that the code will now assume that profileFirstImage does, indeed, have a valid
		//  image URL, and will fail with a broken image (404 error on browser) if it doesn't; thus:
		// TODO(gwyneth): get some sort of default image for when all of the above fails

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
		"ProfileFirstImage"	: profileFirstImage,				// Real life, i.e. 'First Life' image
		"ProfileFirstText"	: profileData.ProfileFirstText,
		"Username"			: username,
		"Libravatar"		: libravatar,
	})
}

// Transformation functions
// These will probably be moved to cache.go or something similar (gwyneth 20200724)

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

// ImageConvert will take sequence of bytes of an image and convert it into a
// another image with minimal compression, possibly resizing it.
// Parameters are []byte of original image, height, width, compression quality
// Returns []byte of converted image
// See https://golangcode.com/convert-pdf-to-jpg/ (gwyneth 20200726)
func ImageConvert(aImage []byte, height, width, compression uint) ([]byte, error) {
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
		return nil, errors.New("Empty image passed to ImageConvert")
	}
	// Now that we have checked all parameters, it's time to setup ImageMagick:
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

    // Load the image into imagemagick
    if err := mw.ReadImageBlob(aImage); err != nil {
		return nil, err
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
		return nil, err
	}

    // Must be *after* ReadImage
    // Flatten image and remove alpha channel, to prevent alpha turning black in jpg
    if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_OFF); err != nil {
        return nil, err
    }

    // Set any compression (100 = max quality)
    if err := mw.SetCompressionQuality(compression); err != nil {
        return nil, err
    }

	// Move to first image
	mw.SetIteratorIndex(0)

    // Convert into PNG
	var formatType string = *config["jp2convertExt"]
	if *config["ginMode"] == "debug" {
		log.Println("[DEBUG] Setting format type to", formatType[1:])
	}
    if err := mw.SetFormat(formatType[1:]); err != nil {
        return nil, err
    }

    // Return []byte for this image
	blob := mw.GetImageBlob()
    return blob, nil
}