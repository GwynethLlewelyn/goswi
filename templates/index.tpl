{{- define "index.tpl" -}}
{{- template "header.tpl" . -}}
{{- template "navigation.tpl" . -}}
	<!-- Content Wrapper -->
	<div id="content-wrapper" class="d-flex flex-column">
		<!-- Main Content -->
		<div id="content">
{{- template "topbar.tpl" . -}}
			<!-- Begin Page Content -->
			<div class="container-fluid">
			<!-- Page Heading -->
				<div class="d-sm-flex align-items-center justify-content-between mb-4">
					<h1 class="h3 mb-0 text-gray-800">Dashboard</h1>
					<a href="/help" class="d-none d-sm-inline-block btn btn-sm btn-primary shadow-sm"><i class="fas fa-question-circle fa-sm text-white-50"></i>&nbsp;Help!</a>
				</div>
				<!-- Content Row -->
				<div class="row">
{{- if .Username -}}
					<div class="col-2">
						<img class="img-profile" src="{{- .Libravatar -}}">
					</div>
					<div class="col-10">
						<p>Welcome,&nbsp;{{- .Username -}}!</p>
						{{- if .Content -}}
						<p>{{ .Content }}</p>
						{{- end -}}
					</div>
{{- else -}}
					<div class="col">
						<p>Nothing yet... please be patient!</p>
					</div>
{{- end -}}
{{- if .BoxMessage -}}
{{ template "infobox.tpl" . }}
{{- end -}}
				</div> <!-- /.row -->
			</div> <!-- /.container-fluid -->
		</div> <!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}