// Implementation 

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
// 	"math/rand"
	"net/http"
	"strings"
// 	"strconv"
	"time"
)

// UserForm is used for capturing form data from the login page and adds decoration for JSON.
// see https://chenyitian.gitbooks.io/gin-tutorials/tdd/2.html
type UserForm struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"-" form:"password" binding:"required"`
	RememberMe string `json:"rememberMe" form:"rememberMe"`
}

// generateSessionToken uses the same approach as OpenSimulator, which is to return a newly created UUID.
func generateSessionToken() string {
	return uuid.New().String()
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
			log.Println("[DEBUG]: user", username, "not in database")
		}
		return false
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Avatar data from database: '%s %s' (%s)", avatarFirstName, avatarLastName, principalID)
	}
	var passwordHash, passwordSalt string
	err = db.QueryRow("SELECT passwordHash, passwordSalt FROM auth WHERE UUID = ?", 
		principalID).Scan(&passwordHash, &passwordSalt)
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG]: password not in database")
		}
		return false
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Authentication data for: '%s %s' (%s): user-submitted password: %q Hash on DB: %q Salt on DB: %q",
			avatarFirstName, avatarLastName, principalID, password, passwordHash, passwordSalt)
	}
	// md5(md5("password") + ":" + passwordSalt) according to http://opensimulator.org/wiki/Auth
	// but the code for OpenSim is different!!
	
	hashed := GetMD5Hash(password + ":" + passwordSalt) // see OpenSimulator source code in /opensim/OpenSim/Services/PasswordAuthenticationService, method Authenticate()
	
	// we'll simplify the above code to a one-liner which will be more legible once we debug this properly! (gwyneth 20200626)

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG]: md5(password + ':' + passwordSalt) = %q, which we must compare with %q", hashed, passwordHash)
		return false
	}
	// compare and see if it matches
	if passwordHash == hashed {
		// authenticated! now set session cookie and do all the magic
		log.Printf("[INFO]: User %q authenticated.", username)
	} else {
		log.Printf("[WARN]: Invalid authentication for %q — either user not found or password is wrong", username)
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
		// probably should do a redirect, we'll see — or at least have a cooler-looking page.
		return
	}	
/*
	username := c.PostForm("username")
	password := c.PostForm("password")
*/
	if strings.TrimSpace(oneUser.Password) == "" {	// this should not happen, as we put the password as 'required' on the decorations
		log.Println("The password can't be empty")
	}
	if *config["ginMode"] == "debug" {
		// warning: this will expose a password!!
		log.Printf("[INFO] User: %q Password: %q Remember me? %q", oneUser.Username, oneUser.Password, oneUser.RememberMe)
	}
	if isUserValid(oneUser.Username, oneUser.Password) {
		c.SetCookie("goswitoken", generateSessionToken(), 3600, "", "", false, false)
		c.SetCookie("goswiusername", oneUser.Username, 3600, "", "", false, false)
	} else {
		 log.Printf("[ERROR] Invalid username/password combination for user %q!", oneUser.Username)
	}
	c.Redirect(http.StatusSeeOther, "/")	// see https://softwareengineering.stackexchange.com/questions/99894/why-doesnt-http-have-post-redirect and https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/303
}

// logout unsets the session/cookie that contains the user authentication data.
func logout(c *gin.Context) {
	c.SetCookie("goswitoken", "", -1, "", "", false, false)
	c.SetCookie("goswiusername", "", -1, "", "", false, false)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// registerNewUser is currently unimplemented but will use Remote Admin to create new users, as opposed to writing to the OpenSimulator database directly.
func registerNewUser(username, password string) (*UserForm, error) {
	return nil, fmt.Errorf("placeholder error")
}

// isUsernameAvailable simply checks the OpenSimulator database table 'UserAccounts' to see if a user exists with this username; note that OpenSimulator considers the username to have two distinct parts, 'first name' and 'last name'.
func isUsernameAvailable(username string) bool {
	// split username into space-separated fields
	theUsername := strings.Fields(strings.TrimSpace(username))
	// We'll just use the first two
	avatarFirstName	:= theUsername[0]
	avatarLastName	:= theUsername[1]
	
	// BUG(gwyneth): I'm not sure this will work if people 'forget' to place a space between first and last name.
	// We might need to do a bit more than this.
	
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var principalID string
	err = db.QueryRow("SELECT PrincipalID FROM UserAccounts WHERE FirstName = ? AND LastName = ?", 
		avatarFirstName, avatarLastName).Scan(&principalID)	// if there is one, this will be nil
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG]: user", username, "not in database")
		}
		return true	// true means: username is available!
	}
	// no errors, username already exists in the database; it's not available, so we return false
	return false
}

// showRegistrationPage is the handler for showing the registration page (currently 404).
func showRegistrationPage(c *gin.Context) {
	// we show a 404 error for now
	c.HTML(http.StatusNotFound, "404.tpl", gin.H{
		"now": formatAsYear(time.Now()),
		"author": *config["author"],
		"description": *config["description"],
		"titleCommon": *config["titleCommon"] + " - Register new user",
	})
}

// register is the method called from the registration page (currently 404).
func register(c *gin.Context) {
	// reply with a 404 for now
	c.HTML(http.StatusNotFound, "404.tpl", gin.H{
		"now": formatAsYear(time.Now()),
		"author": *config["author"],
		"description": *config["description"],
		"titleCommon": *config["titleCommon"] + " - Register new user",
	})
}

ffunc ensureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInInterface, _ := c.Get("Authenticated")
		loggedIn := loggedInInterface.(bool)
		if !loggedIn {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func ensureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInInterface, _ := c.Get("Authenticated")
		loggedIn := loggedInInterface.(bool)
		if loggedIn {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func setUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie("goswitoken"); err == nil || token != "" {
			cookie, err := c.Cookie("goswiusername")

			if err != nil {
				c.Set("Authenticated", "<unknown username>")
			} else {
				c.Set("Authenticated", cookie)
			}
		} else {
			c.Set("Authenticated", false)
		}
	}
}