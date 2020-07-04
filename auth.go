// Implementation

package main

import (
	"database/sql"
//	"fmt"
	"github.com/gin-contrib/sessions"
// 	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
// 	"math/rand"
	"net/http"
	"strings"
	"strk.kbt.io/projects/go/libravatar"
	"time"
)

// UserForm is used for capturing form data from the login page and adds decoration for JSON.
// see https://chenyitian.gitbooks.io/gin-tutorials/tdd/2.html
type UserForm struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"-" form:"password" binding:"required"`
	Email string `json:"-" form:"email"`
	RememberMe string `json:"rememberMe" form:"rememberMe"`
}

type ChangePasswordForm struct {
	OldPassword string `json:"-" form:"oldpassword" binding:"required"`
	NewPassword string `json:"-" form:"newpassword" binding:"required"`
	ConfirmNewPassword string `json:"-" form:"confirmnewpassword" binding:"required"`
}

type ResetPasswordForm struct {
	Email string `json:"email" form:"email" binding:"required"`
}

// generateSessionToken uses the same approach as OpenSimulator, which is to return a newly created UUID.
func generateSessionToken() string {
	return uuid.New().String()
}

// isUserValid checks the database for a valid user/pass combination. It returns a boolean. the email and the UUID. Yes, it's ugly.
// TODO(gwyneth): Figure out a better way to extract the email data and store it safely.
// TODO(gwyneth): Probably, this ought to return (string, string, bool) to be a bit more consistent...
//   or even a UserForm, which could be passed as value and returned with extra fields filled in (gwyneth 20200703)
func isUserValid(username, password string) (bool, string, string) {
	// split username into space-separated fields
	theUsername := strings.Fields(strings.TrimSpace(username))
	if len(theUsername) < 2 {
		log.Printf("[WARN] Invalid first name/last name: %q does not contain spaces! Returning with error", username)
		return false, "", ""
	}
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

	var principalID, email string
	err = db.QueryRow("SELECT PrincipalID, Email FROM UserAccounts WHERE FirstName = ? AND LastName = ?",
		avatarFirstName, avatarLastName).Scan(&principalID, &email)	// there can be only one, or our database is corrupted
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG]: user", username, "not in database")
		}
		return false, "", ""
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Avatar data from database: '%s %s' (%s) Email: %q", avatarFirstName, avatarLastName, principalID, email)
	}

	var passwordHash, passwordSalt string
	err = db.QueryRow("SELECT passwordHash, passwordSalt FROM auth WHERE UUID = ?",
		principalID).Scan(&passwordHash, &passwordSalt)
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Println("[DEBUG]: password not in database")
		}
		return false, "", ""
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Authentication data for: '%s %s' (%s): user-submitted password: %q Hash on DB: %q Salt on DB: %q",
			avatarFirstName, avatarLastName, principalID, password, passwordHash, passwordSalt)
	}
	// md5(md5("password") + ":" + passwordSalt) according to http://opensimulator.org/wiki/Auth

	var hashedPassword, hashed, interior string // make sure they are strings, or comparison might fail

	hashedPassword = GetMD5Hash(password)
	interior = hashedPassword + ":" + passwordSalt
	hashed = GetMD5Hash(interior) // see OpenSimulator source code in /opensim/OpenSim/Services/PasswordAuthenticationService, method Authenticate()

	// we'll simplify the above code to a one-liner which will be more legible once we debug this properly! (gwyneth 20200626)

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] md5(password) = %q, (md5(password) + \":\" + passwordSalt) = %q, md5(md5(password) + \":\" + passwordSalt) = %q, which we must compare with %q",
			hashedPassword, interior, hashed, passwordHash)
	}
	// compare and see if it matches
	if passwordHash == hashed {
		// authenticated! now set session cookie and do all the magic
		log.Printf("[INFO] User %q (%v) authenticated (email: <%s>).", username, principalID, email)
		return true, email,	principalID
	} else {
		log.Printf("[WARN] Invalid authentication for %q — either user not found or password is wrong", username)
		return false, "", ""
	}

	return true, email, principalID
}

// performLogin is what the login form will call as the method to pass username/password.
func performLogin(c *gin.Context) {
	var oneUser UserForm
	session := sessions.Default(c)

	defer func(){
	// TODO(gwyneth): we ought to deal with an empty email and/or an empty avatar_url
		if *config["ginMode"] == "debug" {
			log.Printf("[INFO] Session for %q set: token %q - Also, Libravatar is %q", session.Get("Username"), session.Get("Token"), session.Get("Libravatar"))
		}
	}()

	if c.Bind(&oneUser) != nil { // nil means no errors
//	if err := c.ShouldBind(&oneUser); err != nil {	// this should be working; it's the 'prefered' way. But it doesn't! D'uh! (gwyneth 20200623)
		c.HTML(http.StatusBadRequest, "login.tpl", gin.H{
			"ErrorTitle"	: "Login Failed",
			"ErrorMessage"	: "No form data posted",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "What?",
			"logintemplate"	: true,
		})
		log.Println("No form data posted")

		return
	}
	if strings.TrimSpace(oneUser.Password) == "" {	// this should not happen, as we put the password as 'required' on the decorations
		c.HTML(http.StatusBadRequest, "login.tpl", gin.H{
			"ErrorTitle"	: "Login Failed",
			"ErrorMessage"	: "Empty password, please try again",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "Oh, No!",
			"logintemplate"	: true,
			"WrongUsername"	: oneUser.Username,
			"WrongRememberMe" : oneUser.RememberMe,
		})

		return

		log.Println("The password can't be empty")
	}
	if *config["ginMode"] == "debug" {
		// warning: this will expose a password!!
		log.Printf("[INFO] User: %q Password: %q Remember me? %q", oneUser.Username, oneUser.Password, oneUser.RememberMe)
	}
	if ok, email, uuid := isUserValid(oneUser.Username, oneUser.Password); ok {
		session.Set("Username", oneUser.Username)
		session.Set("UUID", uuid)
		session.Set("Token", generateSessionToken())

		if email != "" {
//			avt.SetSecureFallbackHost("unicornify.pictures")	// possibly not needed, we'll implement it locally
			avt := libravatar.New()
			avt.SetAvatarSize(60)	// for some silly reason, that's what our template has...
			avt.SetUseHTTPS(true)
			avt.SetSecureFallbackHost("unicornify.pictures")
			if avatar_url, err := avt.FromEmail(email); err == nil {
				session.Set("Libravatar", avatar_url)
			} else {
				// couldn't get an image url from the Libravatar service, so get an Unicorn instead!
				session.Set("Libravatar", "https://unicornify.pictures/avatar/" + GetMD5Hash(oneUser.Username) + "?s=60")
				session.Set("Email", email)	// who knows, it might be useful at some point
			}
		} else {
			// if we don't have a valid email, get an Unicorn!
			session.Set("Libravatar", "https://unicornify.pictures/avatar/" + GetMD5Hash(oneUser.Username) + "?s=60")
		}
		session.Set("RememberMe", oneUser.RememberMe)
		session.Save()
	} else {
		// invalid user, do not set cookies!
		log.Printf("[ERROR] Invalid username/password combination for user %q!", oneUser.Username)

		c.HTML(http.StatusBadRequest, "login.tpl", gin.H{
			"ErrorTitle"	: "Login Failed",
			"ErrorMessage"	: "Invalid credentials provided",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "Oh, No!",
			"logintemplate"	: true,
			"WrongUsername"	: oneUser.Username,
			"WrongPassword"	: oneUser.Password,
			"WrongRememberMe" : oneUser.RememberMe,
		})

		return
	}
	c.Redirect(http.StatusSeeOther, "/")	// see https://softwareengineering.stackexchange.com/questions/99894/why-doesnt-http-have-post-redirect and https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/303
}

// logout unsets the session/cookie that contains the user authentication data.
func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Set("Username", "") // this will mark the session as "written" and hopefully remove the username
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0,
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, "/")
//	c.Redirect(http.StatusFound, "/")	// see https://github.com/gin-contrib/sessions/issues/29#issuecomment-376382465
}

// registerNewUser is currently unimplemented (too dangerous).
func registerNewUser(c *gin.Context) {
	var oneUser UserForm	// similar to performLogin()
//	session := sessions.Default(c)	// should not have any session

	if c.Bind(&oneUser) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "register.tpl", gin.H{
			"ErrorTitle"	: "Registration Failed",
			"ErrorMessage"	: "No form data posted",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "What?",
			"logintemplate"	: true,
		})
		log.Println("No form data posted")

		return
	}
//(username, password string) (*UserForm, error) {
	log.Printf("[INFO] Not implemented yet")

	c.HTML(http.StatusBadRequest, "register.tpl", gin.H{
		"ErrorTitle"	: "Registration Failed",
		"ErrorMessage"	: "No new users allowed!",
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"Debug"			: false,
		"titleCommon"	: *config["titleCommon"] + "Sorry!",
		"logintemplate"	: true,
		"WrongUsername"	: oneUser.Username,
		"WrongPassword"	: oneUser.Password,
		"WrongEmail"	: oneUser.Email,
	})

	return
}

func changePassword(c *gin.Context) {
//(oldpassword, newpassword, newpasswordverify string) (*UserForm, error) {
	var aPasswordChange ChangePasswordForm
	// session := sessions.Default(c)

	if c.Bind(&aPasswordChange) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "change-password.tpl", gin.H{
			"ErrorTitle"	: "Password change failed",
			"ErrorMessage"	: "No form data posted",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "Say what?",
			"logintemplate"	: true,
		})
		log.Println("No form data posted")

		return
	}
}

func resetPassword(c *gin.Context) {
//email string) (*UserForm, error) {
	var aPasswordReset ResetPasswordForm
	//session := sessions.Default(c)

	if c.Bind(&aPasswordReset) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "reset-password.tpl", gin.H{
			"ErrorTitle"	: "Password reset failed",
			"ErrorMessage"	: "No form data posted",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "Whut?",
			"logintemplate"	: true,
		})
		log.Println("No form data posted")

		return
	}
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

/***
*	Middleware for dealing with login/session cookies
*/

// ensureLoggedIn tests if the user is logged in, reading in from the context to see if a flag is set.
// Note that this flag is not a boolean any more, I'm using this pseudo-flag to store the username
func ensureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		loggedInInterface := session.Get("Username")
		if loggedInInterface == nil || loggedInInterface == "" {
			if *config["ginMode"] == "debug" {
				log.Printf("[INFO]: ensureNotLoggedIn(): No authenticated user")
			}
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			if *config["ginMode"] == "debug" {
				log.Printf("[INFO]: ensureNotLoggedIn(): Username is %q", loggedInInterface)
			}
		}
	}
}

// ensureNotLoggedIn tests if the user is NOT logged in, reading in from the context to see if a flag is set.
func ensureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		loggedInInterface := session.Get("Username")
		if loggedInInterface != nil && loggedInInterface != "" {
			if *config["ginMode"] == "debug" {
				log.Printf("[INFO]: ensureNotLoggedIn(): Username is %q", loggedInInterface)
			}
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			if *config["ginMode"] == "debug" {
				log.Printf("[INFO]: ensureNotLoggedIn(): No authenticated user")
			}
		}
	}
}

// setUserStatus gets loaded for each page, and sees if the cookie is set. This seems to be the 'correct' way to do this under Gin.
func setUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// Note that all the things below may set everything to empty strings, which is FINE! (gwyneth 20200628)
		c.Set("Username",	session.Get("Username"))
		c.Set("Email", 		session.Get("Email"))
		c.Set("Libravatar",	session.Get("Libravatar"))
		c.Set("Token",		session.Get("Token"))
		c.Set("UUID",		session.Get("UUID"))
		c.Set("RememberMe",	session.Get("RememberMe"))

		if *config["ginMode"] == "debug" {
			log.Printf("[INFO]: setUserStatus(): Authenticated? %q (username) Cookie token: %q Libravatar: %q", session.Get("Username"), session.Get("Token"), session.Get("Libravatar"))
		}
	}
}