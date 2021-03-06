{{ define "change-password.tpl" }}
{{ template "header.tpl" .}}
<body class="bg-gradient-primary">
	<div class="container">
		<!-- Outer Row -->
		<div class="row justify-content-center">
			<div class="col-xl-10 col-lg-12 col-md-9">
				<div class="card o-hidden border-0 shadow-lg my-5">
					<div class="card-body p-0">
					<!-- Nested Row within Card Body -->
						<div class="row">
							<div class="col-lg-6 d-none d-lg-block bg-change-password-image"></div>
							<div class="col-lg-6">
								<div class="p-5">
									<div class="text-center">
										<h1 class="h4 text-gray-900 mb-2">{{- if .BoxTitle -}}Oops!{{- else -}}Change password{{- end -}}</h1>
										{{- if .BoxTitle}}
										{{- template "infobox.tpl" .}}
										{{- end }}
										{{- if not .someTokens }}
										<p class="mb-4">Please enter your old password, the new one and confirm the new one</p>
										{{- else -}}
										<p class="mb-4">Please enter a new password and confirm it</p>
										{{- end -}}
									</div> <!-- ./text-center -->
									<form class="user" action="/user/change-password" method="POST">
										{{- if not .someTokens -}}
										<div class="form-group">
											<input type="password" class="form-control form-control-user" id="oldpassword" name="oldpassword" placeholder="Old Password" value="{{- .WrongOldPassword -}}" required>
										</div>
										{{- else -}}
											<input type="hidden" id="t" value="{{- .someTokens -}}">
										{{- end -}}
										<div class="form-group">
											<input type="password" class="form-control form-control-user" id="newpassword" name="newpassword" placeholder="New Password" value="{{- .WrongNewPassword -}}" minlength="8" minlength="20" required>
										</div>
										<div class="form-group">
											<input type="password" class="form-control form-control-user" id="confirmnewpassword" name="confirmnewpassword" placeholder="Confirm New Password" value="{{- .WrongConfirmNewPassword -}}" minlength="8" minlength="20" required>
										</div>
										<input type="submit" value="Change Password" class="btn btn-primary btn-user btn-block">
									</form>
									<hr>
									<div class="text-center">
										<a class="small" href="/"><i class="fas fa-fw fa-long-arrow-alt-left"></i>&nbsp;Back to Dashboard</a>
									</div>
								</div> <!-- ./p-5 -->
							</div> <!-- ./col-lg-6 -->
						</div> <!-- ./row -->
					</div> <!-- ./card-body -->
				</div> <!-- ./card -->
			</div> <!-- ./col-xl-10 -->
		</div> <!-- ./row justify-content-center -->
	</div> <!-- ./container --->
	{{ template "footer.tpl" .}}
{{ end }}