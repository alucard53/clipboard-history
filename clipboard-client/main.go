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

type CopyPayload struct {
	Id int
}

func(c *Client) loop() {
	rows := giu.Layout{}

	for _, clip := range c.clips {
		rows = append(rows, giu.Row(
			giu.Button(clip.Content).OnClick(func() {
				json.NewEncoder(*c.conn).Encode(CopyPayload{clip.Id})
			}),
		))
	}


        giu.SingleWindow().Layout(
                giu.Label("Clipboard History"),
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

        wnd := giu.NewMasterWindow("Clipboard History", 800, 800, giu.MasterWindowFlagsNotResizable)
	wnd.Run(client.loop)
}
