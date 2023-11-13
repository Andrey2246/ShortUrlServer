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

type DB struct {
	net.Conn
}

func (db *DB) createShortLink(w http.ResponseWriter, r *http.Request) {
	tempReader := make([]byte, 1024)
	n, _ := r.Body.Read(tempReader)
	fmt.Println(string(tempReader[:n]))
}

func main() {
	conn, err := net.Dial("tcp", dataBaseAddr)
	if err != nil {
		fmt.Print("connection to database server cant be established")
		os.Exit(1)
	}
	db := &DB{conn}
	db.Write([]byte("ShortUrlServer"))
	http.HandleFunc("/", db.createShortLink)
	err = http.ListenAndServe("localhost:"+thisServerPort, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed\n")
	} else {
		fmt.Println("error starting server: ", err)
	}
}
