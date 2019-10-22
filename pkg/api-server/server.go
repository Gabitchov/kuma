package api_server

import (
	"context"
	"fmt"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
	"net/http"

	"github.com/Kong/kuma/pkg/api-server/definitions"
	config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/emicklei/go-restful"
)

var (
	log = core.Log.WithName("api-server")
)

type ApiServer struct {
	server *http.Server
}

func (a *ApiServer) Address() string {
	return a.server.Addr
}

func NewApiServer(resManager manager.ResourceManager, defs []definitions.ResourceWsDefinition, config config.ApiServerConfig) *ApiServer {
	container := restful.NewContainer()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: container.ServeMux,
	}

	ws := new(restful.WebService)
	ws.
		Path("/meshes").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	addToWs(ws, defs, resManager, config)
	container.Add(ws)
	container.Add(indexWs())

	configureOpenApi(container, ws, defs)
	return &ApiServer{
		server: srv,
	}
}

func addToWs(ws *restful.WebService, defs []definitions.ResourceWsDefinition, resManager manager.ResourceManager, config config.ApiServerConfig) {
	overviewWs := overviewWs{
		resManager: resManager,
	}
	overviewWs.AddToWs(ws)

	for _, definition := range defs {
		resourceWs := resourceWs{
			resManager:           resManager,
			readOnly:             config.ReadOnly,
			ResourceWsDefinition: definition,
			sample:               definition.Sample,
			sampleList:           definition.SampleList,
		}
		resourceWs.AddToWs(ws)
	}
}

func configureOpenApi(container *restful.Container, webService *restful.WebService, wsDefinitions []definitions.ResourceWsDefinition) {
	openApiConfig := restfulspec.Config{
		WebServices: []*restful.WebService{webService},
		APIPath:     "/apidocs.json",
		PostBuildSwaggerObjectHandler: func(s *spec.Swagger) {
			for _, def := range wsDefinitions {
				if def.Name == "Dataplane" {
					sdef := s.Definitions["definitions.dataplaneRestResource"]
					sdef.Example = def.Sample
					s.Definitions["definitions.dataplaneRestResource"] = sdef
				}
				//if def.Name == "Mesh" {
				//	sdef := s.Definitions["definitions.meshRestResource"]
				//	sdef.Example = def.Sample
				//	s.Definitions["definitions.meshRestResource"] = sdef
				//}
			}
		},
	}
	container.Add(restfulspec.NewOpenAPIService(openApiConfig))

	// todo(jakubdyszkiewicz) figure out how to pack swagger ui dist package and expose swagger ui
	//container.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("path/to/swagger-ui-dist"))))
}

func (a *ApiServer) Start(stop <-chan struct{}) error {
	errChan := make(chan error)
	go func() {
		err := a.server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("Shutting down server")
			default:
				log.Error(err, "Could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	log.Info("starting", "port", a.Address())
	select {
	case <-stop:
		log.Info("Stopping down API Server")
		return a.server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func SetupServer(rt runtime.Runtime) error {
	apiServer := NewApiServer(rt.ResourceManager(), definitions.All, *rt.Config().ApiServer)
	return rt.Add(apiServer)
}
