package main

import (
	"smtp/smtp_server"
)

// 메일 전송 ( Server )
func main() {
	// run smtp server
	go smtp_server.RunServe()
	// sleep forever
	c := make(chan struct{})
	<-c
}
