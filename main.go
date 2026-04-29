package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	core "github.com/virtual-db/core"
	mysqldriver "github.com/virtual-db/mysql-driver"
)

func main() {
	app := core.New(core.Config{
		PluginDir: env("VDB_PLUGIN_DIR", "plugins"),
	})

	cfg := mysqldriver.Config{
		Addr:           env("VDB_LISTEN_ADDR", ":3306"),
		DBName:         env("VDB_DB_NAME", "appdb"),
		SourceDSN:      env("VDB_SOURCE_DSN", ""),
		AuthSourceAddr: env("VDB_AUTH_SOURCE_ADDR", ""),
	}

	if certFile := os.Getenv("VDB_TLS_CERT_FILE"); certFile != "" {
		keyFile := os.Getenv("VDB_TLS_KEY_FILE")
		if keyFile == "" {
			log.Fatal("vdb-mysql: VDB_TLS_CERT_FILE is set but VDB_TLS_KEY_FILE is not")
		}
		tlsCfg, err := loadTLS(certFile, keyFile)
		if err != nil {
			log.Fatalf("vdb-mysql: load TLS cert: %v", err)
		}
		cfg.TLSConfig = tlsCfg
	}

	driver := mysqldriver.NewDriver(cfg, app.DriverAPI())
	app.UseDriver(driver)

	if err := app.Run(); err != nil {
		log.Fatalf("vdb-mysql: %v", err)
	}
}

// loadTLS loads a PEM-encoded certificate and private key and returns a
// *tls.Config suitable for use as a server TLS configuration. The minimum
// TLS version is set to 1.2 to match the MySQL wire protocol convention.
func loadTLS(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		// SystemCertPool fails on some minimal container images; start empty.
		pool = x509.NewCertPool()
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
