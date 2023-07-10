package vast_graphql

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vision-cli/common/execute"
	md "github.com/vision-cli/common/module"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/common/transpiler/model"
)

func CreateGraphqlResolvers(
	service model.Service,
	projectRoot string,
	moduleName string,
	remoteServiceDirectory string,
	servicePath string,
	projectPath string,
	executor execute.Executor) {
	astGoModels := NewCompletlyBlankAst("resolvers", service)
	astGoModels.addImport("github.com/graph-gophers/graphql-go")
	if hasSearchableEnumField(service.Entities) {
		astGoModels.addImport("google.golang.org/protobuf/types/known/structpb")
	}
	astGoModels.addImport(remoteServiceDirectory + "/config")
	astGoModels.addImport(remoteServiceDirectory + "/server")

	astGoModels.AddImportAs(remoteServiceDirectory+"/proto", "pb")
	astGoModels.addGoTypes(service.Entities, moduleName)
	targetDir := filepath.Join(projectRoot, "services", "default", "graphql", "resolvers")
	goResolversPath := filepath.Join(targetDir, cases.Snake(moduleName)+"_"+cases.Snake(service.Name)+".go")
	fmt.Println("Creating resolvers for " + service.Name + " at " + goResolversPath)

	// write the resolver file
	if err := ioutil.WriteFile(goResolversPath, []byte(astGoModels.String()), WriteMode); err != nil {
		panic(err)
	}

	// write the resolver functions
	for _, e := range service.Entities {
		enums := sortEnums(e, goResolversPath)
		writeFunctionWrappers(goResolversPath, e.Name, moduleName, service.Name, enums)
	}
	// write the resolver struct
	writeResolverStruct(goResolversPath, moduleName, service.Name)

	replacement := filepath.Join("../../", servicePath)

	if err := md.Replace(targetDir, remoteServiceDirectory, replacement, executor); err != nil {
		panic(err)
	}

	if err := md.Replace(targetDir, projectPath+"/libs/go/persistence", "../../../libs/go/persistence", executor); err != nil {
		panic(err)
	}

	// run go mod tidy
	if err := md.Tidy(targetDir, executor); err != nil {
		panic(err)
	}
}

func hasSearchableEnumField(entities []model.Entity) bool {
	for _, e := range entities {
		for _, field := range e.Fields {
			if field.IsSearchable && field.Type == model.TypeEnum {
				return true
			}
		}
	}
	return false
}

func (a *AstGo) addGoTypes(entities []model.Entity, moduleName string) {
	for _, e := range entities {
		// -------- Get -----------
		getRequest := model.Entity{
			Name: cases.Pascal(moduleName) + "Get" + cases.Pascal(e.Name) + "Request",
			Fields: []model.Field{
				{Name: "ID", Type: "id", IsArray: false, IsNullable: false, IsSearchable: false}},
		}
		a.file.Decls = append(a.file.Decls, getRequest.GoAstType(model.GoTargetResolver))

		// Get response is not marked as a response message because its treated as the entity type
		getResponse := model.Entity{
			Name: cases.Pascal(moduleName) + cases.Pascal(e.Name),
			Fields: append(
				[]model.Field{{Name: "ID", Type: "id", IsArray: false, IsNullable: false, IsSearchable: false}},
				e.Fields...,
			),
		}
		a.file.Decls = append(a.file.Decls, getResponse.GoAstType(model.GoTargetResolver))

		// -------- List -----------
		listRequest := model.Entity{
			Name: cases.Pascal(moduleName) + "List" + cases.Pascal(e.Name) + "s" + "Request",
			Fields: []model.Field{
				{Name: "Pagination", Type: "PaginationInput", IsArray: false, IsNullable: true, IsSearchable: false},
				{Name: "SearchBy", Type: model.TypeGoString, IsArray: false, IsNullable: true, IsSearchable: false},
			},
		}
		listFilterInput := a.writeFilterInput(e)

		if len(listFilterInput.Fields) > 0 {
			listRequest.Fields = append(listRequest.Fields, model.Field{Name: "FilterBy", Type: "Filter" + cases.Pascal(e.Name), IsArray: false, IsNullable: true, IsSearchable: false})
			a.file.Decls = append(a.file.Decls, listFilterInput.GoAstType(model.GoTargetModel))
		}
		a.file.Decls = append(a.file.Decls, listRequest.GoAstType(model.GoTargetResolver))

		listResponse := model.Entity{
			Name: "Multi" + cases.Pascal(moduleName) + cases.Pascal(e.Name),
			Fields: []model.Field{
				{Name: cases.Camel(moduleName) + cases.Pascal(e.Name) + "s", Type: cases.Pascal(moduleName) + cases.Pascal(e.Name), IsArray: true, IsNullable: true, IsSearchable: false},
				{Name: "pagination", Type: "Pagination", IsArray: false, IsNullable: true, IsSearchable: false},
			},
		}
		a.file.Decls = append(a.file.Decls, listResponse.GoAstType(model.GoTargetResolver))

		// -------- Create -----------
		createRequest := model.Entity{
			Name:   cases.Pascal(moduleName) + "Create" + cases.Pascal(e.Name) + "Request",
			Fields: e.Fields,
		}
		a.file.Decls = append(a.file.Decls, createRequest.GoAstType(model.GoTargetResolver))

		createResponse := model.Entity{
			Name: cases.Pascal(moduleName) + "Create" + cases.Pascal(e.Name) + "Response",
			Fields: []model.Field{
				{Name: "ID", Type: "id", IsArray: false, IsNullable: false, IsSearchable: false}},
		}
		a.file.Decls = append(a.file.Decls, createResponse.GoAstType(model.GoTargetResolver))

		// -------- Update -----------
		updateRequest := model.Entity{
			Name: cases.Pascal(moduleName) + "Update" + cases.Pascal(e.Name) + "Request",
			Fields: append(
				[]model.Field{{Name: "ID", Type: "id", IsArray: false, IsNullable: false, IsSearchable: false}},
				e.Fields...,
			),
		}
		a.file.Decls = append(a.file.Decls, updateRequest.GoAstType(model.GoTargetResolver))

		updateResponse := model.Entity{
			Name: cases.Pascal(moduleName) + "Update" + cases.Pascal(e.Name) + "Response",
			Fields: []model.Field{
				{Name: "message", Type: "string", IsArray: false, IsNullable: false, IsSearchable: false}},
		}
		a.file.Decls = append(a.file.Decls, updateResponse.GoAstType(model.GoTargetResolver))

		// -------- Delete -----------
		deleteRequest := model.Entity{
			Name: cases.Pascal(moduleName) + "Delete" + cases.Pascal(e.Name) + "Request",
			Fields: []model.Field{
				{Name: "ID", Type: "id", IsArray: false, IsNullable: false, IsSearchable: false}},
		}
		a.file.Decls = append(a.file.Decls, deleteRequest.GoAstType(model.GoTargetResolver))

		deleteResponse := model.Entity{
			Name: cases.Pascal(moduleName) + "Delete" + cases.Pascal(e.Name) + "Response",
			Fields: []model.Field{
				{Name: "message", Type: "string", IsArray: false, IsNullable: false, IsSearchable: false}},
		}
		a.file.Decls = append(a.file.Decls, deleteResponse.GoAstType(model.GoTargetResolver))
	}
}

func writeFunctionWrappers(goResolversPath, entityName, moduleName, serviceName string, enums []string) {
	writeFunctionWrapper(goResolversPath, entityName, moduleName, serviceName, "Get", enums)
	writeFunctionWrapper(goResolversPath, entityName, moduleName, serviceName, "List", enums)
	writeFunctionWrapper(goResolversPath, entityName, moduleName, serviceName, "Create", nil)
	writeFunctionWrapper(goResolversPath, entityName, moduleName, serviceName, "Update", nil)
	writeFunctionWrapper(goResolversPath, entityName, moduleName, serviceName, "Delete", nil)
}

func writeFunctionWrapper(goResolversPath, entityName, moduleName, serviceName, verb string, enums []string) {
	f, err := os.OpenFile(goResolversPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, WriteMode)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	codeStr, err := verbFunctionWrapper(entityName, moduleName, serviceName, verb, enums)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(codeStr); err != nil {
		panic(err)
	}
}

func verbFunctionWrapper(entityName, moduleName, serviceName, verb string, enums []string) (string, error) {
	type Tokens struct {
		Entity                 string
		EntityPlural           string
		EnumListChecks         string
		EnumOutputChecks       string
		Resolver               string
		Function               string
		GoRequest              string
		GoResponse             string
		PbRequest              string
		PbResponse             string
		PbResponseStruct       string
		PbPluralResponseStruct string
		PbFunction             string
	}
	var tokens Tokens
	if verb == "Get" {
		enumOutputChecks := generateEnumOutputConversions(enums)
		tokens = Tokens{
			Resolver:         cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver",
			Function:         cases.Pascal(moduleName) + verb + cases.Pascal(entityName),
			GoRequest:        cases.Pascal(moduleName) + verb + cases.Pascal(entityName) + "Request",
			GoResponse:       cases.Pascal(moduleName) + cases.Pascal(entityName),
			PbRequest:        "pb." + verb + cases.Pascal(entityName) + "Request",
			PbResponse:       "pb." + verb + cases.Pascal(entityName) + "Response",
			EnumOutputChecks: enumOutputChecks,
			PbFunction:       "r.srv." + verb + cases.Pascal(entityName),
		}
	} else if verb == "List" {
		listEnumChecks := generateListEnumConversions(enums, entityName)
		enumOutputChecks := generateEnumOutputConversions(enums)
		tokens = Tokens{
			Entity:                 cases.Pascal(entityName),
			EntityPlural:           cases.Pascal(entityName) + "s",
			EnumListChecks:         listEnumChecks,
			Resolver:               cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver",
			Function:               cases.Pascal(moduleName) + verb + cases.Pascal(entityName) + "s",
			GoRequest:              cases.Pascal(moduleName) + verb + cases.Pascal(entityName) + "s" + "Request",
			GoResponse:             "Multi" + cases.Pascal(moduleName) + cases.Pascal(entityName),
			EnumOutputChecks:       enumOutputChecks,
			PbRequest:              "pb." + verb + cases.Pascal(entityName) + "s" + "Request",
			PbResponseStruct:       cases.Pascal(moduleName) + cases.Pascal(entityName),
			PbPluralResponseStruct: cases.Pascal(moduleName) + cases.Pascal(entityName) + "s",
			PbFunction:             "r.srv." + verb + cases.Pascal(entityName) + "s",
		}
	} else {
		tokens = Tokens{
			Resolver:   cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver",
			Function:   cases.Pascal(moduleName) + verb + cases.Pascal(entityName),
			GoRequest:  cases.Pascal(moduleName) + verb + cases.Pascal(entityName) + "Request",
			GoResponse: cases.Pascal(moduleName) + verb + cases.Pascal(entityName) + "Response",
			PbRequest:  "pb." + verb + cases.Pascal(entityName) + "Request",
			PbResponse: "pb." + verb + cases.Pascal(entityName) + "Response",
			PbFunction: "r.srv." + verb + cases.Pascal(entityName),
		}
	}

	var codeTmpl string

	switch verb {
	case "Get":
		codeTmpl = getTmpl
	case "List":
		codeTmpl = listTmpl
	case "Create":
		codeTmpl = createTmpl
	case "Update":
		codeTmpl = updateTmpl
	case "Delete":
		codeTmpl = deleteTmpl
	}

	return tmpl.TmplToString(codeTmpl, tokens)
}

func writeResolverStruct(goResolversPath, moduleName, serviceName string) {
	f, err := os.OpenFile(goResolversPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, WriteMode)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	codeStr, err := resolverStruct(moduleName, serviceName)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(codeStr); err != nil {
		panic(err)
	}
}

func resolverStruct(moduleName, serviceName string) (string, error) {
	type Tokens struct {
		Resolver string
	}
	tokens := Tokens{
		Resolver: cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver",
	}

	codeTmpl := `
type {{.Resolver}} struct {
	srv *server.Server
}

func New{{.Resolver}}() {{.Resolver}} {
	conf := config.MustLoadConfig()
	srv := server.MustLoadServer(conf)

	return {{.Resolver}}{
		srv: srv,
	}
}
`
	return tmpl.TmplToString(codeTmpl, tokens)
}

const (
	getTmpl = `
func (r *{{.Resolver}}) {{.Function}}(args {{.GoRequest}}) (*{{.GoResponse}}, error) {
req := PbToGoStruct[{{.GoRequest}}, {{.PbRequest}}](args, false)
result, err := {{.PbFunction}}(nil, &req)
if err != nil {
	return nil, err
}
resp := PbToGoStruct[{{.PbResponse}}, {{.GoResponse}}](*result, false)
{{.EnumOutputChecks}}
return &resp, nil
}
`

	listTmpl = `
func (r *{{.Resolver}}) {{.Function}}(args {{.GoRequest}}) (*{{.GoResponse}}, error) {
req := {{.PbRequest}}{
	SearchBy: *args.SearchBy,
}
if args.Pagination != nil {
	reqPagination := GoStructToPb[PaginationInput, pb.PaginationRequest](*args.Pagination)
	req.Pagination = &reqPagination
}
{{.EnumListChecks}}
result, err := {{.PbFunction}}(nil, &req)
if err != nil {
	return nil, err
}
resp{{.EntityPlural}} := []*{{.PbResponseStruct}}{}
for _, result := range result.{{.EntityPlural}} {
    resp := PbToGoStruct[pb.{{.Entity}}, {{.PbResponseStruct}}](*result, true)
    {{.EnumOutputChecks}}
    resp{{.EntityPlural}} = append(resp{{.EntityPlural}}, &resp)
}
respPagination := PbToGoStruct[pb.PaginationResponse, Pagination](*result.Pagination, true)
return &{{.GoResponse}}{
	{{.PbPluralResponseStruct}}: resp{{.EntityPlural}},
	Pagination:       &respPagination,
}, nil
}
`

	createTmpl = `
func (r *{{.Resolver}}) {{.Function}}(args {{.GoRequest}}) (graphql.ID, error) {
req := GoStructToPb[{{.GoRequest}}, {{.PbRequest}}](args)
result, err := {{.PbFunction}}(nil, &req)
if err != nil {
	return "", err
}
resp := PbToGoStruct[{{.PbResponse}}, {{.GoResponse}}](*result, false)
return resp.ID, nil
}
`

	updateTmpl = `
func (r *{{.Resolver}}) {{.Function}}(args {{.GoRequest}}) (string, error) {
req := GoStructToPb[{{.GoRequest}}, {{.PbRequest}}](args)
result, err := {{.PbFunction}}(nil, &req)
if err != nil {
	return "", err
}
resp := PbToGoStruct[{{.PbResponse}}, {{.GoResponse}}](*result, true)
return resp.Message, nil
}
`

	deleteTmpl = `
func (r *{{.Resolver}}) {{.Function}}(args {{.GoRequest}}) (string, error) {
req := GoStructToPb[{{.GoRequest}}, {{.PbRequest}}](args)
result, err := {{.PbFunction}}(nil, &req)
if err != nil {
	return "", err
}
resp := PbToGoStruct[{{.PbResponse}}, {{.GoResponse}}](*result, true)
return resp.Message, nil
}
`
)

func (a *AstGo) writeFilterInput(entity model.Entity) *model.Entity {
	filterFields := []model.Field{}
	for _, field := range entity.Fields {
		if field.IsSearchable && (field.Type == model.TypeBoolean || field.Type == model.TypeEnum || field.Type == model.TypeId) {
			filterFields = append(filterFields, model.Field{
				Name: field.Name,
				Type: field.FilterGoType(model.GoTargetResolver),
			})
			if field.Type == model.TypeEnum {
				enumEntity := model.Entity{
					Name: field.Name,
					Fields: []model.Field{
						{
							Name: "Nullable",
							Type: "*" + model.TypeGoString,
						},
						{
							Name: field.Name,
							Type: "*" + model.TypeGoString,
						},
					},
				}
				a.file.Decls = append(a.file.Decls, enumEntity.GoAstType(model.GoTargetModel))
			}
		}
	}
	filterEntity := model.Entity{
		Name:        "Filter" + cases.Pascal(entity.Name),
		Persistence: entity.Persistence,
		Fields:      filterFields,
	}
	return &filterEntity
}

func sortEnums(e model.Entity, goResolversPath string) []string {
	f, err := os.OpenFile(goResolversPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, WriteMode)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var enums []string
	for _, field := range e.Fields {
		if field.IsSearchable && field.Type == model.TypeEnum {
			funcStr, err := buildEnumPbFunc(field.Name)
			if err != nil {
				panic(err)
			}
			if _, err = f.WriteString(funcStr); err != nil {
				panic(err)
			}
			enums = append(enums, field.Name)
		}
	}
	return enums
}

func generateListEnumConversions(enums []string, entityName string) string {
	var listChecks string
	type Tokens struct {
		EnumName string
	}
	listCheckTmpl := `
    if args.FilterBy.{{.EnumName}} != nil {
        req{{.EnumName}} := build{{.EnumName}}NullableStruct(*args.FilterBy.{{.EnumName}})
        reqFilter.{{.EnumName}} = req{{.EnumName}}
    }`
	for _, enum := range enums {
		tokens := Tokens{
			EnumName: cases.Pascal(enum),
		}
		listCheck, err := tmpl.TmplToString(listCheckTmpl, tokens)
		if err != nil {
			panic(err)
		}
		listChecks += listCheck
	}
	if listChecks == "" {
		return listChecks
	}
	filterName := "Filter" + cases.Pascal(entityName)
	return fmt.Sprintf(
		`if args.FilterBy != nil {
			reqFilter := GoStructToPb[%s, pb.%s](*args.FilterBy)
			req.FilterBy = &reqFilter
			%s
}`, filterName, filterName, listChecks)
}

func generateEnumOutputConversions(enums []string) string {
	var enumOutputChecks string
	type Tokens struct {
		EnumName string
	}
	fixEnumOutput := `
    resp.{{.EnumName}} = result.{{.EnumName}}.String()`
	for _, enum := range enums {
		tokens := Tokens{
			EnumName: cases.Pascal(enum),
		}
		enumOutputCheck, err := tmpl.TmplToString(fixEnumOutput, tokens)
		if err != nil {
			panic(err)
		}
		enumOutputChecks += enumOutputCheck
	}
	return enumOutputChecks
}

func buildEnumPbFunc(enumName string) (string, error) {
	type Tokens struct {
		EnumName string
	}
	tokens := Tokens{
		EnumName: cases.Pascal(enumName),
	}

	funcTmpl := `
func build{{.EnumName}}NullableStruct(goStruct {{.EnumName}}) *pb.Nullable{{.EnumName}} {
    if goStruct.{{.EnumName}} != nil {
        return &pb.Nullable{{.EnumName}}{
            Kind: &pb.Nullable{{.EnumName}}_{{.EnumName}}{
                {{.EnumName}}: pb.{{.EnumName}}(pb.{{.EnumName}}_value[*goStruct.{{.EnumName}}]),
            },
        }
    }
    return &pb.Nullable{{.EnumName}}{
        Kind: &pb.Nullable{{.EnumName}}_Null{Null: structpb.NullValue_NULL_VALUE},
    }
}
`
	return tmpl.TmplToString(funcTmpl, tokens)
}
