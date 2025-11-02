{{/*
Expand the name of the chart.
*/}}
{{- define "go-mcp-gateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "go-mcp-gateway.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "go-mcp-gateway.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "go-mcp-gateway.labels" -}}
helm.sh/chart: {{ include "go-mcp-gateway.chart" . }}
{{ include "go-mcp-gateway.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "go-mcp-gateway.selectorLabels" -}}
app.kubernetes.io/name: {{ include "go-mcp-gateway.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "go-mcp-gateway.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "go-mcp-gateway.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Redis host
*/}}
{{- define "go-mcp-gateway.redisHost" -}}
{{- if .Values.redis.external.enabled }}
{{- .Values.redis.external.host }}
{{- else }}
{{- printf "%s-redis" (include "go-mcp-gateway.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Redis port
*/}}
{{- define "go-mcp-gateway.redisPort" -}}
{{- if .Values.redis.external.enabled }}
{{- .Values.redis.external.port }}
{{- else }}
{{- print "6379" }}
{{- end }}
{{- end }}

{{/*
Redis password secret name
*/}}
{{- define "go-mcp-gateway.redisSecretName" -}}
{{- if .Values.redis.external.enabled }}
{{- printf "%s-redis-external" (include "go-mcp-gateway.fullname" .) }}
{{- else }}
{{- printf "%s-redis" (include "go-mcp-gateway.fullname" .) }}
{{- end }}
{{- end }}
