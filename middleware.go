/***
 *	Middleware for dealing with login/session cookies
 */
package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// ensureLoggedIn tests if the user is logged in, reading in from the context to see if a flag is set.
// Note that this flag is not a boolean any more, I'm using this pseudo-flag to store the username
func ensureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		loggedInInterface := session.Get("Username")
		if loggedInInterface == nil || loggedInInterface == "" {
			if *config["ginMode"] == "debug" {
				log.Printf("[INFO]: ensureLoggedIn(): No authenticated user")
			}
			c.Abort()
			c.HTML(http.StatusOK, "404.tpl", environment(c,
				gin.H{
					"errorcode":     http.StatusUnauthorized,
					"errortext":     http.StatusText(http.StatusUnauthorized),
					"errorbody":     "You must be authenticated to continue",
					"titleCommon":   *config["titleCommon"] + " - " + http.StatusText(http.StatusUnauthorized),
					"logintemplate": false,
				}))
			//c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			if *config["ginMode"] == "debug" {
				log.Printf("[INFO]: ensureLoggedIn(): Username is %q", loggedInInterface)
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
			c.Abort()
			c.HTML(http.StatusOK, "404.tpl", environment(c,
				gin.H{
					"errorcode":     http.StatusUnauthorized,
					"errortext":     "Already authenticated",
					"errorbody":     "You have already logged in!",
					"titleCommon":   *config["titleCommon"] + " -  Already authenticated",
					"logintemplate": false,
				}))
			// c.AbortWithStatus(http.StatusUnauthorized)
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
		c.Set("Username", session.Get("Username"))
		c.Set("Email", session.Get("Email"))
		c.Set("Libravatar", session.Get("Libravatar"))
		c.Set("Token", session.Get("Token"))
		c.Set("UUID", session.Get("UUID"))
		c.Set("RememberMe", session.Get("RememberMe"))
		c.Set("sidebarCollapsed", session.Get("sidebarCollapsed"))

		if *config["ginMode"] == "debug" {
			log.Printf("[INFO]: setUserStatus(): Authenticated? %q (username) Cookie token: %q Libravatar: %q", session.Get("Username"), session.Get("Token"), session.Get("Libravatar"))
		}
	}
}
