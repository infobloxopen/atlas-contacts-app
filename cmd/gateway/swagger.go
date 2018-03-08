package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func SwaggerHandler(rw http.ResponseWriter, req *http.Request) {
	p := strings.TrimPrefix(req.URL.Path, "/swagger/")
	p = filepath.Join(SwaggerDir, p)
	p += ".swagger.json"
	log.Printf("serving %s", p)
	http.ServeFile(rw, req, p)
}
