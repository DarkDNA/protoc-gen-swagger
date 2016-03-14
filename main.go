package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/DarkDNA/protoc-gen-swagger/protobuf/darkdna/api"
	"github.com/DarkDNA/protoc-gen-swagger/protobuf/google/api"
)

var used = map[string]bool{}
var schemas = map[string]Schema{}

func makeMethod(output Document, comment, url string, svc *descriptor.ServiceDescriptorProto, opts *descriptor.MethodDescriptorProto) Operation {
	parts := strings.SplitN(comment, "\n", 2)
	op := Operation{
		ID:          *opts.Name,
		Summary:     parts[0],
		Description: parts[1],
		Results:     make(map[string]Response),
		Parameters:  []Parameter{},
	}

	for k, v := range schemas[(*opts.InputType)[1:]].Properties {
		p := Parameter{
			In:          QueryPos,
			Name:        k,
			Type:        v.Type,
			Format:      v.Format,
			Description: v.Description,
		}

		if v.Ref != "" {
			refV := schemas[v.Ref[14:len(v.Ref)]]

			if len(refV.EnumValues) > 0 {
				p.EnumValues = refV.EnumValues
			}
		}

		if strings.Contains(url, "{"+k+"}") {
			p.In = PathPos
			p.Required = true
		}

		op.Parameters = append(op.Parameters, p)
	}
	used[(*opts.OutputType)[1:]] = true

	if proto.HasExtension(opts.GetOptions(), api.E_Tags) {
		tmp, _ := proto.GetExtension(opts.GetOptions(), api.E_Tags)
		op.Tags = tmp.([]string)
	}

	defResult := Response{
		Schema: Schema{
			Ref: "#/definitions/" + (*opts.OutputType)[1:],
		},

		Description: op.Summary,
	}

	if proto.HasExtension(opts.GetOptions(), api.E_Codes) {
		tmp, _ := proto.GetExtension(opts.GetOptions(), api.E_Codes)
		ext := tmp.(*api.CodesRule)

		if ext.Okay != "" {
			defResult.Description = ext.Okay
		}

		if ext.NotFound != "" {
			used["darkdna.api.Error"] = true

			op.Results["404"] = Response{
				Schema: Schema{
					Ref: "#/definitions/darkdna.api.Error",
				},

				Description: ext.NotFound,
			}
		}
	}
	op.Tags = append(op.Tags, *svc.Name)

	op.Results["200"] = defResult

	op.Results["default"] = Response{
		Description: "All other status codes.",
		Schema: Schema{
			Ref: "#/definitions/darkdna.api.Error",
		},
	}

	return op
}

func makeSchema(comment string, opts *descriptor.FieldDescriptorProto) Schema {
	field := Schema{
		Description: comment,
	}

	switch opts.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		field.Type = "bool"

	case descriptor.FieldDescriptorProto_TYPE_INT32:
		field.Type = "integer"
		field.Format = "int32"

	case descriptor.FieldDescriptorProto_TYPE_INT64:
		field.Type = "integer"
		field.Format = "int64"

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		field.Ref = "#/definitions/" + opts.GetTypeName()[1:]

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		field.Type = "string"
		field.Ref = "#/definitions/" + opts.GetTypeName()[1:]

	default:
		field.Type = "string"
	}

	if opts.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		field.Description = ""

		return Schema{
			Description: comment,
			Type:        "array",
			Items:       &field,
		}
	}

	return field
}

func getComment(file *descriptor.FileDescriptorProto, path ...int) string {
	for _, loc := range file.GetSourceCodeInfo().GetLocation() {
		if loc.LeadingComments == nil {
			continue
		}

		if len(loc.Path) != len(path) {
			continue
		}

		found := true
		for i := 0; i < len(path); i++ {
			if int32(path[i]) != loc.Path[i] {
				found = false

				break
			}
		}

		if found {
			return *loc.LeadingComments
		}
	}

	return ""
}

func main() {
	msg := plugin.CodeGeneratorRequest{}
	buff, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err := proto.Unmarshal(buff, &msg); err != nil {
		panic(err)
	}

	ret := &plugin.CodeGeneratorResponse{}
	defer func() {
		buff, _ := proto.Marshal(ret)
		os.Stdout.Write(buff)
	}()

	output := Document{
		Version:  "2.0",
		Schemes:  []string{"http", "https"},
		Produces: []string{"application/json"},
		Consumes: []string{},
		Methods:  make(map[string]Path),
		Schemas:  make(map[string]Schema),
	}

	var services []*descriptor.FileDescriptorProto

	for _, file := range msg.ProtoFile {
		if file.GetOptions() != nil && proto.HasExtension(file.GetOptions(), api.E_Info) {
			tmp, _ := proto.GetExtension(file.GetOptions(), api.E_Info)
			ext, _ := tmp.(*api.ApiRule)

			output.Information.Title = ext.Title
			output.Information.Version = ext.Version

			output.Host = ext.Host
			output.BasePath = ext.BaseUri
		}

		for i, msg := range file.MessageType {
			schema := Schema{
				Type:       "object",
				Properties: make(map[string]Schema),
			}

			for fieldID, field := range msg.Field {
				schema.Properties[*field.Name] = makeSchema(getComment(file, 4, i, 2, fieldID), field)
			}

			schemas[file.GetPackage()+"."+msg.GetName()] = schema
		}

		for _, enum := range file.EnumType {
			schema := Schema{
				Type: "string",
			}

			for _, value := range enum.GetValue() {
				schema.EnumValues = append(schema.EnumValues, value.GetName())
			}

			schemas[file.GetPackage()+"."+enum.GetName()] = schema
		}

		if len(file.Service) > 0 {
			services = append(services, file)
		}
	}

	for _, file := range services {
		for svcID, svc := range file.Service {
			output.Tags = append(output.Tags, Tag{
				Name:        *svc.Name,
				Description: getComment(file, 6, svcID),
			})

			for methID, meth := range svc.Method {
				method := Path{}
				comment := getComment(file, 6, svcID, 2, methID)

				if meth.Options != nil {
					if proto.HasExtension(meth.Options, google_api.E_Http) {
						ext, err := proto.GetExtension(meth.Options, google_api.E_Http)
						if err != nil {
							e := err.Error()
							ret.Error = &e
							ret.File = nil

							return
						}

						var methodSchema struct {
							Verb string
							Path string
						}

						if httpOpts, ok := ext.(*google_api.HttpRule); ok {
							if httpOpts.GetGet() != "" {
								methodSchema.Verb = "GET"
								methodSchema.Path = httpOpts.GetGet()
							} else if httpOpts.GetPost() != "" {
								methodSchema.Verb = "POST"
								methodSchema.Path = httpOpts.GetPost()
							} else if httpOpts.GetPut() != "" {
								methodSchema.Verb = "PUT"
								methodSchema.Path = httpOpts.GetPut()
							} else {
								// TODO: Error?
								fmt.Fprintf(os.Stderr, "Failed to find method type for %#v", meth)
								continue
							}

							if otherMeths, ok := output.Methods[methodSchema.Path]; ok {
								for k, v := range otherMeths {
									method[k] = v
								}
							}

							method[methodSchema.Verb] = makeMethod(output, comment, methodSchema.Path, svc, meth)
							output.Methods[methodSchema.Path] = method
						}
					}
				}
			}
		}

		for {
			startLen := len(used)

			for typ, schema := range schemas {
				if isUsed, _ := used[typ]; isUsed {
					for _, v := range schema.Properties {
						if v.Ref != "" {
							used[v.Ref[14:len(v.Ref)]] = true
						}
					}
				}
			}

			if len(used) == startLen {
				break
			}
		}

		for typ, isUsed := range used {
			if isUsed {
				output.Schemas[typ] = schemas[typ]
			}
		}

		fname := strings.Replace(file.GetName(), ".proto", ".pb.json", 1)
		buff, _ := json.Marshal(output)
		data := string(buff)

		ret.File = append(ret.File, &plugin.CodeGeneratorResponse_File{
			Name:    &fname,
			Content: &data,
		})
	}
}
