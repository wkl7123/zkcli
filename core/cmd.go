package core

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/json"
	"github.com/samuel/go-zookeeper/zk"
	"io/ioutil"
	"reflect"
)

const flag int32 = 0

var acl = zk.WorldACL(zk.PermAll)
var ErrUnknownCmd = errors.New("unknown command")

type Cmd struct {
	Name        string
	Options     []string
	ExitWhenErr bool
	Conn        *zk.Conn
	Config      *Config
}

func NewCmd(name string, options []string, conn *zk.Conn, config *Config) *Cmd {
	return &Cmd{
		Name:    name,
		Options: options,
		Conn:    conn,
		Config:  config,
	}
}

func ParseCmd(cmd string) (name string, options []string) {
	args := make([]string, 0)
	for _, cmd := range strings.Split(cmd, " ") {
		if cmd != "" {
			args = append(args, cmd)
		}
	}
	if len(args) == 0 {
		return
	}

	return args[0], args[1:]
}

func (c *Cmd) addHistory() {
	f, err := os.OpenFile(c.Config.HistoryFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	output := c.Name
	for _, elem := range c.Options {
		output += " " + elem
	}
	f.WriteString(output + "\n")

}

func (c *Cmd) ls() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	p := "/"
	options := c.Options
	if len(options) > 0 {
		p = options[0]
	}
	cleanPath(p)
	children, _, err := c.Conn.Children(p)
	if err != nil {
		return
	}
	fmt.Printf("[%s]\n", strings.Join(children, ", "))
	return
}

func (c *Cmd) get() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	p := "/"
	options := c.Options
	if len(options) > 0 {
		p = options[0]
	}
	p = cleanPath(p)
	value, _, err := c.Conn.Get(p)
	if err != nil {
		return
	}

	if len(options) >= 2 {
		var j interface{}
		er := json.Unmarshal(value, &j)
		if er != nil {
			fmt.Println("not a valid json structure")
			return
		}
		jm := j.(map[string]interface{})
		fields := strings.Split(strings.Trim(options[1], "/"), "/")
		n := 0
		for n <= len(fields)-2 {
			if reflect.TypeOf(jm[fields[n]]).String() == "string" {
				fmt.Println("wrong path")
				return
			}
			jm = (jm[fields[n]]).(map[string]interface{})
			n += 1
		}
		jsonStr, e := json.MarshalIndent(jm[fields[n]], "", "    ")
		if e != nil {
			fmt.Println(e.Error())
			return
		}
		fmt.Println(string(jsonStr))
		return
	}
	fmt.Printf("%+v\n", string(value))
	return
}

func (c *Cmd) gf() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	p := "/"
	options := c.Options
	if len(options) != 2 {
		return
	}

	p = options[0]
	p = cleanPath(p)
	path := options[1]
	value, _, err := c.Conn.Get(p)
	if err != nil {
		return
	}
	fd, err := os.Create(path)
	defer fd.Close()
	_, err = fd.Write(value)
	if err == nil {
		fmt.Println("ok")
	}
	return
}

func (c *Cmd) create() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	p := "/"
	data := ""
	options := c.Options
	if len(options) > 0 {
		p = options[0]
		if len(options) > 1 {
			data = options[1]
		}
	}
	cleanPath(p)
	_, err = c.Conn.Create(p, []byte(data), flag, acl)
	if err != nil {
		return
	}
	fmt.Printf("Created %s\n", p)
	root, _ := splitPath(p)
	suggestCache.del(root)
	return
}

func (c *Cmd) set() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	p := "/"
	data := ""
	options := c.Options
	if len(options) > 0 {
		p = options[0]
		if len(options) > 1 {
			data = options[1]
		}
	}
	cleanPath(p)
	_, err = c.Conn.Set(p, []byte(data), -1)
	return err
}

func (c *Cmd) sf() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	options := c.Options
	if len(options) != 2 {
		return
	}

	p := options[0]
	filePath := options[1]
	data, err := ioutil.ReadFile(filePath)
	cleanPath(p)
	_, err = c.Conn.Set(p, []byte(data), -1)
	if err == nil {
		fmt.Println("ok")
	}
	return err
}

func (c *Cmd) delete() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	p := "/"
	options := c.Options
	if len(options) > 0 {
		p = options[0]
	}
	cleanPath(p)
	err = c.Conn.Delete(p, -1)
	if err != nil {
		return
	}
	fmt.Printf("Deleted %s\n", p)
	root, _ := splitPath(p)
	suggestCache.del(root)
	return
}

func (c *Cmd) close() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	c.Conn.Close()
	if !c.connected() {
		fmt.Println("Closed")
	}
	return
}

func (c *Cmd) connect() (err error) {
	options := c.Options
	var conn *zk.Conn
	if len(options) > 0 {
		cf := NewConfig(strings.Split(options[0], ","), c.Config.HistoryFilePath)
		conn, err = cf.Connect()
		if err != nil {
			return err
		}
	} else {
		conn, err = c.Config.Connect()
		if err != nil {
			return err
		}
	}
	if c.connected() {
		c.Conn.Close()
	}
	c.Conn = conn
	fmt.Println("Connected")
	return err
}

func (c *Cmd) addAuth() (err error) {
	err = c.checkConn()
	if err != nil {
		return
	}

	options := c.Options
	if len(options) < 2 {
		return errors.New("addauth <scheme> <auth>")
	}
	scheme := options[0]
	auth := options[1]
	err = c.Conn.AddAuth(scheme, []byte(auth))
	if err != nil {
		return
	}
	fmt.Println("Added")
	return err
}

func (c *Cmd) connected() bool {
	state := c.Conn.State()
	return state == zk.StateConnected || state == zk.StateHasSession
}

func (c *Cmd) checkConn() (err error) {
	if !c.connected() {
		err = errors.New("connection is disconnected")
	}
	return
}

func (c *Cmd) run() (err error) {
	switch c.Name {
	case "ls":
		c.addHistory()
		return c.ls()
	case "get":
		c.addHistory()
		return c.get()
	case "gf":
		c.addHistory()
		return c.gf()
	case "sf":
		c.addHistory()
		return c.sf()
	case "create":
		c.addHistory()
		return c.create()
	case "set":
		c.addHistory()
		return c.set()
	case "delete":
		c.addHistory()
		return c.delete()
	case "close":
		c.addHistory()
		return c.close()
	case "connect":
		c.addHistory()
		return c.connect()
	case "addauth":
		c.addHistory()
		return c.addAuth()
	case "help":
		printHelp()
		return nil
	case "":
		return nil
	default:
		return ErrUnknownCmd
	}
}

func (c *Cmd) Run() {
	err := c.run()
	if err != nil {
		if err == ErrUnknownCmd {
			printUnsupported()
			if c.ExitWhenErr {
				os.Exit(2)
			}
		} else {
			printRunError(err)
			if c.ExitWhenErr {
				os.Exit(3)
			}
		}
	}
}

func printUnsupported() {
	fmt.Println("error: unsupported command, use 'help' to know more")
}

func printHelp() {
	fmt.Println(`    ls <path>
    get <path> <field[/<subField>][/<subField]>
    set <path> [<data>]
	gf <path> <filePath>
	sf <path> <filePath>
    create <path> [<data>]
    delete <path>
    connect <host:port>
    addauth <scheme> <auth>
    close
    exit`)
}

func printRunError(err error) {
	fmt.Println(err)
}

func cleanPath(p string) string {
	if p == "/" {
		return p
	}
	return strings.TrimRight(p, "/")
}

func GetExecutor(cmd *Cmd) func(s string) {
	return func(s string) {
		name, options := ParseCmd(s)
		cmd.Name = name
		cmd.Options = options
		if name == "exit" {
			os.Exit(0)
		}
		cmd.Run()
	}
}
