package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const dataBaseAddr = "localhost:6379"
const thisServerAddr = "192.168.31.178"
const statServerAddr = "192.168.31.178"

//const thisServerAddr = "10.241.88.151:3333"
//const statServerAddr = "10.241.88.151:3333"

const thisServerPort = ":3333"
const statServerPort = ":3030"

// const thisServerAddr = "10.241.88.151:3333"

var dbConn net.Conn

type LinkFollow struct {
	Url  string
	Ip   string
	Time string
}

func dbWriteRead(command string) string {
	fmt.Printf("dbConn: %v\n", dbConn.RemoteAddr().String())
	_, err := dbConn.Write([]byte(command + " \n"))
	log.Println(err)
	dbResponse := make([]byte, 1024)
	n, _ := dbConn.Read(dbResponse)
	log.Println("Wrote to DB command " + command + ".\nGot response: " + string(dbResponse[:n]) + "\n")
	return string(dbResponse[:n])

}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/shorten", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<form method="post" action="/shorten">
				<input type="url" name="url" placeholder="Enter a URL" required>
				<input type="submit" value="Shorten">
			</form>
		</body>
		</html>
	`)
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	fmt.Println(dbConn.RemoteAddr().String())
	if r.Method != http.MethodPost {

		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	shortKey := strconv.FormatInt(int64(generateShortKey(originalURL)), 10)
	dbWriteRead("HSET " + shortKey + " " + originalURL)

	shortenedURL := fmt.Sprintf("http://%s/s/%s", thisServerAddr+thisServerPort, shortKey) // Serve the result page
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<p>Original URL: `, originalURL, `</p>
			<p>Shortened URL: <a href="`, shortenedURL, `">`, shortenedURL, `</a></p>
		</body>
		</html>
	`)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.String()[3:]
	if shortKey == "" {
		http.Error(w, "Shortened key is missing", http.StatusBadRequest)
		return
	}
	dbResponse := dbWriteRead("HGET " + shortKey)
	if dbResponse == "no such key" {
		http.Error(w, "Short key not found", 404)
		return
	}

	timePeriod := strconv.Itoa(time.Now().Hour()) + ":" + strconv.Itoa(time.Now().Minute())
	timePeriod += "-" + strconv.Itoa(time.Now().Add(time.Minute).Hour()) + ":" + strconv.Itoa(time.Now().Add(time.Minute).Minute())
	stat := LinkFollow{r.RemoteAddr, dbResponse[:len(dbResponse)-1] + " (" + shortKey + ")", timePeriod}
	b, _ := json.Marshal(stat)
	_, err := http.Post("http://"+statServerAddr+statServerPort+"/", "application/json", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
		log.Println("cant POST to stat server")
	}

	http.Redirect(w, r, string(dbResponse), http.StatusMovedPermanently)
}

func generateShortKey(source string) int {
	ans := 1
	for _, l := range source {
		ans = (ans+int(l))%1024 + 1
	}
	return ans
}

func main() {
	var err error
	dbConn, err = net.Dial("tcp", dataBaseAddr)
	if err != nil {
		fmt.Println("connection to database server cant be established")
		fmt.Print(err)
		os.Exit(1)
	}
	dbWriteRead("ShortUrlServer")
	dbWriteRead("HMAKE 1024")

	http.HandleFunc("/", handleForm)
	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/s/", handleRedirect)

	err = http.ListenAndServe(thisServerAddr+thisServerPort, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else {
		fmt.Println("error starting server: ", err)
	}
}
