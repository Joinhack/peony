
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
)

func main() {
	flag.Parse()
	app := peony.NewApp(*srcPath, *importPath)
	app.BindAddr = *bindAddr
	svr := app.NewServer()

	svr.Mapper("/index", (*controllers0.Login)(nil), &peony.TypeAction{Name: "Login.Index", MethodName: "Index", MethodArgs:[]*peony.MethodArgType{&peony.MethodArgType{Name:"user", Type:reflect.TypeOf((*[]*controllers0.Mail)(nil)).Elem()},
		&peony.MethodArgType{Name:"m", Type:reflect.TypeOf((*models0.User)(nil))}}})
	svr.Mapper("/test", controllers0.Index, &peony.MethodAction{Name:"Index", MethodArgs:[]*peony.MethodArgType{&peony.MethodArgType{Name:"s", Type:reflect.TypeOf((*controllers1.S)(nil))},
		&peony.MethodArgType{Name:"ss", Type:reflect.TypeOf((*string)(nil)).Elem()}}})


	go func(){
		time.Sleep(1*time.Second)
		fmt.Println("Server is running, bind at", app.BindAddr)
	}()
	if err := svr.Run(); err != nil {
		panic(err)
	}
}
