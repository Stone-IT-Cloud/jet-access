{{/*
Expand the name of the chart.
*/}}
{{- define "vault-busybox.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "vault-busybox.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "vault-busybox.labels" -}}
helm.sh/chart: {{ include "vault-busybox.chart" . }}
{{ include "vault-busybox.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "vault-busybox.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vault-busybox.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}