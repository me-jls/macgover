package main

import (
	"database/sql"
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	//"google.golang.org/genproto/googleapis/cloud/bigquery/connection/v1"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	getenvs "gitlab.com/avarf/getenvs"
)

var (
	db       *sql.DB
)

// --------------------------- Database

// no swagger information
func DBsqlconnect(engine string) (*sql.DB, error) {
	User := os.Getenv("DB_USER")
	Passwd := os.Getenv("DB_PASSWORD")
	Host := os.Getenv("DB_HOST")
	DBName := os.Getenv("DB_NAME")
	Timeout := getenvs.GetEnvString("DB_TIMEOUT", "5")
	var connection, Port string
	switch engine {
		case "mysql":
			Port = getenvs.GetEnvString("DB_PORT", "3306")
			connection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%ss",User,Passwd,Host,Port,DBName,Timeout)
		case "postgres":
			Port = getenvs.GetEnvString("DB_PORT", "5432")
			connection = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%s sslmode=disable", Host, Port, User, Passwd, DBName, Timeout)
	}
	log.Printf("[%s] INFO : DB_USER=%s", strings.ToUpper(engine), User)
	log.Printf("[%s] INFO : DB_PASSWORD=%s", strings.ToUpper(engine), b64.StdEncoding.EncodeToString([]byte(Passwd)))
	log.Printf("[%s] INFO : DB_HOST=%s", strings.ToUpper(engine), Host)
	log.Printf("[%s] INFO : DB_PORT=%s", strings.ToUpper(engine), Port)
	log.Printf("[%s] INFO : DB_NAME=%s", strings.ToUpper(engine), DBName)
	log.Printf("[%s] INFO : DB_TIMEOUT=%s", strings.ToUpper(engine), Timeout)
	db, err := sql.Open(engine, connection)
	if err != nil {
		log.Printf("[%s] ERROR : open=%s",strings.ToUpper(engine), err.Error())
	}
	// make sure connection is available
	err = db.Ping()
	if err != nil {
		log.Printf("[%s] ERROR : ping=%s",strings.ToUpper(engine), err.Error())
	}
	return db, err
}

// ---- swagger Informations
// @Tags         Database
// @router /v1/db/{engine} [get]
// @summary Test Database connection
// @consume text/plain
// @produce text/plain
// @param engine path string true "Database Engine (ex: mysql, postgres...)"
// @success 200 string OK
// @failure 500 string Internal Server Error
func dbEngineHandler(c *gin.Context) {
	engine := c.Param("engine")
	db, err := DBsqlconnect(engine)
	if err != nil {
		log.Printf("[%s] ERROR : %s",strings.ToUpper(engine), err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Printf("[%s] ERROR : %s",strings.ToUpper(engine), err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer db.Close()
	c.String(http.StatusOK, "Database connection OK (version: "+version+")")
}

// ---- swagger Informations
// @Tags         Database
// @router /v1/db/{engine}/count/{table} [get]
// @summary Count lines in tables
// @consume text/plain
// @produce text/plain
// @param engine path string true "Database Engine (ex: mysql, postgres...)"
// @param table path string true "Table name"
// @success 200 string OK
// @failure 500 string Internal Server Error
func dbHandlerCountRowTable(c *gin.Context) {
	engine := c.Param("engine")
	db, err := DBsqlconnect(engine)
	if err != nil {
		log.Printf("[%s] ERROR : %s",strings.ToUpper(engine), err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	table := c.Param("table")
	request := "SELECT COUNT(*) AS COUNT FROM " + table + ";"
	log.Printf("[%s] REQUEST: %s" ,strings.ToUpper(engine),request)
	var count int
	err = db.QueryRow(request).Scan(&count)
	if err != nil {
		log.Printf("[%s] ERROR : %s",strings.ToUpper(engine), err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer db.Close()
	msg := strconv.Itoa(count) + " row(s) found in table " + strings.ToUpper(table)
	log.Printf("[%s] MSG: %s" ,strings.ToUpper(engine),msg)
	c.String(http.StatusOK, msg)
}


