package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Client struct {
	IP          string
	PORT        string
	Username    string
	IsConnected bool
	Conn        net.Conn
}

type Server struct {
	IP          string
	PORT        string
	Clients     []Client
	Connections int
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func New(IP string, PORT string) *Client {
	return &Client{
		IP:   IP,
		PORT: PORT,
	}
}

const ClearLine = "\033[2K"

func (client *Client) Connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", client.IP, client.PORT))
	handleError(err)
	fmt.Println("Client connected to server", client.IP)
	client.IsConnected = true
	client.Conn = conn
}

func (client *Client) Disconnect() {
	fmt.Println("You are disconnected from server ! Goodbye")
	err := client.Conn.Close()
	handleError(err)
}

func (client *Client) CreateUsername() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choose your username: ")
	username, err := reader.ReadString('\n')
	handleError(err)
	check := CheckUsername(string(username), client.Conn)
	if check {
		client.Username = string(username)
		client.ready()
	} else {
		fmt.Println("Please choose another, this one already exists")
		client.CreateUsername()
	}
}

func (client *Client) ready() {
	fmt.Println("Welcome on this server, enjoy !")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { // client message send to the server
		defer wg.Done()
		for {
			if !client.IsConnected {
				break
			}
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			fmt.Print(ClearLine)
			fmt.Printf("\n")
			handleError(err)
			client.Conn.Write([]byte(text))
		}
	}()

	go func() { // message reception from server
		defer wg.Done()
		for {
			fmt.Print(ClearLine)
			fmt.Printf("\n")
			message, err := bufio.NewReader(client.Conn).ReadString('\n')
			if err != nil {
				client.IsConnected = false
				log.Fatal("Server is offline, please retry later")
				return
			}
			fmt.Print(message)
		}
	}()

	wg.Wait()
}

func CheckUsername(username string, conn net.Conn) bool {
	conn.Write([]byte(username))
	messageBuffer := make([]byte, 4096)
	status, err := conn.Read(messageBuffer)
	if err != nil {
		log.Fatal("Error during your access, server is down")
	}
	if string(messageBuffer[:status]) != "ERROR" {
		return true
	} else {
		return false
	}
}
