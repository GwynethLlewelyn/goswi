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
				<div class="col">
					<!-- Welcome Text -->
					<div class="text-center">
						<h1><i class="fas fa-fw fa-door-open"></i>Welcome to the Beta Technologies OpenSimulator Grid!</h1>
					</div>
				</div>
			</div> <!-- /.row -->
			<!-- Content Row -->
			<div class="row">
				<div class="col-2 mb-4">
					<a href="https://betatechnologies.info" target=_blank><img src="https://betatechnologies.info/wp-content/uploads/2020/05/Beta-Technologies-Vertical-Logo-2008.svg" alt="Beta Technologies Logo"></a>
				</div>
				<div class="col-lg-10 mb-4">
					<p>We still don't have much to show here... it's all under construction!</p>
					<!-- DataTables for Viewer Info -->
					{{ if .viewerInfo }}
					<div class="card shadow mb-4">
						<a href="#viewerInfo" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="viewerInfo">
							<h6 class="m-0 font-weight-bold text-primary">Your Viewer Info</h6>
						</a>
						<div class="collapse show" id="vieweInfo">
							<div class="card-body">
								<div class="table-responsive">
									<table class="table table-bordered table-striped" id="viewerInfo" width="100%" cellspacing="0"
										data-order='[]' data-page-length='25'>
										<thead>
											<tr>
												<th>ViewerName</th>
												<th>Grid</th>
												<th>Language</th>
												<th>LoginContentVersion</th>
												<th>OS</th>
												<th>SourceID</th>
												<th>Version</th>
											</tr>
										</thead>
									</table>
								</div>
							</div>
							<!-- ./card-body -->
						</div>
						<!-- ./collapse -->
					</div>
					{{ end }}
				</div>
			</div> <!-- /.row -->
			<!-- Content Row -->
			<div class="row">
				<div class="col mb-4">
					<!-- DataTables for Region list -->
					<div class="card shadow mb-4">
						<a href="#regionList" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="regionList">
							<h6 class="m-0 font-weight-bold text-primary">List of Regions</h6>
						</a>
						<div class="collapse show" id="regionList">
							<div class="card-body">
								<div class="table-responsive">
									<table class="table table-bordered table-striped" id="regionsTable" width="100%" cellspacing="0"
										data-order='[]' data-page-length='25'>
										<thead>
											<tr>
												<th>regionName</th>
												<th>locX</th>
												<th>locY</th>
											</tr>
										</thead>
									</table>
								</div>
							</div>
							<!-- ./card-body -->
						</div>
						<!-- ./collapse -->
					</div>
				</div>
				{{ if .usersOnline }}
				<div class="col mb-4">
					<!-- DataTables for online Users -->
					<div class="card shadow mb-4">
						<a href="#onlineUsersCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="onlineUsersCard">
							<h6 class="m-0 font-weight-bold text-primary">Users online</h6>
						</a>
						<div class="collapse show" id="onlineUsersCard">
							<div class="card-body">
								<div class="table-responsive">
									<table class="table table-bordered table-striped" id="usersOnline" width="100%" cellspacing="0"
										data-order='[]' data-page-length='25'>
										<thead>
											<tr>
												<th>Avatar Name</th>
											</tr>
										</thead>
									</table>
								</div>
							</div>
							<!-- ./card-body -->
						</div>
						<!-- ./collapse -->
					</div>
				</div>
				{{ end }}
			</div> <!-- /.row -->
			{{ if .Debug }}
			<div class="row">
				<div class="col">
					<h2>Debug info:</h2>
					<p><b>viewerInfo:</b>&nbsp;{{ .viewerInfo -}}</p><hr />
					<p><b>regionsTable:</b>&nbsp;{{ .regionsTable -}}</p><hr />
					<p><b>usersOnline:</b>&nbsp;{{ .usersOnline -}}</p>
				</div>
			</div>
			{{ end }}
			{{ template "map.tpl" }}
			{{ template "back.tpl"}}
		</div>
		<!-- /.container-fluid -->

	</div>
	<!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}