package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	//"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
)

// ---- swagger Informations
// @Tags         Endpoints
// @router /v1/ldap [get]
// @summary LDAP connection and blind test
// @security BasicAuth
// @produce text/plain
// @success 200 string OK
// @failure 401 string Unauthorized
// @failure 500 string Internal Server Error
func ldapHandler(c *gin.Context) {
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		c.Writer.Header().Add("WWW-Authenticate", `Basic realm="Macgover", charset="UTF-8" `)
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[LDAP] INFO : login=" + username)
	
	ldapBindDN := os.Getenv("LDAP_BIND_DN")
	username = "cn=" + strings.ToLower(username) + "," + ldapBindDN
	log.Printf("[LDAP] INFO : BindDN=" + username)

	ldapURL := os.Getenv("LDAP_URL")
	log.Printf("[LDAP] INFO : LDAP_URL=" + ldapURL)

	l, err := ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	defer l.Close()

	if err != nil {
		log.Printf("[LDAP] ERROR : Dial=" + err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		err := l.Bind(strings.ToLower(username), password)
		if err != nil {
			log.Printf("[LDAP] ERROR : Bind=" + err.Error())
			c.Writer.Header().Add("WWW-Authenticate", `Basic realm="Macgover", charset="UTF-8" `)
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	c.String(http.StatusOK, "LDAP Connection and Bind are OK")
}
