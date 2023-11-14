package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
)

const dataBaseAddr = "localhost:6379"
const thisServerPort = "3333"

type myConn struct {
	net.Conn
}

func (dbConn *myConn) createShortLink(w http.ResponseWriter, r *http.Request) {
	tempReader := make([]byte, 1024)
	n, _ := r.Body.Read(tempReader)
	fmt.Println(string(tempReader[:n]))
}

func main() {
	conn, err := net.Dial("tcp", dataBaseAddr)
	if err != nil {
		fmt.Println("connection to database server cant be established")
		fmt.Print(err)
		os.Exit(1)
	}
	dbConn := &myConn{conn}
	dbConn.Write([]byte("ShortUrlServer"))
	http.HandleFunc("/", dbConn.createShortLink)
	err = http.ListenAndServe("localhost:"+thisServerPort, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else {
		fmt.Println("error starting server: ", err)
	}
}
