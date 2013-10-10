package peony

type App struct {
}

func NewApp() *App {
	app := &App{}
	return app
}

func (app *App) Bind(url string) {

}
