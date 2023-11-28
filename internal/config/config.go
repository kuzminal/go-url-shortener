package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

// Параметры запуска приложения
var (
	RunPort         = ":8080"                                    // RunPort порт для запуска приложения
	BaseURL         = "http://localhost" + RunPort               // BaseURL базовый URL для приложения
	PersistFile     = ""                                         // PersistFile файл для хранилища
	AuthSecret      = []byte("ololo-trololo-shimba-boomba-look") // AuthSecret сикрет для авторизации пользователя
	DatabaseDSN     = ""                                         // DatabaseDSN строка подключения к БД
	UseTLS          = false                                      // UseTLS флаг использования TLS
	CertFile        = ""                                         // CertFile пусть к файлу с сертификатом
	KeyFile         = ""                                         // KeyFile путь к файлу с приватным ключом
	ConfigFile      = ""                                         // ConfigFile путь к файлу с конфигурацией приложения
	ShutdownTimeout = 10 * time.Second                           // ShutdownTimeout время ожидания для graceful shutdown
	TrustedSubnet   = ""                                         // TrustedSubnet маска подсети
)

// AppConfig структура для конфигурации приложения
type AppConfig struct {
	RunPort         string `json:"run_port"`         // RunPort порт для запуска приложения
	BaseURL         string `json:"base_url"`         // BaseURL базовый URL для приложения
	PersistFile     string `json:"persist_file"`     // PersistFile файл для хранилища
	DatabaseDSN     string `json:"database_dsn"`     // DatabaseDSN строка подключения к БД
	UseTLS          bool   `json:"use_tls"`          // UseTLS флаг использования TLS
	CertFile        string `json:"cert_file"`        // CertFile пусть к файлу с сертификатом
	KeyFile         string `json:"key_file"`         // KeyFile путь к файлу с приватным ключом
	ShutdownTimeout int    `json:"shutdown_timeout"` // ShutdownTimeout время ожидания для graceful shutdown
	TrustedSubnet   string `json:"trusted_subnet"`   // TrustedSubnet маска подсети
}

// Parse разбарает папаметры запуска приложения
func Parse() {
	flag.StringVar(&RunPort, "a", RunPort, "port to run server")
	flag.StringVar(&BaseURL, "b", BaseURL, "base URL for shorten URL response")
	flag.StringVar(&PersistFile, "f", PersistFile, "file to store shorten URLs")
	flag.StringVar(&DatabaseDSN, "d", DatabaseDSN, "connection string to database")
	flag.BoolVar(&UseTLS, "s", UseTLS, "use TLS for server")
	flag.StringVar(&CertFile, "certfile", "cert.pem", "certificate PEM file")
	flag.StringVar(&KeyFile, "keyfile", "key.pem", "key PEM file")
	flag.StringVar(&ConfigFile, "config", ConfigFile, "path to config file")
	flag.DurationVar(&ShutdownTimeout, "gst", ShutdownTimeout, "graceful shutdown timeout")
	flag.StringVar(&TrustedSubnet, "t", TrustedSubnet, "CIDR")

	flag.Parse()
	if ConfigFile != "" {
		err := ParseJSON()
		if err != nil {
			return
		}
	}

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
	if val := os.Getenv("CONFIG"); val != "" {
		ConfigFile = val
	}
	if val := os.Getenv("TIMEOUT"); val != "" {
		timeout, err := strconv.Atoi(val)
		if err == nil {
			ShutdownTimeout = time.Duration(timeout) * time.Second
		}
	}
	if val := os.Getenv("TRUSTED_SUBNET"); val != "" {
		TrustedSubnet = val
	}

	BaseURL = strings.TrimRight(BaseURL, "/")
}

// ParseJSON парсинг файла конфигурации в структуру и присвоение переменным
func ParseJSON() error {
	var cfg AppConfig
	file, err := os.ReadFile(ConfigFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return err
	}
	if RunPort == "" {
		RunPort = cfg.RunPort
	}
	if BaseURL == "" {
		BaseURL = cfg.BaseURL
	}
	if PersistFile == "" {
		PersistFile = cfg.PersistFile
	}
	if DatabaseDSN == "" {
		DatabaseDSN = cfg.DatabaseDSN
	}

	if CertFile == "" {
		CertFile = cfg.CertFile
	}
	if KeyFile == "" {
		KeyFile = cfg.KeyFile
	}
	if TrustedSubnet == "" {
		TrustedSubnet = cfg.TrustedSubnet
	}

	return nil
}
