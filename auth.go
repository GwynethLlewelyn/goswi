// Implementation

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/sessions"
// 	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	"github.com/philippgille/gokv"
//	"github.com/philippgille/gokv/syncmap"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
// 	"math/rand"
	"net/http"
	"net/smtp"
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
	GPG string	`json:"gpg" form:"gpg"`	// GPG fingerprint to encrypt email, if provided (gwyneth 20200705)
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
		log.Printf("[INFO] User %q authenticated.", username)
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
	if ok, email, principalID := isUserValid(oneUser.Username, oneUser.Password); ok {
		session.Set("Username", oneUser.Username)
		session.Set("UUID", principalID)
		session.Set("Token", generateSessionToken())
		if *config["ginMode"] == "debug" {
			log.Printf("[INFO] User valid with username: %q UUID: %s Email: <%s> Token: %s", oneUser.Username, principalID, email, session.Get("Token"))
		}

		if email != "" {
//			avt.SetSecureFallbackHost("unicornify.pictures")	// possibly not needed, we'll implement it locally
			avt := libravatar.New()
			avt.SetAvatarSize(60)	// for some silly reason, that's what our template has...
			avt.SetUseHTTPS(true)
//			avt.SetSecureFallbackHost("unicornify.pictures")
			if avatar_url, err := avt.FromEmail(email); err == nil {
				session.Set("Libravatar", avatar_url)
			} else {
				if *config["ginMode"] == "debug" {
					log.Println("[WARN]: Libravatar returned error:", err)
				}
				// couldn't get an image url from the Libravatar service, so get an Unicorn instead!
				session.Set("Libravatar", "https://unicornify.pictures/avatar/" + GetMD5Hash(oneUser.Username) + "?s=60")
				session.Set("Email", email)	// who knows, it might be useful at some point
			}
		} else {
			// if we don't have a valid email, get an Unicorn!
			if *config["ginMode"] == "debug" {
				log.Println("[WARN]: Empty email on database, attempting to get a Unicorn")
			}
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

// ResetPasswordTokens are stored in a KV store
type ResetPasswordTokens struct {
	Selector	string		`json:"selector"`	// 15 chars
	Verifier	string		`json:"verifier"`	// 18 chars
	Timestamp	time.Time	`json:"timestamp"`
}

// resetPassword is called with a POST from the form; we act upon it here.
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

	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] aPasswordReset: %+v", aPasswordReset)
	}

	// check if this email address is in the database
	if *config["dsn"] == "" {
		log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
	}
	db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
	checkErrFatal(err)

	defer db.Close()

	var principalID, email string
	err = db.QueryRow("SELECT PrincipalID, Email FROM UserAccounts WHERE Email = ?", aPasswordReset.Email).Scan(&principalID, &email)	// there can be only one, or our database is corrupted
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG] email address %q not in database, but we're not telling. Error was: %v", aPasswordReset.Email, err)
		}
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Password reset: We have email %q from user (empty means: not in database) and UUID %q (empty means: not in database)", email, principalID)
	}

	if (principalID != "" && email != "") {
		// let's test our KV store by pushing some garbage into it!
		// TODO(gwyneth): this has to be done much more carefully...
		// Use Cryptographically Secure Randomly Generated Split-Tokens as shown on the P.I.E. algorithm found here https://paragonie.com/blog/2016/09/untangling-forget-me-knot-secure-account-recovery-made-simple#secure-password-reset-tokens (gwyneth 20200706)

		selector := randomBase64String(15)
		verifier := randomBase64String(18)

		GOSWIstore.Set(principalID, ResetPasswordTokens{
			Selector: selector,
			Verifier: verifier,	// this will have to be changed to a SHA256 hash of the verifier
			Timestamp: time.Now(),
		})
		if *config["ginMode"] == "debug" {
			var someTokens ResetPasswordTokens
			found, err := GOSWIstore.Get(principalID, &someTokens)
			if err == nil {
				if found {
					log.Printf("[DEBUG] What we just stored: %+v", someTokens)
				} else {
					log.Println("[DEBUG]", principalID, "not found in store")
				}
			} else {
				log.Println("[WARN] Nothing stored for", principalID, "error was", err)
			}
		}
		// Now send email!
		// using example from https://riptutorial.com/go/example/20761/sending-email-with-smtp-sendmail-- (gwyneth 20200706)
		if email != "" {
			// Build the actual URL for token
			tokenURL := c.Request.URL.Scheme + "//" + c.Request.URL.Host + "/user/token/" + selector + verifier

			// The grid manager's email is stored in *config["gOSWIemail"]

			// server we are authorised to send email through is stored in *config["SMTPhost"]

			// Create the authentication for the SendMail()
			// using PlainText, but other authentication methods are encouraged
			auth := smtp.PlainAuth("", *config["gOSWIemail"], "", *config["SMTPhost"])

			// NOTE: Using the backtick here ` works like a heredoc, which is why all the
			// rest of the lines are forced to the beginning of the line, otherwise the
			// formatting is wrong for the RFC 822 style
			// TODO(gwyneth): use a template instead?
			message := `To: "Someone" <` + email + `>
From: "` + *config["author"] + `" <` + *config["gOSWIemail"] + `>
Subject: Password reset link

Someone asked for your password to be reset.

If it was you, click on <a href="` + tokenURL + `">` + tokenURL + `</a>.
		`
			if *config["ginMode"] == "debug" {
				fmt.Printf("[DEBUG] Message to be sent: %q\n", message)
			}
			if err := smtp.SendMail(*config["SMTPhost"]+":25", auth, *config["gOSWIemail"], []string{email}, []byte(message)); err != nil {
				fmt.Printf("[ERROR] Sending reset link email to <%s> via SMTP host %q failed: %v\n", email, *config["SMTPhost"], err)
				//os.Exit(1)
			}
			if *config["ginMode"] == "debug" {
				fmt.Println("[INFO] Success in sending reset link to", email)
			}
		}
	}
	c.HTML(http.StatusOK, "reset-password-confirmation.tpl", gin.H{
		"Content"		: fmt.Sprintf("Please check your email address %q for a password reset link; if your email address is in our database, you should get it shortly.", email),
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"Debug"			: false,
		"titleCommon"	: *config["titleCommon"] + "Email sent!",
		"logintemplate"	: true,
	})
}

// checkTokenForPasswordReset is called when the user clicks the link for resetting their password, and we need to check if the token is a valid token to allow authentication
func checkTokenForPasswordReset(c *gin.Context) {
	var token string

	if err := c.ShouldBindUri(&token); err != nil {
			c.JSON(400, gin.H{"msg": err})
			return
	}

	c.HTML(http.StatusNotFound, "404.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"titleCommon"	: *config["titleCommon"] + " - 404",
		"errortext"		: "Token incorrect",
		"errorbody"		: fmt.Sprintf("Either your token %q is invalid or it has expired!", ),
	})
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

