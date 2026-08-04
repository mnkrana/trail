package calendar

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

const (
	authTimeout = 2 * time.Minute
)

func RunAuthServer(ctx context.Context, cfg *oauth2.Config) error {
	state, err := randomState()
	if err != nil {
		return err
	}

	listeners := bindLoopback()
	if len(listeners) == 0 {
		return fmt.Errorf("cannot bind localhost:8080 — port is in use on both IPv4 and IPv6. Stop the process holding it (check: lsof -i :8080) and retry 'trail auth'")
	}
	defer func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}()

	done := make(chan *TokenData, 1)
	serverErr := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/calendar/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("state") != state {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "<h3>Error: state mismatch. Close this tab and retry.</h3>")
			return
		}

		token, err := cfg.Exchange(r.Context(), query.Get("code"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<h3>Error: token exchange failed: %v</h3>", err)
			return
		}

		tokens := &TokenData{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenType:    token.TokenType,
			Expiry:       token.Expiry,
		}
		if err := saveTokens(tokens); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "<h3>Error: failed to save tokens: %v</h3>", err)
			return
		}

		fmt.Fprint(w, "<h2>✅ Authenticated successfully!</h2><p>You can close this tab and return to the terminal.</p>")
		fmt.Println("\n✅ Authenticated successfully! Tokens saved.")
		done <- tokens
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "<p>This is the trail auth server. Nothing to see here — close this tab.</p>")
	})

	server := &http.Server{Handler: mux}
	for _, ln := range listeners {
		go func(l net.Listener) {
			if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
				serverErr <- err
			}
		}(ln)
	}

	time.Sleep(100 * time.Millisecond)

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	fmt.Println("🔐 Starting auth on http://localhost:8080")
	fmt.Println("Opening your browser to authorize the application...")
	fmt.Println()
	fmt.Println("If the browser doesn't open, visit:")
	fmt.Println(authURL)
	fmt.Println()

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("⚠️  Could not auto-open browser: %v\n", err)
	}

	select {
	case <-done:
		return nil
	case err := <-serverErr:
		return err
	case <-time.After(authTimeout):
		return fmt.Errorf("authentication timed out after %v — run 'trail auth' again", authTimeout)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func bindLoopback() []net.Listener {
	var listeners []net.Listener

	ln4, err4 := net.Listen("tcp4", "127.0.0.1:8080")
	if err4 == nil {
		listeners = append(listeners, ln4)
	} else {
		fmt.Printf("⚠️  Could not bind 127.0.0.1:8080 (IPv4): %v\n", err4)
	}

	ln6, err6 := net.Listen("tcp6", "[::1]:8080")
	if err6 == nil {
		listeners = append(listeners, ln6)
	} else {
		fmt.Printf("⚠️  Could not bind [::1]:8080 (IPv6): %v\n", err6)
	}

	return listeners
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
