package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vharitonsky/iniflags"
//	"html/template"
//	"io"
//	"log"
//	"net"
	"net/http"
//	"net/http/fcgi"
//	"runtime"
	"time"
	syslog "github.com/RackSec/srslog"
)

var (
	local	= flag.String("local", "", "serve as webserver, example: 0.0.0.0:8000")
	DSN		= flag.String("dsn", "", "DSN for calling MySQL database")
	wLog, _	= syslog.Dial("", "", syslog.LOG_ERR, "gOSWI")
)
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
	// start parsing configuration
	iniflags.SetConfigFile("./config.ini")
	iniflags.Parse()
	
	router := gin.Default()
	router.Delims("{{", "}}") // stick to default delims for Go templates
/*	router.SetFuncMap(template.FuncMap{
		"formatAsYear": formatAsYear,
	})*/
	router.LoadHTMLGlob("./templates/*.tpl")
	//router.HTMLRender = createMyRender()

	// Static stuff (will probably do it via nginx)
	router.Static("/lib", "./lib")
	router.Static("/images", "./images")
	router.StaticFile("/favicon.ico", "./images/favicons/favicon.ico")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tpl", gin.H{
			"now": formatAsYear(time.Now()),
		})
	})

	router.GET("/welcome", func(c *gin.Context) {
		c.HTML(http.StatusOK, "welcome.tpl", gin.H{})
	})
	// the following are not implemented yet
	router.GET("/economy", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{})
	})
	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.tpl", gin.H{})
	})
	router.GET("/help", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{})
	})
	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{})
	})
	router.GET("/password", func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{})
	})
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{})
	})

	if *local == "" {
		router.Run(":8033")
	} else {
		router.Run(*local)
	}
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

	if *local != "" { // Run as a local web server
		wLog.Info("Run as local web server")
		err = http.ListenAndServe(*local, r)
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