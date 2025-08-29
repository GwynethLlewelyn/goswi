package main

// NOTE: Before compiling, make sure you read the instructions regarding
// compilation with ImageMagick 6 vs. 7.

import (
	"flag"
	"html/template"
	"log"
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
	nrgin "github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/peterbourgon/diskv/v3"
	_ "github.com/philippgille/gokv"
	"github.com/philippgille/gokv/syncmap"
	"github.com/vharitonsky/iniflags"
	"gopkg.in/gographics/imagick.v3/imagick"
	//	syslog "github.com/RackSec/srslog"
)

// Global variables
var (
	//	wLog, _	= syslog.Dial("", "", syslog.LOG_ERR | syslog.LOG_LOCAL0, "gOSWI")	// write to syslog.
	PathToStaticFiles, cacheDir string
	GOSWIstore                  syncmap.Store            // this stores tokens for password reset links
	imageCache                  *diskv.Diskv             // and this is the cache for images (gwyneth 20200726)
	slideshow                   []string                 // slideshow is a slice of strings representing all images for the splash-screen slideshow.
	bluemondaySafeHTML          = bluemonday.UGCPolicy() // Initialise bluemonday: this is the standard, we might do it a little more restrictive (gwyneth 20200815)
)

// Full configuration, which can be retrieved via flags, configure file, environment (TBD)...
var config = map[string]*string{
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
	"NewRelicAppName":    flag.String("NewRelicAppName", "", "Name of your New Relic application (empty: disabled)"),
	"NewRelicLicenseKey": flag.String("NewRelicLicenseKey", "", "Your New Relic license key"),
}

// Note: flag.Tail() offers us all parameters at the end of the command line, we will use that to generate a list of images for the slideshow, but we cannot us that using pkg iniflags (gwyneth 20200711).

// main starts here.
func main() {
	// figure out where the configuration is
	_, callerFile, _, _ := runtime.Caller(0)
	PathToStaticFiles = filepath.Dir(callerFile)
	if *config["ginMode"] == "debug" {
		log.Println("[DEBUG] executable path is now ", PathToStaticFiles, " while the callerFile is ", callerFile)
	}

	// check if we have a config.ini on the same path as the binary; if not, try to get it to wherever PathToStaticFiles is pointing to
	iniflags.SetConfigFile(filepath.Join(PathToStaticFiles, "/config.ini"))
	// start parsing configuration
	iniflags.Parse()
	// initialise slideshow (all the URLs should be at the end of the commandline)
	slideshow = strings.Split(*config["slides"], ",")
	if len(slideshow) == 0 {
		slideshow = append(slideshow, "https://source.unsplash.com/K4mSJ7kc0As/700x300", "https://source.unsplash.com/Mv9hjnEUHR4/700x300", "https://source.unsplash.com/oWTW-jNGl9I/700x300")
	} else {
		for i := 0; i < len(slideshow); i++ {
			slideshow[i] = strings.TrimSpace(slideshow[i]) // this will respect the order
		}
	}
	if *config["ginMode"] == "debug" {
		log.Printf("List of %d slide(s) has been set to: %+v", len(slideshow), slideshow)
	}

	// cookieStore MUST be set to a random string! (gwyneth 20200628)
	// we might also check for weak security strings on the configuration
	if *config["cookieStore"] == "" {
		log.Fatal("[ERROR] Empty random string for 'cookieStore'; please set it either on the .INI file or pass it via a flag!\nAborting for security reasons.")
	}

	// prepare Gin router/render — first, set it to debug or release (release is default)
	if *config["ginMode"] == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
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

	// If we have a valid New Relic configuration, add it to the middleware list first (gwyneth 20210422)
	// @see https://github.com/newrelic/go-agent/blob/v3.11.0/_integrations/nrgin/v1/example/main.go
	// TODO(gwyneth): get New Relic license key from the environment for extra security (gwyneth 20210422)
	if *config["NewRelicAppName"] != "" && *config["NewRelicLicenseKey"] != "" {
		app, err := newrelic.NewApplication(
			newrelic.ConfigAppName(*config["NewRelicAppName"]),
			newrelic.ConfigLicense(*config["NewRelicLicenseKey"]),
			// TODO(gwyneth): figure out how to funnel the logs from New Relic to Gin! (20211111)
			//			newrelic.ConfigDebugLogger(os.Stdout),			// this was sending debug logs to syslog!
			//			newrelic.ConfigInfoLogger(gin.DefaultWriter),	// now sending only info to the gin logger.
			// NO LOGGING, duh! (gwyneth 20210901)
		)
		if nil != err {
			log.Println("Failed to init New Relic", err)
			// os.Exit(1)
		} else {
			router.Use(nrgin.Middleware(app))
		}
	}
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
		log.Println("[WARN] Creating/accessing cache directory", cacheDir, "returned error:", err)
		// we might not be able to use a cache if this doesn't work
		// so we'll try to create a temporary cache instead

		cacheDir = filepath.Join(os.TempDir(), *config["cache"])
		err = os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			log.Println("[WARN] Creating temporary cache directory", cacheDir, "also returned error:", err)
		}
	}

	// Gin router configuration starts here
	// Static stuff (will probably do it via nginx)
	router.Static("/lib", filepath.Join(PathToStaticFiles, "/lib"))
	router.Static("/assets", filepath.Join(PathToStaticFiles, "/assets"))
	if *config["cache"] != "" {
		router.Static("/cache", cacheDir)
		log.Println("[INFO] Cache directory set up at", cacheDir)
	} else {
		log.Println("[ERROR] Could not access or create cache directory with", cacheDir, "— this means there will be trouble ahead... error was (possibly)", err)
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

	// Initialise ImageMagick, which we use to convert JPEG2000 to PNG
	imagick.Initialize()
	defer imagick.Terminate()

	// Deal with the way gOSWI was called, namely if it uses a default port, uses TLS (=HTTPS), etc.
	if *config["local"] == "" {
		if *config["tlsCRT"] != "" && *config["tlsKEY"] != "" {
			err := router.RunTLS(":8033", *config["tlsCRT"], *config["tlsKEY"]) // if it works, it will never return
			if err != nil {
				log.Printf("[WARN] Could not run with TLS; either the certificate %q was not found, or the private key %q was not found, or either [maybe even both] are invalid.\n", *config["tlsCRT"], *config["tlsKEY"])
				log.Println("[INFO] Running _without_ TLS on the usual port")
				log.Fatal(router.Run(":8033"))
			}
		} else {
			log.Println("[INFO] Running with standard HTTP on the usual port, no TLS configuration detected")
			log.Fatal(router.Run(":8033"))
		}
	} else {
		log.Fatal(router.Run(*config["local"]))
	}
	// if we are here, router.Run() failed with an error
	log.Fatal("Boom, something went wrong! (or maybe this was merely stopped, I don't know)")
}

/*

func homeView(w http.ResponseWriter, r *http.Request) {
	wLog.Info("homeView called")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "<html><head><title>It works!</title></head><body><p>FastCGI under Go works!</p></body></html>")
}

func balView(w http.ResponseWriter, r *http.Request) {
	wLog.Info("balView called")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	fastCGIenv := fcgi.ProcessEnv(r)

	io.WriteString(w, "<html><head><title>Balurdio!</title></head><body><p>Returning:</p><p><pre>")
	if fastCGIenv != nil {
		n, err := fmt.Fprintln(w, fastCGIenv)

		// The n and err return values from Fprintln are
		// those returned by the underlying io.Writer.
		if err != nil {
			wLog.Crit(fmt.Sprintf("Fprintln: %v\n", err))
		}
		wLog.Info(fmt.Sprintln(n, "bytes written."))
	}
	io.WriteString(w, "</pre></p></body></html>")
}

func main() {
	var err error

	r := mux.NewRouter()

	r.HandleFunc("/balurdio/", balView)
	r.HandleFunc("/", homeView)

	flag.Parse()

	if *config["local"] != "" { // Run as a local web server
		wLog.Info("Run as local web server")
		err = http.ListenAndServe(*config["local"], r)
	} else { // Run as FCGI via standard I/O

		l, err := net.Listen("unix", "/var/run/fcgiwrap.socket")
		if err != nil {
			wLog.Crit(err.Error())
			log.Fatal(err)
		}
		defer l.Close()


		wLog.Info("Run as FCGI via standard I/O")
		err = fcgi.Serve(nil, r)
	}
	if err != nil {
		wLog.Crit(err.Error())
		log.Fatal(err)
	}
}
*/
