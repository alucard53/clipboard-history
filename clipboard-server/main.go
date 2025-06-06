package main

import (
	"clipboard-server/types"
	"cmp"
	"encoding/json"
	"fmt"
	"net"
	"slices"
	"sync"
	"time"
)

type Payload struct {
	Id   int
	Type string
}

type Server struct {
	conn      *net.Conn
	mutex     *sync.RWMutex
	clipboard *types.Clipboard
}

func (s *Server) clearAll() {
	s.mutex.Lock()
	s.clipboard.ClearAll()
	s.mutex.Unlock()
	s.send(s.conn)
}

func (s *Server) clear(id int) {
	s.clipboard.Clear(id, s.mutex)

	s.send(s.conn)
}

func (s *Server) copy(
	newId int,
) {
	s.mutex.Lock()
	s.clipboard.Copy(newId)
	s.mutex.Unlock()
}

func (s *Server) clipboardListener() {
	for {
		time.Sleep(1 * time.Second)
		content, err := types.GetLatestClipboardContent()

		if err != nil {
			fmt.Println("Failed to read clipboard content", err)
			continue
		}

		s.mutex.RLock()
		if content == s.clipboard.GetClipboard() {
			s.mutex.RUnlock()
			continue
		}
		s.mutex.RUnlock()

		s.mutex.Lock()
		s.clipboard.SetNew(content)

		if s.conn != nil {
			s.send(s.conn)
		}
		s.mutex.Unlock()
	}
}

func initialize() Server {
	server := Server{
		conn:      nil,
		clipboard: types.NewClipboard(),
		mutex:     &sync.RWMutex{},
	}
	content, err := types.GetLatestClipboardContent()

	if err != nil {
		fmt.Println("Failed to read initial clipboard content", err)
		return server
	}

	if content != "" {
		server.clipboard.SetNew(content)
	}

	return server
}

func (s *Server) clientListener() {
	for {
		payload := Payload{}
		err := json.NewDecoder(*s.conn).Decode(&payload)
		if err != nil {
			fmt.Println("Error in decoding client msg", err)
			if s.conn != nil {
				(*s.conn).Close()
			}
			return
		}

		switch payload.Type {
		case "copy":
			s.copy(payload.Id)
		case "clear":
			s.clear(payload.Id)
		case "clearAll":
			s.clearAll()
		default:
			fmt.Println("invalid instruction type from client")
		}
	}
}

func (s *Server) send(conn *net.Conn) {
	data := []types.Clip{}

	for Id, Content := range s.clipboard.Clips {
		data = append(
			data,
			types.Clip{
				Id:      Id,
				Content: Content,
			},
		)
	}

	slices.SortFunc(
		data,
		func(a types.Clip, b types.Clip) int { return cmp.Compare(b.Id, a.Id) },
	)

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

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to handle listener")
		}

		server.conn = &conn

		data := []types.Clip{}

		for Id, Content := range server.clipboard.Clips {
			data = append(
				data,
				types.Clip{
					Id:      Id,
					Content: Content,
				},
			)
		}

		slices.SortFunc(
			data,
			func(a types.Clip, b types.Clip) int { return cmp.Compare(b.Id, a.Id) },
		)

		json.NewEncoder(conn).Encode(data)

		go server.clientListener()
	}
}
