{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "readinessPath" -}}
{{- if contains "v0.2.0" .Values.image.tag -}}/{{- else -}}/readyz{{- end -}}
{{- end -}}
