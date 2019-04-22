package main

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/handler"
	"net/http"
	"strconv"
)

func (project *Project) Deploy() {
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

type InternalType struct {
	name        string
	description string
	string      string
	error       error
}

func NewInternalType(name string, description string) InternalType {
	return InternalType{
		name:        name,
		description: description,
		string:      name,
		error:       nil,
	}
}

func (internalType *InternalType) Name() string {
	return internalType.name
}

func (internalType *InternalType) Description() string {
	return internalType.description
}

func (internalType *InternalType) String() string {
	return internalType.string
}

func (internalType *InternalType) Error() error {
	return internalType.error
}

//var _ graphql.Output = (*InternalType)(nil)
//
//func ResolveType(def ast.Type) graphql.Output {
//	switch def.(type) {
//	case *ast.Named:
//		var ttype graphql.Output
//		switch def.(interface{}).(*ast.Named).Name.Value {
//		case "String":
//			ttype = graphql.String
//		default:
//			fmt.Println(def.(interface{}).(*ast.Named).Name.Value)
//			internalType := NewInternalType(def.(interface{}).(*ast.Named).Name.Value, "")
//			ttype = &internalType
//		}
//		return ttype
//	case *ast.NonNull:
//		return graphql.NewNonNull(ResolveType(def.(interface{}).(*ast.NonNull).Type))
//	case *ast.List:
//		return graphql.NewList(ResolveType(def.(interface{}).(*ast.List).Type))
//	default:
//		return graphql.Boolean
//	}
//}
//
//func (project *Project) resolveTypes() map[string]graphql.Output {
//	types := map[string]graphql.Output{}
//	for _, item := range project.parse().Definitions {
//		switch item.(interface{}).(type) {
//		case *ast.ObjectDefinition:
//			def := item.(interface{}).(*ast.ObjectDefinition)
//			fields := graphql.Fields{}
//			for _, field := range def.Fields {
//				args := graphql.FieldConfigArgument{}
//				for _, arg := range field.Arguments {
//					argInternalType := NewInternalType(arg.Type.String(), "")
//					args[arg.Name.Value] = &graphql.ArgumentConfig{
//						Type:         &argInternalType,
//						Description:  arg.Description.Value,
//						DefaultValue: arg.DefaultValue.GetValue(),
//					}
//				}
//				var description string
//				if field.Description == nil {
//					description = ""
//				} else {
//					description = field.Description.Value
//				}
//				fields[field.Name.Value] = &graphql.Field{
//					Name:        field.Name.Value,
//					Description: description,
//					Type:        ResolveType(field.Type),
//					Args:        args,
//					Resolve: func(p graphql.ResolveParams) (i interface{}, e error) {
//						return nil, nil
//					},
//				}
//			}
//			var name string
//			if def.Name.Value == "Query" {
//				name = "RootQuery"
//			} else {
//				name = def.Name.Value
//			}
//			var description string
//			if def.Description == nil {
//				description = ""
//			} else {
//				description = def.Description.Value
//			}
//			types[def.Name.Value] = graphql.NewObject(graphql.ObjectConfig{
//				Name:        name,
//				Description: description,
//				Fields:      fields,
//			})
//			break
//		}
//
//	}
//	return types
//}
//
//func (project *Project) injectTypes() {
//	for _, def := range project.Types {
//		switch def.(type) {
//		case *graphql.Object:
//			for _, field := range def.(*graphql.Object).Fields() {
//				switch field.Type.(type) {
//				case *InternalType:
//					field.Type = project.Types[field.Type.Name()].(graphql.Output)
//				case *graphql.List:
//					var item graphql.Output
//					switch field.Type.(type) {
//					case *InternalType:
//						item = project.Types[field.Type.Name()].(graphql.Output)
//						break
//					case *graphql.NonNull:
//						item = graphql.NewNonNull(project.Types[field.Type.Name()].(graphql.Output))
//						break
//					}
//					field.Type = graphql.NewList(item)
//					break
//				case *graphql.NonNull:
//					field.Type = graphql.NewNonNull(project.Types[field.Type.Name()].(graphql.Output))
//				}
//			}
//			break
//		}
//	}
//}

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
						project.Types[name] = graphql.NewObject(graphql.ObjectConfig{
							Name:        name,
							Description: description,
							Fields:      fields,
						})
						for _, field := range def.Fields {
							args := graphql.FieldConfigArgument{}
							fields[field.Name.Value] = &graphql.Field{
								Name:        field.Name.Value,
								Description: ReturnDescriptionOrNull(field),
								Args:        args,
								Type:        project.resolveType(field.Type, ""),
								Resolve: func(p graphql.ResolveParams) (i interface{}, e error) {
									return nil, nil
								},
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
						project.Types[name] = graphql.NewInputObject(graphql.InputObjectConfig{
							Name:        name,
							Description: description,
							Fields:      values,
						})
						for _, value := range def.Values {
							values[value.Name.Value] = &graphql.EnumValueConfig{
								Description: ReturnDescriptionOrNull(value),
							}
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
		case *ast.EnumDefinition:
			values := map[string]*graphql.EnumValueConfig{}
			for _, v := range def.(interface{}).(*ast.EnumDefinition).Values {
				values[v.Name.Value] = &graphql.EnumValueConfig{
					Value: v.Name.Value,
				}
			}
			project.Types[def.(interface{}).(*ast.EnumDefinition).Name.Value] = graphql.NewEnum(graphql.EnumConfig{
				Name:   def.(interface{}).(*ast.EnumDefinition).Name.Value,
				Values: values,
			})
		}
	}
	//for _, def := range project.Definitions.Definitions {
	//	switch def.(type) {
	//	case *ast.UnionDefinition:
	//		types := []*graphql.Object{}
	//		project.Types[def.(interface{}).(*ast.UnionDefinition).Name.Value] = graphql.NewUnion(graphql.UnionConfig{
	//			Name: def.(interface{}).(*ast.UnionDefinition).Name.Value,
	//			ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
	//				for _, v := range types {
	//					if v.Name() == p.Value.(interface{}).(*graphql.Object).Name() {
	//						return v
	//					}
	//				}
	//				return nil
	//			},
	//			Types: types,
	//		})
	//		for _, v := range def.(interface{}).(*ast.UnionDefinition).Types {
	//			item := project.resolveType(nil, v.Name.Value)
	//			types = append(types, item.(interface{}).(*graphql.Object))
	//		}
	//	case *ast.InterfaceDefinition:
	//		project.Types[def.(interface{}).(*ast.InterfaceDefinition).Name.Value] = graphql.NewInterface(graphql.InterfaceConfig{
	//			Name:   def.(interface{}).(*ast.InterfaceDefinition).Name.Value,
	//			Fields: graphql.Fields{},
	//		})
	//		for _, v := range def.(interface{}).(*ast.InterfaceDefinition).Fields {
	//			project.Types[def.(interface{}).(*ast.InterfaceDefinition).Name.Value].(interface{}).(*graphql.Interface).AddFieldConfig(v.Name.Value, &graphql.Field{
	//				Name: v.Name.Value,
	//				Type: project.resolveType(v.Type, v.Type.String()),
	//			})
	//		}
	//	case *ast.InputObjectDefinition:
	//		project.Types[def.(interface{}).(*ast.InputObjectDefinition).Name.Value] = graphql.NewInputObject(graphql.InputObjectConfig{
	//			Name:   def.(interface{}).(*ast.InputObjectDefinition).Name.Value,
	//			Fields: graphql.InputObjectConfigFieldMap{},
	//		})
	//		for _, v := range def.(interface{}).(*ast.InputObjectDefinition).Fields {
	//			var defaultValue interface{}
	//			var description string
	//			if v.DefaultValue != nil {
	//				defaultValue = v.DefaultValue.GetValue()
	//			}
	//			if v.Description != nil {
	//				description = v.Description.Value
	//			}
	//			project.Types[def.(interface{}).(*ast.InputObjectDefinition).Name.Value].(interface{}).(*graphql.InputObject).AddFieldConfig(v.Name.Value, &graphql.InputObjectFieldConfig{
	//				Type:         project.resolveType(v.Type, v.Type.String()),
	//				DefaultValue: defaultValue,
	//				Description:  description,
	//			})
	//		}
	//	}
	//}
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
	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
	if err != nil {
		fmt.Println(err)
	}
	return &schema
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
