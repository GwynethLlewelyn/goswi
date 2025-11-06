{{ define "generic.tpl" }}
{{ template "header.tpl" .}}
{{ template "navigation.tpl" .}}
	<!-- Content Wrapper -->
	<div id="content-wrapper" class="d-flex flex-column">
		<!-- Main Content -->
		<div id="content">
{{ template "topbar.tpl" .}}
		<!-- Begin Page Content -->
		<div class="container-fluid">
			<!-- Content Row -->
			<div class="row">
				<div class="col">
					<!-- Title -->
					<div class="text-center">
						<h1>{{ .description -}}</h1>
					</div>
				</div>
			</div> <!-- /.row -->
			<!-- Content Row -->
			<div class="row">
				<div class="col mb-4">
					{{ .content }}
				</div>
			</div> <!-- /.row -->
			{{- template "back.tpl" -}}
		</div> <!-- /.container-fluid -->
	</div> <!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}