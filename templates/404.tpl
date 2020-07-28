{{- define "404.tpl" -}}
{{- template "header.tpl" . -}}
{{- template "navigation.tpl" . -}}
	<!-- Content Wrapper -->
	<div id="content-wrapper" class="d-flex flex-column">
		<!-- Main Content -->
		<div id="content">
{{- template "topbar.tpl" . -}}
			<!-- Begin Page Content -->
			<div class="container-fluid">
				<div class="row">
					<div class="col">
						<!-- 404 Error Text -->
						<div class="text-center">
							<div class="error mx-auto" data-text="{{- if .errorcode -}}{{- .errorcode -}}{{- else -}}404{{- end -}}">{{- if .errorcode -}}{{- .errorcode -}}{{- else -}}404{{- end -}}</div>
							<p class="lead text-gray-800 mb-5">{{- if .errortext -}}{{- .errortext -}}{{- else -}}Page Not Found{{- end -}}</p>
							<p class="text-gray-500 mb-4">{{- if .errorbody -}}{{- .errorbody -}}{{- else -}}It looks like you found a glitch in the matrix...{{- end -}}</p>
						</div>
					</div>
				</div>
				{{ template "back.tpl"}}
			</div> <!-- /.container-fluid -->
		</div>
		<!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}