{{- define "header.tpl" -}}<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<meta name="description" content="{{- .description -}}">
	<meta name="author" content="{{- .author -}}">

	<title>{{- .titleCommon -}}</title>

	<!-- Custom fonts for this template-->
	<link href="../lib/startbootstrap-sb-admin-2/vendor/fontawesome-free/css/all.min.css" rel="stylesheet" type="text/css">

	<!-- Google fonts -->
	<link rel="preconnect" href="https://fonts.gstatic.com">
	<link href="https://fonts.googleapis.com/css2?family=Fira+Code&family=Hind+Guntur:wght@400;600&display=swap" rel="stylesheet">
	<link href="/assets/css/goswi.css" rel="stylesheet" type="text/css">
	{{- if .needsTables -}}
	<!-- Custom styles for this page -->
	<link href="../lib/startbootstrap-sb-admin-2/vendor/datatables/dataTables.bootstrap4.min.css" rel="stylesheet" type="text/css">
	<!-- change the style above for more compact tables -->
	<style>
	table.dataTable.table-compact thead .sorting:before,
	table.dataTable.table-compact thead .sorting:after,
	table.dataTable.table-compact thead .sorting_asc:before,
	table.dataTable.table-compact thead .sorting_asc:after,
	table.dataTable.table-compact thead .sorting_desc:before,
	table.dataTable.table-compact thead .sorting_desc:after,
	table.dataTable.table-compact thead .sorting_asc_disabled:before,
	table.dataTable.table-compact thead .sorting_asc_disabled:after,
	table.dataTable.table-compact thead .sorting_desc_disabled:before,
	table.dataTable.table-compact thead .sorting_desc_disabled:after {
		    position: absolute;
		    bottom: 0.025rem;
		    display: block;
		    opacity: 0.3;
	}
	table.dataTable.table-compact thead th,
	table.dataTable.table-compact thead td {
		padding: 0.1rem;
		font-size: 0.7rem;
		line-height: 1;
		// padding: 1px;
	}
	table.dataTable.table-compact tfoot th,
	table.dataTable.table-compact tfoot td {
		padding: 0.1rem;
		font-size: 0.7rem;
		line-height: 1;
		// padding: 1px;
	}
	table.dataTable.table-compact tbody th,
	table.dataTable.table-compact tbody td {
		padding: 0.1rem;
		font-size: 0.7rem;
		line-height: 1;
		// padding: 1px;
	}
	</style>
	{{- end -}}
	{{- if .needsMap -}}
		<!-- Leaflet -->
		<!-- old version in case we need to fall back
		<link rel="stylesheet" href="https://unpkg.com/leaflet@1.6.0/dist/leaflet.css"
 integrity="sha512-xwE/Az9zrjBIphAcBb3F6JVqxf46+CDLwfLMHloNu6KEQCAWi6HcDUbeOfBIptF7tcCzusKFjFw2yuvEpDL9wQ=="
 crossorigin=""/>
		<script src="https://unpkg.com/leaflet@1.6.0/dist/leaflet.js"
 integrity="sha512-gZwIG9x3wUXg2hdXF6+rVkLF/0Vi9U8D2Ntg4Ga5I5BZpVkVxlJWbSQtXPSiUTtC0TjtGOmxa1AJPuV0CPthew=="
 crossorigin=""></script>-->
 		<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=" crossorigin="">
		<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" integrity="sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo=" crossorigin=""></script>
 		<style>
			#gridMap { height: 40rem; }
			.leaflet-container {
				font-family: "Hind Guntur", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
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
	{{- end -}}
</head>
{{- if not .logintemplate -}}
<body id="page-top">
	<!-- Page Wrapper -->
	<div id="wrapper">
{{- end -}}
{{ end }}