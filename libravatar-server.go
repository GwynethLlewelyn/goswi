// This implements a minimalistic Libravatar-compatible, federated server, which will run as part of gOSWI.
// It returns Profile images from OpenSim to be used as Libravatars.
// For this to work, DNS needs to be properly setup with:
//
//
//
package main

import {
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net"
}

func Libravatar(c *gin.Context) {
	return
}