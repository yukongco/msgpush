package push

import (
	"fmt"
	"sync"

	"github.com/yukongco/msgpush/comet/conf"
	. "github.com/yukongco/msgpush/comet/logs"
	"github.com/yukongco/msgpush/common/check"
)

var (
	HubOrg *Hub
)

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients, 采用安全类型的 map
	Clients *sync.Map

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	HubOrg = &Hub{
		Broadcast:  make(chan []byte, conf.Conf.WebsocketConf.BroadcastMax),
		Register:   make(chan *Client, conf.Conf.WebsocketConf.RegisterMax),
		Unregister: make(chan *Client, conf.Conf.WebsocketConf.UnregisterMax),
		Clients:    &sync.Map{},
	}

	return HubOrg
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// key = devId + code
			h.Clients.LoadOrStore(client.User.Key+client.User.Code, client)
			// notify user off-line message
			go client.NotifyMsg()

		case client := <-h.Unregister:
			key := client.User.Key + client.User.Code
			_, ok := h.Clients.Load(key)
			if ok {
				h.Clients.Delete(key)
				close(client.send)
			}
		case message := <-h.Broadcast: // 推送给所有的client
			h.Clients.Range(func(key, value interface{}) bool {
				client, ok := value.(*Client)
				if !ok {
					Log.Errorf("The key is not type of *Client, key: %v", key)
					return true
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					h.Clients.Delete(client)
				}
				return true
			})
		}
	}
}

// judge dev is on line at this node ?
func (h *Hub) IsOnline(cliKey string) (*Client, bool) {
	value, ok := HubOrg.Clients.Load(cliKey)
	if !ok {
		return nil, false
	}

	cli, ok := value.(*Client)
	if !ok {
		Log.Errorf("The key is not type of *Client, key: %v", cliKey)
		return nil, false
	}

	return cli, true
}

// android and pc msg push
func AndroidPcPush(key string, msg []byte) error {
	client, exist := HubOrg.IsOnline(key)
	if !exist {
		fmt.Println("user not on line")
		return check.User_Not_On_Line
	}

	client.send <- msg

	return nil
}
