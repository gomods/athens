{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "livenessPath" -}}
{{- if eq .Values.image.tag "v0.3.0" -}}/{{- else -}}/healthz{{- end -}}
{{- end -}}
{{- define "readinessPath" -}}
{{- if eq .Values.image.tag "v0.3.0" -}}/{{- else -}}/readyz{{- end -}}
{{- end -}}
