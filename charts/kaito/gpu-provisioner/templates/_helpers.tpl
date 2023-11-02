{{/*
Expand the name of the chart.
*/}}
{{- define "gpu-provisioner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "gpu-provisioner.fullname" -}}
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
{{- define "gpu-provisioner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "gpu-provisioner.labels" -}}
helm.sh/chart: {{ include "gpu-provisioner.chart" . }}
{{ include "gpu-provisioner.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.additionalLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "gpu-provisioner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gpu-provisioner.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
gpu-provisioner image to use
*/}}
{{- define "gpu-provisioner.controller.image" -}}
{{- if .Values.controller.image.digest }}
{{- printf "%s:%s@%s" .Values.controller.image.repository  (default (printf "v%s" .Chart.AppVersion) .Values.controller.image.tag) .Values.controller.image.digest }}
{{- else }}
{{- printf "%s:%s" .Values.controller.image.repository  (default (printf "v%s" .Chart.AppVersion) .Values.controller.image.tag) }}
{{- end }}
{{- end }}


{{/* Get PodDisruptionBudget API Version */}}
{{- define "gpu-provisioner.pdb.apiVersion" -}}
{{- if and (.Capabilities.APIVersions.Has "policy/v1") (semverCompare ">= 1.21-0" .Capabilities.KubeVersion.Version) -}}
{{- print "policy/v1" -}}
{{- else -}}
{{- print "policy/v1beta1" -}}
{{- end -}}
{{- end -}}

{{/*
Patch the label selector on an object
This template will add a labelSelector using matchLabels to the object referenced at _target if there is no labelSelector specified.
The matchLabels are created with the selectorLabels template.
This works because Helm treats dictionaries as mutable objects and allows passing them by reference.
*/}}
{{- define "gpu-provisioner.patchLabelSelector" -}}
{{- if not (hasKey ._target "labelSelector") }}
{{- $selectorLabels := (include "gpu-provisioner.selectorLabels" .) | fromYaml }}
{{- $_ := set ._target "labelSelector" (dict "matchLabels" $selectorLabels) }}
{{- end }}
{{- end }}

{{/*
Patch pod affinity
This template uses the patchLabelSelector template to add a labelSelector to pod affinity objects if there is no labelSelector specified.
This works because Helm treats dictionaries as mutable objects and allows passing them by reference.
*/}}
{{- define "gpu-provisioner.patchPodAffinity" -}}
{{- if (hasKey ._podAffinity "requiredDuringSchedulingIgnoredDuringExecution") }}
{{- range $term := ._podAffinity.requiredDuringSchedulingIgnoredDuringExecution }}
{{- include "gpu-provisioner.patchLabelSelector" (merge (dict "_target" $term) $) }}
{{- end }}
{{- end }}
{{- if (hasKey ._podAffinity "preferredDuringSchedulingIgnoredDuringExecution") }}
{{- range $weightedTerm := ._podAffinity.preferredDuringSchedulingIgnoredDuringExecution }}
{{- include "gpu-provisioner.patchLabelSelector" (merge (dict "_target" $weightedTerm.podAffinityTerm) $) }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Patch affinity
This template uses patchPodAffinity template to add a labelSelector to podAffinity & podAntiAffinity if one isn't specified.
This works because Helm treats dictionaries as mutable objects and allows passing them by reference.
*/}}
{{- define "gpu-provisioner.patchAffinity" -}}
{{- if (hasKey .Values.affinity "podAffinity") }}
{{- include "gpu-provisioner.patchPodAffinity" (merge (dict "_podAffinity" .Values.affinity.podAffinity) .) }}
{{- end }}
{{- if (hasKey .Values.affinity "podAntiAffinity") }}
{{- include "gpu-provisioner.patchPodAffinity" (merge (dict "_podAffinity" .Values.affinity.podAntiAffinity) .) }}
{{- end }}
{{- end }}

{{/*
Patch topology spread constraints
This template uses the patchLabelSelector template to add a labelSelector to topologySpreadConstraints if one isn't specified.
This works because Helm treats dictionaries as mutable objects and allows passing them by reference.
*/}}
{{- define "gpu-provisioner.patchTopologySpreadConstraints" -}}
{{- range $constraint := .Values.topologySpreadConstraints }}
{{- include "gpu-provisioner.patchLabelSelector" (merge (dict "_target" $constraint) $) }}
{{- end }}
{{- end }}

{{/*
Flatten Settings Map using "." syntax
*/}}
{{- define "flattenSettings" -}}
{{- $map := first . -}}
{{- $label := last . -}}
{{- range $key := (keys $map | uniq | sortAlpha) }}
  {{- $sublabel := $key -}}
  {{- $val := (get $map $key) -}}
  {{- if $label -}}
    {{- $sublabel = list $label $key | join "." -}}
  {{- end -}}
  {{/* Special-case "tags" since we want this to be a JSON object */}}
  {{- if eq $key "tags" -}}
    {{- if not (kindIs "invalid" $val) -}}
      {{- $sublabel | quote | nindent 2 }}: {{ $val | toJson | quote }}
    {{- end -}}
  {{- else if kindOf $val | eq "map" -}}
    {{- list $val $sublabel | include "flattenSettings" -}}
  {{- else -}}
  {{- if not (kindIs "invalid" $val) -}}
    {{- $sublabel | quote | nindent 2 -}}: {{ $val | quote }}
  {{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Flatten the stdout logging outputs from args provided
*/}}
{{- define "gpu-provisioner.controller.outputPathsList" -}}
{{ $paths := list -}}
{{- range .Values.controller.outputPaths -}}
    {{- $paths = printf "%s" . | quote  | append $paths -}}
{{- end -}}
{{ $paths | join ", " }}
{{- end -}}

{{/*
Flatten the stderr logging outputs from args provided
*/}}
{{- define "gpu-provisioner.controller.errorOutputPathsList" -}}
{{ $paths := list -}}
{{- range .Values.controller.errorOutputPaths -}}
    {{- $paths = printf "%s" . | quote  | append $paths -}}
{{- end -}}
{{ $paths | join ", " }}
{{- end -}}