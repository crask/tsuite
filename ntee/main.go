package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"
)

var (
	localAddr   string
	remoteAddr  string
	inputDelay  uint
	inputFile   string
	outputDelay uint
	outputFile  string
)

const (
	istream = 0
	ostream = 1
)

func init() {
	flag.StringVar(&localAddr, "l", ":7777", "set local address")
	flag.StringVar(&remoteAddr, "r", "localhost:7777", "set remote address")
	flag.UintVar(&inputDelay, "id", 0, "set input delay ")
	flag.UintVar(&outputDelay, "od", 0, "set output delay ")
	flag.StringVar(&inputFile, "if", "/dev/stdout", "set input log")
	flag.StringVar(&outputFile, "of", "/dev/stdout", "set output log")
}

func main() {
	flag.Parse()

	server, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Print(err)
		} else {
			handleConn(conn)
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	remote, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Print(err)
		return
	}

	defer remote.Close()

	complete := make(chan bool, 2)
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)

	go copyContent(conn, remote, complete, ch1, ch2, istream)
	go copyContent(remote, conn, complete, ch2, ch1, ostream)

	<-complete
	<-complete
}

func copyContent(from net.Conn, to net.Conn, complete chan bool, done chan bool, otherDone chan bool, stream int) {
	var err error = nil
	var bytes []byte = make([]byte, 256)
	var read int = 0
	var delay uint = 0
	var file *os.File
	var filename string

	if stream == istream {
		delay = inputDelay
		filename = inputFile
	} else {
		delay = outputDelay
		filename = outputFile
	}

	file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		select {
		// If we received a done message from the other goroutine, we exit.
		case <-otherDone:
			complete <- true
			return
		default:
			// Read data from the source connection.
			read, err = from.Read(bytes)
			// If any errors occured, write to complete as we are done (one of the
			// connections closed.)
			if err != nil {
				complete <- true
				done <- true
				from.Close()
				to.Close()
				return
			}

			file.Write(bytes[:read])
			time.Sleep(time.Duration(delay) * time.Millisecond)

			// Write data to the destination.
			_, err = to.Write(bytes[:read])
			// Same error checking.
			if err != nil {
				complete <- true
				done <- true
				from.Close()
				to.Close()
				return
			}
		}
	}
}
