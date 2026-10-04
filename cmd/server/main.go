package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	localfinance "local-finance"
	"local-finance/internal/api"
	"local-finance/internal/db"
	"local-finance/internal/service"
)

func main() {
	port := flag.Int("port", 8080, "Port to run HTTP server on")
	dbPath := flag.String("db", "", "Path to SQLite database file (default: ~/.localfinance/local_finance.db)")
	openBrowserFlag := flag.Bool("open", true, "Auto-open web app in default browser (default: true; use -open=false to disable)")
	flag.Parse()

	// Default database path in user home directory if not provided
	finalDBPath := *dbPath
	if finalDBPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			finalDBPath = "./local_finance.db"
		} else {
			appDir := filepath.Join(homeDir, ".localfinance")
			_ = os.MkdirAll(appDir, 0755)
			finalDBPath = filepath.Join(appDir, "local_finance.db")
		}
	}

	log.Printf("📂 Database location: %s", finalDBPath)
	database, err := db.NewDB(finalDBPath)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer database.Close()

	svc := service.NewTransactionService(database)
	router := api.SetupRouter(database, svc, localfinance.GetStaticFS())

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	url := fmt.Sprintf("http://%s", addr)

	// Check if port is already in use
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// Fallback to random available port if 8080 is occupied
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("❌ Failed to start listener: %v", err)
		}
		addr = listener.Addr().String()
		url = fmt.Sprintf("http://%s", addr)
	}
	_ = listener.Close()

	log.Printf("🚀 Local Finance server listening on %s", url)

	if *openBrowserFlag {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(url)
		}()
	}

	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	if err != nil {
		log.Printf("⚠️ Could not open default browser automatically: %v", err)
	}
}
