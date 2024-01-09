package websocketclient

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type WSAPINotification struct {
	ID    int            `json:"id"`
	Type  string         `json:"type"`
	Event WSAPIEventType `json:"event"`
}

type WSAPIEventType struct {
	Type          string                              `json:"type"`
	Notifications map[string]WSAPINotificationDetails `json:"notifications"`
}

type WSAPINotificationDetails struct {
	Message        string `json:"message"`
	NotificationID string `json:"notification_id"`
	Title          string `json:"title"`
	CreatedAt      string `json:"created_at"`
}

type WebSocketClient struct {
	URL  string
	Conn *websocket.Conn
}

func NewWebSocketClient(url string) *WebSocketClient {
	return &WebSocketClient{
		URL: url,
	}
}

func (client *WebSocketClient) Connect() error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	conn, _, err := dialer.Dial(client.URL, http.Header{})
	if err != nil {
		return err
	}

	client.Conn = conn
	return nil
}

type WebsocketMessageAuthBody struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
}

type WebsocketMessageBody struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
}

const WebsocketMessageTypeAuthRequired = "auth_required"
const WebsocketMessageTypeAuthOK = "auth_ok"

func (client *WebSocketClient) FetchNotifications(accessToken string) (*WSAPINotification, error) {
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	defer client.Conn.Close()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		var response WSAPINotification
		err = json.Unmarshal(message, &response)
		if err != nil {
			return nil, err
		}

		if response.Type == WebsocketMessageTypeAuthRequired {
			stringJSON, _ := json.Marshal(WebsocketMessageAuthBody{Type: "auth", AccessToken: accessToken})
			client.Conn.WriteMessage(websocket.TextMessage, []byte(stringJSON))
		} else if response.Type == WebsocketMessageTypeAuthOK {
			stringJSON, _ := json.Marshal(WebsocketMessageBody{Id: 5, Type: "persistent_notification/subscribe"})
			client.Conn.WriteMessage(websocket.TextMessage, []byte(stringJSON))
		} else if response.Type == "event" {
			return &response, nil
		}
	}
}

// func main() {
// 	client := NewWebSocketClient("ws://192.168.1.24:8123/api/websocket")

// 	notification, err := client.FetchNotifications()
// 	if err != nil {
// 		log.Fatalf("Error fetching notifications: %v", err)
// 	}

// 	fmt.Printf("Notification count: %d\n", len(notification.Event.Notifications))

// 	for _, notification := range notification.Event.Notifications {
// 		fmt.Printf("Notification: %s\n\n", notification)
// 	}
// }
