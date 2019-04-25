package main

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/handler"
	"net/http"
	"strconv"
)

var connections map[uint]*mgo.Database
var directives map[uint]map[string][]*ast.Directive
var fieldDirectives map[uint]map[string]map[string][]*ast.Directive

func (project *Project) Deploy() {
	directives[project.ID] = map[string][]*ast.Directive{}
	fieldDirectives[project.ID] = map[string]map[string][]*ast.Directive{}
	h := handler.New(&handler.Config{
		Schema:     project.compile(),
		Pretty:     true,
		Playground: true,
	})

	fmt.Println("/project-" + strconv.FormatUint(uint64(project.ID), 10))
	http.Handle("/project-"+strconv.FormatUint(uint64(project.ID), 10), h)
	http.ListenAndServe(":8080", nil)
}

func ValidateSchema(schema string) error {
	_, err := parser.Parse(parser.ParseParams{
		Source: schema,
	})
	return err
}

func (project *Project) resolveType(ttype ast.Type, name string) graphql.Output {
	if _, ok := project.Types[name]; ok {
		return project.Types[name]
	} else {
		switch ttype.(type) {
		case *ast.Named:
			switch ttype.(interface{}).(*ast.Named).Name.Value {
			case "String":
				return graphql.String
			case "Float":
				return graphql.Float
			case "ID":
				return graphql.ID
			case "Int":
				return graphql.Int
			case "Boolean":
				return graphql.Boolean
			default:
				return project.resolveType(nil, ttype.(interface{}).(*ast.Named).Name.Value)
			}
		case *ast.List:
			return graphql.NewList(project.resolveType(ttype.(interface{}).(*ast.List).Type, ""))
		case *ast.NonNull:
			return graphql.NewNonNull(project.resolveType(ttype.(interface{}).(*ast.NonNull).Type, ""))
		default:
			for _, item := range project.Definitions.Definitions {
				switch item.(type) {
				case *ast.ObjectDefinition:
					if item.(*ast.ObjectDefinition).Name.Value == name {
						def := item.(*ast.ObjectDefinition)
						if def == nil {
							return nil
						}
						fields := graphql.Fields{}
						var description string
						if def.Description != nil {
							description = def.Description.Value
						}
						directives[project.ID][name] = def.Directives
						fieldDirectives[project.ID][name] = map[string][]*ast.Directive{}
						for _, field := range def.Fields {
							fieldDirectives[project.ID][name][field.Name.Value] = field.Directives
							args := graphql.FieldConfigArgument{}
							resolvFunc := func(p graphql.ResolveParams) (i interface{}, e error) {
								if _, ok := connections[project.ID]; !ok {
									ses, err := mgo.Dial(project.Mongo)
									if err != nil {
										return nil, err
									}
									connections[project.ID] = ses.DB("")
								}
								var target *graphql.Object
								var escape = func(t ast.Type) ast.Type {
									return nil
								}
								escape = func(t ast.Type) ast.Type {
									switch t.String() {
									case "NonNull":
										return escape(t.(interface{}).(*ast.NonNull).Type)
									case "List":
										return escape(t.(interface{}).(*ast.List).Type)
									default:
										return t
									}
								}
								switch v := project.resolveType(escape(field.Type), "").(type) {
								case *graphql.Object:
									target = v
								}
								collection := ""
								if target != nil {
									targetDirectives := directives[project.ID][target.Name()]
									for _, directive := range targetDirectives {
										if directive.Name.Value == "model" {
											for _, arg := range directive.Arguments {
												if arg.Name.Value == "collection" {
													collection = arg.Value.GetValue().(string)
												}
											}
										}
									}
								}
								for _, directive := range fieldDirectives[project.ID][name][p.Info.FieldName] {
									switch directive.Name.Value {
									case "create":
										var doc interface{}
										doc = p.Args
										err := connections[project.ID].C(collection).Insert(&doc)
										if err != nil {
											return nil, err
										}
										return doc, nil
									case "update":
										var doc map[string]interface{}
										err := connections[project.ID].C(collection).FindId(bson.ObjectIdHex(p.Args["id"].(string))).One(&doc)
										if err != nil {
											return nil, err
										}
										for k, v := range p.Args {
											doc[k] = v
										}
										err = connections[project.ID].C(collection).UpdateId(bson.ObjectIdHex(p.Args["id"].(string)), doc)
										if err != nil {
											return nil, err
										}
										return doc, nil
									case "delete":
										err := connections[project.ID].C(collection).RemoveId(bson.ObjectIdHex(p.Args["id"].(string)))
										if err != nil {
											return nil, err
										}
										return nil, nil
									case "all":
										var documents []interface{}
										err := connections[project.ID].C(collection).Find(nil).Iter().All(&documents)
										if err != nil {
											return nil, err
										}
										for i, doc := range documents {
											doc.(bson.M)["id"] = doc.(bson.M)["_id"].(bson.ObjectId).Hex()
											documents[i] = doc
										}
										return documents, nil
									}
								}
								return nil, nil
							}
							switch v := project.resolveType(field.Type, "").(type) {
							case *graphql.Object:
								break
							case *graphql.NonNull:
								switch v.OfType.(type) {
								case *graphql.List:
									break
								default:
									if _, ok := project.Types[v.OfType.Name()]; !ok {
										resolvFunc = nil
									}
								}
								break
							case *graphql.List:
								switch v.OfType.(type) {
								case *graphql.List:
									break
								default:
									if _, ok := project.Types[v.OfType.Name()]; !ok {
										resolvFunc = nil
									}
								}
								break
							default:
								resolvFunc = nil
								break
							}
							fields[field.Name.Value] = &graphql.Field{
								Name:        field.Name.Value,
								Description: ReturnDescriptionOrNull(field),
								Args:        args,
								Type:        project.resolveType(field.Type, ""),
								Resolve:     resolvFunc,
							}
							for _, arg := range field.Arguments {
								var defaultsValue interface{}
								if arg.DefaultValue != nil {
									defaultsValue = arg.DefaultValue.GetValue()
								}
								args[arg.Name.Value] = &graphql.ArgumentConfig{
									Type:         project.resolveType(arg.Type, ""),
									Description:  ReturnDescriptionOrNull(arg),
									DefaultValue: defaultsValue,
								}
							}
						}
						project.Types[name] = graphql.NewObject(graphql.ObjectConfig{
							Name:        name,
							Description: description,
							Fields:      fields,
						})
						directives[project.ID][name] = def.Directives
						return project.Types[name]
					}
				case *ast.InputObjectDefinition:
					if item.(*ast.InputObjectDefinition).Name.Value == name {
						def := item.(*ast.InputObjectDefinition)
						if def == nil {
							return nil
						}
						fields := graphql.InputObjectConfigFieldMap{}
						var description string
						if def.Description != nil {
							description = def.Description.Value
						}
						project.Types[name] = graphql.NewInputObject(graphql.InputObjectConfig{
							Name:        name,
							Description: description,
							Fields:      fields,
						})
						for _, field := range def.Fields {
							fields[field.Name.Value] = &graphql.InputObjectFieldConfig{
								Description: ReturnDescriptionOrNull(field),
								Type:        project.resolveType(field.Type, ""),
							}
						}
						return project.Types[name]
					}
				case *ast.InterfaceDefinition:
					if item.(*ast.InterfaceDefinition).Name.Value == name {
						def := item.(*ast.InterfaceDefinition)
						if def == nil {
							return nil
						}
						fields := graphql.Fields{}
						var description string
						if def.Description != nil {
							description = def.Description.Value
						}
						project.Types[name] = graphql.NewInterface(graphql.InterfaceConfig{
							Name:        name,
							Description: description,
							Fields:      fields,
						})
						for _, field := range def.Fields {
							fields[field.Name.Value] = &graphql.Field{
								Description: ReturnDescriptionOrNull(field),
								Type:        project.resolveType(field.Type, ""),
							}
						}
						return project.Types[name]
					}
				case *ast.EnumDefinition:
					if item.(*ast.EnumDefinition).Name.Value == name {
						def := item.(*ast.EnumDefinition)
						if def == nil {
							return nil
						}
						values := graphql.EnumValueConfigMap{}
						var description string
						if def.Description != nil {
							description = def.Description.Value
						}
						for _, value := range def.Values {
							values[value.Name.Value] = &graphql.EnumValueConfig{
								Description: ReturnDescriptionOrNull(value),
							}
						}
						project.Types[name] = graphql.NewEnum(graphql.EnumConfig{
							Name:        name,
							Description: description,
							Values:      values,
						})
						return project.Types[name]
					}
				case *ast.UnionDefinition:
					if item.(*ast.UnionDefinition).Name.Value == name {
						def := item.(*ast.UnionDefinition)
						if def == nil {
							return nil
						}
						values := []*graphql.Object{}
						var description string
						if def.Description != nil {
							description = def.Description.Value
						}
						project.Types[name] = graphql.NewUnion(graphql.UnionConfig{
							Name:        name,
							Description: description,
							Types:       values,
						})
						for _, t := range def.Types {
							values = append(values, project.resolveType(nil, t.Name.Value).(interface{}).(*graphql.Object))
						}
						return project.Types[name]
					}
				}
			}
		}
	}
	return nil
}

func ReturnDescriptionOrNull(item interface{}) string {
	switch item.(type) {
	case *ast.FieldDefinition:
		if item.(*ast.FieldDefinition).Description != nil {
			return item.(*ast.FieldDefinition).Description.Value
		} else {
			return ""
		}
	case *ast.InputValueDefinition:
		if item.(*ast.InputValueDefinition).Description != nil {
			return item.(*ast.InputValueDefinition).Description.Value
		} else {
			return ""
		}
	}
	return ""
}

func (project *Project) compile() *graphql.Schema {
	project.Definitions = project.parse()
	project.Types = map[string]graphql.Output{}
	for _, def := range project.Definitions.Definitions {
		switch def.(type) {
		case *ast.ScalarDefinition:
			project.Types[def.(interface{}).(*ast.ScalarDefinition).Name.Value] = graphql.NewScalar(graphql.ScalarConfig{
				Name: def.(interface{}).(*ast.ScalarDefinition).Name.Value,
				Serialize: func(value interface{}) interface{} {
					if v, ok := value.(*string); ok {
						if v == nil {
							return nil
						}
						return *v
					}
					return fmt.Sprintf("%v", value)
				},
			})

		}
	}
	queryObject := project.resolveType(nil, "Query")
	mutationObject := project.resolveType(nil, "Mutation")
	var query *graphql.Object
	var mutation *graphql.Object
	if queryObject != nil {
		query = queryObject.(interface{}).(*graphql.Object)
	}
	if mutationObject != nil {
		mutation = mutationObject.(interface{}).(*graphql.Object)
	}
	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation, Directives: project.operate()})
	if err != nil {
		fmt.Println(err)
	}
	return &schema
}

func (project *Project) operate() []*graphql.Directive {
	return []*graphql.Directive{
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "check",
			Args: graphql.FieldConfigArgument{
				"code": &graphql.ArgumentConfig{
					Type:         graphql.String,
					Description:  "code that will be executed on request",
					DefaultValue: "false",
				},
			},
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
				graphql.DirectiveLocationEnumValue,
				graphql.DirectiveLocationInputFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "model",
			Args: graphql.FieldConfigArgument{
				"collection": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "model collection",
				},
			},
			Locations: []string{
				graphql.DirectiveLocationObject,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "unique",
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "spread",
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "create",
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "update",
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "delete",
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "all",
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
		graphql.NewDirective(graphql.DirectiveConfig{
			Name: "paginate",
			Args: graphql.FieldConfigArgument{
				"size": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					Description:  "page size",
					DefaultValue: 25,
				},
			},
			Locations: []string{
				graphql.DirectiveLocationFieldDefinition,
			},
		}),
	}
}

func (project *Project) parse() *ast.Document {
	document, err := parser.Parse(parser.ParseParams{
		Source: project.Schema,
	})
	if err != nil {
		fmt.Println(err)
	}
	return document
}
