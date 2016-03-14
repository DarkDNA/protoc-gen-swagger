This is a prototype of a protoc plugin to generate a swagger spec out of the given protobuf, allowing you to export APIs you have otherwise generated using other protobuf-to-http-endpoint systems. (such as [protoc-gen-gokit](https://github.com/AmandaCameron/protoc-gen-gokit).)

The swagger spec is agnostic to the generated HTTP-handling code, and thus can be used with any language / code generator that uses the [Google HTTP API protobuf extensions](https://github.com/googleapis/googleapis/blob/master/google/api/http.proto).

# WARNING:
This is largely untested code, I found and ported the prototype of what we currently use inside DarkDNA to this, and released it as-is as there has been interest in projects such as these. Use at your own risk.

# License
MIT License