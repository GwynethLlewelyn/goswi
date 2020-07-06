{{ define "reset-password-confirmation.tpl" }}
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
										<h1 class="h4 text-gray-900 mb-2">{{- if .ErrorTitle -}}Oh, snap!{{- else -}}Ok!{{- end -}}</h1>
										<p>{{- .Content -}}</p>
									</div> <!-- ./text-center -->
									<hr>
									<div class="text-center">
										<a class="small" href="/user/register">Create an Account!</a>
									</div> <!-- register -->
									<div class="text-center">
										<a class="small" href="/user/login">Already have an account? Login!</a>
									</div> <!-- login -->
									<div class="text-center">
										<a class="small" href="/"><i class="fas fa-fw fa-long-arrow-alt-left"></i>&nbsp;Back to Dashboard</a>
									</div> <!-- ./text-center -->
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