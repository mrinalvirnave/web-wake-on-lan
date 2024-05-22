package main

import (
	"fmt"
	"log"
	"mrinalvirnave/go-web-wol/remoteshut"
	"mrinalvirnave/go-web-wol/wol"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// albums slice to seed record album data.
var wol_macs = strings.Split(os.Getenv("WOL_MACS"), ",")
var shutdown_hosts = strings.Split(os.Getenv("SHUTDOWN_HOSTS"), ",")

func main() {

	router := gin.Default()
	router.POST("/wake", postWake)
	router.GET("/shut", remoteShut)

	router.Run(":5868")
}

// postWake Wakes the machine from JSON received in the request body.
func postWake(c *gin.Context) {

	// Get the array of strings from the context JSON body and assign it to requestMacs.
	var requestMacs []string
	if c.Request.ContentLength <= 0 {
		// Copy the macs slice into the requestMacs slice.
		requestMacs = make([]string, len(wol_macs))
		copy(requestMacs, wol_macs)
	} else {
		if err := c.BindJSON(&requestMacs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Parse through the requestMacs slice and call the wol.Wake function on each mac address.
	for _, mac := range requestMacs {
		if err := wakeMac(mac); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusAccepted, gin.H{"macs": requestMacs})
}

func remoteShut(c *gin.Context) {

	// Loop through the hostnames in the shutdown_hosts slice and ssh to port 22 as the user shut
	// and run the command "sudo shutdown -h now".
	var output []string
	for _, host := range shutdown_hosts {
		conn, err := remoteshut.Connect(host+":22", "shut")
		if err != nil {
			log.Default().Println(err)
			output = append(output, host+" was already off")
		} else {
			_, err = conn.SendCommands("sudo shutdown -h now")
			if err != nil {
				output = append(output, host+" has been Shutdown")
			}
		}
	}

	c.JSON(http.StatusAccepted, gin.H{"output": output})
}

func wakeMac(mac string) error {
	// bcastInterface can be "eth0", "eth1", etc.. An empty string implies
	// that we use the default interface when sending the UDP packet (nil).

	macAddr := mac
	var localAddr *net.UDPAddr
	bcastAddr := "192.168.1.255:7"

	udpAddr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		return err
	}

	// Build the magic packet.
	mp, err := wol.New(macAddr)
	if err != nil {
		return err
	}

	// Grab a stream of bytes to send.
	bs, err := mp.Marshal()
	if err != nil {
		return err
	}

	// Grab a UDP connection to send our packet of bytes.
	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Printf("Attempting to send a magic packet to MAC %s\n", macAddr)
	fmt.Printf("... Broadcasting to: %s\n", bcastAddr)
	n, err := conn.Write(bs)
	if err == nil && n != 102 {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Magic packet sent successfully to %s\n", macAddr)
	return nil
}
