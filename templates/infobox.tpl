{{- define "infobox.tpl" -}}
<div class="alert alert-{{- .BoxType -}} bg-gradient-{{- .BoxType -}} text-white alert-dismissible fade show" role="alert">
	<i class="fas
	{{- if eq .BoxType "primary" .}}fa-flag
	{{- else if eq .BoxType "success" .}}fa-check
	{{- else if eq .BoxType "info" .}}fa-info-circle
	{{- else if eq .BoxType "warning" .}}fa-exclamation-triangle
	{{- else if eq .BoxType "danger" .}}fa-bomb
	{{- else if eq .BoxType "secondary" .}}fa-arrow-right{{- end -}} fa-pull-left fa-border text-white"></i>&nbsp;<strong>{{- .BoxTitle -}}:</strong>{{- .BoxMessage }}
	<button type="button" class="close" data-dismiss="alert" aria-label="Close">
		<span aria-hidden="true"><i class="fas fa-times-circle text-white"></i></span>
	</button>
</div> <!-- ./alert -->
{{- end }}
