{{ define "navigation.tpl" }}
	<!-- Sidebar -->
	<ul class="navbar-nav bg-gradient-primary sidebar sidebar-dark accordion toggled" id="accordionSidebar">

		<!-- Sidebar - Brand -->
		<a class="sidebar-brand d-flex align-items-center justify-content-center" href="/">
			<div class="sidebar-brand-icon">
				<img src="{{- .logo -}}">
			</div>
			<div class="sidebar-brand-text mx-3">{{- .logoTitle -}}</div>
		</a>

		<!-- Divider -->
		<hr class="sidebar-divider my-0">

		<!-- Nav Item - Dashboard -->
		<li class="nav-item active">
			<a class="nav-link" href="/">
				<i class="fas fa-fw fa-tachometer-alt"></i>
				<span>Dashboard</span></a>
		</li>

		<!-- Divider -->
		<hr class="sidebar-divider">

		<!-- Heading -->
		<div class="sidebar-heading">
		Available pages
		</div>

		<!-- Nav Item - Welcome -->
		<li class="nav-item">
			<a class="nav-link" href="/welcome">
				<i class="fas fa-fw fa-door-open"></i>
				<span>Welcome</span></a>
		</li>

		<!-- Nav Item - About -->
		<li class="nav-item">
			<a class="nav-link" href="/about">
				<i class="fas fa-fw fa-address-card"></i>
				<span>About</span></a>
		</li>

		<!-- Nav Item - Help -->
		<li class="nav-item">
			<a class="nav-link" href="/help">
				<i class="fas fa-fw fa-question-circle"></i>
				<span>Help</span></a>
		</li>

		<!-- Nav Item - Economy -->
		<li class="nav-item disabled">
			<a class="nav-link" href="/economy">
				<i class="fas fa-fw fa-hand-holding-usd"></i>
				<span>Economy</span></a>
		</li>
{{ if not .Username }}
		<!-- Nav Item - Register -->
		<li class="nav-item disabled">
			<a class="nav-link" href="/user/register">
				<i class="fas fa-fw fa-user-plus"></i>
				<span>Register new resident</span></a>
		</li>
{{ end }}
{{ if .Username }}
	 <!-- Nav Item - Password -->
		<li class="nav-item">
			<a class="nav-link" href="/user/change-password">
				<i class="fas fa-fw fa-unlock-alt"></i>
				<span>Change password</span></a>
		</li>
{{ end }}

		<!-- Divider -->
		<hr class="sidebar-divider d-none d-md-block">

		<!-- Sidebar Toggler (Sidebar) -->
		<div class="text-center d-none d-md-inline">
			<button class="rounded-circle border-0" id="sidebarToggle"></button>
		</div>
	</ul>
	<!-- End of Sidebar -->
{{ end }}