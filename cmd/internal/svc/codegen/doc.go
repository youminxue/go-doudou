package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/youminxue/odin/toolkit/astutils"
	"github.com/youminxue/odin/toolkit/constants"
	v3 "github.com/youminxue/odin/toolkit/openapi/v3"
	"github.com/youminxue/odin/toolkit/stringutils"
)

func getSchemaNames(vofile string) []string {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, vofile, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	sc := astutils.NewStructCollector(ExprStringP)
	ast.Walk(sc, root)
	structs := sc.DocFlatEmbed()
	var ret []string
	for _, item := range structs {
		if item.IsExport {
			ret = append(ret, item.Name)
		}
	}
	return ret
}

func schemasOf(vofile string) []v3.Schema {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, vofile, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	sc := astutils.NewStructCollector(ExprStringP)
	ast.Walk(sc, root)
	structs := sc.DocFlatEmbed()
	var ret []v3.Schema
	for _, item := range structs {
		ret = append(ret, v3.NewSchema(item))
	}
	return ret
}

const (
	get    = "GET"
	post   = "POST"
	put    = "PUT"
	delete = "DELETE"
)

func operationOf(method astutils.MethodMeta, httpMethod string) v3.Operation {
	var ret v3.Operation
	var params []v3.Parameter

	ret.Description = strings.Join(method.Comments, "\n")

	// If http method is "POST" and each parameters' type is one of v3.Int, v3.Int64, v3.Bool, v3.String, v3.Float32, v3.Float64,
	// then we use application/x-www-form-urlencoded as Content-type, and we make one ref schema from them as request body.
	var simpleCnt int
	for _, item := range method.Params {
		if v3.IsBuiltin(item) || item.Type == "context.Context" {
			simpleCnt++
		}
	}
	if httpMethod == post && simpleCnt == len(method.Params) {
		ret.RequestBody = postFormUrl(method)
	} else {
		// Simple parameters such as v3.Int, v3.Int64, v3.Bool, v3.String, v3.Float32, v3.Float64 and corresponding Array type
		// will be put into query parameter as url search params no matter what http method is.
		// Complex parameters such as structs in vo and dto package, map and corresponding slice/array type
		// will be put into request body as json content type.
		// File and file array parameter will be put into request body as multipart/form-data content type.
		upload := false
		for _, item := range method.Params {
			if item.Type == "context.Context" {
				continue
			}
			pschemaType := v3.SchemaOf(item)
			if reflect.DeepEqual(pschemaType, v3.FileArray) || pschemaType == v3.File {
				upload = true
				break
			}
		}

		if upload {
			ret.RequestBody = uploadFile(method)
			for _, item := range method.Params {
				if item.Type == "context.Context" {
					continue
				}
				if v3.IsBuiltin(item) && item.IsPathVariable {
					pschema := v3.CopySchema(item)
					v3.RefAddDoc(&pschema, strings.Join(item.Comments, "\n"))
					required := !v3.IsOptional(item.Type)
					param := v3.Parameter{
						Name:        strings.ToLower(strcase.ToLowerCamel(item.Name)),
						In:          v3.InPath,
						Schema:      &pschema,
						Description: pschema.Description,
						Required:    required,
					}
					params = append(params, param)
				}
			}
		} else {
			for _, item := range method.Params {
				if item.Type == "context.Context" {
					continue
				}
				pschema := v3.CopySchema(item)
				v3.RefAddDoc(&pschema, strings.Join(item.Comments, "\n"))
				required := !v3.IsOptional(item.Type)
				if v3.IsBuiltin(item) {
					param := v3.Parameter{
						Name:        strcase.ToLowerCamel(item.Name),
						In:          v3.InQuery,
						Schema:      &pschema,
						Description: pschema.Description,
						Required:    required,
					}
					if item.IsPathVariable {
						param.Name = strings.ToLower(param.Name)
						param.In = v3.InPath
					}
					params = append(params, param)
				} else {
					var content v3.Content
					mt := &v3.MediaType{
						Schema: &pschema,
					}
					reflect.ValueOf(&content).Elem().FieldByName("JSON").Set(reflect.ValueOf(mt))
					ret.RequestBody = &v3.RequestBody{
						Content:  &content,
						Required: required,
					}
				}
			}
		}
	}

	ret.Parameters = params
	ret.Responses = response(method)
	return ret
}

func response(method astutils.MethodMeta) *v3.Responses {
	var respContent v3.Content
	var hasFile bool
	var fileDoc string
	for _, item := range method.Results {
		if item.Type == "*os.File" {
			hasFile = true
			fileDoc = strings.Join(item.Comments, "\n")
			break
		}
	}
	if hasFile {
		respContent.Stream = &v3.MediaType{
			Schema: &v3.Schema{
				Type:        v3.StringT,
				Format:      v3.BinaryF,
				Description: fileDoc,
			},
		}
	} else {
		title := method.Name + "Resp"
		respSchema := v3.Schema{
			Type:       v3.ObjectT,
			Title:      title,
			Properties: make(map[string]*v3.Schema),
		}
		for _, item := range method.Results {
			if item.Type == "error" {
				continue
			}
			key := item.Name
			if stringutils.IsEmpty(key) {
				key = item.Type[strings.LastIndex(item.Type, ".")+1:]
			}
			rschema := v3.CopySchema(item)
			v3.RefAddDoc(&rschema, strings.Join(item.Comments, "\n"))
			prop := strcase.ToLowerCamel(key)
			respSchema.Properties[prop] = &rschema
			if !v3.IsOptional(item.Type) {
				respSchema.Required = append(respSchema.Required, prop)
			}
		}
		v3.Schemas[title] = respSchema
		respContent.JSON = &v3.MediaType{
			Schema: &v3.Schema{
				Ref: "#/components/schemas/" + title,
			},
		}
	}
	return &v3.Responses{
		Resp200: &v3.Response{
			Content: &respContent,
		},
	}
}

func uploadFile(method astutils.MethodMeta) *v3.RequestBody {
	title := method.Name + "Req"
	reqSchema := v3.Schema{
		Type:       v3.ObjectT,
		Title:      title,
		Properties: make(map[string]*v3.Schema),
	}
	for _, item := range method.Params {
		if item.Type == "context.Context" {
			continue
		}
		if item.IsPathVariable {
			continue
		}
		pschemaType := v3.SchemaOf(item)
		if reflect.DeepEqual(pschemaType, v3.FileArray) || pschemaType == v3.File || v3.IsBuiltin(item) {
			pschema := v3.CopySchema(item)
			pschema.Description = strings.Join(item.Comments, "\n")
			prop := strcase.ToLowerCamel(item.Name)
			reqSchema.Properties[prop] = &pschema
			if !v3.IsOptional(item.Type) {
				reqSchema.Required = append(reqSchema.Required, prop)
			}
		}
	}
	v3.Schemas[title] = reqSchema
	mt := &v3.MediaType{
		Schema: &v3.Schema{
			Ref: "#/components/schemas/" + title,
		},
	}
	var content v3.Content
	reflect.ValueOf(&content).Elem().FieldByName("FormData").Set(reflect.ValueOf(mt))
	return &v3.RequestBody{
		Content:  &content,
		Required: len(reqSchema.Required) > 0,
	}
}

func postFormUrl(method astutils.MethodMeta) *v3.RequestBody {
	title := method.Name + "Req"
	reqSchema := v3.Schema{
		Type:       v3.ObjectT,
		Title:      title,
		Properties: make(map[string]*v3.Schema),
	}
	for _, item := range method.Params {
		if item.Type == "context.Context" {
			continue
		}
		pschema := v3.CopySchema(item)
		pschema.Description = strings.Join(item.Comments, "\n")
		prop := strcase.ToLowerCamel(item.Name)
		reqSchema.Properties[prop] = &pschema
		if !v3.IsOptional(item.Type) {
			reqSchema.Required = append(reqSchema.Required, prop)
		}
	}
	v3.Schemas[title] = reqSchema
	mt := &v3.MediaType{
		Schema: &v3.Schema{
			Ref: "#/components/schemas/" + title,
		},
	}
	var content v3.Content
	reflect.ValueOf(&content).Elem().FieldByName("FormURL").Set(reflect.ValueOf(mt))
	return &v3.RequestBody{
		Content:  &content,
		Required: len(reqSchema.Required) > 0,
	}
}

// GetShelves_ShelfBooks_Book
// shelves/:shelf/books/:book
func apiPattern(method string) string {
	ret := astutils.Pattern(method)
	splits := strings.Split(ret, "/")
	var partials []string
	for _, v := range splits {
		if strings.HasPrefix(v, ":") {
			partials = append(partials, "{"+strings.TrimPrefix(v, ":")+"}")
		} else {
			partials = append(partials, v)
		}
	}
	return strings.Join(partials, "/")
}

func pathsOf(ic astutils.InterfaceCollector, routePatternStrategy int) map[string]v3.Path {
	if len(ic.Interfaces) == 0 {
		return nil
	}
	pathmap := make(map[string]v3.Path)
	inter := ic.Interfaces[0]
	for _, method := range inter.Methods {
		endpoint := fmt.Sprintf("/%s", apiPattern(method.Name))
		if routePatternStrategy == 1 {
			endpoint = fmt.Sprintf("/%s/%s", strings.ToLower(inter.Name), noSplitPattern(method.Name))
		}
		hm := httpMethod(method.Name)
		op := operationOf(method, hm)
		if val, ok := pathmap[endpoint]; ok {
			reflect.ValueOf(&val).Elem().FieldByName(strings.Title(strings.ToLower(hm))).Set(reflect.ValueOf(&op))
			pathmap[endpoint] = val
		} else {
			var v3path v3.Path
			reflect.ValueOf(&v3path).Elem().FieldByName(strings.Title(strings.ToLower(hm))).Set(reflect.ValueOf(&op))
			pathmap[endpoint] = v3path
		}
	}
	return pathmap
}

var gofileTmpl = `package {{.SvcPackage}}

import "github.com/youminxue/odin/framework/rest"

func init() {
	rest.Oas = ` + "`" + `{{.Doc}}` + "`" + `
}
`

// GenDoc generates OpenAPI 3.0 description json file.
// Not support alias type in vo or dto file.
func GenDoc(dir string, ic astutils.InterfaceCollector, routePatternStrategy int) {
	var (
		err     error
		svcname string
		docfile string
		gofile  string
		fi      os.FileInfo
		api     v3.API
		data    []byte
		paths   map[string]v3.Path
		tpl     *template.Template
		sqlBuf  bytes.Buffer
		source  string
	)
	svcname = ic.Interfaces[0].Name
	docfile = filepath.Join(dir, strings.ToLower(svcname)+"_openapi3.json")
	fi, err = os.Stat(docfile)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if fi != nil {
		logrus.Warningln("file " + docfile + " will be overwritten")
	}
	gofile = filepath.Join(dir, strings.ToLower(svcname)+"_openapi3.go")
	fi, err = os.Stat(gofile)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if fi != nil {
		logrus.Warningln("file " + gofile + " will be overwritten")
	}
	paths = pathsOf(ic, routePatternStrategy)
	api = v3.API{
		Openapi: "3.0.2",
		Info: &v3.Info{
			Title:       svcname,
			Description: strings.Join(ic.Interfaces[0].Comments, "\n"),
			Version:     fmt.Sprintf("v%s", time.Now().Local().Format(constants.FORMAT10)),
		},
		Servers: []v3.Server{
			{
				URL: fmt.Sprintf("http://localhost:%d", 6060),
			},
		},
		Paths: paths,
		Components: &v3.Components{
			Schemas: v3.Schemas,
		},
	}
	data, err = json.Marshal(api)
	err = ioutil.WriteFile(docfile, data, os.ModePerm)
	if err != nil {
		panic(err)
	}
	if tpl, err = template.New("doc.go.tmpl").Parse(gofileTmpl); err != nil {
		panic(err)
	}
	if err = tpl.Execute(&sqlBuf, struct {
		SvcPackage string
		Doc        string
	}{
		SvcPackage: ic.Package.Name,
		Doc:        string(data),
	}); err != nil {
		panic(err)
	}
	source = strings.TrimSpace(sqlBuf.String())
	astutils.FixImport([]byte(source), gofile)
}

func ParseDto(dir string, dtoDir string) {
	var (
		err        error
		vos        []v3.Schema
		allMethods map[string][]astutils.MethodMeta
		allConsts  map[string][]string
	)
	vodir := filepath.Join(dir, dtoDir)
	if _, err = os.Stat(vodir); os.IsNotExist(err) {
		return
	}
	var files []string
	err = filepath.Walk(vodir, astutils.Visit(&files))
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		v3.SchemaNames = append(v3.SchemaNames, getSchemaNames(file)...)
	}
	allMethods = make(map[string][]astutils.MethodMeta)
	allConsts = make(map[string][]string)
	for _, file := range files {
		sc := astutils.EnumsOf(file, ExprStringP)
		for k, v := range sc.Methods {
			allMethods[k] = append(allMethods[k], v...)
		}
		for k, v := range sc.Consts {
			allConsts[k] = append(allConsts[k], v...)
		}
	}
	for k, v := range allMethods {
		if v3.IsEnumType(v) {
			v3.Enums[k] = astutils.EnumMeta{
				Name:   k,
				Values: allConsts[k],
			}
		}
	}
	for _, file := range files {
		vos = append(vos, schemasOf(file)...)
	}
	for _, item := range vos {
		v3.Schemas[item.Title] = item
	}
}
