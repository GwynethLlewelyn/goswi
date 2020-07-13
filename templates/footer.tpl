{{ define "footer.tpl" }}
{{ if .logintemplate }}
<!-- we may include jQuery here, for doing the fancy view password thing -->
{{ end }}
{{ if not .logintemplate }}
		<!-- Footer -->
		<footer class="sticky-footer bg-white">
			<div class="container my-auto">
				<div class="copyright text-center my-auto">
					<span><i class="fab fa-creative-commons fa-fw"></i> {{.now }} by gOSWI and {{.author}}. Some rights reserved. Uses <a href="https://startbootstrap.com/themes/sb-admin-2/" target=_blank>SB Admin 2</a> templates inside <a href="https://golang.org/" target=_blank>Go.</a></span>
				</div>
			</div>
		</footer>	<!-- End of Footer -->
	</div>	<!-- End of Content Wrapper -->
	</div>	<!-- End of Page Wrapper -->
	<!-- Scroll to Top Button-->
	<a class="scroll-to-top rounded" href="#page-top">
	<i class="fas fa-angle-up"></i>
	</a>
	<!-- Logout Modal-->
	<div class="modal fade" id="logoutModal" tabindex="-1" role="dialog" aria-labelledby="exampleModalLabel" aria-hidden="true">
		<div class="modal-dialog" role="document">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="exampleModalLabel">Ready to Leave?</h5>
					<button class="close" type="button" data-dismiss="modal" aria-label="Close">
					<span aria-hidden="true">×</span>
					</button>
				</div>
				<div class="modal-body">Select "Logout" below if you are ready to end your current session.</div>
				<div class="modal-footer">
					<button class="btn btn-secondary" type="button" data-dismiss="modal">Cancel</button>
					<a class="btn btn-primary" href="/user/logout">Logout</a>
				</div>
			</div>
		</div>
	</div>
{{ end }}
	<!-- Bootstrap core JavaScript-->
	<script src="../lib/startbootstrap-sb-admin-2/vendor/jquery/jquery.min.js"></script>
	<script src="../lib/startbootstrap-sb-admin-2/vendor/bootstrap/js/bootstrap.bundle.min.js"></script>

	<!-- Core plugin JavaScript-->
	<script src="../lib/startbootstrap-sb-admin-2/vendor/jquery-easing/jquery.easing.min.js"></script>

	<!-- Custom scripts for all pages-->
	<script src="../lib/startbootstrap-sb-admin-2/js/sb-admin-2.min.js"></script>
{{ if .needsTables }}
	<script src="../lib/startbootstrap-sb-admin-2/vendor/datatables/jquery.dataTables.min.js"></script>
	<script src="../lib/startbootstrap-sb-admin-2/vendor/datatables/dataTables.bootstrap4.min.js"></script>
	<script type="text/javascript">
	// Call the dataTables jQuery plugin
		$(document).ready(function() {
			$('#viewerInfo').dataTable( {
				"searching":	false,
				"ordering":		false,
				"paging":		false,
				"scrollCollapse": true,
				"info":			false,
				"data": {{ .viewerInfo -}},
				"columns": [
					{ "data": "channel" },
					{ "data": "grid" },
					{ "data": "lang" },
					{ "data": "login_content_version" },
					{ "data": "os" },
					{ "data": "sourceid" },
					{ "data": "version" }
				]
			});
			$('#regionsTable').dataTable( {
				"searching":	false,
				"paging": 		false,
				"info":			false,
				"data": {{ .regionsTable -}},
				"columns": [
					{ "data": "regionName" },
					{ "data": "locX" },
					{ "data": "locY" }
				]
			});
			$('#usersOnline').dataTable( {
				"data": {{ .usersOnline -}},
				"columns": [
					{ "data": "Avatar Name" }
				]
			});
		});
	</script>
{{ end }}
</body>
</html>
{{ end }}