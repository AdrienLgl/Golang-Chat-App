package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	SEND       = "send_message"
	CONNECT    = "connection"
	DISCONNECT = "disconnect"
)

type Client struct {
	Connection net.Conn
	Username   string
}

type Server struct {
	IP          string
	PORT        string
	Clients     map[net.Conn]Client
	Connections int
	Listener    net.Listener
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func New(IP string, PORT string) *Server {
	return &Server{
		IP:      IP,
		PORT:    PORT,
		Clients: make(map[net.Conn]Client),
	}
}

func (server *Server) Run() {
	fmt.Println("Lancement du serveur...")
	// listening on 3569 port
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.IP, server.PORT))
	server.Listener = ln
	handleError(err)
	for {
		conn, err := server.Listener.Accept() // port 3569 authorization
		if err != nil {
			server.Clients[conn] = Client{Connection: conn} // add new client
		}
		go server.addClient(conn)
	}
}

func (server *Server) Close() {
	for _, c := range server.Clients {
		fmt.Println("Client was close: ", c.Connection.RemoteAddr())
		err := c.Connection.Close()
		handleError(err)
	}
	err := server.Listener.Close()
	handleError(err)
	fmt.Println("Server is closing...")
}

func (server *Server) catchClientUsername(client net.Conn) (string, error) {
	var err error

	usernameBuffer := make([]byte, 4096)
	length, err := client.Read(usernameBuffer)

	username := string(usernameBuffer[:length]) // remove all unused bytes in the buffer and convert it to string

	if err != nil { // if the client interrupt the username input
		// server.addLog(fmt.Sprintf("Client from %s interrupt the username input\n", client.RemoteAddr()))
		username = "error"
		err = errors.New("client interrupt input")
	}

	return strings.TrimSuffix(username, "\n"), err
}

func (server *Server) isUsernameExists(username string, client net.Conn) bool {
	findUsername := false
	for c := range server.Clients {
		if server.Clients[c].Username == username { // we ignore the client itself comparaison because he is already in the clients Map()
			fmt.Println(fmt.Sprintf("The username %s already exist !\n", username) + "\n")
			findUsername = true
			break
		}
	}
	return findUsername
}

func (server *Server) checkUsernameClient(client net.Conn) bool {
	var (
		username string
		err      error
	)

	for {
		username, err = server.catchClientUsername(client)
		if err != nil {
			return false
		}

		if server.isUsernameExists(username, client) {
			client.Write([]byte("ERROR"))
		} else {
			client.Write([]byte("OK"))
			break
		}
	}

	server.Clients[client] = Client{Username: username, Connection: client}
	return true
}

func (server *Server) addClient(client net.Conn) {
	fmt.Println("New client: ", client.RemoteAddr())
	if !server.checkUsernameClient(client) {
		server.addLog(client, fmt.Sprintf("%s: %s", CONNECT, "incorrect username"))
		return // exit the goroutine because the client interrupt the username input
	}

	message := fmt.Sprintf("[INFO] %s join the server\n", server.Clients[client].Username)
	server.addLog(client, fmt.Sprintf("%s: %s", CONNECT, message))
	fmt.Println(message)

	if len(server.Clients) > 1 {
		client.Write([]byte("List of usernames in the server:\n"))
		for c := range server.Clients {
			client.Write([]byte(fmt.Sprintf("-> %s\n", server.Clients[c].Username)))
		}
	}

	client.Write([]byte("You can start the discussion with guests ...\n\n"))
	server.sendToAll(client, message, true)
	server.receive(client)
}

func (server *Server) receive(client net.Conn) {
	buf := bufio.NewReader(client)
	for {
		if message, err := buf.ReadString('\n'); err != nil {
			server.removeClient(client)
			break
		} else {
			message = fmt.Sprintf("%s: %s", server.Clients[client].Username, message)
			fmt.Println(message)
			server.addLog(client, fmt.Sprintf("%s: %s", SEND, message))
			server.sendToAll(client, message, true)
		}
	}
}

func (server *Server) removeClient(client net.Conn) {
	message := fmt.Sprintf("[INFO] %s is now disconnected\n", server.Clients[client].Username)
	fmt.Println(message)
	server.sendToAll(client, message, true)
	server.addLog(client, fmt.Sprintf("%s: %s", DISCONNECT, message))
	delete(server.Clients, client)
}

func (server *Server) sendToAll(client net.Conn, message string, ignoreItself bool) {
	for c := range server.Clients {
		if ignoreItself {
			if c != client { // we do not send the message back to the same sender
				c.Write([]byte(message))
			}
		} else {
			c.Write([]byte(message))
		}
	}
}

func (server *Server) addLog(client net.Conn, action string) {
	f, err := os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Printf("[%d] - %d - action: %s \n", time.Now(), client.RemoteAddr(), action)
}
