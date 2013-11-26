
package main
import (
	"github.com/joinhack/peony"
	"reflect"
	"time"
	"fmt"
	"flag"
	controllers0 "github.com/joinhack/peony/demos/login/app/controllers"
	controllers1 "github.com/joinhack/peony/demos/login/app/controllers/controllers"
	models0 "github.com/joinhack/peony/demos/login/app/models"
)

var (
	runMode    *string = flag.String("runMode", "", "Run mode.")
	bindAddr   *string = flag.String("bindAddr", ":8080", "By default, read from app.conf")
	importPath *string = flag.String("importPath", "", "Go Import Path for the app.")
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
	svr.Mapper("/", (*controllers0.Login)(nil), &peony.MethodAction{Name: "Login.Index", MethodName: "Index", MethodArgs:[]*peony.ArgType{&peony.ArgType{Name:"user", Type:reflect.TypeOf((*[]*controllers0.Mail)(nil)).Elem()},
		&peony.ArgType{Name:"m", Type:reflect.TypeOf((*models0.User)(nil))}}})
	svr.Mapper("/test", controllers0.Index, &peony.FuncAction{Name:"Index", MethodArgs:[]*peony.ArgType{&peony.ArgType{Name:"s", Type:reflect.TypeOf((*controllers1.S)(nil))},
		&peony.ArgType{Name:"ss", Type:reflect.TypeOf((*string)(nil)).Elem()}}})


	go func(){
		time.Sleep(1*time.Second)
		fmt.Println("Server is running, listening on", app.BindAddr)
	}()
	if err := svr.Run(); err != nil {
		panic(err)
	}
}
