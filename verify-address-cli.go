package main

// https://www.webdigi.co.uk/blog/2009/how-to-check-if-an-email-address-exists-without-sending-an-email/
// TODO: Library?
// Cache DNS lookup and mail sever results
// Throttle per server? Use all MX servers?

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/oyvindsk/go-email-address-check/verify"
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

	addrOK, smtpMsg, err := verify.VerifyAddress(emailAddress)

	if err != nil {
		log.Fatal(err)
	}

	if addrOK {
		log.Printf("OK, Address %q seems valid (smtp msg: %q)", emailAddress, smtpMsg)
	} else {
		log.Printf("InvalidOK, Address %q seems invalid (smtp msg: %q)", emailAddress, smtpMsg)
	}

}
