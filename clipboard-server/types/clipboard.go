package types

import (
	"fmt"
	"os/exec"
	"sync"
	"runtime"
)

type Command interface {
	copy () string
	paste () string
	process (content []byte) string
}

type MacOSCopy struct {
}

func (c MacOSCopy) copy() string {
	return "pbcopy"
}
func (c MacOSCopy) paste() string {
	return "pbpaste"
}

func (c MacOSCopy) process(content []byte) string {
	return string(content)
}

type WLCopy struct {
}

func (c WLCopy) copy() string {
	return "wl-copy"
}

func (c WLCopy) paste() string {
	return "wl-paste"
}

func (c WLCopy) process(content []byte) string {
	return string(content[:len(content) - 1])
}

// TODO: X11Copy

type Clip struct {
	Id      int
	Content string
}

type Clipboard struct {
	id          int
	Clips       map[int]string
	idGenerator IdGenerator
	command	    Command
}

func (c *Clipboard) GetLatestClipboardContent() (string, error) {
	cmd := exec.Command(c.command.paste())

	if bytes, err := cmd.Output(); err != nil {
		fmt.Println("Failed to run wl-paste", err)
		return "", nil
	} else {
		return c.command.process(bytes), nil
	}
}

func (c *Clipboard) GetClipboard() string {
	return c.Clips[c.id]
}

func (c *Clipboard) SetClipboard(content string) {
	// TODO: get command based on OS/rendered
	cmd := exec.Command(c.command.copy())
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

func (c *Clipboard) Clear(id int, mutex *sync.Mutex) {
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

	var command Command

	osType := runtime.GOOS

	if osType == "linux" {
		command = WLCopy{}
	} else if osType == "darwin" {
		command = MacOSCopy{}
	}


	return &Clipboard{
		id:          idGenerator.Next(),
		Clips:       make(map[int]string),
		idGenerator: idGenerator,
		command: command,
	}
}
