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
										<h1 class="h4 text-gray-900 mb-4">{{- if .BoxTitle -}}Oh, no!{{- else -}}Welcome Back!{{- end -}}</h1>
									</div>
									{{- if .BoxTitle -}}
									{{ template "infobox.tpl" .}}
									{{- end -}}
									<form class="user" action="/user/login" method="POST">
										<div class="form-group">
											<input type="username" class="form-control form-control-user" id="username" name="username" aria-describedby="usernameHelp" placeholder="Your Avatar username..." value="{{- .WrongUsername -}}" title="The avatar name is a first name, one space, and a last name" pattern="\w*\s\w*" required>
										</div>
										<div class="form-group">
												<input type="password" class="form-control form-control-user" id="password" name="password" placeholder="Password" value="{{- .WrongPassword -}}" required>
<!--										    <div class="input-group" id="show_hide_password">
												<div class="input-group-addon">
													<a href=""><i class="fa fa-eye-slash" aria-hidden="true"></i></a>
    											</div>
										    </div>
										</div>
										<script type="text/javascript">
											$(document).ready(function() {
													$("#show_hide_password a").on('click', function(event) {
														event.preventDefault();
														if($('#show_hide_password input').attr("type") == "text"){
															$('#show_hide_password input').attr('type', 'password');
															$('#show_hide_password i').addClass( "fa-eye-slash" );
															$('#show_hide_password i').removeClass( "fa-eye" );
														}else if($('#show_hide_password input').attr("type") == "password"){
															$('#show_hide_password input').attr('type', 'text');
															$('#show_hide_password i').removeClass( "fa-eye-slash" );
															$('#show_hide_password i').addClass( "fa-eye" );
														}
													});
											});
										</script> -->
										<div class="form-group">
											<div class="custom-control custom-checkbox small">
												<input type="checkbox" class="custom-control-input" id="rememberMe" name="rememberMe" {{- if .WrongRememberMe -}}checked{{- end -}} disabled="disabled">&nbsp;<!-- disabled for now --><label class="custom-control-label" for="rememberMe">Remember Me</label>
											</div>
										</div><input type="submit" class="btn btn-primary btn-user btn-block" value="Login">
<!--									<hr>
										<a href="#" class="btn btn-google btn-user btn-block"> Login with Google</a>
										<a href="#" class="btn btn-facebook btn-user btn-block"> Login with Facebook</a> -->
									</form>
									<hr>
									<div class="text-center">
										<a class="small" href="/user/reset-password">Forgot Password?</a>
									</div>
									<div class="text-center">
										<a class="small" href="/user/register">Create an Account!</a>
									</div>
									<div class="text-center">
										<a class="small" href="/"><i class="fas fa-fw fa-long-arrow-alt-left"></i>&nbsp;Back to Dashboard</a>
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