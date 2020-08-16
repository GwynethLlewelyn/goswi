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
					Grid Status: {{ if eq .GridStatus "ONLINE"}}<span class="text-success">ONLINE</span>{{ else if .GridStatus eq "OFFLINE" }}<span class="text-danger">OFFLINE</span>{{ else }}<span class="text-secondary">{{ .GridStatus }}</span>{{ end }}<br>
					Online Now: {{ .Online_Now }}<br>
					HG Visitors Last 30 Days: {{ .HG_Visitors_Last_30_Days }}<br>
					Local Users Last 30 Days: {{ .Local_Users_Last_30_Days }}<br>
					Total Active Last 30 Days: {{ .Total_Active_Last_30_Days }}<br>
					Registered Users: {{ .Registered_Users }}<br>
					Regions: {{ .Regions }}<br>
					VarRegions: {{ .Var_Regions }}<br>
					Single Regions: {{ .Single_Regions }}<br>
					Total LandSize (km<sup>2</sup>): {{ .Total_LandSize(km2) }}<br>
					Login URL: {{ .Login_URL }}<br>
					Website: {{ .Website }}<br>
					Login Screen: {{ .Login_Screen }}<br>
				</div>
			</div> <!-- /.row -->
			{{- template "back.tpl" -}}
		</div> <!-- /.container-fluid -->
	</div> <!-- End of Main Content -->
{{ template "footer.tpl" .}}
{{ end }}