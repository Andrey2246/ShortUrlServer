package main

import (
	"errors"
	"fmt"
	"io"
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
	
}

func main() {
	conn, err := net.Dial("tcp", dataBaseAddr)
	if err != nil {
		fmt.Print("connection to database server cant be established")
		os.Exit(1)
	}
	db := &DB{conn}
	db.Write([]byte("ShortUrlServer"))
	server := http.Server{Addr: thisServerPort, Handler: nil}
	http.Get()
	http.Get("http://localhost:" + thisServerPort)
	http.HandleFunc("/", db.createShortLink)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed\n")
	} else {
		fmt.Println("error starting server: ", err)
	}
}
