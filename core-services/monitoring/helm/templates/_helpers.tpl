{{/*
太上老君监控系统 Helm Chart 辅助模板
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "taishanglaojun-monitoring.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "taishanglaojun-monitoring.fullname" -}}
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
{{- define "taishanglaojun-monitoring.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "taishanglaojun-monitoring.labels" -}}
helm.sh/chart: {{ include "taishanglaojun-monitoring.chart" . }}
{{ include "taishanglaojun-monitoring.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: taishanglaojun
{{- with .Values.global.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "taishanglaojun-monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "taishanglaojun-monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: monitoring
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "taishanglaojun-monitoring.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "taishanglaojun-monitoring.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the image name
*/}}
{{- define "taishanglaojun-monitoring.image" -}}
{{- $registry := .Values.app.image.registry -}}
{{- $repository := .Values.app.image.repository -}}
{{- $tag := .Values.app.image.tag | toString -}}
{{- $digest := .Values.app.image.digest -}}
{{- if .Values.global.imageRegistry }}
{{- $registry = .Values.global.imageRegistry -}}
{{- end }}
{{- if $digest }}
{{- printf "%s/%s@%s" $registry $repository $digest }}
{{- else }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}
{{- end }}

{{/*
Return the proper Storage Class
*/}}
{{- define "taishanglaojun-monitoring.storageClass" -}}
{{- $storageClass := .Values.persistence.storageClass -}}
{{- if .Values.global.storageClass -}}
{{- $storageClass = .Values.global.storageClass -}}
{{- end -}}
{{- if $storageClass -}}
{{- if (eq "-" $storageClass) -}}
{{- printf "storageClassName: \"\"" -}}
{{- else }}
{{- printf "storageClassName: %s" $storageClass -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return true if cert-manager required annotations for TLS signed certificates are set in the Ingress annotations
Ref: https://cert-manager.io/docs/usage/ingress/#supported-annotations
*/}}
{{- define "taishanglaojun-monitoring.ingress.certManagerRequest" -}}
{{ if or (hasKey . "cert-manager.io/cluster-issuer") (hasKey . "cert-manager.io/issuer") }}
    {{- true -}}
{{ end }}
{{- end }}

{{/*
Compile all warnings into a single message.
*/}}
{{- define "taishanglaojun-monitoring.validateValues" -}}
{{- $messages := list -}}
{{- $messages := append $messages (include "taishanglaojun-monitoring.validateValues.persistence" .) -}}
{{- $messages := append $messages (include "taishanglaojun-monitoring.validateValues.resources" .) -}}
{{- $messages := without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message -}}
{{- end -}}
{{- end -}}

{{/*
Validate values of Taishanglaojun Monitoring - Persistence
*/}}
{{- define "taishanglaojun-monitoring.validateValues.persistence" -}}
{{- if and .Values.persistence.enabled (not .Values.persistence.size) -}}
taishanglaojun-monitoring: persistence.size
    A size must be provided when persistence is enabled
{{- end -}}
{{- end -}}

{{/*
Validate values of Taishanglaojun Monitoring - Resources
*/}}
{{- define "taishanglaojun-monitoring.validateValues.resources" -}}
{{- if not .Values.app.resources -}}
taishanglaojun-monitoring: app.resources
    Resource limits and requests should be defined for production deployments
{{- end -}}
{{- end -}}

{{/*
Get the password secret.
*/}}
{{- define "taishanglaojun-monitoring.secretName" -}}
{{- if .Values.auth.existingSecret -}}
{{- printf "%s" .Values.auth.existingSecret -}}
{{- else -}}
{{- printf "%s" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Return true if a secret object should be created
*/}}
{{- define "taishanglaojun-monitoring.createSecret" -}}
{{- if not .Values.auth.existingSecret -}}
{{- true -}}
{{- end -}}
{{- end -}}

{{/*
Get the configuration ConfigMap name.
*/}}
{{- define "taishanglaojun-monitoring.configmapName" -}}
{{- if .Values.existingConfigmap -}}
{{- printf "%s" .Values.existingConfigmap -}}
{{- else -}}
{{- printf "%s-config" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Return true if a configmap object should be created
*/}}
{{- define "taishanglaojun-monitoring.createConfigmap" -}}
{{- if not .Values.existingConfigmap -}}
{{- true -}}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified prometheus name.
*/}}
{{- define "taishanglaojun-monitoring.prometheus.fullname" -}}
{{- printf "%s-prometheus" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}

{{/*
Create a default fully qualified grafana name.
*/}}
{{- define "taishanglaojun-monitoring.grafana.fullname" -}}
{{- printf "%s-grafana" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}

{{/*
Create a default fully qualified influxdb name.
*/}}
{{- define "taishanglaojun-monitoring.influxdb.fullname" -}}
{{- printf "%s-influxdb" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}

{{/*
Create a default fully qualified jaeger name.
*/}}
{{- define "taishanglaojun-monitoring.jaeger.fullname" -}}
{{- printf "%s-jaeger" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}

{{/*
Create a default fully qualified alertmanager name.
*/}}
{{- define "taishanglaojun-monitoring.alertmanager.fullname" -}}
{{- printf "%s-alertmanager" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}

{{/*
Create a default fully qualified node-exporter name.
*/}}
{{- define "taishanglaojun-monitoring.nodeExporter.fullname" -}}
{{- printf "%s-node-exporter" (include "taishanglaojun-monitoring.fullname" .) -}}
{{- end -}}