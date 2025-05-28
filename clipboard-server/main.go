package main

import (
	"fmt"
	"os/exec"
	"time"
	"net"
	"encoding/json"
	"clipboard-server/idGen"
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
	idGenerator  idGen.IdGenerator
	channel      chan bool
}

func getCurrentClipboardContent() (string, error) {
	cmd := exec.Command("wl-paste")

	if bytes, err := cmd.Output();  err != nil {
		fmt.Println("Failed to run wl-paste", err)
		return "", err
	} else {
		return string(bytes[:len(bytes) - 1]), nil
	}
}

func (s *Server) copy(
	newId int,
) {
	cmd := exec.Command("wl-copy")
	stdin, err := cmd.StdinPipe()
	
	if err != nil {
		fmt.Println("Failed to get wl-copy stdin pipe", err)
		return
	}
	
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
		defer time.Sleep(1 * time.Second)
		content, err := getCurrentClipboardContent()

		if err != nil {
			fmt.Println("Failed to read clipboard content", err)
			continue
		}

		if content != s.clips[s.id] {
			nextId := s.idGenerator.Next()
			s.clips[nextId] = content
			s.id = nextId
			s.channel <- true
		}
	}
}

func initialize() Server {
	idGenerator := idGen.New()
	id := idGenerator.Next()
	clips := make(map[int]string)

	server := Server{
		id,
		clips,
		idGenerator,
		make(chan bool),
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
	conn *net.Conn,
) {
	for {
		payload := CopyPayload{}
		err := json.NewDecoder(*conn).Decode(&payload)
		if err != nil {
			fmt.Println("Error in decoding client msg", err)
			s.channel <- false
			return
		}
		s.copy(payload.Id)
	}
}


func(s *Server) clientSender(conn *net.Conn) {
	for {
		stop := !<-s.channel
		if stop {
			fmt.Println("cleaning up")
			return
		}
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

		go server.clientListener(&conn)
		go server.clientSender(&conn)
	}
}
