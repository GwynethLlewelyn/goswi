package main

// NOTE: Before compiling, make sure you read the instructions regarding
// compilation with ImageMagick 6 vs. 7.

import (
	"flag"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/peterbourgon/diskv/v3"
	_ "github.com/philippgille/gokv"
	"github.com/philippgille/gokv/syncmap"
	"github.com/vharitonsky/iniflags"
)

// Global variables

var (
	//	wLog, _	= syslog.Dial("", "", syslog.LOG_ERR | syslog.LOG_LOCAL0, "gOSWI")	// write to syslog.
	PathToStaticFiles, cacheDir string
	GOSWIstore                  syncmap.Store            // Stores tokens for password reset links.
	imageCache                  *diskv.Diskv             // As the name says: this is the image cache (gwyneth 20200726)
	slideshow                   []string                 // Slideshow is a slice of strings representing all images for the splash-screen slideshow.
	bluemondaySafeHTML          = bluemonday.UGCPolicy() // Initialise bluemonday: this is the standard, we might do it a little more restrictive (gwyneth 20200815)
	router                      *gin.Engine              // Declared as global because it's supposed to be a singleton Gin router, available on all files here.
)

// Notification methods (optional, build with `systemd` to include the notification system).
// Since v0.9.3

// Is a valid `systemd` running?
var activeSystemd bool

// Enum of possible notification types sent by app to systemd.
type notificationType int8

const (
	// Sent when app is launching and loading configuration, not yet ready
	appReloading     notificationType = iota // Configuration being loaded, initialisation being done...)
	appReady                                 // Application is now ready to accept requests.
	appStopping                              // Application had a (planned) shutdown.
	appStoppingError                         // Emergency shutdown, emits fatal error, possibly exiting with code 126.
)

// Configure the complex flagging system, which will also require loading from `config.ini`
// Note: flag.Tail() offers us all parameters at the end of the command line, we will use that to generate a list of images for the slideshow, but we cannot us that using pkg iniflags (gwyneth 20200711).

// main starts here.
func main() {
	// Full configuration, which can be retrieved via flags, configure file, environment (TBD)...
	// Note: this is now a singleton, assigned on configtype.go
	config = Config{
		"local":              flag.String("local", "", "serve as webserver, example: 0.0.0.0:8000"),
		"dsn":                flag.String("dsn", "", "DSN for calling MySQL database"),
		"templatePath":       flag.String("templatePath", "", "Path to where the templates are stored (with trailing slash) - leave empty for autodetect"),
		"ginMode":            flag.String("ginMode", "release", "Default is 'release' (production-level logging) but you can set it to 'debug' (more logging)"),
		"tlsCRT":             flag.String("tlsCRT", "", "Absolute path for CRT certificate for TLS; leave empty for HTTP"),
		"tlsKEY":             flag.String("tlsKEY", "", "Absolute path for private key for TLS; leave empty for HTTP"),
		"author":             flag.String("author", "--nobody--", "Author name"),
		"description":        flag.String("description", "gOSWI", "Description for each page"),
		"titleCommon":        flag.String("titleCommon", "gOSWI", "Common part of the title for each page (usually the brand)"),
		"cookieStore":        flag.String("cookieStore", randomBase64String(64), "Secret random string required for the cookie store (will be generated randomly if unset)"),
		"SMTPhost":           flag.String("SMTPhost", "localhost", "Hostname of the SMTP server (for sending password reset tokens via email)"),
		"gOSWIemail":         flag.String("gOSWIemail", "manager@localhost", "Email address for the grid manager (must be valid and accepted by SMTPhost)"),
		"gOSWIpassword":      flag.String("gOSWIpassword", "", "Password for the grid manager (must be valid and accepted by SMTPhost)"),
		"logo":               flag.String("logo", "/assets/logos/gOSWI%20logo.svg", "Logo (SVG preferred); defaults to gOSWI logo"),
		"logoTitle":          flag.String("logoTitle", "gOSWI", "Title for the URL on the logo"),
		"sidebarCollapsed":   flag.String("sidebarCollapsed", "false", "true for a collapsed sidebar on startup"),
		"slides":             flag.String("slides", "", "Comma-separated list of URLs for slideshow images"),
		"convertExt":         flag.String("convertExt", ".png", "Filename extension or type for cached resources (depends on the converter actually supporting this particular extension; if not, conversion will fail)"),
		"cache":              flag.String("cache", "/cache/", "File path to the assets cache"),
		"assetServer":        flag.String("assetServer", "http://localhost:8003", "URL to OpenSimulator asset server (no trailing slash)"),
		"ROBUSTserver":       flag.String("ROBUSTserver", "http://localhost:8002", "URL to OpenSimulator ROBUST server (no trailing slash)"),
		"gridstats":          flag.String("gridstats", "/stats", "Relative path to where the Grid statistics are stored (default: /stats)"),
		"ImageMagickCommand": flag.String("ImageMagickCommand", "", "Absolute path to ImageMagick command `imagick`; empty means search $PATH"),
		"ImageMagickParams":  flag.String("ImageMagickParams", "", "Parameters to be sent to ImageMagick command (EXPERIMENTAL!)"),
		"NewRelicAppName":    flag.String("NewRelicAppName", "", "Name of your New Relic application (empty: disabled)"),
		"NewRelicLicenseKey": flag.String("NewRelicLicenseKey", "", "Your New Relic license key"),
		"OTelServiceName":    flag.String("OTelServiceName", "", "Name of your OpenTelemetry service (empty: disabled)"),
		"OTelCollectorURL":   flag.String("OTelCollectorURL", "", "URL for OpenTelemetry Collector (default: none)"),
		"OTelInsecureMode":   flag.String("OTelInsecureMode", "", "OpenTelemetry Insecure Mode (default: empty, i.e. secure mode)"),
	}

	notify(appReloading)

	// figure out where the configuration is
	_, callerFile, _, _ := runtime.Caller(0)
	PathToStaticFiles = filepath.Dir(callerFile)
	config.LogDebug("executable path is now ", PathToStaticFiles, " while the callerFile is ", callerFile)

	// check if we have a config.ini on the same path as the binary; if not, try to get it to wherever PathToStaticFiles is pointing to
	iniflags.SetConfigFile(filepath.Join(PathToStaticFiles, "/config.ini"))
	// start parsing configuration
	iniflags.Parse()
	// initialise slideshow (all the URLs should be at the end of the commandline)
	slideshow = strings.Split(*config["slides"], ",")
	if len(slideshow) == 0 {
		slideshow = append(
			slideshow,
			"https://source.unsplash.com/K4mSJ7kc0As/700x300",
			"https://source.unsplash.com/Mv9hjnEUHR4/700x300",
			"https://source.unsplash.com/oWTW-jNGl9I/700x300",
		)
	} else {
		for i := range len(slideshow) {
			slideshow[i] = strings.TrimSpace(slideshow[i]) // this will respect the order
		}
	}
	config.LogTracef("%d slide(s) have been set to: %+v", len(slideshow), slideshow)

	// cookieStore MUST be set to a random string! (gwyneth 20200628)
	// we might also check for weak security strings on the configuration
	if *config["cookieStore"] == "" {
		config.LogFatal("[ERROR] Empty random string for 'cookieStore'; please set it either on the .INI file or pass it via a flag!\nAborting for security reasons.")
	}

	// prepare Gin router/render — first, set it to debug or release (release is default).
	// Note: incidentally, this will actually also set the logging level.
	if *config["ginMode"] == "debug" || *config["ginMode"] == "trace" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()
	router.Delims("{{", "}}") // stick to default delims for Go templates
	router.SetFuncMap(template.FuncMap{
		"bitTest": bitTest,
	})
	// figure out where the templates are
	if *config["templatePath"] != "" {
		if !strings.HasSuffix(*config["templatePath"], "/") {
			*config["templatePath"] += "/"
		}
	} else {
		*config["templatePath"] = "/templates/"
	}

	router.LoadHTMLGlob(filepath.Join(PathToStaticFiles, *config["templatePath"], "*.tpl"))
	//router.HTMLRender = createMyRender()
	//	router.Use(setUserStatus())	// this will allow us to 'see' if the user is authenticated or not

	// If we have a telemetry option compiled in (New Relic, Open Telemetry...) we call it hre.
	// Note that this is *not* a requirement, the default is a no-op.
	// But if set, it will add some middleware to Gin.
	initTelemetry()

	store := memstore.NewStore([]byte(*config["cookieStore"])) // now using sessions (Gorilla sessions via Gin extension) stored in memory (gwyneth 20200812)
	router.Use(sessions.Sessions("goswisession", store))

	// Initialise the diskv storage on the cache directory (gwyneth 20200724)
	imageCache = diskv.New(diskv.Options{
		// BasePath:		  *config["cache"],
		BasePath:          PathToStaticFiles,
		AdvancedTransform: imageCacheTransform, // currently defined on profile.go (gwyneth 20200724)
		InverseTransform:  imageCacheInverseTransform,
		CacheSizeMax:      100 * 1024 * 1024, // possibly will become a config.ini option
	})

	// Prepare a directory for the cache (i.e. create it if it doesn't exist) (gwyneth 20200718)
	// Note: in the future we might use diskv for the cache and pretty much ignore this
	// Note 2: We *are* using diskv for the cache, but allegedly this is still 'needed'. (gwyneth 20200909)
	cacheDir := filepath.Join(PathToStaticFiles, *config["cache"])
	err := os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		config.LogWarn("Creating/accessing cache directory", cacheDir, "returned error:", err)
		// we might not be able to use a cache if this doesn't work
		// so we'll try to create a temporary cache instead

		cacheDir = filepath.Join(os.TempDir(), *config["cache"])
		err = os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			config.LogWarn("Creating temporary cache directory", cacheDir, "also returned error:", err)
		}
	}

	// Gin router configuration starts here
	// Static stuff (will probably do it via nginx)
	router.Static("/lib", filepath.Join(PathToStaticFiles, "/lib"))
	router.Static("/assets", filepath.Join(PathToStaticFiles, "/assets"))
	if *config["cache"] != "" {
		router.Static("/cache", cacheDir)
		config.LogInfo("Cache directory set up at", cacheDir)
	} else {
		config.LogError("Could not access or create cache directory with", cacheDir, "— this means there will be trouble ahead... error was (possibly)", err)
	}
	router.StaticFile("/favicon.ico", filepath.Join(PathToStaticFiles, "/assets/favicons/favicon.ico"))
	router.StaticFile("/browserconfig.xml", filepath.Join(PathToStaticFiles, "/assets/favicons/browserconfig.xml"))
	router.StaticFile("/site.webmanifest", filepath.Join(PathToStaticFiles, "/assets/favicons/site.webmanifest"))

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tpl", environment(c,
			gin.H{
				"titleCommon": *config["titleCommon"] + " - Home",
			}))
	})

	router.GET("/welcome", GetStats)
	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.tpl", environment(c,
			gin.H{
				"titleCommon": *config["titleCommon"] + " - About",
			}))
	})
	router.GET("/help", func(c *gin.Context) {
		c.HTML(http.StatusOK, "help.tpl", environment(c,
			gin.H{
				"titleCommon": *config["titleCommon"] + " - Help",
			}))
	})
	// the following are not implemented yet
	router.GET("/economy", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", environment(c,
			gin.H{
				"titleCommon": *config["titleCommon"] + " - Economy",
			}))
	})
	router.GET("/search", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", environment(c,
			gin.H{
				"titleCommon": *config["titleCommon"] + " - Search results",
			}))
	})

	// adding routers for grid statistics; this requires a bit of tweaking, since:
	// 1) we may have conflicting /stats directories (i.e. /stats for the reverse proxy server itself);
	// 2) We need to handle both the case when someone wants the HTML page, and the cases when requiring a parameter for JSON, XML, YAML etc.;
	// 3) Gin does not give us the headers by default, so we need to add middleware for that;
	// 4) The middleware seems to be buggy and the default returns the IP address of the client, not of the server!
	//
	var locationMiddleware = location.New(location.Config{
		Headers: location.Headers{Host: "X-Forwarded-Host"},
	})

	router.GET(*config["gridstats"]+"/:ResponseFormatType", locationMiddleware, OSSimpleStats) // plus middleware to get hostname
	router.GET(*config["gridstats"], locationMiddleware, OSSimpleStats)                        // router without any specific format returns HTML (hopefully)

	userRoutes := router.Group("/user")
	{
		userRoutes.POST("/register", ensureNotLoggedIn(), registerNewUser)
		userRoutes.GET("/register", ensureNotLoggedIn(), func(c *gin.Context) {
			// we show a 404 error for now
			c.HTML(http.StatusOK, "404.tpl", environment(c,
				gin.H{
					"errorcode":     http.StatusForbidden,
					"errortext":     "Access denied",
					"errorbody":     "Sorry, this grid is not accepting new registrations.",
					"titleCommon":   *config["titleCommon"] + " - Register new user",
					"logintemplate": false,
				}))
		})
		userRoutes.POST("/change-password", ensureLoggedIn(), changePassword)
		userRoutes.GET("/change-password", ensureLoggedIn(), func(c *gin.Context) {
			c.HTML(http.StatusOK, "change-password.tpl", environment(c, gin.H{
				"titleCommon":   *config["titleCommon"] + " - Change Password",
				"logintemplate": true,
			}))
		})
		userRoutes.POST("/reset-password", ensureNotLoggedIn(), resetPassword)
		userRoutes.GET("/reset-password", ensureNotLoggedIn(), func(c *gin.Context) {
			c.HTML(http.StatusOK, "reset-password.tpl", environment(c, gin.H{
				"titleCommon":   *config["titleCommon"] + " - Reset Password",
				"logintemplate": true,
			}))
		})
		userRoutes.GET("/token/:token", ensureNotLoggedIn(), checkTokenForPasswordReset)
		userRoutes.POST("/login", ensureNotLoggedIn(), performLogin)
		userRoutes.GET("/login", ensureNotLoggedIn(), func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.tpl", environment(c, gin.H{
				"Debug":         false, // probably unnecessary
				"titleCommon":   *config["titleCommon"] + "Welcome!",
				"logintemplate": true,
			}))
		})
		userRoutes.GET("/logout", ensureLoggedIn(), logout)
		userRoutes.GET("/profile", ensureLoggedIn(), GetProfile)
		userRoutes.POST("/profile", ensureLoggedIn(), saveProfile)
		userRoutes.GET("/offline-messages", ensureLoggedIn(), getOfflineMessages)
		userRoutes.GET("/feed-messages", ensureLoggedIn(), getFeedMessages)
	}
	router.GET("/mapdata", GetMapData)
	router.GET("/avatar/:hash", Libravatar) // note that there might be further parameters beyond the hash (gwyneth 20200908)

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", environment(c, gin.H{
			"titleCommon": *config["titleCommon"] + " - 404",
		}))
	})
	router.NoMethod(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", environment(c, gin.H{
			"titleCommon": *config["titleCommon"] + " - 404",
		}))
	})
	// Ping handler (who knows, it might be useful in some contexts... such as Let's Encrypt certificates
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// initialize our KV store (gwyneth 20200705)
	// Note that the current implementation is a goroutine-safe, in-memory solution, without persistent storage —
	//  for now, we will not really need that, since it only stores relatively 'temporary' things, and, if all else fails,
	//  you can redo those things again (e.g. tokens for password reset)

	GOSWIstore = syncmap.NewStore(syncmap.DefaultOptions)
	defer GOSWIstore.Close() // according to the developer, stores should be closed when not in usage, since certain store implementations may require an explicit close to deallocate memory, free database resources, etc. (20200705)

	// NOTE: ImageMagick is now initialised separately, with the build tag `imagick`.
	// See imagick_compiled.go and the README.md. (gwyneth 20251027)

	// Inform systemd (if running from systemd) that we've finished initialisation
	// and configuration. and are now ready to start accepting requests.
	notify(appReady)

	// Deal with the way gOSWI was called, namely if it uses a default port, uses TLS (=HTTPS), etc.
	if *config["local"] == "" {
		if *config["tlsCRT"] != "" && *config["tlsKEY"] != "" {
			err := router.RunTLS(":8033", *config["tlsCRT"], *config["tlsKEY"]) // if it works, it will never return
			if err != nil {
				config.LogWarnf("Could not run with TLS; either the certificate %q was not found, or the private key %q was not found, or either [maybe even both] are invalid.\n", *config["tlsCRT"], *config["tlsKEY"])
				config.LogInfo("Running _without_ TLS on the usual port")
				config.LogFatal(router.Run(":8033"))
			}
		} else {
			config.LogInfo("Running with standard HTTP on the usual port, no TLS configuration detected")
			config.LogFatal(router.Run(":8033"))
		}
	} else {
		config.LogFatal(router.Run(*config["local"]))
	}
	// if we are here, router.Run() failed with an error
	config.LogFatal("Boom, something went wrong! (or maybe this was merely stopped, I don't know)")
	notify(appStoppingError)
}
