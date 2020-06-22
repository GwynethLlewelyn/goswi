// Implementation 

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"strconv"
	"time"
)

// see https://chenyitian.gitbooks.io/gin-tutorials/tdd/2.html
type UserForm struct {
    Username string `json:"username" form:"username" binding:"required"`
    Password string `json:"-" form:"password" binding:"required"`
    RememberMe string `json:"rememberMe" form:"rememberMe"`
}

// generateSessionToken generates a session token; we will need to do something better than this.
func generateSessionToken() string {
    // We're using a random 16 character string as the session token
    // This is NOT a secure way of generating session tokens
    // DO NOT USE THIS IN PRODUCTION
    return strconv.FormatInt(rand.Int63(), 16)
}

// isUserValid checks the database for a valid user/pass combination.
func isUserValid(username, password string) bool {
	// split username into space-separated fields
	theUsername := strings.Fields(strings.TrimSpace(username))
	// We'll just use the first two
	avatarFirstName	:= theUsername[0]
	avatarLastName	:= theUsername[1]
	
	// check if user exists; we'll do password checking later, as soon as we figure out how
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var principalID string
	err = db.QueryRow("SELECT PrincipalID FROM UserAccounts WHERE FirstName = ? AND LastName = ?", 
		avatarFirstName, avatarLastName).Scan(&principalID)	// there can be only one, or our database is corrupted
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG]", username, "not in database")
		}
		return false
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Avatar data from database: '%s %s' (%s)", avatarFirstName, avatarLastName, principalID)
	}
	return true
}

// showLoginPage does exactly what it says, being retrieved with a simple GET request.
func showLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"Debug"			: false,
		"titleCommon"	: *config["titleCommon"] + "Welcome!",
		"logintemplate"	: true,
	})
}

// performLogin is what the form above will call as the method to pass username/password.
func performLogin(c *gin.Context) {
	var oneUser UserForm
	
	if c.Bind(&oneUser) != nil { // nil means no errors
//	if err := c.ShouldBind(&oneUser); err != nil {	// this should be working; it's the 'prefered' way. But it doesn't! D'uh! (gwyneth 20200623)
		c.String(http.StatusBadRequest, "Bad request, no post data found")
		// probably should do a redirect, we'll see
		return
	}
	
/*
	username := c.PostForm("username")
    password := c.PostForm("password")
*/
    if strings.TrimSpace(oneUser.Password) == "" {	// this should not happen, as we put the password as 'required' on the decorations
        log.Println("The password can't be empty")
    }
    log.Printf("[INFO] User: '%s' Password: '%s' Remember me? '%s'", oneUser.Username, oneUser.Password, oneUser.RememberMe)
    log.Printf("[INFO] Is this user valid? '%t'", isUserValid(oneUser.Username, oneUser.Password))
    c.Redirect(http.StatusSeeOther, "/")	// see https://softwareengineering.stackexchange.com/questions/99894/why-doesnt-http-have-post-redirect and https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/303
}

// logout will (eventually) kill the session/cookie that contains the user authentication data.
func logout(c *gin.Context) {}


func registerNewUser(username, password string) (*UserForm, error) {
    return nil, fmt.Errorf("placeholder error")
}

func isUsernameAvailable(username string) bool {
    return false
}

func showRegistrationPage(c *gin.Context) {
	// we show a 404 error for now
	c.HTML(http.StatusNotFound, "404.tpl", gin.H{
		"now": formatAsYear(time.Now()),
		"author": *config["author"],
		"description": *config["description"],
		"titleCommon": *config["titleCommon"] + " - Register new user",
	})
}

func register(c *gin.Context) {
	// reply with a 404 for now
	c.HTML(http.StatusNotFound, "404.tpl", gin.H{
		"now": formatAsYear(time.Now()),
		"author": *config["author"],
		"description": *config["description"],
		"titleCommon": *config["titleCommon"] + " - Register new user",
	})
}