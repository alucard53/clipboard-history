package main

import (
	"fmt"
	"net"
	"github.com/AllenDang/giu"
	"encoding/json"
)

type Clip struct {
	Id int
	Content string
}

type Client struct {
	conn *net.Conn
	clips []Clip
}

type Payload struct {
	Id   int
	Type string
}

func(c *Client) loop() {
	rows := giu.Layout{}

	for _, clip := range c.clips {
		rows = append(rows, giu.Row(
			giu.Button(clip.Content).OnClick(func() {
				json.NewEncoder(*c.conn).Encode(Payload{
					clip.Id,
					"copy",
				})
			}),
			giu.Button("clear").OnClick(func() {
				json.NewEncoder(*c.conn).Encode(Payload{
					clip.Id,
					"clear",
				})
			}),
		))
	}

	giu.SingleWindow().Layout(
		giu.Row(
			giu.Label("Clipboard History"),
			giu.Button("Clear all").OnClick(func() {
				c.clips = []Clip{}
				json.NewEncoder(*c.conn).Encode(Payload{
					Type: "clearAll",
				})
			}),
		),
		rows,
	)
}

func (c *Client) listener() {
	for {
		clips := []Clip{}
		json.NewDecoder(*c.conn).Decode(&clips)
		c.clips = clips
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		fmt.Println("Failed to connect to server", err)
		return
	}

	defer conn.Close()
	client := Client{&conn, []Clip{}}
	go client.listener()

	wnd := giu.NewMasterWindow("Clipboard History", 500, 500, 0)
	wnd.Run(client.loop)
}
