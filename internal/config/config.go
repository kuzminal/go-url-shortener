package config

import (
	"encoding/json"
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
	ConfigFile  = ""                                         // ConfigFile путь к файлу с конфигурацией приложения
)

// AppConfig структура для конфигурации приложения
type AppConfig struct {
	RunPort     string `json:"run_port"`     // RunPort порт для запуска приложения
	BaseURL     string `json:"base_url"`     // BaseURL базовый URL для приложения
	PersistFile string `json:"persist_file"` // PersistFile файл для хранилища
	DatabaseDSN string `json:"database_dsn"` // DatabaseDSN строка подключения к БД
	UseTLS      bool   `json:"use_tls"`      // UseTLS флаг использования TLS
	CertFile    string `json:"cert_file"`    // CertFile пусть к файлу с сертификатом
	KeyFile     string `json:"key_file"`     // KeyFile путь к файлу с приватным ключом
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
	if val := os.Getenv("CONFIG"); val != "" {
		ConfigFile = val
	}

	BaseURL = strings.TrimRight(BaseURL, "/")
	if ConfigFile != "" {
		err := ParseJSON()
		if err != nil {
			return
		}
	}
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
	RunPort = cfg.RunPort
	BaseURL = cfg.BaseURL
	PersistFile = cfg.PersistFile
	DatabaseDSN = cfg.DatabaseDSN
	UseTLS = cfg.UseTLS
	CertFile = cfg.CertFile
	KeyFile = cfg.KeyFile
	return nil
}
