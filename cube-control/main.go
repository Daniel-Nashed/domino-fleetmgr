package main

import (
    "context"
    "crypto/tls"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "strings"
    "syscall"
    "time"
)


const (
    version   = "0.9.0"
    copyright = "Copyright 2026 Nash!Com/Daniel Nashed. All rights reserved."

    defaultListenAddr = ":8443"
    maxBodySize       = 5 << 20 // 5 MB
)


type Status string

const (
    StatusSuccess Status = "success"
    StatusError   Status = "error"
)

type ApplyResponse struct {
    Timestamp string `json:"timestamp"`
    Status    Status `json:"status"`
    Output    string `json:"output,omitempty"`
    Error     string `json:"error,omitempty"`
}



func main() {
    listenAddr := getenv("CUBE_CONTROL_LISTEN_ADDR", defaultListenAddr)
    apiToken := loadToken()

    const (
        tlsCertFile = "/tls/tls.crt"
        tlsKeyFile  = "/tls/tls.key"
    )

    log.SetFlags(0)

    if len(os.Args) > 1 && os.Args[1] == "--version" {
        log.Println(version)
        return
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/apply", applyHandler(apiToken))

    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
        MaxVersion: tls.VersionTLS13,
    }

    server := &http.Server{
        Addr:      listenAddr,
        Handler:   mux,
        TLSConfig: tlsConfig,
    }

    log.Println("")
    log.Printf("Cube Control %s", version)
    log.Println(copyright)
    log.Println("")

    log.Printf("Waiting for TLS certificate and key in /tls ...")

    for {
        if fileExists(tlsCertFile) && fileExists(tlsKeyFile) {
            break
        }
        time.Sleep(2 * time.Second)
    }

    log.Printf("TLS certificate and key found. Starting Cube Control on %s (TLS)", listenAddr)

    go func() {
        if err := server.ListenAndServeTLS(tlsCertFile, tlsKeyFile); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Listen error: %v", err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
    <-stop

    log.Println("Shutting down Cube Control ...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}


func applyHandler(expectedToken string) http.HandlerFunc {

    return func(w http.ResponseWriter, r *http.Request) {

        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }

        // Token validation (optional)
        if expectedToken != "" {
            if r.Header.Get("Authorization") != "Bearer "+expectedToken {
                w.WriteHeader(http.StatusForbidden)
                return
            }
        }

        // Limit request body size
        r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        start := time.Now()

        cmd := exec.Command("kubectl", "apply", "-f", "-")

        stdin, err := cmd.StdinPipe()
        if err != nil {
            http.Error(w, "Internal error", http.StatusInternalServerError)
            return
        }

        go func() {
            defer stdin.Close()
            stdin.Write(body)
        }()

        output, err := cmd.CombinedOutput()

        resp := ApplyResponse{
            Timestamp: time.Now().Format(time.RFC3339),
        }

        resp.Output = strings.TrimSpace(string(output))

        if err != nil {

            resp.Status = "error"
            resp.Error  = err.Error()
            w.WriteHeader(http.StatusInternalServerError)
        } else {

            resp.Status = "success"
            w.WriteHeader(http.StatusOK)
        }

        log.Printf(
            "apply request status=%s duration=%s",
            resp.Status,
            time.Since(start),
        )

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)

    }
}


func fileExists(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return !info.IsDir()
}


func getenv(key, fallback string) string {
    val := os.Getenv(key)
    if val == "" {
        return fallback
    }
    return val
}


func loadToken() string {
    token := os.Getenv("CUBE_CONTROL_TOKEN")
    if token != "" {
        return token
    }

    tokenFile := os.Getenv("CUBE_CONTROL_TOKEN_FILE")
    if tokenFile == "" {
        tokenFile = "/var/run/secrets/cube-control/token"
    }

    data, err := os.ReadFile(tokenFile)
    if err != nil {
        // Not fatal — only log if explicitly set
        if os.Getenv("CUBE_CONTROL_TOKEN_FILE") != "" {
            log.Printf("Failed to read token file %s: %v", tokenFile, err)
        }
        return ""
    }

    return strings.TrimSpace(string(data))
}
