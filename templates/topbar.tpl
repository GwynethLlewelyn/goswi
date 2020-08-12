{{ define "topbar.tpl" }}
		<!-- Topbar -->
		<nav class="navbar navbar-expand navbar-light bg-white topbar mb-4 static-top shadow">

		  <!-- Sidebar Toggle (Topbar) -->
		  <button id="sidebarToggleTop" class="btn btn-link d-md-none rounded-circle mr-3">
			<i class="fa fa-bars"></i>
		  </button>

		  <!-- Topbar Search -->
		  <form class="d-none d-sm-inline-block form-inline mr-auto ml-md-3 my-2 my-md-0 mw-100 navbar-search" action="/search" method="post">
			<div class="input-group">
				<input type="text" class="form-control bg-light border-0 small" id="mainSearch" name="mainSearch" placeholder="Search for..." aria-label="Search" data-toggle="tooltip" data-placement="bottom-right" title="Type something to search in the OpenSimulator database (not all fields will be searched)">
			    <div class="input-group-append">
					<button class="btn btn-primary" type="submit">
						<i class="fas fa-search fa-sm"></i>
					</button>
				</div>
			</div>
		  </form>

		  <!-- Topbar Navbar -->
		  <ul class="navbar-nav ml-auto">
			<!-- Nav Item - Search Dropdown (Visible Only XS) -->
			<li class="nav-item dropdown no-arrow d-sm-none">
			  <a class="nav-link dropdown-toggle" href="#" id="searchDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
				<i class="fas fa-search fa-fw"></i>
			  </a>
			  <!-- Dropdown - Search -->
			  <div class="dropdown-menu dropdown-menu-right p-3 shadow animated--grow-in" aria-labelledby="searchDropdown">
				<form class="form-inline mr-auto w-100 navbar-search" action="/search" method="post">
					<div class="input-group">
						<input type="text" class="form-control bg-light border-0 small" id="search" name="search" placeholder="Search for..." aria-label="Search" data-toggle="tooltip" data-placement="bottom-right" title="Type something to search in the OpenSimulator database (not all fields will be searched)">
					<div class="input-group-append">
					  <button class="btn btn-primary" type="submit">
						<i class="fas fa-search fa-sm"></i>
					  </button>
					</div>
				  </div>
				</form>
			  </div>
			</li>
{{ if .Username }}
			<!-- Nav Item - Alerts -->
			<li class="nav-item dropdown no-arrow mx-1 disabled">
			  <a class="nav-link dropdown-toggle" href="#" id="alertsDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
				<i class="fas fa-bell fa-fw"></i>
				<!-- Counter - Alerts -->
				<span class="badge badge-danger badge-counter">3+</span>
			  </a>
			  <!-- Dropdown - Alerts -->
			  <div class="dropdown-list dropdown-menu dropdown-menu-right shadow animated--grow-in" aria-labelledby="alertsDropdown">
				<h6 class="dropdown-header">
				  Alerts Center
				</h6>
				<a class="dropdown-item d-flex align-items-center" href="#">
				  <div class="mr-3">
					<div class="icon-circle bg-primary">
					  <i class="fas fa-file-alt text-white"></i>
					</div>
				  </div>
				  <div>
					<div class="small text-gray-500">December 12, 2019</div>
					<span class="font-weight-bold">A new monthly report is ready to download!</span>
				  </div>
				</a>
				<a class="dropdown-item d-flex align-items-center" href="#">
				  <div class="mr-3">
					<div class="icon-circle bg-success">
					  <i class="fas fa-donate text-white"></i>
					</div>
				  </div>
				  <div>
					<div class="small text-gray-500">December 7, 2019</div>
					$290.29 has been deposited into your account!
				  </div>
				</a>
				<a class="dropdown-item d-flex align-items-center" href="#">
				  <div class="mr-3">
					<div class="icon-circle bg-warning">
					  <i class="fas fa-exclamation-triangle text-white"></i>
					</div>
				  </div>
				  <div>
					<div class="small text-gray-500">December 2, 2019</div>
					Spending Alert: We've noticed unusually high spending for your account.
				  </div>
				</a>
				<a class="dropdown-item text-center small text-gray-500" href="#">Show All Alerts</a>
			  </div>
			</li>

			<!-- Nav Item - Messages -->
			<li class="nav-item dropdown no-arrow mx-1 disabled">
			  <a class="nav-link dropdown-toggle" href="#" id="messagesDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
				<i class="fas fa-envelope fa-fw"></i>
				<!-- Counter - Messages -->
				<span class="badge badge-danger badge-counter">{{- if .Messages -}}{{- len .Messages -}}{{- end -}}</span>
			  </a>
			  {{ if .Messages }}
			  <!-- Dropdown - Messages -->
			  <div class="dropdown-list dropdown-menu dropdown-menu-right shadow animated--grow-in" aria-labelledby="messagesDropdown" >
				<h6 class="dropdown-header">
				  Offline Instant Message Center
				</h6>
				{{ range .Messages }}
				<a class="dropdown-item d-flex align-items-center" href="#">
					<div class="dropdown-list-image mr-3">
						<img class="rounded-circle" src="{{- .Libravatar -}}" alt="{{- .FromID -}}">
						<div class="status-indicator bg-success"></div>
					</div>
					<div class="font-weight-bold">
						<div class="text-truncate">{{- .Message -}}</div>
						<div class="small text-gray-500">{{- .Username -}} · {{- .TMStamp -}}</div>
					</div>
				</a>
				{{ else }}
				<a class="dropdown-item d-flex align-items-center" href="#">
					<div class="dropdown-list-image mr-3">
						<i class="fas fa-ban fa-3x"></i>
						<div class="status-indicator bg-danger"></div>
						<div class="font-weight-bold">
							<div class="text-truncate">No Offline Instant Messages</div>
							<div class="small text-gray-500">&nbsp;</div>
						</div>
					</div>
				</a>
				<a class="dropdown-item text-center small text-gray-500" href="#">Read More Messages</a>
				{{ end }}
			  </div>
			  {{ end }}
			</li>

			<li class="topbar-divider d-none d-sm-block"></li>

			<!-- Nav Item - User Information -->
			<li class="nav-item dropdown no-arrow">
				<a class="nav-link dropdown-toggle" href="#" id="userDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
					<span class="mr-2 d-none d-lg-inline text-gray-600 small">{{- .Username -}}</span>
					<img class="img-profile rounded-circle" alt="Avatar" src="{{- .Libravatar -}}">
				</a>
				  <!-- Dropdown - User Information -->
				<div class="dropdown-menu dropdown-menu-right shadow animated--grow-in" aria-labelledby="userDropdown">
					<a class="dropdown-item" href="/user/profile">
						<i class="fas fa-user fa-sm fa-fw mr-2 text-gray-400"></i>
						Profile
					</a>
					<a class="dropdown-item disabled" href="#">
						<i class="fas fa-cogs fa-sm fa-fw mr-2 text-gray-400"></i>
						Settings
					</a>
					<a class="dropdown-item disabled" href="#">
						<i class="fas fa-list fa-sm fa-fw mr-2 text-gray-400"></i>
						Activity Log
					</a>
					<a class="dropdown-item" href="/user/change-password">
						<i class="fas fa-unlock-alt fa-sm fa-fw mr-2 text-gray-400"></i>
						Change password
					</a>
					<div class="dropdown-divider"></div>
					<a class="dropdown-item" href="/user/logout" data-toggle="modal" data-target="#logoutModal">
						<i class="fas fa-sign-out-alt fa-sm fa-fw mr-2 text-gray-400"></i>
						Logout
					</a>
				</div> <!-- ./Dropdown -->
			</li>
{{ else }}
			<li>
				<a href="/user/login" class="d-none d-sm-inline-block btn btn-sm btn-primary shadow-sm"><i class="fas fa-sign-in-alt fa-sm text-white-50"></i>&nbsp;Log in</a>
			</li>
{{ end }}
		  </ul>

		</nav>
		<!-- End of Topbar -->
{{ end }}