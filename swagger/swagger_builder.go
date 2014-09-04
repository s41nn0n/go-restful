package swagger

import (
	"reflect"
	"strings"

	"github.com/emicklei/go-restful"
)

type SwaggerApiDeclarationBuilder struct {
	Config  Config
	Service *restful.WebService
}

func (b SwaggerApiDeclarationBuilder) ApiDeclaration() (key string, decl ApiDeclaration) {
	rootPath := b.Service.RootPath()
	if rootPath == "" || rootPath == "/" {
		// use routes
		for _, route := range b.Service.Routes() {
			key := b.staticPathFromRoute(route)
			decl = b.ComposeDeclaration(key)
		}
	} else { // use root path
		key = b.Service.RootPath()
		decl = b.ComposeDeclaration(key)
	}
	return key, decl
}

func (b SwaggerApiDeclarationBuilder) ComposeDeclaration(pathPrefix string) ApiDeclaration {
	decl := ApiDeclaration{
		SwaggerVersion: swaggerVersion,
		BasePath:       b.Config.WebServicesUrl,
		ResourcePath:   b.Service.RootPath(),
		Models:         map[string]Model{}}

	// collect any path parameters
	rootParams := []Parameter{}
	for _, param := range b.Service.PathParameters() {
		rootParams = append(rootParams, asSwaggerParameter(param.Data()))
	}
	// aggregate by path
	pathToRoutes := map[string][]restful.Route{}
	for _, other := range b.Service.Routes() {
		if strings.HasPrefix(other.Path, pathPrefix) {
			routes := pathToRoutes[other.Path]
			pathToRoutes[other.Path] = append(routes, other)
		}
	}
	for path, routes := range pathToRoutes {
		api := Api{Path: path, Description: b.Service.Documentation()}
		for _, route := range routes {
			operation := Operation{HttpMethod: route.Method,
				Summary:  route.Doc,
				Type:     asDataType(route.WriteSample),
				Nickname: route.Operation}

			operation.Consumes = route.Consumes
			operation.Produces = route.Produces

			// share root params if any
			for _, swparam := range rootParams {
				operation.Parameters = append(operation.Parameters, swparam)
			}
			// route specific params
			for _, param := range route.ParameterDocs {
				operation.Parameters = append(operation.Parameters, asSwaggerParameter(param.Data()))
			}
			b.addModelsFromRouteTo(&operation, route, &decl)
			api.Operations = append(api.Operations, operation)
		}
		decl.Apis = append(decl.Apis, api)
	}
	return decl
}

// addModelsFromRoute takes any read or write sample from the Route and creates a Swagger model from it.
func (b SwaggerApiDeclarationBuilder) addModelsFromRouteTo(operation *Operation, route restful.Route, decl *ApiDeclaration) {
	if route.ReadSample != nil {
		b.addModelFromSampleTo(operation, false, route.ReadSample, decl.Models)
	}
	if route.WriteSample != nil {
		b.addModelFromSampleTo(operation, true, route.WriteSample, decl.Models)
	}
}

// addModelFromSample creates and adds (or overwrites) a Model from a sample resource
func (b SwaggerApiDeclarationBuilder) addModelFromSampleTo(operation *Operation, isResponse bool, sample interface{}, models map[string]Model) {
	st := reflect.TypeOf(sample)
	isCollection := false
	if st.Kind() == reflect.Slice || st.Kind() == reflect.Array {
		st = st.Elem()
		isCollection = true
	} else {
		if st.Kind() == reflect.Ptr {
			if st.Elem().Kind() == reflect.Slice || st.Elem().Kind() == reflect.Array {
				st = st.Elem().Elem()
				isCollection = true
			}
		}
	}
	modelName := modelBuilder{}.keyFrom(st)
	if isResponse {
		if isCollection {
			modelName = "array[" + modelName + "]"
		}
		operation.Type = modelName
	}
	modelBuilder{models}.addModel(reflect.TypeOf(sample), "")
}

func (b SwaggerApiDeclarationBuilder) staticPathFromRoute(r restful.Route) string {
	static := r.Path
	bracket := strings.Index(static, "{")
	if bracket <= 1 { // result cannot be empty
		return static
	}
	if bracket != -1 {
		static = r.Path[:bracket]
	}
	if strings.HasSuffix(static, "/") {
		return static[:len(static)-1]
	} else {
		return static
	}
}

func asSwaggerParameter(param restful.ParameterData) Parameter {
	return Parameter{
		Name:        param.Name,
		Description: param.Description,
		ParamType:   asParamType(param.Kind),
		Type:        param.DataType,
		DataType:    param.DataType,
		Format:      asFormat(param.DataType),
		Required:    param.Required}
}

func asFormat(name string) string {
	return "" // TODO
}

func asParamType(kind int) string {
	switch {
	case kind == restful.PathParameterKind:
		return "path"
	case kind == restful.QueryParameterKind:
		return "query"
	case kind == restful.BodyParameterKind:
		return "body"
	case kind == restful.HeaderParameterKind:
		return "header"
	case kind == restful.FormParameterKind:
		return "form"
	}
	return ""
}

func asDataType(any interface{}) string {
	if any == nil {
		return "void"
	}
	return reflect.TypeOf(any).Name()
}
