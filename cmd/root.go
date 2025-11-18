// Package cmd defines the command-line interface commands
package cmd

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/ayn2op/tview"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/diamondburned/ningen/v3"
	"github.com/involvex/disgo-cli/internal/config"
	"github.com/involvex/disgo-cli/internal/keyring"
	"github.com/involvex/disgo-cli/internal/logger"
	"github.com/joho/godotenv"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

var (
	discordState *ningen.State
	app          *application
)

func Run() error {
	// Load .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("❌ Warning: Failed to load .env file: %v\n", err)
		} else {
			fmt.Println("✅ Environment file .env loaded successfully")
		}
	} else {
		fmt.Println("⚠️ No .env file found - using system environment variables")
	}

	// Initialize OAuth configuration with environment variables
	initOAuthConfig()

	// Debug OAuth configuration
	clientID := os.Getenv("DISGO_CLI_CLIENT_ID")
	clientSecret := os.Getenv("DISGO_CLI_CLIENT_SECRET")
	fmt.Printf("🔍 OAuth Debug: ClientID='%s' (length: %d), ClientSecret='%s' (length: %d)\n",
		clientID, len(clientID), strings.Repeat("*", len(clientSecret)), len(clientSecret))
	fmt.Printf("🔍 OAuth Config: ClientID='%s', RedirectURL='%s'\n",
		oauthConfig.ClientID, oauthConfig.RedirectURL)

	tokenEnvVar := os.Getenv("DISGO_CLI_TOKEN")
	tokenFlag := flag.String("token", tokenEnvVar, "authentication token")
	serveFlag := flag.Bool("serve", false, "start web server for QR code login")
	debugFlag := flag.Bool("debug", false, "enable debug mode")
	dFlag := flag.Bool("d", false, "enable debug mode (short)")
	portFlag := flag.String("port", "4444", "web server port (default 4444)")
	hostFlag := flag.String("host", "localhost", "web server host (default localhost)")

	configPath := flag.String("config-path", config.DefaultPath(), "path of the configuration file")
	logPath := flag.String("log-path", logger.DefaultPath(), "path of the log file")
	logLevel := flag.String("log-level", "info", "log level")
	flag.Parse()

	var level slog.Level
	if *debugFlag || *dFlag {
		ws.EnableRawEvents = true
		level = slog.LevelDebug
		fmt.Println("🔧 Debug mode enabled")
	} else {
		switch *logLevel {
		case "debug":
			ws.EnableRawEvents = true
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}

	if err := logger.Load(*logPath, level); err != nil {
		return fmt.Errorf("failed to load logger: %w", err)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	token := *tokenFlag
	if token == "" {
		token, err = keyring.GetToken()
		if err != nil {
			slog.Info("failed to retrieve token from keyring", "err", err)
		}
	}

	if *serveFlag {
		return startWebServer(token, *portFlag, *hostFlag)
	}

	tview.Styles = tview.Theme{}
	app = newApplication(cfg)
	return app.run(token)
}

func startWebServer(existingToken string, port string, host string) error {
	if existingToken != "" {
		fmt.Println("Already logged in. Token found in keyring.")
		return nil
	}

	// Update OAuth config with the actual host and port
	clientID := os.Getenv("DISGO_CLI_CLIENT_ID")
	clientSecret := os.Getenv("DISGO_CLI_CLIENT_SECRET")
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("http://%s:%s/oauth/callback", host, port),
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}

	// Start web server on specified host and port
	hostPort := host + ":" + port
	fmt.Printf("🌐 Binding to %s...\n", hostPort)
	server := &http.Server{
		Addr:    hostPort,
		Handler: nil, // use default mux
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the login page
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Disgo CLI - Discord Login</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin: 50px; background: #36393f; color: #dcddde; }
        .container { max-width: 500px; margin: 0 auto; background: #2f3136; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.3); }
        .status { margin: 20px; color: #b9bbbe; font-size: 14px; }
        .button { background: #5865f2; color: white; border: none; padding: 12px 24px; border-radius: 4px; cursor: pointer; font-size: 16px; margin: 10px; }
        .button:hover { background: #4752c4; }
        .button:disabled { background: #404eed; cursor: not-allowed; }
        h1 { color: #ffffff; margin-bottom: 10px; }
        p { color: #b9bbbe; }
        .qr-placeholder { margin: 30px; padding: 40px; background: white; color: black; border-radius: 8px; font-family: monospace; display: inline-block; }
        .success { color: #57f287; font-weight: bold; }
        .error { color: #f04747; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Disgo CLI</h1>
        <p>Login with Discord</p>
        <div class="status" id="status">Choose a login method</div>

        <div id="login-options">
            <button class="button" onclick="showQR()">Login with QR Code</button>
            <br>
            <button class="button" onclick="startOAuth()">Login with Discord Auth</button>
        </div>

        <div id="qr-section" style="display: none;">
            <p>Scan the QR code below with your Discord mobile app:</p>
            <div style="margin: 30px;">
                <img id="qr-image" src="/qr" alt="QR Code for Discord Login" style="border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.3);" width="256" height="256" onerror="showQRError()" />
                <div id="qr-error" style="display: none; margin: 20px; color: #f04747; font-weight: bold;">
                    Unable to generate QR code. OAuth not configured properly.
                </div>
            </div>
            <button class="button" onclick="hideQR()">Back</button>
        </div>

        <div id="oauth-section" style="display: none;">
            <p>Login with Discord OAuth2:</p>
            <p>Note: This requires the server to be configured with Discord OAuth2 credentials.</p>
            <button class="button" onclick="startOAuthFlow()">Continue with Discord</button>
            <br>
            <button class="button" onclick="hideOAuth()">Back</button>
        </div>

        <div id="loading-section" style="display: none;">
            <p>Redirecting to Discord...</p>
        </div>
    </div>
    <script>
        function showQR() {
            document.getElementById('login-options').style.display = 'none';
            document.getElementById('qr-section').style.display = 'block';
            document.getElementById('status').textContent = 'Scan the QR code with your Discord mobile app to login.';
        }

        function hideQR() {
            document.getElementById('login-options').style.display = 'block';
            document.getElementById('qr-section').style.display = 'none';
            document.getElementById('status').textContent = 'Choose a login method';
        }

        function startOAuth() {
            document.getElementById('login-options').style.display = 'none';
            document.getElementById('oauth-section').style.display = 'block';
            document.getElementById('status').textContent = 'Login with your Discord account.';
        }

        function startOAuthFlow() {
            document.getElementById('oauth-section').style.display = 'none';
            document.getElementById('loading-section').style.display = 'block';
            document.getElementById('status').textContent = 'Redirecting you to Discord...';

            // Redirect to OAuth URL
            window.location.href = '/oauth/login';
        }

        function hideOAuth() {
            document.getElementById('login-options').style.display = 'block';
            document.getElementById('oauth-section').style.display = 'none';
            document.getElementById('loading-section').style.display = 'none';
            document.getElementById('status').textContent = 'Choose a login method';
        }

        function showQRError() {
            document.getElementById('qr-error').style.display = 'block';
        }

        // Check URL parameters for success/error messages
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('success')) {
            document.getElementById('status').innerHTML = '<span class="success">✓ You have been successfully logged in! You can close this window.</span>';
        } else if (urlParams.get('error')) {
            document.getElementById('status').innerHTML = '<span class="error">✗ Login failed: ' + urlParams.get('error') + '</span>';
        }
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// OAuth login endpoint
	http.HandleFunc("/oauth/login", func(w http.ResponseWriter, r *http.Request) {
		state := generateState()
		url := oauthConfig.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	// OAuth callback endpoint
	http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check for errors in callback
		if r.URL.Query().Get("error") != "" {
			errorMsg := r.URL.Query().Get("error_description")
			if errorMsg == "" {
				errorMsg = r.URL.Query().Get("error")
			}
			http.Redirect(w, r, "/?error="+errorMsg, http.StatusTemporaryRedirect)
			return
		}

		// Exchange code for token
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Redirect(w, r, "/?error=No authorization code received", http.StatusTemporaryRedirect)
			return
		}

		token, err := exchangeCodeForToken(code)
		if err != nil {
			slog.Error("Failed to exchange OAuth code for token", "error", err)
			http.Redirect(w, r, "/?error=Failed to authenticate with Discord", http.StatusTemporaryRedirect)
			return
		}

		// Store the token in keyring
		if err := keyring.SetToken(token); err != nil {
			slog.Error("Failed to store token in keyring", "error", err)
			http.Redirect(w, r, "/?error=Failed to save login session", http.StatusTemporaryRedirect)
			return
		}

		slog.Info("User successfully logged in via OAuth2")

		// Redirect back to homepage with success message
		http.Redirect(w, r, "/?success=1", http.StatusTemporaryRedirect)
	})

	// QR Code endpoint
	http.HandleFunc("/qr", func(w http.ResponseWriter, r *http.Request) {
		// Generate a QR code that links to Discord OAuth login
		clientID := os.Getenv("DISGO_CLI_CLIENT_ID")
		clientSecret := os.Getenv("DISGO_CLI_CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			fmt.Printf("🔴 QR Debug: Missing OAuth credentials - ClientID: '%s', ClientSecret: '%s'\n",
				clientID, strings.Repeat("*", len(clientSecret)))
			http.Error(w, "Discord OAuth not configured - missing client_id or client_secret", http.StatusServiceUnavailable)
			return
		}

		fmt.Printf("🟢 QR Debug: OAuth configured - generating QR code...\n")

		// Refresh OAuth config with current host and port
		oauthConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  fmt.Sprintf("http://%s:%s/oauth/callback", host, port),
			Scopes:       []string{"identify"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		}

		state := generateState()
		oauthURL := oauthConfig.AuthCodeURL(state)

		fmt.Printf("🟢 QR Debug: OAuth URL: %s...\n", oauthURL[:100]+"...")

		// Generate QR code
		qrCode, err := qrcode.New(oauthURL, qrcode.Medium)
		if err != nil {
			fmt.Printf("🔴 QR Debug: Failed to generate QR code: %v\n", err)
			http.Error(w, "Failed to generate QR code: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get QR code as PNG
		png, err := qrCode.PNG(256)
		if err != nil {
			fmt.Printf("🔴 QR Debug: Failed to encode QR code: %v\n", err)
			http.Error(w, "Failed to encode QR code: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("🟢 QR Debug: QR code generated successfully (%d bytes)\n", len(png))

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(png)
	})

		fmt.Printf("🌐 Starting web server on http://%s:%s...\n", host, port)
		fmt.Printf("🌐 Open your browser at http://%s:%s\n", host, port)

		if host != "localhost" && host != "127.0.0.1" {
			fmt.Println("📱 IMPORTANT: For QR codes to work on mobile devices,")
			fmt.Println("📱 The redirect URI in your Discord App OAuth2 settings")
			fmt.Printf("📱 must be: http://%s:%s/oauth/callback\n", host, port)
		}

		fmt.Println("Press Ctrl+C to stop the server")

		// Check if port is available on specified host
		fmt.Print("✓ Checking server availability... ")
		serverAddr := host + ":" + port
		listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Port %s already in use: %v\n", port, err)
		return fmt.Errorf("web server failed: %w", err)
	}
	listener.Close()
	fmt.Printf("✓ Port %s is available\n", port)

	fmt.Println("✓ Starting web server...")
	fmt.Printf("✓ Web server is now running at http://%s:%s\n", host, port)
	fmt.Println("✓ Open your browser and scan QR codes with Discord mobile app")
	fmt.Println("✓ Press Ctrl+C to stop the web server")

	// Start server (blocks until interrupted)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("Web server error: %v\n", err)
	}

	fmt.Println("✓ Web server stopped")
	return nil
}
