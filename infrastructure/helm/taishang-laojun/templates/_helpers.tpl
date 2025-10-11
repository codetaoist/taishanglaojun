{{/*
Expand the name of the chart.
*/}}
{{- define "taishang-laojun.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "taishang-laojun.fullname" -}}
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
{{- define "taishang-laojun.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "taishang-laojun.labels" -}}
helm.sh/chart: {{ include "taishang-laojun.chart" . }}
{{ include "taishang-laojun.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: taishang-laojun
{{- end }}

{{/*
Selector labels
*/}}
{{- define "taishang-laojun.selectorLabels" -}}
app.kubernetes.io/name: {{ include "taishang-laojun.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "taishang-laojun.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "taishang-laojun.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the namespace to use
*/}}
{{- define "taishang-laojun.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride }}
{{- end }}

{{/*
Create a default fully qualified frontend name.
*/}}
{{- define "taishang-laojun.frontend.fullname" -}}
{{- printf "%s-frontend" (include "taishang-laojun.fullname" .) }}
{{- end }}

{{/*
Create a default fully qualified backend name.
*/}}
{{- define "taishang-laojun.backend.fullname" -}}
{{- printf "%s-backend" (include "taishang-laojun.fullname" .) }}
{{- end }}

{{/*
Create a default fully qualified api gateway name.
*/}}
{{- define "taishang-laojun.apiGateway.fullname" -}}
{{- printf "%s-api-gateway" (include "taishang-laojun.fullname" .) }}
{{- end }}

{{/*
Create a default fully qualified postgresql name.
*/}}
{{- define "taishang-laojun.postgresql.fullname" -}}
{{- printf "%s-postgresql" (include "taishang-laojun.fullname" .) }}
{{- end }}

{{/*
Create a default fully qualified redis name.
*/}}
{{- define "taishang-laojun.redis.fullname" -}}
{{- printf "%s-redis" (include "taishang-laojun.fullname" .) }}
{{- end }}

{{/*
Create the database URL
*/}}
{{- define "taishang-laojun.databaseUrl" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "postgres://%s:%s@%s:5432/%s" .Values.postgresql.auth.username .Values.postgresql.auth.password (include "taishang-laojun.postgresql.fullname" .) .Values.postgresql.auth.database }}
{{- else }}
{{- printf "postgres://%s:%s@%s:5432/%s" .Values.externalServices.database.username .Values.externalServices.database.password .Values.externalServices.database.host .Values.externalServices.database.name }}
{{- end }}
{{- end }}

{{/*
Create the Redis URL
*/}}
{{- define "taishang-laojun.redisUrl" -}}
{{- if .Values.redis.enabled }}
{{- if .Values.redis.auth.enabled }}
{{- printf "redis://:%s@%s:6379" .Values.redis.auth.password (include "taishang-laojun.redis.fullname" .) }}
{{- else }}
{{- printf "redis://%s:6379" (include "taishang-laojun.redis.fullname" .) }}
{{- end }}
{{- else }}
{{- if .Values.externalServices.redis.password }}
{{- printf "redis://:%s@%s:6379" .Values.externalServices.redis.password .Values.externalServices.redis.host }}
{{- else }}
{{- printf "redis://%s:6379" .Values.externalServices.redis.host }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create image pull policy
*/}}
{{- define "taishang-laojun.imagePullPolicy" -}}
{{- if .Values.global.imageRegistry }}
{{- .Values.image.pullPolicy | default "IfNotPresent" }}
{{- else }}
{{- .Values.image.pullPolicy | default "Always" }}
{{- end }}
{{- end }}

{{/*
Create security context
*/}}
{{- define "taishang-laojun.securityContext" -}}
runAsNonRoot: true
runAsUser: 1001
runAsGroup: 1001
fsGroup: 1001
seccompProfile:
  type: RuntimeDefault
{{- end }}

{{/*
Create container security context
*/}}
{{- define "taishang-laojun.containerSecurityContext" -}}
allowPrivilegeEscalation: false
capabilities:
  drop:
  - ALL
readOnlyRootFilesystem: true
runAsNonRoot: true
runAsUser: 1001
runAsGroup: 1001
{{- end }}

{{/*
Create resource limits
*/}}
{{- define "taishang-laojun.resources" -}}
{{- if .resources }}
{{- toYaml .resources }}
{{- else }}
limits:
  cpu: 500m
  memory: 512Mi
requests:
  cpu: 100m
  memory: 128Mi
{{- end }}
{{- end }}

{{/*
Create node selector
*/}}
{{- define "taishang-laojun.nodeSelector" -}}
{{- if .Values.nodeAffinity.enabled }}
{{- toYaml .Values.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms }}
{{- end }}
{{- if .nodeSelector }}
{{- toYaml .nodeSelector }}
{{- end }}
{{- end }}

{{/*
Create tolerations
*/}}
{{- define "taishang-laojun.tolerations" -}}
{{- if .tolerations }}
{{- toYaml .tolerations }}
{{- end }}
{{- end }}

{{/*
Create affinity
*/}}
{{- define "taishang-laojun.affinity" -}}
{{- if .Values.podAntiAffinity.enabled }}
podAntiAffinity:
  {{- toYaml .Values.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution | nindent 2 }}
{{- end }}
{{- if .affinity }}
{{- toYaml .affinity }}
{{- end }}
{{- end }}

{{/*
Create ingress annotations
*/}}
{{- define "taishang-laojun.ingress.annotations" -}}
cert-manager.io/cluster-issuer: "letsencrypt-prod"
nginx.ingress.kubernetes.io/ssl-redirect: "true"
nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
nginx.ingress.kubernetes.io/proxy-body-size: "10m"
nginx.ingress.kubernetes.io/rate-limit: "100"
nginx.ingress.kubernetes.io/rate-limit-window: "1m"
{{- end }}

{{/*
Create monitoring labels
*/}}
{{- define "taishang-laojun.monitoring.labels" -}}
prometheus.io/scrape: "true"
prometheus.io/port: "9090"
prometheus.io/path: "/metrics"
{{- end }}

{{/*
Create backup labels
*/}}
{{- define "taishang-laojun.backup.labels" -}}
backup.taishanglaojun.ai/enabled: "true"
backup.taishanglaojun.ai/schedule: {{ .Values.backup.schedule | quote }}
backup.taishanglaojun.ai/retention: {{ .Values.backup.retention | quote }}
{{- end }}

{{/*
Create environment variables
*/}}
{{- define "taishang-laojun.env" -}}
- name: ENVIRONMENT
  value: {{ .Values.environment | quote }}
- name: REGION
  value: {{ .Values.region | quote }}
- name: DOMAIN
  value: {{ .Values.domain | quote }}
- name: KUBERNETES_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
{{- end }}

{{/*
Create volume mounts for security
*/}}
{{- define "taishang-laojun.volumeMounts" -}}
- name: tmp
  mountPath: /tmp
{{- end }}

{{/*
Create volumes for security
*/}}
{{- define "taishang-laojun.volumes" -}}
- name: tmp
  emptyDir: {}
{{- end }}