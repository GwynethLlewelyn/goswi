{{- define "infobox.tpl" -}}
<div class="alert alert-danger bg-gradient-danger text-white alert-dismissible fade show" role="alert">
	<i class="fas fa-exclamation-triangle text-white"></i>&nbsp;<strong>{{- .BoxTitle -}}:</strong>{{- .BoxMessage }}
	<button type="button" class="close" data-dismiss="alert" aria-label="Close">
		<span aria-hidden="true"><i class="fas fa-times-circle text-white"></i></span>
	</button>
</div>
{{- end -}}