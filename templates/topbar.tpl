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
				<a class="nav-link dropdown-toggle" href="#" id="alertsDropdown" role="button" data-toggle="dropdown" aria haspopup="true" aria-expanded="false">
					<i class="fas fa-bell fa-fw"></i>
					<!-- Counter - Alerts -->
					{{- if .FeedMessages -}}<span class="badge badge-danger badge-counter">{{- .numberFeedMessages -}}</span>{{- end -}}
				</a>
				{{ if .FeedMessages }}
				<!-- Dropdown - Alerts -->
				<div class="dropdown-list dropdown-menu dropdown-menu-right shadow animated--grow-in" aria-labelledby="alertsDropdown">
					<h6 class="dropdown-header">
						Feed Notifications Center
					</h6>
					{{ range $i, $e := .FeedMessages }}
					<a class="dropdown-item d-flex align-items-center not-msg-{{- $i -}}" href="#">
						<div class="mr-3">
							<img class="rounded-circle" src="{{- .Libravatar -}}" alt="{{- .Username -}}">
							<div class="small text-gray-500 text-center">{{- .Feedgroup -}}</div>
						</div>
						<div>
							<span class="font-weight-normal">From:&nbsp;</span><span class="font-weight-bolder">{{- .Username -}}</span>
							<div class="small text-gray-500 mb-1">{{- .Chronostamp -}}</div>
							<span class="font-weight-normal">{{- .PostMarkup -}}</span>
						</div>
					</a>
					{{ else }}
					<a class="dropdown-item d-flex align-items-center" href="#">
						<div class="mr-3">
							<div class="icon-circle bg-primary">
								<i class="fas fa-bell-slash text-white bg-danger"></i>
							</div>
						</div>
						<div>
							<span class="font-weight-normal">No notifications</span>
						</div>
					</a>
					{{ end }}
					{{- if gt .numberFeedMessages 0 -}}<a class="dropdown-item text-center small text-gray-500" href="#">Show All Notifications</a>{{- end -}}
				</div>
				{{ end }}
			</li>

			<!-- Nav Item - Messages -->
			<li class="nav-item dropdown no-arrow mx-1 disabled">
				<a class="nav-link dropdown-toggle" href="#" id="messagesDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
					<i class="fas fa-envelope fa-fw"></i>
					<!-- Counter - Messages -->
					{{- if .Messages -}}<span class="badge badge-danger badge-counter">{{- .numberMessages -}}</span>{{- end -}}
				</a>
			{{ if .Messages }}
				<!-- Dropdown - Messages -->
				<div class="dropdown-list dropdown-menu dropdown-menu-right shadow animated--grow-in" aria-labelledby="messagesDropdown" >
					<h6 class="dropdown-header">
					Offline Instant Message Center
					</h6>
					{{ range $i, $e := .Messages }}
					<a class="dropdown-item d-flex align-items-center im-msg-{{- $i -}}" href="#">
						<div class="dropdown-list-image mr-3">
							<img class="rounded-circle" src="{{- .Libravatar -}}" alt="{{- .Username -}}">
							<div class="status-indicator bg-success"></div>
						</div>
						<div class="font-weight-normal">
							<div class="text-truncate">{{- .Message -}}</div>
							<div class="small text-gray-500">{{- .Username -}}{{- if .TMStamp }}&nbsp;·&nbsp;{{ .TMStamp -}}{{- end -}}</div>
						</div>
					</a>
					{{ else }}
					<a class="dropdown-item d-flex align-items-center" href="#">
						<div class="dropdown-list-image mr-3">
							<i class="fas fa-comment-slash fa-3x text-white bg-danger"></i>
						</div>
						<div class="font-weight-normal">
							<div class="text-truncate">No Offline Instant Messages</div>
							<div class="small text-gray-500">&nbsp;</div>
						</div>
					</a>
					{{ end }}
					{{- if gt .numberMessages 0 -}}<a class="dropdown-item text-center small text-gray-500" href="#">Read More Messages</a>{{- end -}}
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