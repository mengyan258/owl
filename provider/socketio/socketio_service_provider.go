package socketio

import (
	"net/http"

	"bit-labs.cn/owl/contract/foundation"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

type SocketIOServiceProvider struct {
	app foundation.Application
}

var _ foundation.ServiceProvider = (*SocketIOServiceProvider)(nil)

func (s SocketIOServiceProvider) Register() {
	s.app.Register(func() *socketio.Server {
		server := socketio.NewServer(&engineio.Options{
			Transports: []transport.Transport{
				&websocket.Transport{
					CheckOrigin: func(r *http.Request) bool {
						return true
					},
				},
			},
		})
		return server
	})

}

func (s SocketIOServiceProvider) Boot() {

}

func (s SocketIOServiceProvider) GenerateConf() map[string]string {
	return nil
}
