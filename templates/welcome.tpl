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
				<!-- Content Row for carousel-->
				<div class="row">
					<div class="card shadow col mb-4">
						<a href="#collapseCarousel" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="collapseCarousel">
							<h6 class="m-0 font-weight-bold text-primary">{{- .description -}}</h6>
						</a>
						<div class="col mb-4">
							<div class="collapse show" id="collapseCarousel">
								<!-- Carousel -->
								<div id="welcomeSlideshow" class="carousel slide" data-ride="carousel" data-interval="2000">
									<ol class="carousel-indicators">
										{{ range $index, $slideURL := .slideshow -}}
										<li data-target="#welcomeSlideshow" data-slide-to="{{- $index -}}"{{- if eq $index 0 }} class="active"{{- end -}}></li>
										{{ end }}
									</ol> <!-- ./carousel-indicators -->
									<div class="carousel-inner">
										{{ range $index, $slideURL := .slideshow -}}
										<div class="carousel-item{{- if eq $index 0 }} active{{- end -}}">
											<img class="d-block w-100" src="{{- $slideURL -}}" alt="Slide {{ $index -}}">
										</div>
										{{ end }}
									</div> <!-- ./carousel-inner -->
									<a class="carousel-control-prev" href="#welcomeSlideshow" role="button" data-slide="prev">
										<span class="carousel-control-prev-icon" aria-hidden="true"></span>
										<span class="sr-only">Previous</span>
									</a>
									<a class="carousel-control-next" href="#welcomeSlideshow" role="button" data-slide="next">
										<span class="carousel-control-next-icon" aria-hidden="true"></span>
										<span class="sr-only">Next</span>
									</a>
								</div>	<!-- ./carousel -->
							</div>	<!-- ./collapse -->
						</div>	<!-- ./col -->
					</div>	<!-- ./card shadow -->
					<div class="col-4 mb-4">
						<!-- DataTables for Region list -->
						<div class="card shadow mb-4">
							<a href="#regionsTableCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="regionsTableCard">
								<h6 class="m-0 font-weight-bold text-primary">List of Regions</h6>
							</a>
							<div class="collapse show" id="regionsTableCard">
								<div class="card-body text-secondary">
									<div class="table-responsive">
										<table class="table table-bordered table-striped table-compact table-squeezed" id="regionsTable" data-order='[]' data-page-length='35'>
											<thead>
												<tr>
													<th>regionName</th>
													<th>locX</th>
													<th>locY</th>
												</tr>
											</thead>
										</table>
									</div>	<!-- ./table-responsive -->
								</div>	<!-- ./card-body -->
							</div>	<!-- ./collapse -->
						</div>	<!-- ./card shadow -->
					</div>	<!-- ./col -->
				</div> <!-- /.row -->
				{{- if .viewerInfo -}}
				<!-- Content Row -->
				<div class="row">
					<div class="col-8 mb-4">
						<!-- DataTables for Viewer Info -->
						<div class="card shadow mb-4">
							<a href="#viewerInfoCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="viewerInfoCard">
								<h6 class="m-0 font-weight-bold text-primary">Your Viewer Info</h6>
							</a>
							<div class="collapse show" id="viewerInfoCard">
								<div class="card-body">
									<div class="table-responsive">
										<table class="table table-bordered table-compact table-striped table-squeezed" id="viewerInfo" data-order='[]' data-page-length='25'>
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
					</div>
				</div> <!-- /.row -->
				{{- end -}}
				<!-- Content Row -->
				<div class="row">
					{{- if .usersOnline -}}
					<div class="col mb-4">
						<!-- DataTables for online Users -->
						<div class="card shadow mb-4">
							<a href="#onlineUsersCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="onlineUsersCard">
								<h6 class="m-0 font-weight-bold text-primary">Users online</h6>
							</a>
							<div class="collapse show" id="onlineUsersCard">
								<div class="card-body">
									<div class="table-responsive">
										<table class="table table-bordered table-compact table-striped table-squeezed" id="usersOnline" data-order="[]" data-page-length="25">
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
					{{- end -}}
				</div> <!-- /.row -->
				{{- if .Debug -}}
				<div class="row">
					<div class="col">
						<h2>Debug info:</h2>
						<p><b>viewerInfo:</b>&nbsp;{{ .viewerInfo -}}</p><hr />
						<p><b>regionsTable:</b>&nbsp;{{ .regionsTable -}}</p><hr />
						<p><b>usersOnline:</b>&nbsp;{{ .usersOnline -}}</p>
					</div>
				</div>
				{{- end -}}
				{{ template "map.tpl" .}}
				{{ template "back.tpl" .}}
			</div>
			<!-- /.container-fluid -->
		</div>
		<!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}