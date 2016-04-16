package main

// https://www.webdigi.co.uk/blog/2009/how-to-check-if-an-email-address-exists-without-sending-an-email/
// TODO: Library?
// Cache DNS lookup and mail sever results
// Throttle per server? Use all MX servers?

import (
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	var emailAddress string
	if len(os.Args) < 2 {
		fmt.Println("Usage:", filepath.Base(os.Args[0]), "email-address")
		os.Exit(0)
	}
	emailAddress = os.Args[1]
	if emailAddress == "" {
		fmt.Println("Usage:", filepath.Base(os.Args[0]), "email-address")
		os.Exit(0)
	}

	// Parse the address into a user and a host part
	parts := strings.Split(emailAddress, "@")
	if len(parts) != 2 {
		log.Fatal("Can not parse this email address:", emailAddress)
	}
	log.Println(parts)

	// Lookup an MX server for this address
	// Pick one at random? Or??
	servers, err := net.LookupMX(parts[1])
	if err != nil {
		log.Fatal("LookupMX:", err)
	}
	//for _, s := range servers {
	//	fmt.Println(s.Host)
	//}
	server := servers[0].Host
	log.Println("Mail server:", server)

	// Connect to the remote SMTP server.
	c, err := smtp.Dial(fmt.Sprintf("%s:%d", server, 25))
	if err != nil {
		log.Fatal(err)
	}

	// Be polite, say HELLO
	if err = c.Hello("lilleole.odots.org"); err != nil {
		log.Fatal("Hello:", err)
	}

	// Set the sender and recipient first
	if err := c.Mail("os@odots.org"); err != nil {
		log.Fatal(err)
	}
	if err := c.Rcpt(emailAddress); err != nil {
		log.Fatal(err)
	}

	fmt.Println("OK, Address", emailAddress, "seems valid!")

}
