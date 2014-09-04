package swagger

import (
	"github.com/emicklei/go-restful"
	// "github.com/emicklei/hopwatch"
	"log"
	"net/http"
)

type SwaggerService struct {
	config            Config
	apiDeclarationMap map[string]ApiDeclaration
}

func newSwaggerService(config Config) *SwaggerService {
	return &SwaggerService{
		config:            config,
		apiDeclarationMap: map[string]ApiDeclaration{}}
}

// LogInfo is the function that is called when this package needs to log. It defaults to log.Printf
var LogInfo = log.Printf

// InstallSwaggerService add the WebService that provides the API documentation of all services
// conform the Swagger documentation specifcation. (https://github.com/wordnik/swagger-core/wiki).
func InstallSwaggerService(aSwaggerConfig Config) {
	RegisterSwaggerService(aSwaggerConfig, restful.DefaultContainer)
}

// RegisterSwaggerService add the WebService that provides the API documentation of all services
// conform the Swagger documentation specifcation. (https://github.com/wordnik/swagger-core/wiki).
func RegisterSwaggerService(config Config, wsContainer *restful.Container) {
	sws := newSwaggerService(config)
	ws := new(restful.WebService)
	ws.Path(config.ApiPath)
	ws.Produces(restful.MIME_JSON)
	if config.DisableCORS {
		ws.Filter(enableCORS)
	}
	ws.Route(ws.GET("/").To(sws.getListing))
	ws.Route(ws.GET("/{a}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}/{e}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}/{e}/{f}").To(sws.getDeclarations))
	ws.Route(ws.GET("/{a}/{b}/{c}/{d}/{e}/{f}/{g}").To(sws.getDeclarations))
	LogInfo("[restful/swagger] listing is available at %v%v", config.WebServicesUrl, config.ApiPath)
	wsContainer.Add(ws)

	// Build all ApiDeclarations
	for _, each := range config.WebServices {
		rootPath := each.RootPath()
		// skip the api service itself
		if rootPath != config.ApiPath {
			b := SwaggerApiDeclarationBuilder{
				Config:  config,
				Service: each,
			}
			key, decl := b.ApiDeclaration()
			sws.apiDeclarationMap[key] = decl
		}
	}

	// Check paths for UI serving
	if config.StaticHandler == nil && config.SwaggerFilePath != "" && config.SwaggerPath != "" {
		swaggerPathSlash := config.SwaggerPath
		// path must end with slash /
		if "/" != config.SwaggerPath[len(config.SwaggerPath)-1:] {
			LogInfo("[restful/swagger] use corrected SwaggerPath ; must end with slash (/)")
			swaggerPathSlash += "/"
		}

		LogInfo("[restful/swagger] %v%v is mapped to folder %v", config.WebServicesUrl, swaggerPathSlash, config.SwaggerFilePath)
		wsContainer.Handle(swaggerPathSlash, http.StripPrefix(swaggerPathSlash, http.FileServer(http.Dir(config.SwaggerFilePath))))

		//if we define a custom static handler use it
	} else if config.StaticHandler != nil && config.SwaggerPath != "" {
		swaggerPathSlash := config.SwaggerPath
		// path must end with slash /
		if "/" != config.SwaggerPath[len(config.SwaggerPath)-1:] {
			LogInfo("[restful/swagger] use corrected SwaggerFilePath ; must end with slash (/)")
			swaggerPathSlash += "/"

		}
		LogInfo("[restful/swagger] %v%v is mapped to custom Handler %T", config.WebServicesUrl, swaggerPathSlash, config.StaticHandler)
		wsContainer.Handle(swaggerPathSlash, config.StaticHandler)

	} else {
		LogInfo("[restful/swagger] Swagger(File)Path is empty ; no UI is served")
	}
}

func enableCORS(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if origin := req.HeaderParameter(restful.HEADER_Origin); origin != "" {
		// prevent duplicate header
		if len(resp.Header().Get(restful.HEADER_AccessControlAllowOrigin)) == 0 {
			resp.AddHeader(restful.HEADER_AccessControlAllowOrigin, origin)
		}
	}
	chain.ProcessFilter(req, resp)
}

func (sws SwaggerService) getListing(req *restful.Request, resp *restful.Response) {
	listing := ResourceListing{SwaggerVersion: swaggerVersion}
	for k, v := range sws.apiDeclarationMap {
		ref := ApiRef{Path: k}
		if len(v.Apis) > 0 { // use description of first (could still be empty)
			ref.Description = v.Apis[0].Description
		}
		listing.Apis = append(listing.Apis, ref)
	}
	resp.WriteAsJson(listing)
}

func (sws SwaggerService) getDeclarations(req *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(sws.apiDeclarationMap[composeRootPath(req)])
}

// Between 1..7 path parameters is supported
func composeRootPath(req *restful.Request) string {
	path := "/" + req.PathParameter("a")
	b := req.PathParameter("b")
	if b == "" {
		return path
	}
	path = path + "/" + b
	c := req.PathParameter("c")
	if c == "" {
		return path
	}
	path = path + "/" + c
	d := req.PathParameter("d")
	if d == "" {
		return path
	}
	path = path + "/" + d
	e := req.PathParameter("e")
	if e == "" {
		return path
	}
	path = path + "/" + e
	f := req.PathParameter("f")
	if f == "" {
		return path
	}
	path = path + "/" + f
	g := req.PathParameter("g")
	if g == "" {
		return path
	}
	return path + "/" + g
}
