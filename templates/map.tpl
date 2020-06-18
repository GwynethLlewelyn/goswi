{{ define "map.tpl" }}
			<div class="row">
				<div class="col mb-4">
					<!-- DataTables for online Users -->
					<div class="card shadow mb-4">
						<div class="card-header py-3">
							<h6 class="m-0 font-weight-bold text-primary">Grid Map</h6>
						</div>
						<div class="card-body">
							<!-- Grid Map will be shown below, using LeafletJS -->
							<div id="gridMap"></div>
							<script type="text/javascript" src="../assets/js/leaflet-gridmap.js"></script>
						</div>
						<!-- /.card-body -->
					</div>
					<!-- /.card-shadow -->
				</div>
				<!-- /.col -->
			</div>
			<!-- /.row -->
{{ end }}