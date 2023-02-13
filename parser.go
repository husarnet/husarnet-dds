package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-xmlfmt/xmlfmt"
)

func HusarnetPresent() bool {
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		if i.Name == "hnet0" {
			return true
		}
	}
	return false
}

type hostTable struct {
	Result struct {
		HostTable     map[string]string `json:"host_table"`
		LocalHostname string            `json:"local_hostname"`
		LocalIPv6     string            `json:"local_ip"`
	} `json:"result"`
}

func HusarnetAPIrequest(endpoint string) []byte {
	// create a new HTTP client
	client := &http.Client{}

	// create a new HTTP request
	req, err := http.NewRequest("GET", "http://localhost:16216/"+endpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	// make the GET request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	return body
}

func GetHostIPv6(hostname string) string {
	body := HusarnetAPIrequest("api/status")

	var data hostTable
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error unmarshaling response: %v\r\n", err)
		os.Exit(1)
		return "error"
	}

	ipv6Address, ok := data.Result.HostTable[hostname]
	if !ok {
		fmt.Printf("Host not found: %s", hostname)
		os.Exit(1)
		return "error"
	}

	return ipv6Address
}

func GetOwnHusarnetIPv6() string {

	body := HusarnetAPIrequest("api/status")

	var data hostTable
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error unmarshaling response: %v\r\n", err)
		os.Exit(1)
		return "error"
	}

	ipv6Address := data.Result.LocalIPv6

	return ipv6Address

}

func ParseCycloneDDSSimple(templateXML string) string {

	body := HusarnetAPIrequest("api/status")

	var data hostTable
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error unmarshaling response: %v\r\n", err)
		os.Exit(1)
		return "error"
	}

	// Initialize an empty buffer to hold the output
	var output bytes.Buffer

	// Split the input file into lines
	inputLines := strings.Split(templateXML, "\n")

	// Iterate over the lines of the input file
	for _, line := range inputLines {
		// Append the line to the output
		output.WriteString(line + "\n")

		// Check if the line contains the <Peers> tag
		if strings.Contains(line, "<Peers>") {
			for _, ipv6Address := range data.Result.HostTable {
				output.WriteString(fmt.Sprintf("\t<Peer address='%s'/>\n", ipv6Address))
			}
		}
	}

	prettyXML := xmlfmt.FormatXML(string(output.Bytes()), "", "  ", true)

	return prettyXML
}

func ParseFastDDSSimple(templateXML string) string {

	body := HusarnetAPIrequest("api/status")

	var data hostTable
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error unmarshaling response: %v\r\n", err)
		os.Exit(1)
		return "error"
	}

	// Initialize an empty buffer to hold the output
	var output bytes.Buffer

	// Split the input file into lines
	inputLines := strings.Split(string(templateXML), "\n")

	// Iterate over the lines of the input file
	for _, line := range inputLines {
		// Append the line to the output
		output.WriteString(line + "\n")

		// Check if the line contains the <Peers> tag
		if strings.Contains(line, "<initialPeersList>") {
			for _, ipv6Address := range data.Result.HostTable {
				output.WriteString(fmt.Sprintf("\t<locator><udpv6><address>%s</address></udpv6></locator>\n", ipv6Address))
			}

		} else if strings.Contains(line, "<defaultUnicastLocatorList>") || strings.Contains(line, "<metatrafficUnicastLocatorList>") {
			// Append the IP address to the output as a <Peer> tag
			output.WriteString(fmt.Sprintf("\t<locator><udpv6><address>%s</address></udpv6></locator>\n", GetOwnHusarnetIPv6()))
		}
	}

	prettyXML := xmlfmt.FormatXML(string(output.Bytes()), "", "  ", true)

	return prettyXML
}
