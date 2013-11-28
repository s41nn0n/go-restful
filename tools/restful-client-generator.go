package main

// Takes a Swagger JSON file and generates a Service Http client in Go

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/emicklei/go-restful/swagger"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// go run *.go -url http://localhost:8080/apidocs.json -pkg users && cat /tmp/service.go

var apidocsUrl string
var packageName string
var models map[string]swagger.Model

func main() {
	flag.StringVar(&apidocsUrl, "url", "", "endpoint of a REST service (e.g. http://myservice/apidocs.json)")
	flag.StringVar(&packageName, "pkg", "", "name of the package for the generated Service.")
	flag.Parse()
	if len(apidocsUrl) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	models = map[string]swagger.Model{}
	listing := new(swagger.ResourceListing)
	fetch(apidocsUrl, &listing)
	out, _ := os.Create("/tmp/service.go")
	defer out.Close()
	// import and declarations
	io.WriteString(out, package_source(packageName))
	io.WriteString(out, import_service_new_source())
	// all methods
	for _, each := range listing.Apis {
		generateForApi(each, out)
	}
	// all models
	for _, model := range models {
		generateForModel(model, out)
	}
	// helpers
	io.WriteString(out, uribuilder_source())
}

func generateForApi(apiRef swagger.ApiRef, out io.Writer) {
	declaration := new(swagger.ApiDeclaration)
	fetch(apidocsUrl+apiRef.Path, &declaration)

	for _, each := range declaration.Apis {
		for _, op := range each.Operations {
			generateForOperation(each.Path, op, out)
		}
	}
	// collect all models
	for _, model := range declaration.Models {
		models[model.Id] = model
	}
}

func generateForModel(model swagger.Model, out io.Writer) {
	io.WriteString(out, "type "+noPkg(model.Id)+" struct {\n")
	for name, each := range model.Properties {
		generateForModelProperty(name, each, out)
	}
	io.WriteString(out, "}\n")
}

func generateForModelProperty(name string, prop swagger.ModelProperty, out io.Writer) {
	io.WriteString(out, fmt.Sprintf("\t%s	%s	`json:\"%s,omitempty\"`\n", name, prop.Type, name))
}

func generateForOperation(path string, op swagger.Operation, out io.Writer) {
	io.WriteString(out, "func (c Service) ")
	io.WriteString(out, op.Nickname)
	io.WriteString(out, "(")
	for i, each := range op.Parameters {
		if i > 0 {
			io.WriteString(out, ", ")
		}
		writeParameterSignature(each, out)
	}
	io.WriteString(out, ") ")
	if "void" == op.Type {
		io.WriteString(out, "(interface{}, error)")
	} else {
		io.WriteString(out, "(*"+noPkg(op.Type)+", error)")
	}
	io.WriteString(out, " {\n")
	io.WriteString(out, newuri_source(path))
	for _, each := range op.Parameters {
		writeSetUriParameter(each, out)
	}
	io.WriteString(out, createrequest_source(op.HttpMethod))
	io.WriteString(out, accept_source(toCommaSeparated(op.Consumes)))
	io.WriteString(out, dorequest_source())
	if "void" != op.Type {
		io.WriteString(out, decoderesponse_source(noPkg(op.Type)))
	} else {
		io.WriteString(out, "	return nil,nil")
	}
	io.WriteString(out, "\n}\n")
}

func writeSetUriParameter(param swagger.Parameter, out io.Writer) {
	if "path" == param.ParamType {
		io.WriteString(out, fmt.Sprintf("\turi.pathParam(\"%s\",%s)\n", param.Name, toVar(param.Name)))
	} else if "query" == param.ParamType {
		io.WriteString(out, fmt.Sprintf("\turi.queryParam(\"%s\",%s)\n", param.Name, toVar(param.Name)))
	}
}

func writeParameterSignature(param swagger.Parameter, out io.Writer) {
	fmt.Fprintf(out, "%s %s", toVar(noPkg(param.Name)), noPkg(param.Type))
}

func toVar(varName string) string {
	return strings.ToLower(strings.NewReplacer("-", "_").Replace(varName))
}

func toCommaSeparated(names []string) string {
	var buf bytes.Buffer
	for i, each := range names {
		if i > 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(each)
	}
	return buf.String()
}

func noPkg(name string) string {
	dot := strings.LastIndex(name, ".")
	if dot == -1 {
		return name
	} else {
		return name[dot+1:]
	}
}

func fetch(url string, model interface{}) {
	log.Printf("fetching %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Unable to fetch Swagger specification:%v", err)
	}
	defer resp.Body.Close()
	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Unable to read Swagger specification:%v", err)
	}
	json.Unmarshal(buffer, model)
}
