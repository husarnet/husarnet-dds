package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
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

func GetHostIPv6(hostname string) string {
	// Read the hosts file
	hosts, _ := ioutil.ReadFile("/etc/hosts")
	// Iterate over the lines of the hosts file
	hostLines := strings.Split(string(hosts), "\n")
	for _, hostLine := range hostLines {
		// Check if the line ends with " managed by Husarnet"
		match, _ := regexp.MatchString(hostname+" # managed by Husarnet", hostLine)
		if match {
			// Extract the IP address
			fields := strings.Fields(hostLine)
			ip := fields[0]
			return ip
		}
	}

	fmt.Println("Err: no such host")
	os.Exit(1)
	return "error"
}

func GetOwnHusarnetIPv6() string {
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {

		if i.Name == "hnet0" {
			addrs, _ := i.Addrs()
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if strings.HasPrefix(ip.String(), "fc94") {
					return ip.String()
				}
			}
		}
	}

	fmt.Println("no hnet0 interface, or Husarnet has not started yet")
	os.Exit(1)
	return "0000:0000:0000:0000:0000:0000:0000:0000"
}

func ParseCycloneDDSSimple(templateXML string) string {

	// Read the hosts file
	hosts, _ := ioutil.ReadFile("/etc/hosts")

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

			// Iterate over the lines of the hosts file
			hostLines := strings.Split(string(hosts), "\n")
			for _, hostLine := range hostLines {
				// Check if the line ends with " managed by Husarnet"
				match, _ := regexp.MatchString(`.* managed by Husarnet$`, hostLine)
				if match {
					// Extract the IP address
					fields := strings.Fields(hostLine)
					ip := fields[0]

					// Append the IP address to the output as a <Peer> tag
					output.WriteString(fmt.Sprintf("\t<Peer address='%s'/>\n", ip))
				}
			}
		}
	}

	prettyXML := xmlfmt.FormatXML(string(output.Bytes()), "", "  ", true)

	return prettyXML
}

func ParseFastDDSSimple(templateXML string) string {

	// Read the hosts file
	hosts, _ := ioutil.ReadFile("/etc/hosts")

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

			// Iterate over the lines of the hosts file
			hostLines := strings.Split(string(hosts), "\n")
			for _, hostLine := range hostLines {
				// Check if the line ends with " managed by Husarnet"
				match, _ := regexp.MatchString(`.* managed by Husarnet$`, hostLine)
				if match {
					// Extract the IP address
					fields := strings.Fields(hostLine)
					ip := fields[0]

					// Append the IP address to the output as a <Peer> tag
					output.WriteString(fmt.Sprintf("\t<locator><udpv6><address>%s</address></udpv6></locator>\n", ip))
				}
			}
		} else if strings.Contains(line, "<defaultUnicastLocatorList>") || strings.Contains(line, "<metatrafficUnicastLocatorList>") {
			// Append the IP address to the output as a <Peer> tag
			output.WriteString(fmt.Sprintf("\t<locator><udpv6><address>%s</address></udpv6></locator>\n", GetOwnHusarnetIPv6()))
		}
	}

	prettyXML := xmlfmt.FormatXML(string(output.Bytes()), "", "  ", true)

	return prettyXML
}
