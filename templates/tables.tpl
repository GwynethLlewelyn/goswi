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
								<table class="table table-bordered table-striped"  id="{{- if .offlineMessages -}}offline-messages{{- else if .feedMessages -}}feed-messages{{- else -}}broken{{- end -}}" data-order="[]" data-page-length="25" width="100%" cellspacing="0">
								{{- if .offlineMessages -}}
									<thead>
										<tr>
											<th>ID</th>
											<th>Username</th>
											<th>Avatar</th>
											<th>Message</th>
											<th>Date</th>
										</tr>
									</thead>
									{{- if gt .numberMessages 20 -}}
									<tfoot>
										<tr>
											<th>ID</th>
											<th>Username</th>
											<th>Avatar</th>
											<th>Message</th>
											<th>Date</th>
										</tr>
									</tfoot>
									{{- end -}}
								{{- end -}}
								{{- if .feedMessages -}}
									<thead>
										<tr>
											<th>PostParentID</th>
											<th>PosterID</th>
											<th>PostID</th>
											<th>Username</th>
											<th>Avatar</th>
											<th>Message</th>
											<th>Date</th>
											<th class="hidden">Visibility</th>
											<th class="hidden">Comment</th>
											<th class="hidden">Commentlock</th>
											<th class="hidden">Editlock</th>
											<th class="hidden">Feedgroup</th>
										</tr>
									</thead>
									{{- if gt .numberFeedMessages 20 -}}
									<tfoot>
										<tr>
											<th>PostParentID</th>
											<th>PosterID</th>
											<th>PostID</th>
											<th>Username</th>
											<th>Avatar</th>
											<th>Message</th>
											<th>Date</th>
											<th class="hidden">Visibility</th>
											<th class="hidden">Comment</th>
											<th class="hidden">Commentlock</th>
											<th class="hidden">Editlock</th>
											<th class="hidden">Feedgroup</th>
										</tr>
									</tfoot>
									{{- end -}}
								{{- end -}}
								</table>
							</div>
						</div>
					</div>
					{{ if .Debug }}
					{{ template "infobox.tpl" . }}
					{{ end }}
					{{ template "back.tpl"}}
				</div>
				<!-- /.container-fluid -->
			</div>
			<!-- End of Main Content -->
{{ template "footer.tpl" . -}}
{{- end -}}