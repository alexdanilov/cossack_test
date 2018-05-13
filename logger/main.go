package main

import (
	"fmt"
	"net"
	"flag"
	"os"
	"time"
	"log"
	"bufio"
	"sync"
	"crypto/aes"
	"io"
	"crypto/rand"
	"crypto/cipher"
)


type Message struct {
	Client string
	Value string
	Time time.Time
}

var (
	speed int
	content chan Message

	mu sync.RWMutex

	key = flag.String("key", "", "Encrypt key")
	bufferSize = flag.Int("buffer_size", 1, "Buffer size")
	flowSpeed = flag.Int("flow_speed", 2, "Logger flow speed")
	logFile = flag.String("file_path", "logger.log", "Path to logger file")
	tcpAddress = flag.String("addr", ":8080", "logger listen :port")
)


// Encrypt message content
// Example got from: https://gist.github.com/kkirsche/e28da6754c39d5e7ea10
func (m *Message) Encrypt() {
	block, err := aes.NewCipher([]byte(*key))
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	// encrypt message.Value
	m.Value = string(aesgcm.Seal(nil, nonce, []byte(m.Value), nil))
}


// Writer opens log file and write all messages from queue
func writer() {
	logFileHandler, err := os.Create(*logFile)
	if err != nil {
		panic(err)
	}

	// close logFileHandler on writer exit
	defer func() {
		if err := logFileHandler.Close(); err != nil {
			panic(err)
		}
	}()

	log.Println("Flush buffer content to log file...")

	for {
		if len(content) == 0 {
			// finish writer
			break
		}

		msg := <- content

		if len(*key) > 0 {
			msg.Encrypt()
		}
		fmt.Fprintf(logFileHandler, "%s [%s]: %s", msg.Client, msg.Time.Format(time.RFC3339), msg.Value)
	}
}


// Serve gets new connection and run loop to get messages from connection and puts to a queue
func serve(conn net.Conn) {
	defer conn.Close()

	client := conn.RemoteAddr().String()
	log.Println("New client connected:", client)

	for {
		// listen for message ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Close connection for", client)
			break
		}

		// update speed counter
		mu.Lock()
		speed += 1
		if (speed > *flowSpeed) {
			log.Println("WARNING: Flow speed is exited. Speed", speed, "rps")
		}
		mu.Unlock()

		// output received message
		log.Print("Message Received: ", message)

		// write message to queue
		content <- Message{client, message, time.Now()}

		if len(content) >= *bufferSize {
			// buffer is full. Write content to a file
			go writer()
		}
	}
}


func main() {
	flag.Parse()
	log.Println("Staring logger server...")

	// init queue with given buffer size
	content = make(chan Message, *bufferSize)

	// run timer to reset speed counter
	tick := time.Tick(time.Second)
	go func() {
		for {
			select {
			case <-tick:
				mu.Lock()
				speed = 0
				mu.Unlock()
			}
		}
	}()

	// listen port
	ln, _ := net.Listen("tcp", *tcpAddress)

	// run loop to accept clients connections
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serve(c)
	}
}