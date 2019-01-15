package web

import (
	"fmt"
	"github.com/CloudyKit/jet"
	"github.com/buaazp/fasthttprouter"
	"github.com/danielparks/code-manager-dashboard/codemanager"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type webServer struct {
	StateFilePath string
	CodeState     *codemanager.CodeState
	View          *jet.Set
}

var server webServer

func Serve(listenOn string, stateFilePath string) {
	/// FIXME bindata
	server = webServer{
		View:          jet.NewHTMLSet("./web/templates"),
		StateFilePath: stateFilePath,
	}

	codeState, err := codemanager.LoadCodeState(stateFilePath)
	if err != nil {
		log.Fatal(err)
	}

	server.CodeState = &codeState

	router := fasthttprouter.New()
	router.GET("/", Home)
	/// FIXME bindata
	router.ServeFiles("/static/*filepath", "web/static")

	log.Infof("Listening on %v", listenOn)
	log.Fatal(fasthttp.ListenAndServe(listenOn, router.Handler))
}

func render(ctx *fasthttp.RequestCtx, templateName string, context interface{}) error {
	template, err := server.View.GetTemplate(templateName)
	if err != nil {
		ctx.SetStatusCode(500)
		fmt.Fprintf(ctx, "Error loading template: %v", err)
		log.Errorf("%v: loading %v", ctx.URI(), err)
		return err
	}

	vars := make(jet.VarMap)
	vars.Set("Ascending", codemanager.Ascending)
	vars.Set("Descending", codemanager.Descending)

	err = template.Execute(ctx, vars, context)
	if err != nil {
		ctx.SetStatusCode(500)
		fmt.Fprintf(ctx, "Error evaluating template: %v", err)
		log.Errorf("%v: evaluating %v", ctx.URI(), err)
		return err
	}

	ctx.SetContentType("text/html; charset=utf-8")
	return nil
}

func Home(ctx *fasthttp.RequestCtx) {
	log.Infof("Home: %v", ctx.URI())

	// Errors are handled within render
	render(ctx, "home.jet", server.CodeState)
}
