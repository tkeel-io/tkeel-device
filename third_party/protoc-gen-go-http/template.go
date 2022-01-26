package main

import (
	"bytes"
	"strings"
	"text/template"
)

var httpTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

type {{.ServiceType}}HTTPHandler struct {
	srv {{.ServiceType}}HTTPServer
}

func new{{.ServiceType}}HTTPHandler(s {{.ServiceType}}HTTPServer) *{{.ServiceType}}HTTPHandler {
	return &{{.ServiceType}}HTTPHandler{srv: s}
}

{{- range .MethodSets}}

func (h *{{$svrType}}HTTPHandler) {{.Name}}(req *go_restful.Request, resp *go_restful.Response) {
	in := {{.Request}}{}

	{{- if .HasBody}}
	if err := transportHTTP.GetBody(req, &in{{.Body}}); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest,err.Error(),nil),"application/json")
		return
	}
	{{- if ne .Body ""}}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest,err.Error(),nil),"application/json")
		return
	}
	{{- end}}
	{{- else}}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest,err.Error(),nil),"application/json")
		return
	}
	{{- end}}
	{{- if .HasVars}}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteHeaderAndJson(http.StatusBadRequest,
			result.Set(http.StatusBadRequest,err.Error(),nil),"application/json")
		return
	}
	{{- end}}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out,err := h.srv.{{.Name}}(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteHeaderAndJson(httpCode,
			result.Set(httpCode, tErr.Message, out), "application/json")
		return
	}
	{{- if .RawDataResponse}}
	resp.WriteHeaderAndJson(http.StatusOK, out, "application/json")
	{{- else}}	
	anyOut, err := anypb.New(out)
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}

	outB, err := protojson.MarshalOptions{
		UseProtoNames: true,
		EmitUnpopulated: true,
	}.Marshal(&result.Http{
		Code: http.StatusOK,
		Msg: "ok",
		Data: anyOut,
	})
	if err != nil {
		resp.WriteHeaderAndJson(http.StatusInternalServerError,
			result.Set(http.StatusInternalServerError, err.Error(), nil), "application/json")
		return
	}
	resp.WriteHeader(http.StatusOK)

	var remain int
	for {
		outB = outB[remain:]
		remain, err = resp.Write(outB)
		if err != nil {
			return
		}
		if remain == 0 {
			break
		}
	}
	{{- end}}
}

{{- end}}

func Register{{.ServiceType}}HTTPServer(container *go_restful.Container, srv {{.ServiceType}}HTTPServer) {
	var ws *go_restful.WebService
	for _, v := range container.RegisteredWebServices() {
		if v.RootPath() == "/{{.ServiceVersion}}" {
			ws = v
			break
		}
	}
	if ws == nil {
		ws = new(go_restful.WebService)
		ws.ApiVersion("/{{.ServiceVersion}}")
		ws.Path("/{{.ServiceVersion}}").Produces(go_restful.MIME_JSON)
		container.Add(ws)
	}

	handler := new{{.ServiceType}}HTTPHandler(srv)

	{{- range .Methods}}
	ws.Route(ws.{{.Method}}("{{.Path}}").
		To(handler.{{.Name}}))
	{{- end}}
}
`

type serviceDesc struct {
	ServiceType    string // Greeter
	ServiceName    string // helloworld.Greeter
	Metadata       string // api/helloworld/helloworld.proto
	ServiceVersion string // v1
	Methods        []*methodDesc
	MethodSets     map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name    string
	Num     int
	Request string
	Reply   string
	// http_rule
	Path                  string
	Method                string
	HasVars               bool
	HasBody               bool
	Body                  string
	BodyFieldName         string
	ResponseBody          string
	ResponseBodyFieldName string
	// comment
	RawDataResponse bool
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
