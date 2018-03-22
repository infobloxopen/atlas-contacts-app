package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/Infoblox-CTO/ngp.api.toolkit/gw"
)

var (
	Addr         string
	ContactsAddr string
	SwaggerDir   string
)

func main() {
	mux := http.NewServeMux()

	errHandler := runtime.WithProtoErrorHandler(gw.ProtoMessageErrorHandler)

	contactsHandler, err := NewContactsHandler(context.Background(), ContactsAddr, errHandler)
	if err != nil {
		log.Fatalln(err)
	}
	mux.Handle("/contacts/v1/", http.StripPrefix("/contacts/v1", contactsHandler))

	mux.HandleFunc("/swagger/", SwaggerHandler)

	http.ListenAndServe(Addr, mux)
}

func init() {
	flag.StringVar(&Addr, "listen", "0.0.0.0:8080", "")
	flag.StringVar(&ContactsAddr, "contacts", "127.0.0.1:9091", "")
	flag.StringVar(&SwaggerDir, "swagger-dir", "share", "")
	flag.Parse()
}
