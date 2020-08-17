{{ define "stats.tpl" }}
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
					<!-- About -->
					<div class="text-center">
						<h1>Grid Stats</h1>
					</div>
				</div>
			</div> <!-- /.row -->
			<!-- Content Row -->
			<div class="row">
				<div class="col mb-4">
					<strong>Grid Status:</strong> {{ if eq .GridStatus "ONLINE" -}}<span class="text-success">ONLINE</span>{{- else if .GridStatus eq "OFFLINE" -}}<span class="text-danger">OFFLINE</span>{{- else -}}<span class="text-secondary">{{- .GridStatus -}}</span>{{- end -}}<br>
					<strong>Online Now:</strong> {{ .Online_Now -}}<br>
					<strong>HG Visitors Last 30 Days:</strong> {{ .HG_Visitors_Last_30_Days -}}<br>
					<strong>Local Users Last 30 Days:</strong> {{ .Local_Users_Last_30_Days -}}<br>
					<strong>Total Active Last 30 Days:</strong> {{ .Total_Active_Last_30_Days -}}<br>
					<strong>Registered Users:</strong> {{ .Registered_Users -}}<br>
					<strong>Regions:</strong> {{ .Regions -}}<br>
					<strong>VarRegions:</strong> {{ .Var_Regions -}}<br>
					<strong>Single Regions:</strong> {{ .Single_Regions -}}<br>
					<strong>Total Land Size (km<sup>2</sup>):</strong> {{ .Total_LandSize -}}<br>
					<strong>Login URL:</strong> <a href="{{- .Login_URL -}}">{{- .Login_URL -}}</a><br>
					<strong>Website:</strong> <a href="{{- .Website -}}">{{- .Website -}}</a><br>
					<strong>Login Screen:</strong> <a href="{{- .Login_Screen -}}">{{- .Login_Screen -}}</a><br>
				</div>
			</div> <!-- /.row -->
			<div class="row">
				<div class="col mb-4">
					<span class="text-sm-left"><em>Last updated:&nbsp;{{- .timestamp -}}</em></span>
				</div>
			</div>
			{{- template "back.tpl" -}}
		</div> <!-- /.container-fluid -->
	</div> <!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}