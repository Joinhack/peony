
package main
import (
	"reflect"
	"time"
	"fmt"
	"flag"
	peony "github.com/joinhack/peony"
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
	svr.FuncMapper("/static/<string:path>", []string{"GET","POST","PUT","DELETE"}, controllers0.FileDown, 
		&peony.Action{Name:"FileDown", Args:[]*peony.ArgType{&peony.ArgType{Name:"path", Type:reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.MethodMapper("/", []string{"POST"}, (*controllers0.Login).Index, 
		&peony.Action{Name: "Login.Index", Args:[]*peony.ArgType{&peony.ArgType{Name:"user", Type:reflect.TypeOf((*[]*controllers0.Mail)(nil)).Elem()},
		&peony.ArgType{Name:"m", Type:reflect.TypeOf((*models0.User)(nil))}}})
	svr.InterceptMethod((*controllers0.Login).Before, 0, 0)
	svr.FuncMapper("/test", []string{"GET","POST","PUT","DELETE"}, controllers0.Test, 
		&peony.Action{Name:"Test", })
	svr.FuncMapper("/index", []string{"GET","POST","PUT","DELETE"}, controllers0.Index, 
		&peony.Action{Name:"Index", Args:[]*peony.ArgType{&peony.ArgType{Name:"s", Type:reflect.TypeOf((*controllers1.S)(nil))},
		&peony.ArgType{Name:"ss", Type:reflect.TypeOf((*string)(nil)).Elem()}}})
	svr.MethodMapper("/login/login", []string{"GET","POST","PUT","DELETE"}, (*controllers0.Login).Login, 
		&peony.Action{Name: "Login.Login", Args:[]*peony.ArgType{&peony.ArgType{Name:"user", Type:reflect.TypeOf((*[]*controllers0.Mail)(nil)).Elem()},
		&peony.ArgType{Name:"m", Type:reflect.TypeOf((*models0.User)(nil))}}})


	go func(){
		time.Sleep(1*time.Second)
		fmt.Println("Server is running, listening on", app.BindAddr)
	}()
	if err := svr.Run(); err != nil {
		panic(err)
	}
}
