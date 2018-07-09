package core

import (
	"io/ioutil"
	"os"
	"strings"
)

func File2history(filename string) (history []string) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			os.Create(filename)
		}
		history = []string{}
		return
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		history = []string{}
		return
	}
	history = strings.Split(string(content), "\n")
	length := len(history)
	if history[length-1] == "" {
		history = history[:length-1]
		length -= 1
	}
	if length > 1000 {
		history = history[length-1000 : length]
		fd,_ := os.Create(filename)
		defer fd.Close()
		var newOutput string
		for _, elem := range history {
			newOutput += elem + "\n"
		}
		fd.WriteString(newOutput)
	}
	newHistory := make([]string, 0, 1000)
	lastElem := history[0]
	for i:=1; i<len(history);i++  {
		if lastElem != history[i] {
			newHistory = append(newHistory, history[i])
			lastElem = history[i]
		}
	}
	history = newHistory
	return
}

