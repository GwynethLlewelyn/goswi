{{- define "profile.tpl" -}}
{{- template "header.tpl" . -}}
{{- template "navigation.tpl" . -}}
	<!-- Content Wrapper -->
	<div id="content-wrapper" class="d-flex flex-column">
		<!-- Main Content -->
		<div id="content">
{{- template "topbar.tpl" . -}}
		<!-- Begin Page Content -->
		<div class="container-fluid">
			<!-- Content Row -->
			<div class="row">
				<div class="col-4 mb-4">
					<!-- this will be the user's mugshot -->
					{{ if .ProfileURL}}
					<a href="{{- .ProfileURL -}}" target="_blank">
					{{ end -}}
					{{- if .ProfileImage }}
					<img src="{{- .ProfileImage -}}" alt="{{- .Username }} ({{- .UserUUID -}})" height="256" width="256">
					{{ else }}
					<img src="{{- .Libravatar -}}" alt="{{ .Username }} ({{- .UserUUID -}}" height="256" width="256">
					{{ end -}}
					{{- if .ProfileURL}}
					</a>
					{{ end }}
				</div>
				<div class="col-lg-8 mb-4">
					<form class="well form-horizontal" action="/user/profile" method="post"  id="profileForm">
						<fieldset>
							<!-- Your Profile: UserName -->
							<legend class="text-center"><h2>{{- if .Username -}}{{- .Username -}}{{- else -}}Your{{- end -}}&nbsp;Profile</h2></legend><br />
							<!-- About -->
							<div class="form-group">
							<label class="col-md-4 control-label">About You</label>
								<div class="col-md-4 inputGroupContainer">
									<div class="input-group">
										<span class="input-group-addon"><i class="fas fa-user"></i></span>
										<input name="ProfileURL" placeholder="{{- .ProfileURL -}}" class="form-control" type="text">
									</div>
								</div>
							</div>
							<div class="form-group">
    							<label for="AboutText">About</label>
    							<textarea class="form-control" id="AboutText" rows="10">{{- .ProfileAboutText -}}</textarea>
  							</div>
							<!-- ProfileURL -->
							<div class="form-group">
								<label class="col-md-4 control-label">Profile URL</label>
									<div class="col-md-4 inputGroupContainer">
										<div class="input-group">
										<span class="input-group-addon"><i class="fas fa-globe"></i></span>
										<input name="ProfileURL" placeholder="{{- .ProfileURL -}}" class="form-control" type="text">
									</div>
								</div>
							</div>
							<!-- Partner -->
							<div class="form-group">
							<label class="col-md-4 control-label">Partner</label>
								<div class="col-md-4 inputGroupContainer">
									<div class="input-group">
										<span class="input-group-addon"><i class="fas fa-user"></i></span>
										<input name="ProfilePartner" placeholder="{{- .ProfilePartner -}}" class="form-control" type="text">
									</div>
								</div>
							</div>
							<!-- Select example
							<div class="form-group">
								<label class="col-md-4 control-label">Department / Office</label>
								<div class="col-md-4 selectContainer">
									<div class="input-group">
										<span class="input-group-addon"><i class="glyphicon glyphicon-list"></i></span>
										<select name="department" class="form-control selectpicker">
											<option value="">Select your Department/Office</option>
											<option>Department of Engineering</option>
											<option>Department of Agriculture</option>
											<option >Accounting Office</option>
											<option >Tresurer's Office</option>
											<option >MPDC</option>
											<option >MCTC</option>
											<option >MCR</option>
											<option >Mayor's Office</option>
											<option >Tourism Office</option>
										</select>
									</div>
								</div>
							</div> -->

							<!-- Text input-->
							<div class="form-group">
								<label class="col-md-4 control-label">Languages spoken</label>
								<div class="col-md-4 inputGroupContainer">
									<div class="input-group">
										<span class="input-group-addon"><i class="fas fa-language"></i></span>
										<input name="ProfileLanguages" placeholder="{{- .ProfileLanguages -}}" class="form-control" type="text">
									</div>
								</div>
							</div>

							<!-- Text input-->
							<div class="form-group">
							<label class="col-md-4 control-label">Languages spoken</label>
								<div class="col-md-4 inputGroupContainer">
									<div class="input-group">
										<span class="input-group-addon"><i class="fas fa-language"></i></span>
										<input name="ProfileLanguages" placeholder="{{- .ProfileLanguages -}}" class="form-control" type="text">
									</div>
								</div>
							</div>
							<div class="col-2 mb-4">
								{{ if .ProfileFirstImage -}}
								<img src="{{- .ProfileFirstImage -}}" alt="Real Life Image for {{- .Username -}}" height="128" width="128"><br />
								{{- end }}
								<div class="form-group">
									<label for="ProfileFirstText">About your real life</label>
									<textarea class="form-control" id="ProfileFirstText" rows="10">{{- .ProfileFirstText -}}</textarea>
								</div>
							</div>
							<!-- Success message -->
							<div class="alert alert-success invisible" role="alert" id="success_message">Success<i class="fas fa-thumbs-up"></i>Success!</div>
							<!-- Submit Button -->
							<div class="form-group">
								<label class="col-md-4 control-label"></label>
								<div class="col-md-4"><br>
									<button type="submit" class="btn btn-warning">Submit<span class="fas fa-paper-plane"></span></button>
								</div>
							</div>
						</fieldset>
					</form>
					{{- if .ProfileData }}
					<!-- Raw data: {{- .ProfileData -}}-->
					{{ end -}}
				</div> <!-- /.col -->
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
{{ template "footer.tpl" . -}}
{{- end -}}