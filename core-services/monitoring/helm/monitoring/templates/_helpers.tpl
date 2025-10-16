{{/*
Expand the name of the chart.
*/}}
{{- define "monitoring.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "monitoring.fullname" -}}
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
{{- define "monitoring.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "monitoring.labels" -}}
helm.sh/chart: {{ include "monitoring.chart" . }}
{{ include "monitoring.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "monitoring.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "monitoring.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database host
*/}}
{{- define "monitoring.databaseHost" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "monitoring.fullname" .) }}
{{- else }}
{{- .Values.database.host }}
{{- end }}
{{- end }}

{{/*
Redis host
*/}}
{{- define "monitoring.redisHost" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master" (include "monitoring.fullname" .) }}
{{- else }}
{{- .Values.redis.host }}
{{- end }}
{{- end }}

{{/*
Create a default fully qualified postgresql name.
*/}}
{{- define "monitoring.postgresql.fullname" -}}
{{- printf "%s-postgresql" (include "monitoring.fullname" .) }}
{{- end }}

{{/*
Create a default fully qualified redis name.
*/}}
{{- define "monitoring.redis.fullname" -}}
{{- printf "%s-redis" (include "monitoring.fullname" .) }}
{{- end }}

{{/*
Return the proper image name
*/}}
{{- define "monitoring.image" -}}
{{- $registryName := .Values.image.registry -}}
{{- $repositoryName := .Values.image.repository -}}
{{- $tag := .Values.image.tag | toString -}}
{{- if .Values.global }}
    {{- if .Values.global.imageRegistry }}
        {{- $registryName = .Values.global.imageRegistry -}}
    {{- end -}}
{{- end -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else }}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end }}
{{- end }}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "monitoring.imagePullSecrets" -}}
{{- include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global) -}}
{{- end }}

{{/*
Compile all warnings into a single message.
*/}}
{{- define "monitoring.validateValues" -}}
{{- $messages := list -}}
{{- $messages := append $messages (include "monitoring.validateValues.database" .) -}}
{{- $messages := append $messages (include "monitoring.validateValues.redis" .) -}}
{{- $messages := without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message -}}
{{- end -}}
{{- end -}}

{{/*
Validate database configuration
*/}}
{{- define "monitoring.validateValues.database" -}}
{{- if and (not .Values.postgresql.enabled) (not .Values.database.host) -}}
monitoring: database.host
    You must provide a database host when postgresql is disabled.
    Please set database.host or enable postgresql.
{{- end -}}
{{- end -}}

{{/*
Validate redis configuration
*/}}
{{- define "monitoring.validateValues.redis" -}}
{{- if and (not .Values.redis.enabled) (not .Values.redis.host) -}}
monitoring: redis.host
    You must provide a redis host when redis is disabled.
    Please set redis.host or enable redis.
{{- end -}}
{{- end -}}

{{/*
Return the secret name for database password
*/}}
{{- define "monitoring.databaseSecretName" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "monitoring.fullname" .) }}
{{- else }}
{{- printf "%s-db" (include "monitoring.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Return the secret key for database password
*/}}
{{- define "monitoring.databaseSecretKey" -}}
{{- if .Values.postgresql.enabled }}
{{- print "postgres-password" }}
{{- else }}
{{- print "password" }}
{{- end }}
{{- end }}

{{/*
Return the secret name for redis password
*/}}
{{- define "monitoring.redisSecretName" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis" (include "monitoring.fullname" .) }}
{{- else }}
{{- printf "%s-redis" (include "monitoring.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Return the secret key for redis password
*/}}
{{- define "monitoring.redisSecretKey" -}}
{{- if .Values.redis.enabled }}
{{- print "redis-password" }}
{{- else }}
{{- print "password" }}
{{- end }}
{{- end }}