{{ define "404.tpl" }}
{{ template "header.tpl" .}}
{{ template "navigation.tpl" .}}
	<!-- Content Wrapper -->
	<div id="content-wrapper" class="d-flex flex-column">

		<!-- Main Content -->
		<div id="content">

{{ template "topbar.tpl" .}}

		<!-- Begin Page Content -->
		<div class="container-fluid">
			<div class="row">
				<div class="col">
					<!-- 404 Error Text -->
					<div class="text-center">
						<div class="error mx-auto" data-text="404">404</div>
						<p class="lead text-gray-800 mb-5">Page Not Found</p>
						<p class="text-gray-500 mb-0">It looks like you found a glitch in the matrix...</p>
					</div>
				</div>
			</div>
			{{ template "back.tpl"}}
		</div>
		<!-- /.container-fluid -->

		</div>
		<!-- End of Main Content -->

{{ template "footer.tpl" .}}
{{ end }}