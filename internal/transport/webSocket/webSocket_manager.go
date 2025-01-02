package webSocket

import (
	"context"
	"fmt"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/enums"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/url"
	"time"
)

type Manager struct {
	conn           *websocket.Conn
	subscriptionID string
}

func New(schema, host, path string) (*Manager, error) {
	u := url.URL{Scheme: schema, Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect WebSocket: %w", err)
	}
	return &Manager{conn: conn}, nil
}

func (w *Manager) Subscribe(ctx context.Context, action enums.SubscriptionAction) (string, error) {
	subscribeRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  string(action),
		"params": []interface{}{
			map[string]interface{}{
				"mentions": []string{"any"},
			},
		},
	}

	err := w.writeJSONWithContext(ctx, subscribeRequest)
	if err != nil {
		return "", fmt.Errorf("failed to send subscribe request: %w", err)
	}

	var response struct {
		Result struct {
			Subscription string `json:"subscription"`
		} `json:"result"`
	}
	err = w.readJSONWithContext(ctx, &response)
	if err != nil {
		return "", fmt.Errorf("failed to read subscription response: %w", err)
	}

	w.subscriptionID = response.Result.Subscription
	return w.subscriptionID, nil
}

func (w *Manager) Unsubscribe(ctx context.Context, action enums.SubscriptionAction) error {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  string(action),
		"params":  []interface{}{w.subscriptionID},
	}

	err := w.writeJSONWithContext(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to send unsubscribe request: %w", err)
	}

	log.Infof("Unsubscribed from %s", action)
	return nil
}

// Helper methods for context-aware WriteJSON and ReadJSON
func (w *Manager) writeJSONWithContext(ctx context.Context, v interface{}) error {
	done := make(chan error, 1)
	go func() {
		done <- w.conn.WriteJSON(v)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (w *Manager) readJSONWithContext(ctx context.Context, v interface{}) error {
	done := make(chan error, 1)
	go func() {
		done <- w.conn.ReadJSON(v)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (w *Manager) IsConnected() bool {
	return w.conn != nil && w.pingConnection()
}

func (w *Manager) pingConnection() bool {
	if err := w.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second)); err != nil {
		log.WithError(err).Debug("WebSocket connection ping failed")
		return false
	}
	return true
}

func (w *Manager) ReadMessage() ([]byte, error) {
	_, message, err := w.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading WebSocket message: %w", err)
	}
	return message, nil
}

func (w *Manager) Close() error {
	if w.conn != nil {
		if err := w.conn.Close(); err != nil {
			return fmt.Errorf("failed to close WebSocket connection: %w", err)
		}
		log.Info("WebSocket connection closed")
	}
	return nil
}
