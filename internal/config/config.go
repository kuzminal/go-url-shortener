package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

var (
	RunPort     = ":8080"
	BaseURL     = "http://localhost:8080/"
	PersistFile = ""
	AuthSecret  = []byte("ololo-trololo-shimba-boomba-look")
	DatabaseDSN = ""
	UseTLS      = false
	CertFile    = ""
	KeyFile     = ""
)

func Parse() {
	flag.StringVar(&RunPort, "a", RunPort, "port to run server")
	flag.StringVar(&BaseURL, "b", BaseURL, "base URL for shorten URL response")
	flag.StringVar(&PersistFile, "f", PersistFile, "file to store shorten URLs")
	flag.StringVar(&DatabaseDSN, "d", DatabaseDSN, "connection string to database")
	flag.BoolVar(&UseTLS, "s", false, "use TLS for server")
	flag.StringVar(&CertFile, "certfile", "cert.pem", "certificate PEM file")
	flag.StringVar(&KeyFile, "keyfile", "key.pem", "key PEM file")

	flag.Parse()

	if val := os.Getenv("SERVER_ADDRESS"); val != "" {
		RunPort = val
	}
	if val := os.Getenv("BASE_URL"); val != "" {
		BaseURL = val
	}
	if val := os.Getenv("FILE_STORAGE_PATH"); val != "" {
		PersistFile = val
	}
	if val := os.Getenv("DATABASE_DSN"); val != "" {
		DatabaseDSN = val
	}
	if val := os.Getenv("ENABLE_HTTPS"); val != "" {
		parsedVal, err := strconv.ParseBool(val)
		if err == nil {
			UseTLS = parsedVal
		}
	}
	if val := os.Getenv("CERT_FILE"); val != "" {
		CertFile = val
	}
	if val := os.Getenv("KEY_FILE"); val != "" {
		KeyFile = val
	}

	BaseURL = strings.TrimRight(BaseURL, "/")
}
