{{- define "header.tpl" -}}<!DOCTYPE html>
<html lang="en">

<head>

	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<meta name="description" content="{{ .description }}">
	<meta name="author" content="{{ .author }}">

	<title>{{ .titleCommon }}</title>

	<!-- Custom fonts for this template-->
	<link href="../lib/startbootstrap-sb-admin-2/vendor/fontawesome-free/css/all.min.css" rel="stylesheet" type="text/css">
	<link href="https://fonts.googleapis.com/css?family=Nunito:200,200i,300,300i,400,400i,600,600i,700,700i,800,800i,900,900i" rel="stylesheet" type="text/css">

	<!-- Custom styles for this template-->
	<link href="../lib/startbootstrap-sb-admin-2/css/sb-admin-2.min.css" rel="stylesheet" type="text/css">
	<style>
		// Colour settings for main components of SB Admin 2. We're replacing them with Nord colours!
		@import url("../assets/css/nord.css");
		:root {
			--blue: var(--nord10);
			--indigo: var(--nord2);
			--purple: var(--nord3);
			--pink: var(--nord15);
			--red: var(--nord11);
			--orange: var(--nord12);
			--yellow: var(--nord13);
			--green: var(--nord14);
			--teal: var(--nord7);
			--cyan: var(--nord9);
			--white: var(--nord6);
			--gray: var(--nord3);
			--gray-dark: var(--nord2);
			--primary: var(--nord10);
			--secondary: var(--nord3);
			--success: var(--nord14);
			--info: var(--nord9);
			--warning: var(--nord13);
			--danger: var(--nord11);
			--light: var(--nord6);
			--dark: var(--nord0);
		}
		.sidebar .sidebar-brand .sidebar-brand-icon img {
			height: 2rem;
		}
		@font-face {
			font-display: swap;
			font-family: CCSymbols;
			font-synthesis: none;
			src: url(../assets/fonts/CCSymbols.woff2) format(woff2),
				 url(../assets/fonts/CCSymbols.woff)format(woff);
			unicode-range: u+a9, u+229c,
							u+1f10d-f, u+1f16d-f;
		}
		/* #### Generated By: http://www.cufonfonts.com #### */
		@font-face {
			font-family:'Segoe UI Regular';
			font-style:normal;
			font-weight:400;
			src:local('Segoe UI Regular'),url('../assets/fonts/Segoe UI.woff') format("woff");
		}

		@font-face {
			font-family:'Segoe UI Italic';
			font-style:normal;
			font-weight:400;
			src:local('Segoe UI Italic'),url('../assets/fonts/Segoe UI Italic.woff') format("woff");
		}

		@font-face {
			font-family:'Segoe UI Bold';
			font-style:normal;
			font-weight:400;
			src:local('Segoe UI Bold'),url('../assets/fonts/Segoe UI Bold.woff') format("woff");
		}

		@font-face {
			font-family:'Segoe UI Bold Italic';
			font-style:normal;
			font-weight:400;
			src:local('Segoe UI Bold Italic'),url('../assets/fonts/Segoe UI Bold Italic.woff') format("woff");
		}
	</style>
	{{ if .needsTables }}
	<!-- Custom styles for this page -->
	<link href="../lib/startbootstrap-sb-admin-2/vendor/datatables/dataTables.bootstrap4.min.css" rel="stylesheet" type="text/css">
	{{ end }}
	{{ if .needsMap }}
		<!-- Leaflet -->
		<link rel="stylesheet" href="https://unpkg.com/leaflet@1.6.0/dist/leaflet.css"
 integrity="sha512-xwE/Az9zrjBIphAcBb3F6JVqxf46+CDLwfLMHloNu6KEQCAWi6HcDUbeOfBIptF7tcCzusKFjFw2yuvEpDL9wQ=="
 crossorigin=""/>
		<script src="https://unpkg.com/leaflet@1.6.0/dist/leaflet.js"
 integrity="sha512-gZwIG9x3wUXg2hdXF6+rVkLF/0Vi9U8D2Ntg4Ga5I5BZpVkVxlJWbSQtXPSiUTtC0TjtGOmxa1AJPuV0CPthew=="
 crossorigin=""></script>
 		<style>
			#gridMap { height: 40rem; }
			.leaflet-container {
				font-family: "Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
				background-color: #1e465f;
			}
			.regionNameTile {display: flex;}
			.regionNameTile span {
				font-size: small;
				text-align: left;
				color: white;
				margin: auto;
				margin-bottom: 0em;
				margin-left: 0em;
		}
		</style>
	{{ end }}
</head>
{{- if not .logintemplate -}}
<body id="page-top">

	<!-- Page Wrapper -->
	<div id="wrapper">
{{- end -}}
{{ end }}