package main

import (
	"clipboard-server/idGen"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"
)

type CopyPayload struct {
	Id int
}

type Clip struct {
	Id int
	Content string
}

type Server struct {
	id           int
	clips        map[int]string
	conn		 *net.Conn
	idGenerator  idGen.IdGenerator
	channel      chan bool
	mutex	     *sync.RWMutex
}

func getCurrentClipboardContent() (string, error) {
	// TODO: get command based on OS/rendered
	cmd := exec.Command("pbpaste")

	if bytes, err := cmd.Output();  err != nil {
		fmt.Println("Failed to run wl-paste", err)
		return "", err
	} else {
		return string(bytes), nil
	}
}

func (s *Server) copy(
	newId int,
) {
	// TODO: get command based on OS/rendered
	cmd := exec.Command("pbcopy")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		fmt.Println("Failed to get wl-copy stdin pipe", err)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id = newId
	stdin.Write([]byte(s.clips[newId]))
	stdin.Close()

	if err = cmd.Run(); err != nil {
		fmt.Println("failed to run wl-copy", err)
		return
	}
}

func(s *Server) clipboardListener() {
	for {
		time.Sleep(1 * time.Second)
		content, err := getCurrentClipboardContent()

		if err != nil {
			fmt.Println("Failed to read clipboard content", err)
			continue
		}

		s.mutex.RLock()
		if content == s.clips[s.id] {
			s.mutex.RUnlock()
			continue
		}
		s.mutex.RUnlock()

		s.mutex.Lock()
		nextId := s.idGenerator.Next()
		s.clips[nextId] = content
		s.id = nextId

		if s.conn != nil {
			s.send(s.conn)
		}
		s.mutex.Unlock()
	}
}

func initialize() Server {
	idGenerator := idGen.New()
	id := idGenerator.Next()
	clips := make(map[int]string)

	server := Server{
		id,
		clips,
		nil,
		idGenerator,
		make(chan bool),
		&sync.RWMutex{},
	}
	content, err := getCurrentClipboardContent()

	if err != nil {
		fmt.Println("Failed to read initial clipboard content", err)
		return server
	}

	server.clips[id] = content

	return server
}

func(s *Server) clientListener(
) {
	for {
		payload := CopyPayload{}
		err := json.NewDecoder(*s.conn).Decode(&payload)
		if err != nil {
			fmt.Println("Error in decoding client msg", err)
			if s.conn != nil {
				(*s.conn).Close()
			}
			return
		}
		s.copy(payload.Id)
	}
}


func(s *Server) send(conn *net.Conn) {
	data := []Clip{}

	for Id, Content := range s.clips {
		data = append(
			data,
			Clip{
				Id,
				Content,
			},
		)
	}

	json.NewEncoder(*conn).Encode(data)
}

func main() {
	server := initialize()

	go server.clipboardListener()

	listener, err := net.Listen("tcp", ":6969")

	if err != nil {
		fmt.Println("Failed to bind", err)
	} else {
		fmt.Println("Listening at port 6969")
	}

	for  {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to handle listener")
		}

		server.conn = &conn

		data := []Clip{}

		for Id, Content := range server.clips {
			data = append(
				data,
				Clip{
					Id,
					Content,
				},
			)
		}

		json.NewEncoder(conn).Encode(data)

		go server.clientListener()
	}
}
