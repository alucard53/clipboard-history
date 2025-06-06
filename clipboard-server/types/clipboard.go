package types

import (
	"fmt"
	"os/exec"
	"sync"
)

type Clip struct {
	Id      int
	Content string
}

type Clipboard struct {
	id          int
	Clips       map[int]string
	idGenerator IdGenerator
}

func GetLatestClipboardContent() (string, error) {
	// TODO: get command based on OS/rendered
	cmd := exec.Command("pbpaste")

	if bytes, err := cmd.Output(); err != nil {
		fmt.Println("Failed to run wl-paste", err)
		return "", err
	} else {
		return string(bytes), nil
	}
}

func (c *Clipboard) GetClipboard() string {
	return c.Clips[c.id]
}

func (c *Clipboard) SetClipboard(content string) {
	// TODO: get command based on OS/rendered
	cmd := exec.Command("pbcopy")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		fmt.Println("Failed to get wl-copy stdin pipe", err)
		return
	}

	stdin.Write([]byte(content))
	stdin.Close()

	if err = cmd.Run(); err != nil {
		fmt.Println("failed to run wl-copy", err)
		return
	}
}

func (c *Clipboard) SetNew(content string) {
	nextId := c.idGenerator.Next()
	c.Clips[nextId] = content
	c.id = nextId
}

func (c *Clipboard) ClearAll() {
	clear(c.Clips)
	c.id = -1
	c.SetClipboard("")
}

func (c *Clipboard) Copy(newId int) {
	c.id = newId
	c.SetClipboard(c.Clips[newId])
}

func (c *Clipboard) Clear(id int, mutex *sync.RWMutex) {
	delete(c.Clips, id)

	if c.id == id {
		mutex.Lock()
		c.id = -1
		c.SetClipboard("")
		mutex.Unlock()
	}
}

func NewClipboard() *Clipboard {
	idGenerator := NewIdGenerator()
	return &Clipboard{
		id:          idGenerator.Next(),
		Clips:       make(map[int]string),
		idGenerator: idGenerator,
	}
}
