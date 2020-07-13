// Implementation

package main

import (
	"crypto/sha256"
	"crypto/subtle"
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
	t string `form:"t"`	// this field is only used when someone changes the password via email-sent link: it includes the link token
}

type ResetPasswordForm struct {
	Email string `json:"email" form:"email" binding:"required"`
	GPG string	`json:"gpg" form:"gpg"`	// GPG fingerprint to encrypt email, if provided (gwyneth 20200705)
}

// Token is required by the ShouldBindURI() call, it cannot be a simple string for some reason...
type Token struct {
	Payload string `uri:"token"`
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

	// check if user exists
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
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
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
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
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
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
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
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
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

// changePassword asks for the current password (unless it comes from an email link) and a new one (plus validation). Not done yet.
func changePassword(c *gin.Context) {
	var aPasswordChange ChangePasswordForm
	var token string

	// Note: this can be called either by a logged-in user which has authenticated themselves properly
	// OR it can come from a user who has a valid token from an email to reset the link, and can thus enter without password

	session := sessions.Default(c)

	if c.Bind(&aPasswordChange) != nil { // nil means no errors
		c.HTML(http.StatusBadRequest, "change-password.tpl", gin.H{
			"ErrorTitle"	: "Password change failed!",
			"ErrorMessage"	: "No form data posted",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "Say what?",
			"logintemplate"	: true,
		})
		log.Println("[WARN] No form data posted for password change")

		return
	}
	// Ok, we got a form; so do simple checks first
	if aPasswordChange.NewPassword != aPasswordChange.ConfirmNewPassword {
		c.HTML(http.StatusBadRequest, "change-password.tpl", gin.H{
			"ErrorTitle"	: "Password change failed!",
			"ErrorMessage"	: "Confirmation password does not match new password",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "No way, José!",
			"logintemplate"	: true,
		})
		log.Println("[WARN] Confirmation password does not match new password")

		return
	}
	if aPasswordChange.t == "" && (aPasswordChange.NewPassword == aPasswordChange.OldPassword) {
		c.HTML(http.StatusBadRequest, "change-password.tpl", gin.H{
			"ErrorTitle"	: "Password change failed!",
			"ErrorMessage"	: "New password must be different from the old one",
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"Debug"			: false,
			"titleCommon"	: *config["titleCommon"] + "No deal!",
			"logintemplate"	: true,
		})
		log.Println("[WARN] New password must be different from the old one")

		return
	}
	var someTokens ResetPasswordTokens	// this will be used to retrieve data from the KV store; we will check later on (scope issues!) if we have a valid token or not

	isCurrentPasswordValid := false
	if aPasswordChange.t == "" {	// do we have a token here?
		// no token; so we assume that the password is valid and we can proceed to change <script charset="utf-8">
		isCurrentPasswordValid = true
	} else {
		// we still have a token, but to make things more secure, we validate the token again
		// again, first split the token
		token 	 := aPasswordChange.t
		selector := token[:15]
		verifier := token[15:]
		sha256	 := sha256.Sum256([]byte(verifier))
		if *config["ginMode"] == "debug" {
			fmt.Printf("[DEBUG] Got token %q, this is selector %q and verifier %q and SHA256 %q\n", token, selector, verifier, sha256)
		}
		found, err := GOSWIstore.Get(selector, &someTokens)
		if err == nil {
			if found {
				log.Printf("[INFO] What we just stored: %+v", someTokens)
				// check if it is still valid
				if time.Since(someTokens.Timestamp) < (2 * time.Hour) {
					// valid, log user in, move to password change template
					if subtle.ConstantTimeCompare(sha256[:], someTokens.Verifier[:]) == 1 {
						log.Printf("[INFO] Token still valid for user %q (%q)!", someTokens.Username, someTokens.UserUUID)
						isCurrentPasswordValid = true
					}
				}
			}
		}
		// whatever happened, consume the token, that is, delete it from the store, so it can't be reused
		if err := GOSWIstore.Delete(selector); err != nil {
			// this will rarely happen (unless selector == "", which should not occur) since Store.Delete() will NOT throw
			//  errors if the key doesn't exist
			log.Printf("[WARN] Deleting %q from the store threw an error\n", err)
		}
	}
	if isCurrentPasswordValid {
		// We now need to figure out who is the user requesting this!
		// 1) Either this is called via the token sent by email, and it means that someTokens.UserUUID has been set;
		// 2) or this was called by a logged-in user changing their password, and c.Get(UUID) or session.Get(UUID) will have the UUID.
		var UUID string
		if someTokens.UserUUID != "" {
			UUID = someTokens.UserUUID
		} else if UUID, ok := c.Get("UUID"); ok {
			/* no-op */
		} else if UUID = session.Get("UUID"); UUID != nil || UUID != "" {
			/* no-op */
		} else {
			log.Println("[ERROR] Cannot change password because we cannot get a UUID for this user! Hack attempt?")

			c.HTML(http.StatusForbidden, "404.tpl", gin.H{
				"now"			: formatAsYear(time.Now()),
				"author"		: *config["author"],
				"description"	: *config["description"],
				"logo"			: *config["logo"],
				"logoTitle"		: *config["logoTitle"],
				"sidebarCollapsed" : *config["sidebarCollapsed"],
				"titleCommon"	: *config["titleCommon"] + " - 403",
				"errorcode"		: "403",
				"errortext"		: "User not found",
				"errorbody"		: "Unknown or invalid user, cannot proceed, please try later.",
			})
			return
		}
		// ok, we ought to have a valid UUID, at last we can update the password!
		if *config["dsn"] == "" {
			log.Fatal("Please configure the DSN for accessing your OpenSimulator database; this application won't work without that")
		}
		db, err := sql.Open("mysql", *config["dsn"]) // presumes mysql for now
		checkErrFatal(err)

		defer db.Close()
		// using the same variables as in isUserValid(), because of the Principle of Least Surprise
		var hashedPassword, passwordSalt, hashed, interior string
		// generate salt
		passwordSalt = randomBase64String(32)
		hashedPassword = GetMD5Hash(aPasswordChange.NewPassword)
		interior = hashedPassword + ":" + passwordSalt
		hashed = GetMD5Hash(interior)

		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG] md5(password) = %q, (md5(password) + \":\" + passwordSalt) = %q, md5(md5(password) + \":\" + passwordSalt) = %q",
			hashedPassword, interior, hashed)
		}


		res, err := db.Exec(`UPDATE auth SET passwordHash = ?, passwordSalt = ? WHERE UUID = ?"`, hashed, passwordSalt, UUID)
		checkErr(err)

		numRowsAffected, err := res.RowsAffected()
		checkErr(err)

		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG] Changing password resulted in %d row(s) affected\n", numRowsAffected)
		}
		// we're finished here, redirect to the Dashboard
		c.Redirect(http.StatusTemporaryRedirect, "/")

		return
	}

	c.HTML(http.StatusForbidden, "404.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"logo"			: *config["logo"],
		"logoTitle"		: *config["logoTitle"],
		"sidebarCollapsed" : *config["sidebarCollapsed"],
		"titleCommon"	: *config["titleCommon"] + " - 403",
		"errorcode"		: "403",
		"errortext"		: "Token incorrect",
		"errorbody"		: fmt.Sprintf("Either your token %q is invalid or it has expired!", token),	// token may be empty
	})
}

// ResetPasswordTokens are stored in a KV store, the key of which is the Selector.
type ResetPasswordTokens struct {
	UserUUID	string		`json:"uuid"`
	Username	string		`json:"username"`
	Email		string		`json:"email"`
	Verifier	[32]byte	`json:"verifier"`	// 18 chars encoded to 32-byte SHA256 hash
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
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
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

	var principalID, email, firstName, lastName string
	err = db.QueryRow("SELECT PrincipalID, Email, FirstName, LastName FROM UserAccounts WHERE Email = ?", aPasswordReset.Email).Scan(&principalID, &email, &firstName, &lastName)	// there can be only one, or our database is corrupted
	if err != nil { // db.QueryRow() will return ErrNoRows, which will be passed to Scan()
		if *config["ginMode"] == "debug" {
			log.Printf("[DEBUG] email address %q not in database, but we're not telling. Error was: %v", aPasswordReset.Email, err)
		}
	}
	if *config["ginMode"] == "debug" {
		log.Printf("[DEBUG] Password reset: We have email %q (empty means: not in database) from user %q [UUID %q] (empty means: not in database)", email, firstName + " " + lastName, principalID)
	}

	if (principalID != "" && email != "") {
		// let's test our KV store by pushing some garbage into it!
		// TODO(gwyneth): this has to be done much more carefully...
		// Use Cryptographically Secure Randomly Generated Split-Tokens as shown on the P.I.E. algorithm found here https://paragonie.com/blog/2016/09/untangling-forget-me-knot-secure-account-recovery-made-simple#secure-password-reset-tokens (gwyneth 20200706)

		selector := randomBase64String(15)
		verifier := randomBase64String(18)
		sha256	 := sha256.Sum256([]byte(verifier))

		GOSWIstore.Set(selector, ResetPasswordTokens{
			UserUUID:	principalID,
			Username:	firstName + " " + lastName,
			Email:		email,
			Verifier:	sha256,
			Timestamp:	time.Now(),
		})
		if *config["ginMode"] == "debug" {
			var someTokens ResetPasswordTokens
			found, err := GOSWIstore.Get(selector, &someTokens)
			if err == nil {
				if found {
					log.Printf("[DEBUG] What we just stored for selector %q: %+v\n", selector, someTokens)
				} else {
					log.Println("[DEBUG]", selector, "not found in store")
				}
			} else {
				log.Println("[WARN] Nothing stored for", selector, "error was", err)
			}
		}
		// Now send email!
		// using example from https://riptutorial.com/go/example/20761/sending-email-with-smtp-sendmail-- (gwyneth 20200706)
		if email != "" {
			if *config["ginMode"] == "debug" {
				log.Printf("[DEBUG] Request: %+v\n", c.Request)
			}
			// Build the actual URL for token
			scheme := "https:"
			if c.Request.TLS == nil {
				scheme = "http:"
			}
			tokenURL := scheme + "//" + c.Request.Host + "/user/token/" + selector + verifier

			// The grid manager's email is stored in *config["gOSWIemail"]
			//
			// server we are authorised to send email through is stored in *config["SMTPhost"]
			//
			// Create the authentication for the SendMail()
			// using PlainText, but other authentication methods are encouraged
			auth := smtp.PlainAuth("", *config["gOSWIemail"], *config["gOSWIpassword"],*config["SMTPhost"])

			// NOTE: Using the backtick here ` works like a heredoc, which is why all the
			// rest of the lines are forced to the beginning of the line, otherwise the
			// formatting is wrong for the RFC 822 style
			// TODO(gwyneth): use a template instead?
			message := `To: "` + firstName + " " + lastName + `" <` + email + `>
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
	var params Token

	// token = c.Param("token")	// Wouldn't this be more obvious?
	if err := c.ShouldBindUri(&params); err != nil {	// this should never fail, but... 	(gwyneth 20200713)
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"titleCommon"	: *config["titleCommon"] + " - 404",
			"errorcode"		: "404",
			"errortext"		: "Token not sent",
			"errorbody"		: fmt.Sprintf("Invalid token or token not sent. Error was: %v", err),
		})
		log.Println("[ERROR] Invalid token or token not sent. Error was:", err)
		return
	}
	// assign token with the content of the parameter...
	token = params.Payload
	if token == "" { // one of those errors that should never happen... (gwyneth 20200713)
		c.HTML(http.StatusNotFound, "404.tpl", gin.H{
			"now"			: formatAsYear(time.Now()),
			"author"		: *config["author"],
			"description"	: *config["description"],
			"logo"			: *config["logo"],
			"logoTitle"		: *config["logoTitle"],
			"sidebarCollapsed" : *config["sidebarCollapsed"],
			"titleCommon"	: *config["titleCommon"] + " - 404",
			"errorcode"		: "404",
			"errortext"		: "Empty token payload",
			"errorbody"		: "Invalid token or empty token payload.",
		})
		log.Println("[ERROR] Invalid token or empty token payload.")
		return

	}
	// split token
	selector := token[:15]
	verifier := token[15:]
	sha256	 := sha256.Sum256([]byte(verifier))
	if *config["ginMode"] == "debug" {
		fmt.Printf("[DEBUG] Got token %q, this is selector %q and verifier %q and SHA256 %q\n", token, selector, verifier, sha256)
	}

	var someTokens ResetPasswordTokens
	found, err := GOSWIstore.Get(selector, &someTokens)
	if err == nil {
		if found {
			log.Printf("[INFO] What we just stored: %+v", someTokens)
			// check if it is still valid
			if time.Since(someTokens.Timestamp) < (2 * time.Hour) {
				// valid, log user in, move to password change template
				if subtle.ConstantTimeCompare(sha256[:], someTokens.Verifier[:]) == 1 {
					log.Printf("[INFO] Token valid for user %q (%q)!", someTokens.Username, someTokens.UserUUID)
					// user authenticated, get them a session cookie! (gwyneth 20200712)
					session := sessions.Default(c)
					session.Set("Username", someTokens.Username)
					session.Set("UUID", someTokens.UserUUID)
					session.Set("Token", generateSessionToken())
					if *config["ginMode"] == "debug" {
						log.Printf("[INFO] Password change link: User valid with username: %q UUID: %s Email: <%s> Token: %s", someTokens.Username, someTokens.UserUUID, someTokens.Email, session.Get("Token"))
					}
					if someTokens.Email != "" {
						//			avt.SetSecureFallbackHost("unicornify.pictures")	// possibly not needed, we'll implement it locally
						avt := libravatar.New()
						avt.SetAvatarSize(60)	// for some silly reason, that's what our template has...
						avt.SetUseHTTPS(true)
						//			avt.SetSecureFallbackHost("unicornify.pictures")
						if avatar_url, err := avt.FromEmail(someTokens.Email); err == nil {
							session.Set("Libravatar", avatar_url)
						} else {
							if *config["ginMode"] == "debug" {
								log.Println("[WARN]: Libravatar returned error:", err)
							}
							// couldn't get an image url from the Libravatar service, so get an Unicorn instead!
							session.Set("Libravatar", "https://unicornify.pictures/avatar/" + GetMD5Hash(someTokens.Username) + "?s=60")
							session.Set("Email", someTokens.Email)	// who knows, it might be useful at some point
						}
					} else {
						// if we don't have a valid email, get an Unicorn!
						if *config["ginMode"] == "debug" {
							log.Println("[WARN]: Empty email on database, attempting to get a Unicorn")
						}
						session.Set("Libravatar", "https://unicornify.pictures/avatar/" + GetMD5Hash(someTokens.Username) + "?s=60")
					}
//						session.Set("RememberMe", ???)	// we may not be able to set this here
					session.Save()

					// move user to password change template
					c.HTML(http.StatusOK, "change-password.tpl", gin.H{
						"now"			: formatAsYear(time.Now()),
						"author"		: *config["author"],
						"description"	: *config["description"],
						"logo"			: *config["logo"],
						"logoTitle"		: *config["logoTitle"],
						"sidebarCollapsed" : *config["sidebarCollapsed"],
						"Debug"			: false,
						"titleCommon"	: *config["titleCommon"] + "New password!",
						"logintemplate"	: true,
						"someTokens"	: someTokens,	// this will allow us to ignore the 'old' password and just ask for new ones.
					})
					// that's all, folks!
					return
				}
			} else {
				log.Println("[ERROR] Token expired!")
			}
		} else {
			log.Println("[ERROR]", selector, "not found in store")
		}
	} else {
		log.Println("[ERROR] Nothing stored for", selector, "error was", err)
	}

	c.HTML(http.StatusForbidden, "404.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"logo"			: *config["logo"],
		"logoTitle"		: *config["logoTitle"],
		"sidebarCollapsed" : *config["sidebarCollapsed"],
		"titleCommon"	: *config["titleCommon"] + " - 403",
		"errorcode"		: "403",
		"errortext"		: "Token incorrect",
		"errorbody"		: fmt.Sprintf("Either your token %q is invalid or it has expired!", token),
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

