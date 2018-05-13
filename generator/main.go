package main

import (
	"flag"
	"time"
	"net"
	"strconv"
	"log"
)


var (
	content chan int

	generationSpeed = flag.Int("speed", 1, "Speed in seconds")
	tcpAddress = flag.String("addr", "127.0.0.1:8080", "logger IP:port address")
)


// fibonacci generator. Returns fibonacci integers
func fibonacci() func() int {
	x, y := 0, 1
	return func() (r int) {
		r = x
		x, y = y, x + y
		return
	}
}


// Returns new connection or wait until connect
func getConnection() (conn net.Conn) {
	var err error
	for {
		conn, err = net.Dial("tcp", *tcpAddress)
		if err != nil {
			log.Println("Cant connect to logger:", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}

	log.Println("Connected to:", *tcpAddress)
	return conn
}


// Sender opens TCP connection and sends messages from queue
func sender() {
	var conn net.Conn

	defer func() {
		if conn == nil {
			return
		}
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

	conn = getConnection()

	// run loop to read values from queue
	for {
		value := <-content
		log.Println("Received new value", value)

		_, err := conn.Write([]byte(strconv.Itoa(value) + "\n"))
		if err != nil {
			content <- value
			conn = getConnection()
		}
	}
}


func main() {
	flag.Parse()

	content = make(chan int, 1)

	// run sender in background
	go sender()

	log.Println("Starting generator")

	// create content generator
	generator := fibonacci()

	// run generator loop
	for {
		content <- generator()
		time.Sleep(time.Second / time.Duration(*generationSpeed))
	}
}
