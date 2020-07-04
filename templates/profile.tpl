{{ define "profile.tpl" }}
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
						<h1><i class="fas fa-user fa-sm fa-fw"></i><h2>{{- if .Username -}}{{- .Username -}}{{- else -}}Your{{- end -}}&nbsp;Profile</h1>
					</div>
				</div>
			</div> <!-- /.row -->
			<!-- Content Row -->
			<div class="row">
				<div class="col-4 mb-4">
					<!-- this will be the user's mugshot -->
					{{ if .ProfileURL}}
					<a href="{{- .ProfileURL -}}" target="_blank">
					{{ end }}
					{{ if .ProfileImage }}
					<img src="{{- .ProfileImage -}}" alt="{{- .Username -}}" height="256" width="256">
					{{ else }}
					<img src="{{- .Libravatar -}}" alt="{{ .Username }}" height="256" width="256">
					{{ end }}
					{{ if .ProfileURL}}
					</a>
					{{ end }}
				</div>
				<div class="col-lg-8 mb-4">
					<p>
					{{ if .ProfileData }}
					{{- .ProfileData -}}
					{{- else -}}
					One day, your profile will be here!
					{{- end -}}
					</p>
				</div>
			</div> <!-- /.row -->
			<div class="row">
				{{ if .usersOnline }}
				<div class="col mb-4">
					<!-- DataTables for online Friends -->
					<div class="card shadow mb-4">
						<a href="#onlineUsersCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="onlineUsersCard">
							<h6 class="m-0 font-weight-bold text-primary">Your online friends</h6>
						</a>
						<div class="collapse show" id="onlineUsersCard">
							<div class="card-body">
								<div class="table-responsive">
									<table class="table table-bordered" id="usersOnline" width="100%" cellspacing="0"
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
					<p><b>usersOnline:</b>&nbsp;{{ .usersOnline -}}</p>
				</div>
			</div>
			{{ end }}
			{{ template "back.tpl"}}
		</div>
		<!-- /.container-fluid -->
	</div>
	<!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}