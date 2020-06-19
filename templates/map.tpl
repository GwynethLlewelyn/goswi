{{ define "map.tpl" }}
			<div class="row">
				<div class="col mb-4">
					<!-- DataTables for online Users -->
					<div class="card shadow mb-4">
						<a href="#gridMapCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="gridMapCard">
							<h6 class="m-0 font-weight-bold text-primary">Grid Map</h6>
						</a>
						<div class="collapse show" id="gridMapCard">
							<div class="card-body">
								<!-- Grid Map will be shown below, using LeafletJS -->
								<div id="gridMap"></div>
								<script type="text/javascript" src="../assets/js/leaflet-gridmap.js"></script>
							</div>
							<!-- /.card-body -->
						</div>
						<!-- /.collapse-show -->
					</div>
					<!-- /.card-shadow -->
				</div>
				<!-- /.col -->
			</div>
			<!-- /.row -->
{{ end }}