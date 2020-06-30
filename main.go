package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/vharitonsky/iniflags"
//	"html/template"
//	"io"
	"log"
//	"net"
	"net/http"
//	"net/http/fcgi"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	syslog "github.com/RackSec/srslog"
)

var (
	wLog, _		= syslog.Dial("", "", syslog.LOG_ERR, "gOSWI")
	PathToStaticFiles string
)

var config = map[string]*string	{// just a place to keep them all together
	"local"			: flag.String("local", "", "serve as webserver, example: 0.0.0.0:8000"),
	"dsn"			: flag.String("dsn", "", "DSN for calling MySQL database"),
	"templatePath"	: flag.String("templatePath", "", "Path to where the templates are stored (with trailing slash) - leave empty for autodetect"),
	"ginMode"		: flag.String("ginMode", "debug", "Default is 'debug' (more logging) but you can set it to 'release' (production-level logging)"),
	"tlsCRT"		: flag.String("tlsCRT", "", "Absolute path for CRT certificate for TLS; leave empty for HTTP"),
	"tlsKEY"		: flag.String("tlsKEY", "", "Absolute path for private key for TLS; leave empty for HTTP"),
	"author"		: flag.String("author", "--nobody--", "Author name"),
	"description"	: flag.String("description", "gOSWI", "Description for each page"),
	"titleCommon"	: flag.String("titleCommon", "gOSWI", "Common part of the title for each page (usually the brand)"),
	"cookieStore"	: flag.String("cookieStore", "", "Secret random string required for the cookie store; mandatory!"),
}

// goswiSession represents our encapsulated session. We need to store a bit more than just a string token, so this needs to be encoded later.
/*
type goswiSession struct {
	Username string `form:"username" json:"username"`,
	Email string `form:"email" json:"email"`,
	Libravatar string `form:"libravatar" json:"libravatar"`,
	Token string, // probably never exported
}
*/

// formatAsDate is a function for the templating system, which will be registered below.
func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

// formatAsYear is another function for the templating system, which will be registered below.
func formatAsYear(t time.Time) string {
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
}


// main starts here.
func main() {
	// figure out where the configuration is
	_, callerFile, _, _ := runtime.Caller(0)
	PathToStaticFiles := filepath.Dir(callerFile)
	if *config["ginMode"] == "debug" {
		fmt.Fprintln(os.Stderr, "[DEBUG] executable path is now ", PathToStaticFiles, " while the callerFile is ", callerFile)
	}

	// check if we have a config.ini on the same path as the binary; if not, try to get it to wherever PathToStaticFiles is pointing to	
	iniflags.SetConfigFile(path.Join(PathToStaticFiles, "/config.ini"))
	// start parsing configuration
	iniflags.Parse()

	// cookieStore MUST be set to a random string! (gwyneth 20200628)
	// we might also check for weak security strings on the configuration
	if *config["cookieStore"] == "" {
		log.Fatal("Make sure that a random string for 'cookieStore' is set either on the .INI file or pass it via a flag!\nAborting for security reasons.")	
	}

	// prepare Gin router/render — first, set it to debug or release (debug is default)
	if *config["ginMode"] == "release" { gin.SetMode(gin.ReleaseMode) }
		
	router := gin.Default()
	router.Delims("{{", "}}") // stick to default delims for Go templates
/*	router.SetFuncMap(template.FuncMap{
		"formatAsYear": formatAsYear,
	})*/
	// figure out where the templates are
	if (*config["templatePath"] != "") {
		if (!strings.HasSuffix(*config["templatePath"], "/")) { 
			*config["templatePath"] += "/"
		}
	} else {
		*config["templatePath"] = "/templates/"
	}
	
	router.LoadHTMLGlob(path.Join(PathToStaticFiles, *config["templatePath"], "*.tpl"))
	//router.HTMLRender = createMyRender()
	//	router.Use(setUserStatus())	// this will allow us to 'see' if the user is authenticated or not
	store := cookie.NewStore([]byte(*config["cookieStore"]))	// now using sessions (Gorilla sessions via Gin extension)
	router.Use(sessions.Sessions("goswisession", store))
	
	// Static stuff (will probably do it via nginx)
	router.Static("/lib", path.Join(PathToStaticFiles, "/lib"))
	router.Static("/assets", path.Join(PathToStaticFiles, "/assets"))
	router.StaticFile("/favicon.ico", path.Join(PathToStaticFiles, "/assets/favicons/favicon.ico"))
	router.StaticFile("/browserconfig.xml", path.Join(PathToStaticFiles, "/assets/favicons/browserconfig.xml"))
	router.StaticFile("/site.webmanifest", path.Join(PathToStaticFiles, "/assets/favicons/site.webmanifest"))

	router.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)

		c.HTML(http.StatusOK, "index.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"titleCommon"	: *config["titleCommon"] + " - Home",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})

	router.GET("/welcome", GetStats)
	router.GET("/about", func(c *gin.Context) {
		session := sessions.Default(c)

		c.HTML(http.StatusOK, "about.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
//			"needsTables"	: true,	// not really needed? (gwyneth 20200612)
			"titleCommon"	: *config["titleCommon"] + " - About",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})
	router.GET("/help", func(c *gin.Context) {
		session := sessions.Default(c)

		c.HTML(http.StatusOK, "help.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"titleCommon"	: *config["titleCommon"] + " - Help",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})
	// the following are not implemented yet
	router.GET("/economy", func(c *gin.Context) {
		session := sessions.Default(c)

		c.HTML(http.StatusNotFound, "404.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"titleCommon"	: *config["titleCommon"] + " - Economy",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})
	router.GET("/search", func(c *gin.Context) {
		session := sessions.Default(c)

		c.HTML(http.StatusNotFound, "404.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"titleCommon"	: *config["titleCommon"] + " - Economy",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})

	userRoutes := router.Group("/user")
	{
		userRoutes.POST("/register",	ensureNotLoggedIn(), registerNewUser)
		userRoutes.GET("/register",		ensureNotLoggedIn(), func(c *gin.Context) {
			session := sessions.Default(c)
	
			// we show a 404 error for now
			c.HTML(http.StatusOK, "404.tpl", gin.H{
				"errorcode"		: http.StatusForbidden,
				"errortext"		: "Access denied",
				"errorbody"		: "Sorry, this grid is not accepting new registrations.",
				"now"			: formatAsYear(time.Now()),
				"author"		: *config["author"],
				"description"	: *config["description"],
				"titleCommon"	: *config["titleCommon"] + " - Register new user",
				"logintemplate"	: false,
				"Username"		: session.Get("Username"),
				"Libravatar"	: session.Get("Libravatar"),
			})
		})
		userRoutes.POST("/change-password",	ensureLoggedIn(), changePassword)
		userRoutes.GET("/change-password",	ensureLoggedIn(), func(c *gin.Context) {
			session := sessions.Default(c)

			c.HTML(http.StatusOK, "change-password.tpl", gin.H{
				"now"			: formatAsYear(time.Now()),
				"author"		: *config["author"],
				"description"	: *config["description"],
				"titleCommon"	: *config["titleCommon"] + " - Change Password",
				"logintemplate"	: true,
				"Username"		: session.Get("Username"),
				"Libravatar"	: session.Get("Libravatar"),
			})
		})
		userRoutes.POST("/reset-password",	ensureNotLoggedIn(), resetPassword)
		userRoutes.GET("/reset-password",	ensureNotLoggedIn(), func(c *gin.Context) {
			session := sessions.Default(c)

			c.HTML(http.StatusOK, "reset-password.tpl", gin.H{
				"now"			: formatAsYear(time.Now()),
				"author"		: *config["author"],
				"description"	: *config["description"],
				"titleCommon"	: *config["titleCommon"] + " - Reset Password",
				"logintemplate"	: true,
				"Username"		: session.Get("Username"),
				"Libravatar"	: session.Get("Libravatar"),
			})
		})
		userRoutes.POST("/login",	ensureNotLoggedIn(), performLogin)
		userRoutes.GET("/login", 	ensureNotLoggedIn(), func(c *gin.Context) {
			session := sessions.Default(c)
				
			c.HTML(http.StatusOK, "login.tpl", gin.H{
				"now"			: formatAsYear(time.Now()),
				"author"		: *config["author"],
				"description"	: *config["description"],
				"Debug"			: false,	// probably unnecessary
				"titleCommon"	: *config["titleCommon"] + "Welcome!",
				"logintemplate"	: true,
				"Username"		: session.Get("Username"),	// very likely not set!!
				"Libravatar"	: session.Get("Libravatar"),
			})
		})
		userRoutes.GET("/logout",	ensureLoggedIn(), logout)
	}
	router.GET("/mapdata", GetMapData)
	router.NoRoute(func(c *gin.Context) {
		session := sessions.Default(c)
		
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"titleCommon"	: *config["titleCommon"] + " - 404",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})
	router.NoMethod(func(c *gin.Context) {
		session := sessions.Default(c)

		c.HTML(http.StatusNotFound, "404.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"titleCommon"	: *config["titleCommon"] + " - 404",
			"Username"		: session.Get("Username"),
			"Libravatar"	: session.Get("Libravatar"),
		})
	})
	// Ping handler (who knows, it might be useful in some contexts... such as Let's Encrypt certificates
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	if *config["local"] == "" {
		if (*config["tlsCRT"] != "" && *config["tlsKEY"] != "") {
			err := router.RunTLS(":8033", *config["tlsCRT"], *config["tlsKEY"]) // if it works, it will never return
			if (err != nil) {
				log.Println("[WARN] Could not run with TLS; either the certificate", *config["tlsCRT"], "was not found, or the private key",
					*config["tlsKEY"], "was not found, or either [maybe even both] are invalid.")
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
	log.Fatal("Boom, something went wrong! (or maybe this was merely stopped, I don't know")
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