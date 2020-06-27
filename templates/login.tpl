{{ define "login.tpl" }}
{{ template "header.tpl" .}}
<body>

	<div class="container">
		<!-- Outer Row -->

		<div class="row justify-content-center">
			<div class="col-xl-10 col-lg-12 col-md-9">
				<div class="card o-hidden border-0 shadow-lg my-5">
					<div class="card-body p-0">
						<!-- Nested Row within Card Body -->

						<div class="row">
							<div class="col-lg-6 d-none d-lg-block bg-login-image"></div>

							<div class="col-lg-6">
								<div class="p-5">
									<div class="text-center">
										<h1 class="h4 text-gray-900 mb-4">{{- if .ErrorTitle -}}Oh, no!{{- else -}}Welcome Back!{{- end -}}</h1>
									</div>
								    {{ if .ErrorTitle}}
									<div class="bg-gradient-danger text-white">
										<i class="fas fa-exclamation-triangle text-white"></i>&nbsp;{{.ErrorTitle}}: {{.ErrorMessage}}
									</div>
									{{end}}
									<form class="user" action="/user/login" method="POST">
										<div class="form-group">
											<input type="username" class="form-control form-control-user" id="username" name="username" aria-describedby="usernameHelp" placeholder="Your Avatar username...">
										</div>

										<div class="form-group">
											<input type="password" class="form-control form-control-user" id="password" name="password" placeholder="Password">
										</div>

										<div class="form-group">
											<div class="custom-control custom-checkbox small">
												<input type="checkbox" class="custom-control-input" id="rememberMe" name="rememberMe"> <label class="custom-control-label" for="rememberMe">Remember Me</label>
											</div>
										</div><input type="submit" class="btn btn-primary btn-user btn-block" value="Login">
<!--									<hr>
										<a href="#" class="btn btn-google btn-user btn-block"> Login with Google</a>
										<a href="#" class="btn btn-facebook btn-user btn-block"> Login with Facebook</a> -->
									</form>
<!--									<hr> -->

									<div class="text-center">
										<a class="small" href="/user/forgot-password">Forgot Password?</a>
									</div>

									<div class="text-center">
										<a class="small" href="/user/register">Create an Account!</a>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
   {{ template "footer.tpl" .}}
{{ end }}