{{ define "header.tpl" }}
<!DOCTYPE html>
<html lang="en">

<head>

	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<meta name="description" content="{{ .description }}">
	<meta name="author" content="{{ .author }}">

	<title>Beta Technologies OpenSimulator Grid - Dashboard</title>

	<!-- Custom fonts for this template-->
	<link href="../lib/startbootstrap-sb-admin-2/vendor/fontawesome-free/css/all.min.css" rel="stylesheet" type="text/css">
	<link href="https://fonts.googleapis.com/css?family=Nunito:200,200i,300,300i,400,400i,600,600i,700,700i,800,800i,900,900i" rel="stylesheet">

	<!-- Custom styles for this template-->
	<link href="../lib/startbootstrap-sb-admin-2/css/sb-admin-2.min.css" rel="stylesheet">
	<style>
		.sidebar .sidebar-brand .sidebar-brand-icon img {
			height: 2rem;
		}
		@font-face {
			font-display: swap;
			font-family: CCSymbols;
			font-synthesis: none;
			src: url(../images/fonts/CCSymbols.woff2) format(woff2),
				 url(../images/fonts/CCSymbols.woff)  format(woff);
			unicode-range: u+a9, u+229c,
			               u+1f10d-f, u+1f16d-f;
		}
	</style>
	{{ if .needsTables }}
	<!-- Custom styles for this page -->
	<link href="../lib/startbootstrap-sb-admin-2/vendor/datatables/dataTables.bootstrap4.min.css" rel="stylesheet">
	{{ end }}
</head>

<body id="page-top">

	<!-- Page Wrapper -->
	<div id="wrapper">
{{ end }}