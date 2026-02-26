package main


import (
    "context"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/tls"
    "crypto/x509"
    "encoding/hex"
    "encoding/json"
    "encoding/pem"
    "errors"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "strconv"
    "strings"
    "sync/atomic"
    "syscall"
    "time"
)

const (
    version   = "0.9.2"
    copyright = "Copyright 2026 Nash!Com/Daniel Nashed. All rights reserved."

    defaultListenAddr  = ":8443"
    defaultServername  = "cube-control.domino.svc.cluster.local"
    defaultTokenFile   = "/var/run/secrets/cube-control/token"
    defaultCfgCheckIntervalSec = 120
    maxBodySize        = 5 << 20 // 5 MB

    env_CubeControl_ListenAddr         = "CUBE_CONTROL_LISTEN_ADDR"
    env_CubeControl_ServerName         = "CUBE_CONTROL_SERVER_NAME"
    env_CubeControl_CertMgrServer      = "CUBE_CONTROL_CERTMGR_SERVER"
    env_CubeControl_Token              = "CUBE_CONTROL_TOKEN"
    env_CubeControl_TokenFile          = "CUBE_CONTROL_TOKEN_FILE"
    env_CubeControl_CfgCheckInterval   = "CUBE_CONTROL_CFG_CHECK_INTERVAL"
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

var globalApiToken atomic.Value


func showHelpEnv() {

    log.Printf("");
    log.Printf("Environment variables\n");
    log.Printf("---------------------\n");
    log.Printf("");
    log.Printf("%-35s   TLS listen address (default: %s)\n", env_CubeControl_ListenAddr, defaultListenAddr);
    log.Printf("%-35s   Server name (default: %s)\n", env_CubeControl_ServerName, defaultServername);
    log.Printf("%-35s   CertMgr to connect to when checking for certificate updates\n", env_CubeControl_CertMgrServer);
    log.Printf("%-35s   File name to read the authentication token from (default: %s)\n", env_CubeControl_TokenFile, defaultTokenFile);
    log.Printf("%-35s   Authentication token stored in environment variable (takes overrides secrets file)\n", env_CubeControl_Token);
    log.Printf("%-35s   Certificate and token update check interval (default: %ds)\n", env_CubeControl_CfgCheckInterval, defaultCfgCheckIntervalSec);
    log.Printf("");

}


func main() {

    log.SetFlags(0)

    var showVersion = flag.Bool("version", false, "show version")
    var showEnv     = flag.Bool("env", false, "show environment variable help")

    flag.Parse()

    if *showVersion {
        log.Println(version)
        return
    }

    if *showEnv {
        showHelpEnv()
        return
    }

    listenAddr  := getenv(env_CubeControl_ListenAddr, defaultListenAddr)
    serverName  := getenv(env_CubeControl_ServerName, defaultServername)
    certMgrName := getenv(env_CubeControl_CertMgrServer, "")

    globalApiToken.Store(loadToken())

    // Fix TLS certificate and key locations
    const (
        tlsCertFile = "/tls/tls.crt"
        tlsKeyFile  = "/tls/tls.key"
    )

    cfgCheckInterval := getDurationEnv(env_CubeControl_CfgCheckInterval, defaultCfgCheckIntervalSec * time.Second)

    log.Println("")
    log.Println("--------------------------------------------------------------------------------")
    log.Printf("Cube Control %s", version)
    log.Println(copyright)
    log.Println("--------------------------------------------------------------------------------")
    log.Println("")

    log.Printf("Waiting for TLS key: %s", tlsKeyFile)

    for {
        if fileExists(tlsKeyFile) {
            break
        }
        time.Sleep(2 * time.Second)
    }

    time.Sleep(2 * time.Second)

    keyPEM, err := os.ReadFile(tlsKeyFile)
    if err != nil {
        panic(fmt.Errorf("Failed to read private key: %w", err))
    }

    log.Printf("\nChecking for TLS certificate in /tls first")

    chain, err := loadCertificateFromDisk(tlsCertFile)
    if err != nil {
        if certMgrName == "" {
            panic(fmt.Errorf("No certificate found and no CertMgr server configured to fetch certificate. Configure: %s", env_CubeControl_CertMgrServer))
        }

        log.Printf("\nNo local certificate file provided. Downloading certificate from CertMgr server\n\n")
        log.Printf("%s \n", certMgrName)
        log.Printf("Using SNI     : %s\n", serverName)

        // Try to download certs and wait for 30 to try again if not successful
        for {
            chain, err = fetchChain(certMgrName, "443", serverName, 30*time.Second)
            if err == nil {
                log.Printf("\nReceived %d certificates\n\n", len(chain))
                break
            }
            time.Sleep(30 * time.Second)
        }
    }

    dumpCertificateChain(chain)

    // Build + validate initial cert
    tlsCertPtr, fp, err := buildAndValidateTLSCert(chain, keyPEM, serverName)
    if err != nil {
        panic(err)
    }

    // Atomic storage for live reload
    var currentCert atomic.Value // stores *tls.Certificate
    var currentFP atomic.Value   // stores string
    currentCert.Store(tlsCertPtr)
    currentFP.Store(fp)

    // Web server configuration
    mux := http.NewServeMux()
    mux.HandleFunc("/apply", applyHandler())

    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
        GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
            v := currentCert.Load()
            if v == nil {
                return nil, fmt.Errorf("No certificate loaded")
            }
            return v.(*tls.Certificate), nil
        },
    }

    server := &http.Server{
        Addr:      listenAddr,
        Handler:   mux,
        TLSConfig: tlsConfig,
        ReadHeaderTimeout:  10 * time.Second,
        ReadTimeout:        30 * time.Second,
        WriteTimeout:       60 * time.Second,
        IdleTimeout:       120 * time.Second,
    }

    // Signal-aware context (also used to stop reload loop)
    runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
    defer stop()

    log.Printf("\nStarting Cube Control on %s (TLS)\n\n", listenAddr)
    fmt.Printf("Config check interval: %v\n\n", cfgCheckInterval)

    go func() {
        if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Listen error: %v", err)
        }
    }()

    // Cert reload loop
    go func() {
        ticker := time.NewTicker(cfgCheckInterval)
        defer ticker.Stop()

        for {
            select {
            case <-runCtx.Done():
                return
            case <-ticker.C:
            }

            var newChain []*x509.Certificate
            var loadErr error
            var newApiToken string
            var apiToken    string

            // Check for API token reload

            newApiToken = loadToken()
            apiToken    = globalApiToken.Load().(string)

            if (newApiToken != apiToken) {

                globalApiToken.Store(newApiToken)
                apiToken = newApiToken

                log.Printf("\nInfo: API Token updated\n")
            }

            // Prefer local cert file if present

            if fileExists(tlsCertFile) {
                newChain, loadErr = loadCertificateFromDisk(tlsCertFile)
                if loadErr != nil {
                    log.Printf("\nCert reload: Failed reading local cert %s: %v\n", tlsCertFile, loadErr)
                    continue
                }
            } else {
                if certMgrName == "" {
                    // nothing to do
                    continue
                }

                newChain, loadErr = fetchChain(certMgrName, "443", serverName, 30*time.Second)
                if loadErr != nil {
                    log.Printf("\nCert reload: Remote fetch failed (%s): %v\n", certMgrName, loadErr)
                    continue
                }
            }

            newFP := chainFingerprint(newChain)
            oldFPv := currentFP.Load()
            oldFP, _ := oldFPv.(string)

            if newFP == oldFP {
                continue
            }

            // Validate + build tls.Certificate (must match key + hostname)
            newCertPtr, validatedFP, vErr := buildAndValidateTLSCert(newChain, keyPEM, serverName)
            if vErr != nil {
                log.Printf("\nCert reload: new cert invalid (not applying): %v\n", vErr)
                continue
            }

            currentCert.Store(newCertPtr)
            currentFP.Store(validatedFP)

            log.Printf("\nCert reload: Updated active certificate (%d)\n", len(newChain))
            dumpCertificateChain(newChain)
        }
    }()

    <-runCtx.Done()

    log.Println("\nShutting down Cube Control ...\n")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}

func buildAndValidateTLSCert(chain []*x509.Certificate, keyPEM []byte, serverName string) (*tls.Certificate, string, error) {
    if len(chain) == 0 {
        return nil, "", fmt.Errorf("No certificate chain provided")
    }

    leaf := chain[0]

    if err := leaf.VerifyHostname(serverName); err != nil {
        return nil, "", fmt.Errorf("Hostname verification failed: %w", err)
    }

    tlsCert, err := buildTLSCertificate(chain, keyPEM)
    if err != nil {
        return nil, "", fmt.Errorf("Certificate/key mismatch: %w", err)
    }

    if len(tlsCert.Certificate) == 0 || tlsCert.PrivateKey == nil {
        return nil, "", fmt.Errorf("Invalid tls.Certificate (empty cert or missing key)")
    }

    // Populate Leaf for easier introspection/debug later
    if tlsCert.Leaf == nil {
        if cert, err := x509.ParseCertificate(tlsCert.Certificate[0]); err == nil {
            tlsCert.Leaf = cert
        }
    }

    fp := chainFingerprint(chain)
    return &tlsCert, fp, nil
}

func chainFingerprint(chain []*x509.Certificate) string {
    h := sha256.New()
    for _, c := range chain {
        h.Write(c.Raw)
    }
    return fmt.Sprintf("%x", h.Sum(nil))
}

func applyHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }

        expectedToken := globalApiToken.Load().(string)

        // Token validation (optional)
        if expectedToken != "" {
            if r.Header.Get("Authorization") != "Bearer " + expectedToken {
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
            resp.Error = err.Error()
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
    token := os.Getenv(env_CubeControl_Token)
    if token != "" {
        return token
    }

    tokenFile := os.Getenv(env_CubeControl_TokenFile)
    if tokenFile == "" {
        tokenFile = defaultTokenFile
    }

    data, err := os.ReadFile(tokenFile)
    if err != nil {
        // Not fatal — only log if explicitly set
        if os.Getenv(env_CubeControl_TokenFile) != "" {
            log.Printf("Failed to read token file %s: %v", tokenFile, err)
        }
        return ""
    }

    return strings.TrimSpace(string(data))
}


func getIntEnv(name string, defaultValue int) int {
    valStr := os.Getenv(name)
    if valStr == "" {
        return defaultValue
    }

    val, err := strconv.Atoi(valStr)
    if err != nil {
        log.Printf("Invalid value for %s: %s\n", name, valStr)
        return defaultValue
    }

    return val
}

func getDurationEnv(name string, defaultValue time.Duration) time.Duration {
    valStr := os.Getenv(name)
    if valStr == "" {
        return defaultValue
    }

    d, err := time.ParseDuration(valStr)
    if err != nil {
        return defaultValue
    }

    return d
}


func fetchChain(host string, port string, serverName string, timeout time.Duration) ([]*x509.Certificate, error) {
    dialer := &net.Dialer{
        Timeout: timeout,
    }

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    conf := &tls.Config{
        ServerName:         serverName, // overridden SNI
        InsecureSkipVerify: true,       // manual verification later
    }

    conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), conf)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    state := conn.ConnectionState()

    if len(state.PeerCertificates) == 0 {
        return nil, errors.New("no certificates received")
    }

    return state.PeerCertificates, nil
}

func buildTLSCertificate(chain []*x509.Certificate, keyPEM []byte) (tls.Certificate, error) {
    var certPEM []byte

    for _, cert := range chain {
        block := &pem.Block{
            Type:  "CERTIFICATE",
            Bytes: cert.Raw,
        }
        certPEM = append(certPEM, pem.EncodeToMemory(block)...)
    }

    return tls.X509KeyPair(certPEM, keyPEM)
}

func loadCertificateFromDisk(certPath string) ([]*x509.Certificate, error) {
    pemData, err := os.ReadFile(certPath)
    if err != nil {
        return nil, err
    }

    var certs []*x509.Certificate
    for {
        block, rest := pem.Decode(pemData)
        if block == nil {
            break
        }

        if block.Type == "CERTIFICATE" {
            cert, err := x509.ParseCertificate(block.Bytes)
            if err != nil {
                return nil, err
            }
            certs = append(certs, cert)
        }

        pemData = rest
    }

    if len(certs) == 0 {
        return nil, fmt.Errorf("No certificates found in %s", certPath)
    }

    return certs, nil
}


func dumpCertificateChain(chain []*x509.Certificate) {
    log.Printf("")
    log.Printf("---------------------------------------------------")
    log.Printf("Certificates: %d", len(chain))
    log.Printf("---------------------------------------------------")
    log.Printf("")

    for i, cert := range chain {
        log.Printf("----- Certificate %d -----\n\n", i)

        if i == 0 {
            fmt.Printf("%-15s : %s\n", "Type", "Leaf")
        } else {
            fmt.Printf("%-15s : %s\n", "Type", "Intermediate/CA")
        }

        fmt.Printf("%-15s : %s\n", "Subject", cert.Subject.String())
        fmt.Printf("%-15s : %s\n", "Issuer", cert.Issuer.String())

        if len(cert.DNSNames) > 0 {
            fmt.Printf("%-15s : %s\n", "DNS SANs", strings.Join(cert.DNSNames, " "))
        }

        if len(cert.IPAddresses) > 0 {
            ips := make([]string, len(cert.IPAddresses))
            for i, ip := range cert.IPAddresses {
                ips[i] = ip.String()
            }

            fmt.Printf("%-15s : %s\n", "IP SANs", strings.Join(ips, " "))
        }

        if len(cert.EmailAddresses) > 0 {
            fmt.Printf("%-15s : %s\n", "Email SANs", strings.Join(cert.EmailAddresses, " "))
        }

        if len(cert.URIs) > 0 {
            uris := make([]string, len(cert.URIs))
            for i, uri := range cert.URIs {
                uris[i] = uri.String()
            }

            fmt.Printf("%-15s : %s\n", "URI SANs", strings.Join(uris, " "))
        }

        fmt.Printf("%-15s : %s\n", "Serial", cert.SerialNumber.String())
        fmt.Printf("%-15s : %s\n", "SHA256 FP", formatFingerprintSHA256(cert))
        fmt.Printf("%-15s : %s\n", "SHA1 FP", formatFingerprintSHA1(cert))

        if len(cert.SubjectKeyId) > 0 {
            fmt.Printf("%-15s : %s\n", "SKI", formatHexWithColon(cert.SubjectKeyId))
        }

        if len(cert.AuthorityKeyId) > 0 {
            fmt.Printf("%-15s : %s\n", "AKI", formatHexWithColon(cert.AuthorityKeyId))
        }

        fmt.Printf("%-15s : %s\n", "NotBefore", cert.NotBefore.Format(time.RFC3339))
        fmt.Printf("%-15s : %s\n", "NotAfter", cert.NotAfter.Format(time.RFC3339))

        log.Printf("")
    }
}


func formatFingerprintSHA256(cert *x509.Certificate) string {
    sum := sha256.Sum256(cert.Raw)
    return formatHexWithColon(sum[:])
}

func formatFingerprintSHA1(cert *x509.Certificate) string {
    sum := sha1.Sum(cert.Raw)
    return formatHexWithColon(sum[:])
}

func formatHexWithColon(data []byte) string {
    hexStr := hex.EncodeToString(data)
    var result string

    for i := 0; i < len(hexStr); i += 2 {
        if i > 0 {
            result += ":"
        }
        result += hexStr[i : i+2]
    }

    return result
}
