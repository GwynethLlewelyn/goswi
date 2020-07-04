{{ define "reset-password.tpl" }}
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
							<div class="col-lg-6 d-none d-lg-block bg-password-image"></div>
							<div class="col-lg-6">
								<div class="p-5">
									<div class="text-center">
										<h1 class="h4 text-gray-900 mb-2">{{- if .ErrorTitle -}}Oops!{{- else -}}Forgot Your Password?{{- end -}}</h1>
										{{ if .ErrorTitle}}
										<div class="alert alert-danger bg-gradient-danger text-white alert-warning alert-dismissible fade show" role="alert">
											<i class="fas fa-exclamation-triangle text-white"></i>&nbsp;<strong>{{.ErrorTitle}}:</strong>{{.ErrorMessage}}
											<button type="button" class="close" data-dismiss="alert" aria-label="Close">
												<span aria-hidden="true"><i class="fas fa-times-circle text-white"></i></span>
											</button>
										</div>
										{{else}}
										<p class="mb-4">We get it, stuff happens. Just enter your email address below and we'll send you a way to reset your password!</p>
										{{ end }}
									</div> <!-- ./text-center -->
									<form class="user" action="/user/reset-password" method="POST">
										<div class="form-group">
											<input type="email" class="form-control form-control-user" id="email" name="email" aria-describedby="emailHelp" placeholder="Enter Email Address..." value="{{- .WrongEmail -}}">
										</div>
										<input type="submit" class="btn btn-primary btn-user btn-block" value="Reset">
									</form>
									<hr>
									<div class="text-center">
										<a class="small" href="/user/register">Create an Account!</a>
									</div> <!-- register -->
									<div class="text-center">
										<a class="small" href="/user/login">Already have an account? Login!</a>
									</div> <!-- login -->
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