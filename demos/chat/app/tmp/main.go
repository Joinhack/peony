
package main
import (
	"reflect"
	"time"
	"fmt"
	"flag"
	websocket0 "code.google.com/p/go.net/websocket"
	peony "github.com/joinhack/peony"
	controllers0 "github.com/joinhack/peony/demos/chat/app/controllers"
)

var (
	_ = reflect.Ptr
	runMode    *string = flag.String("runMode", "", "Run mode.")
	bindAddr   *string = flag.String("bindAddr", ":8080", "By default, read from app.conf")
	importPath *string = flag.String("importPath", "", "Go ImportPath for the app.")
	srcPath    *string = flag.String("srcPath", "", "Path to the source root.")
	devMode    *bool    = flag.Bool("devMode", false, "Run mode")
)

func main() {
	flag.Parse()
	app := peony.NewApp(*srcPath, *importPath)
	app.BindAddr = *bindAddr
	if devMode != nil {
		app.DevMode = *devMode
	}
	svr := app.NewServer()
	svr.Init()

	svr.MethodMapper(`/`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.Application).Index, &peony.Action{
			Name: "Application.Index",
			},
	)

	svr.MethodMapper(`/application/enterdemo`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.Application).EnterDemo, &peony.Action{
			Name: "Application.EnterDemo",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			
				&peony.ArgType{
					Name: "demo", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/longpolling/room`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.LongPolling).Room, &peony.Action{
			Name: "LongPolling.Room",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/longpolling/room/messages`, []string{"POST"}, 
		(*controllers0.LongPolling).Say, &peony.Action{
			Name: "LongPolling.Say",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			
				&peony.ArgType{
					Name: "message", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/longpolling/room/messages`, []string{"GET"}, 
		(*controllers0.LongPolling).WaitMessages, &peony.Action{
			Name: "LongPolling.WaitMessages",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "lastReceived", 
					Type: reflect.TypeOf((*int)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/longpolling/room/leave`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.LongPolling).Leave, &peony.Action{
			Name: "LongPolling.Leave",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/refresh`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.Refresh).Index, &peony.Action{
			Name: "Refresh.Index",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/refresh/room`, []string{"GET"}, 
		(*controllers0.Refresh).Room, &peony.Action{
			Name: "Refresh.Room",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/refresh/room`, []string{"POST"}, 
		(*controllers0.Refresh).Say, &peony.Action{
			Name: "Refresh.Say",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			
				&peony.ArgType{
					Name: "message", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/refresh/room/leave`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.Refresh).Leave, &peony.Action{
			Name: "Refresh.Leave",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/websocket/room`, []string{"GET", "POST", "PUT", "DELETE"}, 
		(*controllers0.WebSocket).Room, &peony.Action{
			Name: "WebSocket.Room",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			}},
	)

	svr.MethodMapper(`/websocket/room/socket`, []string{"WS"}, 
		(*controllers0.WebSocket).RoomSocket, &peony.Action{
			Name: "WebSocket.RoomSocket",
			
			Args: []*peony.ArgType{ 
				
				&peony.ArgType{
					Name: "user", 
					Type: reflect.TypeOf((*string)(nil)).Elem(),
				},
			
				&peony.ArgType{
					Name: "ws", 
					Type: reflect.TypeOf((*websocket0.Conn)(nil)),
				},
			}},
	)


	svr.Router.Refresh()

	go func(){
		time.Sleep(1*time.Second)
		fmt.Println("Server is running, listening on", app.BindAddr)
	}()
	if err := <- svr.Run(); err != nil {
		panic(err)
	}
}
