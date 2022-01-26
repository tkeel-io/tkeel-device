package add

import (
	"bytes"
	"strings"
	"text/template"
)

const protoTemplate = `
{{$srvPath := .Service | toLower}}
syntax = "proto3";

package {{.Package}};

import "google/api/annotations.proto";

option go_package = "{{.GoPackage}}";
option java_multiple_files = true;
option java_package = "{{.JavaPackage}}";

service {{.Service}} {
	rpc Create{{.Service}} (Create{{.Service}}Request) returns (Create{{.Service}}Response) {
		option (google.api.http) = {
			post : "/{{$srvPath}}"
			body : "*"
		};
	};
	rpc Update{{.Service}} (Update{{.Service}}Request) returns (Update{{.Service}}Response) {
		option (google.api.http) = {
			put : "/{{$srvPath}}/{uid}"
			body : "*"
		};
	};
	rpc Delete{{.Service}} (Delete{{.Service}}Request) returns (Delete{{.Service}}Response) {
		option (google.api.http) = {
			delete : "/{{$srvPath}}/{uid}"
		};
	};
	rpc Get{{.Service}} (Get{{.Service}}Request) returns (Get{{.Service}}Response) {
		option (google.api.http) = {
			get : "/{{$srvPath}}/{uid}"
		};
	};
	rpc List{{.Service}} (List{{.Service}}Request) returns (List{{.Service}}Response) {
		option (google.api.http) = {
			get : "/{{$srvPath}}"
		};
	};
}

message {{.Service}}Object {
	string uid = 1;
}

message Create{{.Service}}Request { {{.Service}}Object obj = 1; }
message Create{{.Service}}Response {}

message Update{{.Service}}Request { {{.Service}}Object obj = 1;  string uid = 2; }
message Update{{.Service}}Response {}

message Delete{{.Service}}Request { string uid = 1; }
message Delete{{.Service}}Response { {{.Service}}Object obj = 1; }

message Get{{.Service}}Request { string uid = 1;}
message Get{{.Service}}Response { {{.Service}}Object obj = 1; }

message List{{.Service}}Request {}
message List{{.Service}}Response { repeated {{.Service}}Object objList = 1; }
`

func (p *Proto) execute() ([]byte, error) {
	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("proto").Funcs(funcMap).Parse(strings.TrimSpace(protoTemplate))
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
