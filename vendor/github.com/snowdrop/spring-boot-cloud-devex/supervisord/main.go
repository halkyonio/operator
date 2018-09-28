package main

// CMDS="echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble" go run main.go

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

const (
	templateFile = "conf/supervisord.tmpl"
	outFile      = "conf/supervisor.conf"
)

type Program struct {
	Name    string
	Command string
}

func main() {
	// Read Supervisord.tmpl file
	log.Println("Read Supervisord.tmpl file")
	currentDir, err := os.Getwd()
	f, err := os.Open(currentDir + "/" + templateFile)
	if err != nil {
		log.Fatal("Can't read template file", err.Error())
	}
	// Close file on exit
	defer f.Close()

	// Read file content
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Recuperate ENV vars and split them / command
	log.Println("Recuperate ENV vars and split them / command")
	m := make(map[string][]Program)
	if cmdsEnv := os.Getenv("CMDS"); cmdsEnv != "" {
		cmds := strings.Split(cmdsEnv, ";")
		for i := range cmds {
			cmd := strings.Split(cmds[i], ":")
			log.Println("Command : ", cmd)
			p := Program{cmd[0], cmd[1]}
			m["cmd-"+string(i)] = append(m["cmd-"+string(i)], p)
		}
	} else {
		log.Fatal("No commands provided !")
	}

	// Create a template to parse supervisord file
	log.Println("Create a template to parse supervisord file")
	t := template.New("Supervisord template")
	t, _ = t.Parse(string(data)) //parse it

	// Open a new file to save the result
	outFile, err := os.OpenFile(
		outFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	// Write template result to the supervisord.conf
	log.Println("Parse template file and generate result")
	error := t.Execute(outFile, m)
	if error != nil {
		log.Fatal(error)
	}
}
