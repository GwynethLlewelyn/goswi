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
											<th>Email</th>th>
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
											<th>Email</th>th>
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
							<p><b>messageTotal:</b>&nbsp;{{ .messageTotal -}}</p>
						</div>
					</div>
					{{ end }}
					{{ template "back.tpl"}}
				</div>
				<!-- /.container-fluid -->
			</div>
			<!-- End of Main Content -->
			<script src="../lib/startbootstrap-sb-admin-2/vendor/datatables/jquery.dataTables.min.js"></script>
			<script src="../lib/startbootstrap-sb-admin-2/vendor/datatables/dataTables.bootstrap4.min.js"></script>
			<script>
			// Call the dataTables jQuery plugin
				$(document).ready(function() {
					{{ if .offlineMessages -}}
					$('#offline-messages').dataTable( {
						"searching":	true,
						"ordering":		true,
						"paging":		false,
						"scrollCollapse": true,
						"info":			false,
						"data": {{ .offlineMessages -}},
						"columnDefs": [
						{
							target: 0,
							visible: false,
							searchable: false
						},
						"columns": [
							{ "data": "ID" },
							{ "data": "im_offline" },
							{ "data": "FromID" },
							{ "data": "Message" },
							{ "data": "TMStamp" },
							{ "data": "FirstName" },
							{ "data": "LastName" },
							{ "data": "Email" }
						]
					});
					{{ end }}
					{{ if .feedMessages -}}
					$('#feed-messages').dataTable( {
						"searching":	true,
						"ordering":		true,
						"paging":		false,
						"scrollCollapse": true,
						"info":			false,
						"data": {{ .feedMessages -}},
						"columns": [
							{ "data": "PostParentID" },
							{ "data": "PosterID" },
							{ "data": "PostID" },
							{ "data": "PostMarkup" },
							{ "data": "Chronostamp" },
							{ "data": "Visibility" },
							{ "data": "Comment" },
							{ "data": "Commentlock" },
							{ "data": "Editlock" },
							{ "data": "Feedgroup" },
							{ "data": "FirstName" },
							{ "data": "LastName" },
							{ "data": "Email" }
						]
					});
					{{ end }}
				}
			</script>
{{ template "footer.tpl" . -}}
{{- end -}}