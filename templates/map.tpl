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
								<!-- 
==============================================================================
Parts of the map code have been inspired/changed from the code that Linden Lab
uses on their own maps website (https://maps.secondlife.com(),
under the following MIT-like license:
==============================================================================
License and Terms of Use

Copyright 2016 Linden Research, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

This javascript makes use of the Second Life Map API, which is documented
at http:wiki.secondlife.com/wiki/Map_API

Use of the Second Life Map API is subject to the Second Life API Terms of Use:
  https:wiki.secondlife.com/wiki/Linden_Lab_Official:API_Terms_of_Use

Questions regarding this javascript, and any suggested improvements to it,
should be sent to the mailing list opensource-dev@list.secondlife.com
==============================================================================

								-->
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