{{ define "welcome.tpl" }}
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
				<div class="col-lg-6 mb-4">
					<!-- Welcome Text -->
					<div class="text-center">
						<h1><i class="fas fa-fw fa-door-open"></i>Welcome to the Beta Technologies OpenSimulator Grid!</h1>
					</div>
				</div>
			</div> <!-- /.row -->
			<!-- Content Row -->
			<div class="row">
				<div class="col-lg-6 mb-4">
					<a href="https://betatechnologies.info" target=_blank><img src="https://betatechnologies.info/wp-content/uploads/2020/05/Beta-Technologies-Vertical-Logo-2008.svg" alt="Beta Technologies Logo" align="left" width="250"></a>
				</div>
				<div class="col-lg-6 mb-4">
					<p>We still don't have much to show here... it's all under construction!</p>
				</div>
			</div> <!-- /.row -->
			{{ template "back.tpl"}}
		</div>
		<!-- /.container-fluid -->

	</div>
	<!-- End of Main Content -->

{{ template "footer.tpl" .}}
{{ end }}