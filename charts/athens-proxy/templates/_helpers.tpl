{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "livenessPath" -}}
{{- if contains "v0.2" .Values.image.tag -}}/healthz{{- else -}}/{{- end -}}
{{- end -}}
{{- define "readinessPath" -}}
{{- if contains "v0.2" .Values.image.tag -}}/readyz{{- else -}}/{{- end -}}
{{- end -}}
