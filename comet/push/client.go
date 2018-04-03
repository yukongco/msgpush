package push

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yukongco/msgpush/comet/conf"
	. "github.com/yukongco/msgpush/comet/logs"
	"github.com/yukongco/msgpush/common/check"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 20 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		originStr := r.Header.Get("Origin")
		webVersion := r.Header.Get("Sec-WebSocket-Version")
		upgradeStr := r.Header.Get("Upgrade")

		fmt.Printf("origin=%v, webVersion=%v, upgradeStr=%v\n", originStr, webVersion, upgradeStr)

		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// 存储用户的一些信息
	User *UserInfo
}

type UserInfo struct {
	Terminal string `json:"terminal"` // 终端, pc, ios, android
	Key      string `json:"key"`      // 设备 id
	Code     string `json:"code"`     // 运行商 code, 标识
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	Log.Info("Connet.............")

	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		fmt.Println("read message before")
		// 使用 readMessage 会造成阻塞
		msmType, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println("Read msg err: ", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Log.Errorf("msmType=%v message=%v, error: %v", msmType, string(message), err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			//fmt.Println("write message")
			//Log.Info("write message: ", string(message))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(check.Newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServerWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Log.Errorf("Connet upgrade is err: %v", err)
		fmt.Println(err)
		return
	}

	Log.Info("ServerWs..............")
	userInfo, err := ParseParms(r)
	if err != nil {
		Log.Errorf("Parse parms is err: %v", err)
		return
	}

	client := &Client{hub: hub,
		conn: conn,
		send: make(chan []byte, conf.Conf.WebsocketConf.ClientSendMax),
		User: userInfo}
	client.hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// 解析客户端的参数
func ParseParms(r *http.Request) (*UserInfo, error) {
	info := UserInfo{}

	devType := GetDevType(r.Header.Get("User-Agent"))

	code, err := UrlParseParm(r, "code")
	if err != nil {
		Log.Errorf("Url parse is err: %v", err)
		return nil, fmt.Errorf("Url parse is err: %v", err)
	}

	key, err := UrlParseParm(r, "key")
	if err != nil {
		Log.Errorf("Dev id is err: %v", err)
		return nil, fmt.Errorf("Url parse is err: %v", err)
	}

	info.Terminal = devType
	info.Key = key
	info.Code = code

	return &info, nil
}

// 获取 url 里面的参数
func UrlParseParm(r *http.Request, key string) (string, error) {
	urlValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		Log.Errorf("Parse url is err: %v", err)
		return "", fmt.Errorf("Parse url is err: %v", err)
	}

	value := urlValues.Get(key)

	return value, nil
}

// 判断终端类型
func GetDevType(userAgent string) string {
	flag := strings.Contains(userAgent, "iphone")
	if flag == true {
		return check.IOS
	}

	flag = strings.Contains(userAgent, "android")
	if flag == true {
		return check.Android
	}

	flag = strings.Contains(userAgent, "Windows")
	if flag == true {
		return check.PC
	}

	flag = strings.Contains(userAgent, "ipad")
	if flag == true {
		return check.IOS
	}

	return check.Android
}
