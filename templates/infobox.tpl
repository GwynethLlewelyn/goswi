{{- define "infobox.tpl" -}}
<div class="alert alert-{{- .BoxType }} bg-gradient-{{- .BoxType }} text-white alert-dismissible fade show" role="alert">
	<i class="fas {{ if .BoxType }}
	{{ if (eq .BoxType "primary") }}fa-flag
	{{ else if (eq .BoxType "success") }}fa-check
	{{ else if (eq .BoxType "info") }}fa-info-circle
	{{ else if (eq .BoxType "warning") }}fa-exclamation-triangle
	{{ else if (eq .BoxType "danger") }}fa-bomb
	{{ else if (eq .BoxType "secondary") }}fa-arrow-right{{- end }}
{{ else }}fa-skull-crossbones{{ end }} fa-pull-left text-white"></i>&nbsp;<strong>{{- .BoxTitle -}}:&nbsp;</strong>{{- .BoxMessage }}
	<button type="button" class="close" data-dismiss="alert" aria-label="Close">
		<span aria-hidden="true"><i class="fas fa-times-circle text-white"></i></span>
	</button>
</div> <!-- ./alert -->
{{- end }}
