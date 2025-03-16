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
	port       = "8080"
	bufferSize = 4048
)

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

func main() {
	// Optimize CPU usage
	runtime.GOMAXPROCS(runtime.NumCPU())

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
