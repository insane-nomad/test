package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

const sendData = "{command:info}"
const PORT = "4028"

func getAddr(from, to string) []string {
	start := net.ParseIP(from).To4()
	stop := net.ParseIP(to).To4()
	result := make([]string, 0)
	for {
		result = append(result, start.String())
		if start[0] == stop[0] && start[1] == stop[1] && start[2] == stop[2] && start[3] == stop[3] {
			break
		}
		prev0, prev1, prev2, prev3 := start[0], start[1], start[2], start[3]
		start[3]++

		if start[3] == 0 && prev3 == 255 {
			start[2]++
		}
		if start[2] == 0 && prev2 == 255 {
			start[1]++
		}
		if start[1] == 0 && prev1 == 255 {
			start[0]++
		}
		if start[0] == 0 && prev0 == 255 {
			break
		}
	}
	return result
}

func getData() []string {
	contents, err := os.ReadFile("config.txt")
	res := make([]string, 0)
	if err != nil {
		fmt.Println("File reading error", err)
		return nil
	}
	addr := strings.Split(string(contents), ";")

	for _, val := range addr {
		if val != "" {
			val := strings.TrimSpace(val)
			if strings.Contains(val, "-") {
				tmp := strings.Split(val, "-")
				res = append(res, getAddr(tmp[0], tmp[1])...)
			}
			if strings.Contains(val, "x") {
				tmp1 := strings.ReplaceAll(val, "x", "0")
				tmp2 := strings.ReplaceAll(val, "x", "255")
				res = append(res, getAddr(tmp1, tmp2)...)
			}
		}
	}
	return res
}

func main() {
	f, err := os.Create("result.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	table := tablewriter.NewWriter(f)
	table.SetHeader([]string{"IP", "Miner"})

	data := [][]string{}
	for _, addr := range getData() {
		//fmt.Println(addr)
		tcpServer, err := net.ResolveTCPAddr("tcp", addr+":"+PORT)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		conn, err := net.DialTCP("tcp", nil, tcpServer)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer conn.Close()

		_, err = conn.Write([]byte(sendData))
		if err != nil {
			fmt.Println("Write data failed:", err.Error())
			os.Exit(1)
		}

		received := make([]byte, 1024)
		_, err = conn.Read(received)
		if err != nil {
			println("Read data failed:", err.Error())
			os.Exit(1)
		}

		if strings.Contains(string(received), "data") && strings.Contains(string(received), "type") {
			out := strings.Split(string(received), "type:\"")
			replacer := strings.NewReplacer("\"", "", "}", "")
			miner := replacer.Replace(out[1])
			data = append(data, []string{addr, miner})
		}
	}
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
