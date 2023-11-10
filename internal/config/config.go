package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

// Параметры запуска приложения
var (
	RunPort     = ":8080"                                    // RunPort порт для запуска приложения
	BaseURL     = "http://localhost:8080/"                   // BaseURL базовый URL для приложения
	PersistFile = ""                                         // PersistFile файл для хранилища
	AuthSecret  = []byte("ololo-trololo-shimba-boomba-look") // AuthSecret сикрет для авторизации пользователя
	DatabaseDSN = ""                                         // DatabaseDSN строка подключения к БД
	UseTLS      = false                                      // UseTLS флаг использования TLS
	CertFile    = ""                                         // CertFile пусть к файлу с сертификатом
	KeyFile     = ""                                         // KeyFile путь к файлу с приватным ключом
)

// Parse разбарает папаметры запуска приложения
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
