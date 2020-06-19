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
				
// #### Function to return information for infoWindow ####
function getRegionInfo(x, y, xjump, yjump) {
	if (__items == null) { 
		console.log("No data yet!");
		return "[No data yet]"; 
	}
// 	console.log("Items are:", __items, "x is:", x, "y is:", y);
	var response = "";
	var i;
	var xmllocX;
	var xmllocY;
	var xmluuid;
	var xmlregionname;
	for (i = 0; i < __items.length; i++/* += 1*/) {
		if (__items[i].nodeType === 1) {
			xmllocX = __items[i].getElementsByTagName("LocX")[0].firstChild.nodeValue;
			xmllocY = __items[i].getElementsByTagName("LocY")[0].firstChild.nodeValue;
			if (xmllocX == x && xmllocY == y) {
				xmluuid = __items[i].getElementsByTagName("Uuid")[0].firstChild.nodeValue;
				xmlregionname = __items[i].getElementsByTagName("RegionName")[0].firstChild.nodeValue;
				response = "<table>";
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
			}
		}
	}
	return response;
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
	layers: [tiles],
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
					console.log("Full Grid Data:", xmlGridData);	// For debugging purposes
					var root = xmlGridData.getElementsByTagName('Map')[0];
					if (root == null) { return; }
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

	var popup = L.popup()
		.setLatLng(popLocation)
		.setContent(getRegionInfo(grid_x, grid_y, local_x, local_y))
		.openOn(map);
})
.setView([3650, 3650], mapMaxZoom);