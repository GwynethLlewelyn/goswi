{{- define "tables.tpl" -}}
{{- template "header.tpl" . -}}
{{- template "navigation.tpl" . -}}
		<!-- Content Wrapper -->
		<div id="content-wrapper" class="d-flex flex-column">
			<!-- Main Content -->
			<div id="content">
{{- template "topbar.tpl" . -}}
				<!-- Begin Page Content -->
				<div class="container-fluid">
					<!-- Page Heading -->
					<h1 class="h3 mb-2 text-gray-800">{{ .tableName }}</h1>
					{{- if .BoxTitle -}}
					{{ template "infobox.tpl" .}}
					{{- end -}}
					<!-- DataTables Example -->
					<div class="card shadow mb-4">
						<div class="card-header py-3">
							<h6 class="m-0 font-weight-bold text-primary">WiP</h6>
						</div>
						<div class="card-body">
							<div class="table-responsive">
								<table class="table table-bordered table-compact table-striped table-squeezed"  id="offline-messages" data-order="[]" data-page-length="25" width="100%" cellspacing="0">
									<thead>
										<tr>
											<th>ID</th>
											<th>im_offline</th>
											<th>FromID</th>
											<th>Message</th>
											<th>TMStamp</th>
											<th>FirstName</th>
											<th>LastName</th>
											<th>Email</th>
										</tr>
									</thead>
									<tfoot>
										<tr>
											<th>ID</th>
											<th>im_offline</th>
											<th>FromID</th>
											<th>Message</th>
											<th>TMStamp</th>
											<th>FirstName</th>
											<th>LastName</th>
											<th>Email</th>
										</tr>
									</tfoot>
								</table>
							</div>
						</div>
					</div>
					{{ if .Debug }}
					<div class="row">
						<div class="col">
							<h2>Debug info:</h2>
							<p><b>Number of messages:</b>&nbsp;{{ .numberMessages -}}</p>
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