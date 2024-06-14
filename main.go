package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	getenvs "gitlab.com/avarf/getenvs"

	// doc api swagger : https://github.com/swaggo/swag
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "github.com/me-jls/macgover/docs"
)
//fake

//go:embed assets/* templates/*
var fs embed.FS

// ---- swagger Informations
// @title Macgover
// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

var (
	port     string
	name     string
	mode     string
	job      string
	argument string
)

type jsonMetric struct {
	Job   string `json:"job"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

func init() {
	flag.StringVar(&port, "port", getenvs.GetEnvString("MACGOVER_PORT", "3000"), "give me a port number")
	flag.StringVar(&name, "name", os.Getenv("MACGOVER_NAME"), "give me a name")
	flag.StringVar(&mode, "mode", getenvs.GetEnvString("MACGOVER_MODE", "server"), "give me a mode to start")
	flag.StringVar(&job, "job", getenvs.GetEnvString("MACGOVER_JOB", "metrics"), "give me a job name")
	flag.StringVar(&argument, "argument", getenvs.GetEnvString("MACGOVER_ARGUMENT", "{}"), "give me a argument")
}

func updateTitleSwagger(c *ginSwagger.Config) {
	c.Title = "Macgover"
}

func main() {
	log.Println("------------------------ Start MACGOVER ------------------------")
	flag.Parse()

	mode = strings.ToLower(mode)
	log.Printf("Mode=%s", mode)

	switch strings.ToLower(mode) {
	case "server":

		router := gin.Default()

		//router.LoadHTMLGlob("templates/*.html.tmpl")
		//router.LoadHTMLFiles("templates/index.html.tmpl")
		tmpl := template.Must(template.New("").ParseFS(fs, "templates/*.tmpl"))
		router.SetHTMLTemplate(tmpl)

		router.GET("/", redirectIndex)

		docs.SwaggerInfo.BasePath = "/"
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, updateTitleSwagger))

		//router.Static("/images", "./images")
		router.StaticFS("/public", http.FS(fs))

		router.Static("/favicon.ico", "https://go.dev/favicon.ico")

		router.GET("/macgover", macgoverHandler)

		v1 := router.Group("/v1")
		{
			v1.GET("/whoami", whoamiHandler)
			v1.GET("/ping", pingHandler)
			v1.GET("/echo", echoHandler)
			v1.POST("/echo", echoHandler)
			v1.GET("/ldap", ldapHandler)
			v1.GET("/db/:engine", dbEngineHandler)
			v1.GET("/db/:engine/count/:table", dbHandlerCountRowTable)
			v1.GET("/healthcheck", healthcheckHandler)
			v1.GET("/metrics", prometheusMetricsHandler)
			v1.POST("/metrics", metricsHandler)
			v1.GET("/url", testUrlHandler)
			v1.POST("/jwt/login", jwtLoginHandler)
			v1.GET("/jwt/test", jwtTestHandler)
			v1.GET("/network", networkHandler)
		}

		// for example new group /v2 ...
		//  - Do not forget to add a router in the swagger definition
		//  - Update the readme.md file, with this new version

		// v2 := router.Group("/v2")
		// {
		// 	v2.GET("/ping", pingHandler)
		// 	v2.GET("/whoami", whoamiHandler)
		// 	v2.GET("/metrics", prometheusMetricsHandler)
		// }

		// Handle error response when a route is not defined
		router.NoRoute(func(c *gin.Context) {
			// In gin this is how you return a JSON response
			c.HTML(404, "404.tmpl", gin.H{"message": "Page not found ..."})
		})

		router.Run(":" + port)
	case "batch":
		switch strings.ToLower(job) {
		case "metrics":
			batchJobMetrics(argument)
		}
	}
}

// Redirect
func redirectIndex(c *gin.Context) {
	c.Redirect(301, "/macgover")
}

// index.html
func macgoverHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":    "Demo",
		"hostname": os.Getenv("HOSTNAME"),
		"commit": os.Getenv("MACGOVER_COMMIT"),
	})
}

// ---- swagger Informations
// @Tags         Testing
// @router /v1/ping [get]
// @summary pong
// @consume plain
// @produce plain
// @success 200 string OK
func pingHandler(c *gin.Context) {

	pHostname := os.Getenv("HOSTNAME") //string
	pDate := time.Now().Unix()         //int64
	pMessage := "pong"                 //string

	if c.Query("format") != "json" {

		c.String(http.StatusOK, "Hostname: "+pHostname+"\nDate: "+strconv.FormatInt(pDate, 10)+"\nMessage: "+pMessage)
	} else {
		var data struct {
			Hostname string `json:"hostname,omitempty"`
			Date     int64  `json:"date,omitempty"`
			Message  string `json:"message,omitempty"`
		}
		data.Hostname = pHostname
		data.Date = pDate
		data.Message = pMessage
		c.JSON(http.StatusOK, data)
		// c.JSON(http.StatusOK, gin.H{
		// 	"hostname": 	os.Getenv("HOSTNAME"),
		// 	"date":  		time.Now().Unix(),
		// 	"url":      	c.Request.RequestURI,
		// 	"method":   	c.Request.Method,
		// 	"message":		"pong",
		// })
	}
}

// ---- swagger Informations
// @Tags         Debugging
// @router /v1/echo [post]
// @summary Display the request in log
// @consume application/json
// @param request body string true "json"
// @produce text/plain
// @success 200 string OK
func echoHandler(c *gin.Context) {
	var statusCodeStr string
	if len(c.Query("code")) > 0 && c.Query("code") != "200" {
		statusCodeStr = c.Query("code")
	}

	statusCodeInt, err := strconv.Atoi(statusCodeStr)
	if err != nil {
		statusCodeInt = 200
	}

	params := c.Request.URL.Query()
	log.Printf("[ECHO] query: %s", params)

	if jsonData, err := ioutil.ReadAll(c.Request.Body); err == nil {
		log.Printf("[ECHO] body: %s", string(jsonData))
	}

	c.String(statusCodeInt, "OK message received !")
}

// ---- swagger Informations
// @Tags         Testing
// @router /v1/whoami [get]
// @summary Return a web page with some informations (Hostname, IP address, User-Agent ...)
// @consume text/plain
// @param wait query string false "Request timeout (ex: 5s)"
// @produce text/plain
// @success 200 string OK
// @failure 500 string Internal Server Error
func whoamiHandler(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	var req *http.Request = c.Request
	wait := c.Query("wait")
	if len(wait) > 0 {
		duration, err := time.ParseDuration(wait)
		log.Printf("[WHOAMI] wait to : %s", duration)
		if err == nil {
			time.Sleep(duration)
		}
	}
	if name != "" {
		_, _ = fmt.Fprintln(w, "Name:", name)
	}
	hostname := os.Getenv("HOSTNAME")
	_, _ = fmt.Fprintln(w, "Hostname:", hostname)
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			_, _ = fmt.Fprintln(w, "IP:", ip)
		}
	}
	_, _ = fmt.Fprintln(w, "RemoteAddr:", req.RemoteAddr)
	if err := req.Write(w); err != nil {
		c.String(http.StatusInternalServerError, "Errors")
		log.Printf(err.Error())
		return
	}
}

// ---- swagger Informations
// @Tags         Testing
// @router /v1/healthcheck [get]
// @summary Healthcheck with custom status
// @consume text/plain
// @produce text/plain
// @param code query string false "Custom return status code (ex: 404)"
// @success 200 string OK
func healthcheckHandler(c *gin.Context) {
	var statusCodeStr string
	if len(c.Query("code")) > 0 && c.Query("code") != "200" {
		statusCodeStr = c.Query("code")
	} else {
		statusCodeStr = "200"
	}
	statusCodeInt, err := strconv.Atoi(statusCodeStr)
	if err != nil {
		c.String(http.StatusInternalServerError, "Errors")
		log.Printf(err.Error())
		return
	}
	if statusCodeInt > 200 {
		log.Printf("Update health check status code [%d]\n", statusCodeInt)
	}
	c.String(statusCodeInt, "Healthcheck return code: "+statusCodeStr)
}

// ---- swagger Informations, workarround for /metrics [get]
// @Tags         Testing
// @router /v1/url [get]
// @summary test url website
// @param test query string false "enter url to test (ex: https://www.ecosia.org/)"
// @consume plain
// @produce plain
// @success 200 string OK
func testUrlHandler(c *gin.Context) {
	url := getenvs.GetEnvString("TEST_URL", "https://www.ecosia.org/")
	if len(c.Query("test")) > 0 {
		url = c.Query("test")
	}
	log.Printf("[URL] INFO : test=%s", url)
	// post on url metrics
	resp, err := http.Get(url)
	if err != nil {
		c.String(http.StatusBadRequest, "Error sending request to "+url)
		log.Printf("[URL] ERROR : " + err.Error())
		return
	}
	log.Printf("[URL] INFO : Response Status: %s", resp.Status)
	c.String(resp.StatusCode, url+" Return code : %s", resp.Status)
}

// ---- swagger Informations, workarround for /metrics [get]
// @Tags         Metrics
// @router /v1/metrics [get]
// @summary get metrics
// @consume plain
// @produce plain
// @success 200 string OK
func prometheusMetricsHandler(c *gin.Context) {
	handler := promhttp.Handler()
	handler.ServeHTTP(c.Writer, c.Request)
}

// ---- swagger Informations,
// @Tags         Metrics
// @router /v1/metrics [post]
// @summary post metrics
// @consume application/json
// @param data body jsonMetric false "Your Json Metric"
// @produce text/plain
// @success 200 string OK
// @failure 400 string Bad request
// @failure 500 string Internal Server Error
func metricsHandler(c *gin.Context) {
	url := getenvs.GetEnvString("PUSHMETRICS_URL", "http://localhost:9091")
	var data *jsonMetric = &jsonMetric{"macgover_server_job", "macgover_server_label", 1}
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading body :"+err.Error())
		return
	}
	log.Printf("[METRICS] INFO : parameters=%v", data)
	log.Printf("[METRICS] INFO : url=%s", url)
	// post on url metrics
	urlToPostMetrics := url + "/metrics/job/" + strings.ReplaceAll(data.Job, " ", "") + "/" + strings.ReplaceAll(data.Label, " ", "") + "/" + strconv.Itoa(data.Value)
	log.Printf("[METRICS] INFO : post on %s", urlToPostMetrics)
	resp, err := http.Post(urlToPostMetrics, "", nil)
	if err != nil {
		c.String(http.StatusBadRequest, "Error sending request to "+url)
		log.Printf("[METRICS] ERROR : " + err.Error())
		return
	}
	log.Printf("[METRICS] INFO : Response Status: %s", resp.Status)
	c.String(resp.StatusCode, "Return code : %s", resp.Status)
}

// function for Job
func batchJobMetrics(argValues string) {
	url := getenvs.GetEnvString("PUSHMETRICS_URL", "http://localhost:9091")
	var data *jsonMetric = &jsonMetric{"macgover_batch_job", "macgover_batch_label", 1}
	err := json.Unmarshal([]byte(argValues), &data) // convert string to json
	log.Printf("[BATCH/METRICS] INFO : argument=%v", data)
	if err != nil {
		log.Printf("[BATCH/METRICS] ERROR : json : %s", err.Error())
		os.Exit(1)
	}
	log.Printf("[BATCH/METRICS] INFO : url=%s", url)
	// post on url metrics
	urlToPostMetrics := url + "/metrics/job/" + strings.ReplaceAll(data.Job, " ", "") + "/" + strings.ReplaceAll(data.Label, " ", "") + "/" + strconv.Itoa(data.Value)
	log.Printf("[BATCH/METRICS] INFO : post on %s", urlToPostMetrics)
	resp, err := http.Post(urlToPostMetrics, "", nil)
	if err != nil {
		log.Printf("[BATCH/METRICS] ERROR : " + err.Error())
		os.Exit(1)
	}
	log.Printf("[BATCH/METRICS] INFO : Response Status: %s", resp.Status)
	if resp.StatusCode == 200 {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// ---- swagger Informations
// @Tags         Networks
// @router /v1/network [get]
// @summary Scan port on ip/hosts
// @consume text/plain
// @param host query string false "host/ip address"
// @param port query string false "port number"
// @param protocol query string false "protocol"
// @produce text/plain
// @success 200 string OK
// @failure 500 string Internal Server Error
func networkHandler(c *gin.Context) {
	timeout := getenvs.GetEnvString("NETWORK_TIMEOUT", "5s")
	ptimeout, _ := time.ParseDuration(timeout)
	host := c.Query("host")
	port := c.Query("port")
	protocol := c.Query("protocol")

	if len(protocol) == 0 {
		protocol = "tcp"
	}

	log.Printf("[NETWORK] INFO : Parameters: host=%s, port=%s, protocol=%s, timeout=%s", host, port, protocol,timeout)

	testInputIP := net.ParseIP(host)
	if testInputIP.To4() != nil {
		addr, err := net.LookupAddr(host)
		log.Printf("[NETWORK] INFO : DNS name = %s", addr)
		if err != nil {
			log.Printf("[NETWORK] ERROR : dns : %s", err.Error())
		}
	}

	ip, err := net.LookupHost(host)
	log.Printf("[NETWORK] INFO : ip address = %s", ip)
	if err != nil {
		log.Printf("[NETWORK] ERROR : ip : %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
	} else {
		hosts := ip
		var ipV4 []string
	
		testInputIPHost := net.ParseIP(host)
		if testInputIPHost.To4() != nil { // ipv4 format
			ipV4 = append(ipV4, host)
		}
	
		for _, item := range hosts {
			testInputIP := net.ParseIP(item)
			if testInputIP.To4() != nil { // ipv4 format
				if item != host {
					ipV4 = append(ipV4, item)
				}
			}
		}
	
		log.Printf("[NETWORK] Checking : %v", strings.Join(ipV4, ","))
	
		// host -> the remote host
		// timeoutSecs -> the timeout value
		var resultConn []string
		for _, s := range ipV4 {
			conn, err := net.DialTimeout(protocol, s+":"+port, ptimeout)
			if err != nil {
				log.Printf("[NETWORK] ERROR : " + err.Error())
				resultConn = append(resultConn, fmt.Sprintf("Connection to %s on %s/%s is KO : %s", s, port, protocol, err.Error()))
			} else {
				conn.Close()
				log.Printf("[NETWORK] Connection to %s on %s/%s is OK", s, port, protocol)
				resultConn = append(resultConn, fmt.Sprintf("Connection to %s on %s/%s is OK", s, port, protocol))
			}
		}

		c.String(http.StatusOK, "Checking "+host+" on "+port+"/"+protocol+" : \n"+strings.Join(resultConn, "\n"))
	}


}
