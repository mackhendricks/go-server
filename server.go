package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	hep "github.com/dOpensource/hep"
)

const (
	port            = "8080"
	bufferSize      = 4048
	DEFAULT_DB_HOST = "localhost:3306"
	DEFAULT_DB_NAME = "go-server"
	DEFAULT_DB_USER = "root"
	DEFAULT_DB_PASS = ""
	DEFAULT_DEBUG   = "false"
)

type CONFIG struct {
	db_host string
	db_name string
	db_user string
	db_pass string
	debug   string
}

// HandleClient manages individual client connections
func handleClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	buffer := make([]byte, bufferSize)

	// Read data from client
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading from client: %v", err)
		return
	}

	// Echo the raw message back
	message := string(buffer[:n])
	log.Printf("Received: %s", message)

	// Parse the HEP Packet
	hep_message, err := hep.NewHepMsg(buffer[:n])
	if hep_message != nil {
		log.Printf("HEP Source Address %s", hep_message.IP4SourceAddress)
		log.Printf("HEP Dest Address %s", hep_message.IP4DestinationAddress)
		log.Printf("HEP Body:\n %s", hep_message.Body)
		log.Printf("_+_+_+_+_+_+_+_+_+_+_+_+_+_++_+_+_+_+_+_+_+_+_+_+")
	}
	//sipmsg := hep_message.SipMsg
	//log.Printf("HEP From %s", sipmsg.From.Val)
	//log.Printf("HEP To %s", sipmsg.To.Val)

	//response := fmt.Sprintf("Server received: %s", hep_message.Body)
	//conn.Write([]byte(response))
	buffer = nil
}

// Get Configuration from Environment Variables
func getEnvionmentVariables() *CONFIG {

	var config CONFIG

	config.db_host = os.Getenv("DB_HOST")
	if config.db_host == "" {
		config.db_host = DEFAULT_DB_HOST
	}

	config.db_name = os.Getenv("DB_NAME")
	if config.db_name == "" {
		config.db_name = DEFAULT_DB_NAME
	}

	config.db_user = os.Getenv("DB_USER")
	if config.db_user == "" {
		config.db_user = DEFAULT_DB_USER
	}

	config.db_pass = os.Getenv("DB_PASS")
	if config.db_pass == "" {
		config.db_pass = DEFAULT_DB_PASS
	}

	config.debug = os.Getenv("DEBUG")
	if config.debug == "" {
		config.debug = DEFAULT_DEBUG

	}

	return &config
}

func main() {
	// Optimize CPU usage
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Print Environment Variable if DEBUG is enabled
	config := getEnvionmentVariables()
	if config.debug == "true" || config.debug == "TRUE" {
		log.Printf("+Debug Enabled\n")
		log.Printf("Environment Variables:\n")
		log.Printf("-----------------------")
		log.Printf("DB_HOST: %s\nDB_NAME: %s\nDB_USER: %s\nDB_PASS: %s", config.db_host, config.db_name, config.db_user, config.db_pass)
	}

	// Listen on TCP port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP server is running on port %s...", port)

	// Wait group to track active connections
	var wg sync.WaitGroup

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-stop:
					return // Stop accepting new connections
				default:
					log.Printf("Connection accept error: %v", err)
				}
				continue
			}

			wg.Add(1)
			go handleClient(conn, &wg)
		}
	}()

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down server...")

	// Stop accepting new connections
	listener.Close()

	// Wait for all active connections to finish
	wg.Wait()
	log.Println("Server exited properly")
}
