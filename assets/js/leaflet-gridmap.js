/**
*	Based on code by Linden Lab for Second Life® and code by hawddamor for opensimmaps
*	The rest is basically copy & pasting by Gwyneth Llewelyn
*/

// Make sure that leaflet.js and leaflet.css have already been loaded!

/**
*	Set the MAP_URL to wherever you have your tiles; you can call them on :8002 if you have gOSWI *and* ROBUST running under http
*	(or both under https); Chrome will complain otherwise.
*/
const GRID_PROTOCOL = 'http://';
const GRID_URL = 'opensim.betatechnologies.info';
const GRID_PORT = ':8002';
const MAP_PROTOCOL = 'https://';
const MAP_URL = MAP_PROTOCOL + GRID_URL + '/maptiles/00000000-0000-0000-0000-000000000000';
const ATTRIBUTION_LINE = 'Map data © 2020 by Beta Technologies';
const showUUID = false;	// if true, this will also show region UUIDs on the popups
const LOADING = "(Loading...)";
const NO_DATA_YET = "[No data yet]";


// Leave those min and max zoom levels as they are, since these are the default zoom levels for OpenSimulator grids.
const mapMinZoom = 1;
const mapMaxZoom = 8;

/**
*	Fetch region data on initialise
*	This is a clever workaround which is acceptable for small grids
*	LL calls a special function/script which returns the region name based on the x,y coords; hawddamor prefers to load
*	the _whole_ grid data as a XML object. The advantage is that we can cache it at some point
*	To-do: we could run a Go timer to save the grid data periodically and simply include the output in a Go template;
*		this would be WAY more efficient for large grids!
*/
var __items;	// this includes the grid data, namely, a way to get region names from coordinates
				// it will be populated once the map loads (hopefully!)

/**
*	getRegionInfo will return a properly formatted table with a region's name, coords, and links (SLURLs).
*	Mostly inspired by https://github.com/hawddamor/opensimmaps/blob/master/url.js (lines 502 ff.)
*/
function getRegionInfo(x, y, xjump, yjump) {
	if (__items == null) {
//		console.log(NO_DATA_YET);
		return NO_DATA_YET;
	}
// 	console.log("Items are:", __items, "x is:", x, "y is:", y);
	var xmllocX;
	var xmllocY;
	for (var i = 0; i < __items.length; i++) {
		if (__items[i].nodeType === 1) {
			xmllocX = __items[i].getElementsByTagName("LocX")[0].firstChild.nodeValue;
			xmllocY = __items[i].getElementsByTagName("LocY")[0].firstChild.nodeValue;
			if (xmllocX == x && xmllocY == y) {
				var xmluuid = __items[i].getElementsByTagName("Uuid")[0].firstChild.nodeValue;
				var xmlregionname = __items[i].getElementsByTagName("RegionName")[0].firstChild.nodeValue;
				var response = "<table>";
				response += "<tr><td colspan='3'><span id='name'><strong>" + xmlregionname + "</strong></span>"
					+ "&nbsp;<span id='loc'>(" + xmllocX + ", " + xmllocY + ")</span></td></tr>";
				if (showUUID === true) {
					response += "<tr><td>Region UUID:\n" + xmluuid + "</td></tr>";
				}
				response += "<tr><td colspan='3'></td></tr>";
				response += "<tr><td><a class='add' href='secondlife://" + GRID_URL + GRID_PORT + "/" + xmlregionname
					+ "/" + xjump + "/" + yjump + "/'>Hypergrid</a>&nbsp;&nbsp;</td>";
				xmlregionname = xmlregionname.replace(" ", "+"); // fix for V3 HG URL
				response += "<td><a class='add' href='secondlife://http|!!" + GRID_URL + GRID_PORT + "/+" + xmlregionname
					+ "'>V3 HG</a>&nbsp;&nbsp;</td>";
				xmlregionname = xmlregionname.replace("+"," "); // change back for local URL
				response += "<td><a class='add' href='secondlife://" + xmlregionname + "/" + xjump + "/" + yjump
					+ "/'>Local</a></td></tr>";
				if (xjump > 255 || yjump > 255) {
					response += "</table><table><tr><td colspan='3'>Viewer may restrict login within SE 256x256 corner </td></tr><tr><td>of larger regions in OpenSim/WhiteCore/Aurora</td></tr>";
				}
				response += "</table>";
				return response;
			}
		}
	}
	return "";	// if we got here, that region is _not_ in our grid...
}

/**
*	Same code as above, but just for getting the region name and nothing else
**/
function getRegionName(x, y) {
	if (__items == null) {
//		console.log(LOADING);
		return LOADING;
	}
	var xmllocX;
	var xmllocY;
	for (var i = 0; i < __items.length; i++) {
		if (__items[i].nodeType === 1) {
			xmllocX = __items[i].getElementsByTagName("LocX")[0].firstChild.nodeValue;
			xmllocY = __items[i].getElementsByTagName("LocY")[0].firstChild.nodeValue;
			if (xmllocX == x && xmllocY == y) {
				return __items[i].getElementsByTagName("RegionName")[0].firstChild.nodeValue;
			}
		}
	}
	return "";
}

/**
*	This extension of the Leaflet tile layer is needed because the tile numbering is not obvious and quite different
*	from 'normal' Leaflet-compatible maps (such as Google Maps, OpenStreetMap, etc. or
*	any tiles generated from automated tile generators.
*/
var OSTileLayer = L.TileLayer.extend({
	getTileUrl: function(coords) {
		var data = {
			r: (this.options.detectRetina && L.Browser.retina && this.options.maxZoom > 0) ? '@2x' : '',
			s: this._getSubdomain(coords),
			z: this._getZoomForUrl()
		};

		var regionsPerTileEdge = Math.pow(2, data.z - 1);
		data.x = coords.x * regionsPerTileEdge;
		data.y = (Math.abs(coords.y) - 1) * regionsPerTileEdge;

		return L.Util.template(this._url, L.extend(data, this.options));
	}
});

/**
*	And this is a new layer just to place the tile names (i. e. the region names)
*	See https://stackoverflow.com/a/48080981/1035977
*/
var GridInfo = L.GridLayer.extend({
	// called for each tile
	// returns a DOM node containing whatver you want
	createTile: function (coords) {
		// create a div
		var tile = document.createElement('div');
		tile.className = "regionNameTile";
		// tile.style.outline = '1px solid black';

		// make sure we have this right
		var osX = Math.abs(coords.x);
		var osY = Math.abs(coords.y) - 1;

		// lookup the piece of data you want
		var regionName = getRegionName(osX, osY);
		if (regionName == "" || regionName == LOADING) {
			tile.className += " water";
		} else {
			tile.className += " " + encodeURI(regionName);	// this is to avoid class names with spaces... bug: what about 2 regions with the same name??
		}
//		console.log("Tile overlay at coords ", osX, ",", osY, "regionName:", regionName);

		// let's add the lat/lng of the center of the tile
		var tileBounds = this._tileCoordsToBounds(coords);
		var center = tileBounds.getCenter();


		// If we _know_ that the region does not exist, don't print anything (it won't be clickable, either)
		// If we are still loading the region names, we don't know if the region exists or not, so print (Loading...) and the coords
		// If we have all data loaded, then print the regionName and the coords
		// Note that a tile refresh, which happens occasionally, will slowly change this layer's data
		tile.innerHTML = '<span>' + ((regionName == "") ? "" : (regionName + "<br /><span style='font-size: smaller'>(" + osX + "," + osY + ")</font>")) +
//			'<br /><span style="font-size: smaller">Lat:&nbsp;' + center.lat + '&nbsp;Lng: ' + center.lng + '&nbsp;Zoom: ' + coords.z + '</span>' +
			'</span>';

		return tile;
	}
});
// Prepare the layer which will contain all labels with the region names and coords
var gridInfoLayer = new GridInfo();

// Prepare tile map, using extended class above
var tiles = new OSTileLayer(MAP_URL + '/map-{z}-{x}-{y}-objects.jpg', {
	crs: L.CRS.Simple,	// "Toto, I've a feeling we're not in Kansas anymore."
	maxZoom: mapMaxZoom,
	minZoom: mapMinZoom,
	zoomOffset: 1,		// under SL and OpenSim, the zoom goes from 0-7
	zoomReverse: true,	// SL and OpenSim count the zoom levels backwards!
	bounds: [
		[0, 0],
		[1048576, 1048576] // according to LL, this is the 'maximum' size for the whole grid
	],
//	continuousWorld: true,
//	noWrap: true,
	attribution: ATTRIBUTION_LINE
});
// Create map using the toles above
var map = L.map('gridMap', {
	crs: L.CRS.Simple,
	minZoom: mapMinZoom,
	maxZoom: mapMaxZoom,
	maxBounds: [
		[0, 0],
		[1048576, 1048576] // see comments above
	],
	layers: [tiles, gridInfoLayer],
	attributionControl: true
})
.on('load', function(event) {
// 	console.log('inside on map event load, trying to call', MAP_PROTOCOL + GRID_URL + "/mapdata");
	var request = new XMLHttpRequest();

	if (request) {
		console.log('XML request succeeded, inside handler');
		request.onreadystatechange = function() {
			if (request.readyState == 4) {
				if (request.status == 200 || request.status == 304) {
					var xmlGridData = request.responseXML;
//					console.log("Full Grid Data:", xmlGridData);	// For debugging purposes
					console.log("Grid data has arrived!");
					var root = xmlGridData.getElementsByTagName('Map')[0];
					if (root == null) {
						console.log("[ERROR]: Grid data has no root element 'Map'");
						return;
					}
					__items = root.getElementsByTagName("Grid");
				}
			}
		};
// 		console.log('Trying to GET', MAP_PROTOCOL + GRID_URL + "/mapdata");
		request.open("GET", MAP_PROTOCOL + GRID_URL + "/mapdata", true);
		request.send(null);
	} else {
		console.log("Getting a new XMLHttpRequest failed!");
	}
})
.on('click', function(event) {
	// calculations to get the grid coords & region coords
	// Note: varregions not supported for now, only 256x256 ones
	var popLocation = event.latlng;

	var x = event.latlng.lng;
	var y = event.latlng.lat;
	// Work out region co-ords, and local co-ords within region
	var grid_x = Math.floor(x);
	var grid_y = Math.floor(y);

	var local_x = Math.round((x - grid_x) * 256);
	var local_y = Math.round((y - grid_y) * 256);

	var regionInfo = getRegionInfo(grid_x, grid_y, local_x, local_y);

	if (regionInfo != "") {	// we have no info for this region, which is displayed as empty water
		var popup = L.popup()
			.setLatLng(popLocation)
			.setContent(regionInfo)
			.openOn(map);
	}
})
.setView([3650, 3650], mapMinZoom)
.on('zoomend', function () { // inspired by https://stackoverflow.com/a/23021470/1035977
	if (this.getZoom() < mapMaxZoom && this.hasLayer(gridInfoLayer)) {
		this.removeLayer(gridInfoLayer);
	}
	if (this.getZoom() == mapMaxZoom && this.hasLayer(gridInfoLayer) == false)
	{
		this.addLayer(gridInfoLayer);
	}
})
.flyTo([3650, 3650], mapMaxZoom);	// do a cute animation to give time for the grid data to be loaded...
