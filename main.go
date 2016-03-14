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
var schemas = map[string]*Schema{}

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

func messageSchema(file *descriptor.FileDescriptorProto, message *descriptor.DescriptorProto, path ...int) *Schema {
	if message.GetOptions() != nil && message.GetOptions().GetMapEntry() {
		return &Schema{
			Type:        "object",
			Description: getComment(file, path...),

			AdditionalProperties: fieldSchema(file, message.Field[1], path...),
		}
	}

	ret := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},

		Description: getComment(file, path...),
	}

	for fid, field := range message.Field {
		fieldPath := append(path[:], 2, fid)

		ret.Properties[field.GetName()] = fieldSchema(file, field, fieldPath...)
	}

	return ret
}

func fieldSchema(file *descriptor.FileDescriptorProto, field *descriptor.FieldDescriptorProto, path ...int) *Schema {
	f := &Schema{}

	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		f.Type = "boolean"

	case descriptor.FieldDescriptorProto_TYPE_INT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		f.Type = "integer"

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		f.Type = "integer"

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		f.Type = "number"

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Enumerate the enum types.

		f.Type = "string"

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		f.Type = "string"

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if field.GetTypeName() == "" {
			return nil
		}

		f.Type = "object"
		f.Ref = "#/definitions/" + field.GetTypeName()[1:]
	}

	if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		if !strings.HasSuffix(f.Ref, "Entry") {
			f = &Schema{
				Type:  "array",
				Items: f,
			}
		}
	}

	if f.Ref == "" {
		f.Title = field.GetName()
		f.Description = getComment(file, path...)
	}

	return f
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
		Schemas:  make(map[string]*Schema),
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

		for messageID, message := range file.GetMessageType() {
			typ, sch := file.GetPackage()+"."+message.GetName(), messageSchema(file, message, 2, messageID)
			schemas[typ] = sch

			for msgID, msg := range message.GetNestedType() {
				typ, sch := typ+"."+msg.GetName(), messageSchema(file, msg, 2, messageID, 3, msgID)
				schemas[typ] = sch
			}
		}

		for _, enum := range file.EnumType {
			schema := &Schema{
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
								methodSchema.Verb = "get"
								methodSchema.Path = httpOpts.GetGet()
							} else if httpOpts.GetPost() != "" {
								methodSchema.Verb = "post"
								methodSchema.Path = httpOpts.GetPost()
							} else if httpOpts.GetPut() != "" {
								methodSchema.Verb = "put"
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

		for _, schema := range schemas {
			if schema.Properties != nil {
				for name, prop := range schema.Properties {
					if strings.HasSuffix(prop.Ref, "Entry") && prop.Type == "object" {
						refSchema := schemas[prop.Ref[14:]]
						if refSchema == nil {
							// This will be caught later on.
							continue
						}

						if refSchema.Type == "object" && refSchema.AdditionalProperties != nil {
							schema.Properties[name] = refSchema
						}
					}
				}
			}
		}

		for {
			startLen := len(used)

			for name, _ := range used {
				schema := schemas[name]
				if schema == nil {
					fmt.Fprintf(os.Stderr, "Failed to look-up schema: %q\n", name)
					continue
				}

				if schema.Properties != nil {
					for _, field := range schema.Properties {
						if field.Type == "array" {
							if field.Items.Ref != "" {
								used[field.Items.Ref[14:]] = true
							}
						} else if field.Type == "object" {
							if field.Ref != "" {
								used[field.Ref[14:]] = true
							}

							if field.AdditionalProperties != nil {
								if field.AdditionalProperties.Ref != "" {
									used[field.AdditionalProperties.Ref[14:]] = true
								}
							}
						}
					}
				} else if schema.AdditionalProperties != nil {
					if schema.AdditionalProperties.Ref != "" {
						used[schema.AdditionalProperties.Ref[14:]] = true
					}
				} else if schema.Type == "array" {
					if schema.Ref != "" {
						used[schema.Ref[14:]] = true
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