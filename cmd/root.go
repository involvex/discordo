// Package cmd defines the command-line interface commands
package cmd

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ayn2op/tview"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/diamondburned/ningen/v3"
	"github.com/involvex/disgo-cli/internal/config"
	"github.com/involvex/disgo-cli/internal/keyring"
	"github.com/involvex/disgo-cli/internal/logger"
)

var (
	discordState *ningen.State
	app          *application
)

func Run() error {
	tokenEnvVar := os.Getenv("DISGO_CLI_TOKEN")
	tokenFlag := flag.String("token", tokenEnvVar, "authentication token")
	serveFlag := flag.Bool("serve", false, "start web server for QR code login")

	configPath := flag.String("config-path", config.DefaultPath(), "path of the configuration file")
	logPath := flag.String("log-path", logger.DefaultPath(), "path of the log file")
	logLevel := flag.String("log-level", "info", "log level")
	flag.Parse()

	var level slog.Level
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
		return startWebServer(token)
	}

	tview.Styles = tview.Theme{}
	app = newApplication(cfg)
	return app.run(token)
}

func startWebServer(existingToken string) error {
	if existingToken != "" {
		fmt.Println("Already logged in. Token found in keyring.")
		return nil
	}

	// Start web server on port 8080
	server := &http.Server{Addr: ":8080"}

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
        h1 { color: #ffffff; margin-bottom: 10px; }
        p { color: #b9bbbe; }
        .qr-placeholder { margin: 30px; padding: 40px; background: white; color: black; border-radius: 8px; font-family: monospace; display: inline-block; }
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
            <button class="button" onclick="startOAuth()">Login with Discord API</button>
        </div>

        <div id="qr-section" style="display: none;">
            <p>Scan the QR code below with your Discord mobile app:</p>
            <div class="qr-placeholder" id="qr-code">
                █▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░█<br>
                █▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄█<br>
            </div>
            <button class="button" onclick="hideQR()">Back</button>
        </div>

        <div id="oauth-section" style="display: none;">
            <p>OAuth login is not implemented yet.</p>
            <p>Please use QR code login for now.</p>
            <button class="button" onclick="hideOAuth()">Back</button>
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
            document.getElementById('status').textContent = 'OAuth login is not implemented yet.';
        }

        function hideOAuth() {
            document.getElementById('login-options').style.display = 'block';
            document.getElementById('oauth-section').style.display = 'none';
            document.getElementById('status').textContent = 'Choose a login method';
        }
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	fmt.Println("Starting web server on http://localhost:8080")
	fmt.Println("Open your browser and click 'Login with QR Code' to see the QR code interface")
	fmt.Println("Note: QR code generation requires implementation of the Discord auth gateway")

	return server.ListenAndServe()
}
