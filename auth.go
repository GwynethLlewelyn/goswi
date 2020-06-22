// Implementation 

package main

import (
	"database/sql"
// 	"encoding/json"
//	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
//	jsoniter "github.com/json-iterator/go"
//	"html/template"
	"log"
	"net/http"
//	"strings"
	"time"
)

func showLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tpl", gin.H{
		"now"			: formatAsYear(time.Now()),
		"author"		: *config["author"],
		"description"	: *config["description"],
		"viewerInfo"	: viewerInfo,
		"Debug"			: false,
		"titleCommon"	: *config["titleCommon"] + "Welcome!",
		"logintemplate"	: true,
	})
}

func performLogin(c *gin.Context) {}

func logout(c *gin.Context) {}