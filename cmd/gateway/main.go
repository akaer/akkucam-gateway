package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"akkucam-gateway/internal/gateway"
	streammanager "akkucam-gateway/internal/stream"
	"akkucam-gateway/internal/version"
)

func main() {
	listen := env("AKKUCAM_LISTEN", gateway.DefaultListen)
	webRoot := env("AKKUCAM_WEB_ROOT", gateway.DefaultWebRoot)
	streams := streammanager.New(gateway.DefaultHLSRoot)
	server := gateway.New(webRoot, streams)

	httpServer := &http.Server{
		Addr:              listen,
		Handler:           gateway.RequestLogger(server.Routes()),
		ReadHeaderTimeout: gateway.ReadHeaderTimeout,
	}

	log.Printf("AkkuCam gateway %s listening on %s", version.Display(), listen)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix(fmt.Sprintf("[%d] ", os.Getpid()))
}
