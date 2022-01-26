| Name | Type | Description | 
| ---- | ---- | ----------- |  {{range $code, $resp := .Definition.Properties}}  {{if eq $resp.Type  "array" }}   {{if eq $resp.Items.Ref  "" }} 
| {{$code}} | Array[ {{FilterSchema $resp.Items.Type}} ] | {{$resp.Description}} | {{else}}  
| {{$code}} | Array[{{FilterSchema $resp.Items.Ref}}] | {{$resp.Description}} {{template "schema_description.md" $resp.Items.Ref}} | {{end}}  {{else if eq $resp.Type  "object"}}
| {{$code}} | Object | {{$resp.Description}} {{template "schema_description.md" $resp.Ref}}  |  {{else}} 
| {{$code}} | {{$resp.Type}} | {{$resp.Description}} |  {{end}} {{end}}

{{$definitions := .Definitions}}
{{- range $code, $resp := .Definition.Properties -}}  
    {{- if eq $resp.Type  "array" -}}   
        {{- if ne $resp.Items.Ref  "" -}}
            {{- $nextRefName := (FilterSchema $resp.Items.Ref) -}}
            {{- if ne $nextRefName $.TopRef -}}
### {{$nextRefName}}
{{template "schema.md" CollectSchema $definitions  $resp.Items.Ref}}
            {{- end -}}
        {{- end -}}  
    {{- else if eq $resp.Type  "object" -}}
        {{- if ne $resp.Ref  ""  -}}
            {{- $nextRefName := (FilterSchema $resp.Items.Ref) -}}
            {{- if ne $nextRefName $.TopRef -}}
### {{$nextRefName}}
{{template "schema.md" CollectSchema $definitions  $resp.Ref}}
            {{- end -}}
        {{- end -}}  
    {{- end -}} 
 {{- end -}}

 
 

