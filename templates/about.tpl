{{ define "about.tpl" }}
{{ template "header.tpl" .}}
{{ template "navigation.tpl" .}}
    <!-- Content Wrapper -->
    <div id="content-wrapper" class="d-flex flex-column">

      <!-- Main Content -->
      <div id="content">

{{ template "topbar.tpl" .}}

        <!-- Begin Page Content -->
        <div class="container-fluid">

          <!-- 404 Error Text -->
          <div class="text-center">
		  	<h1><i class="fas fa-fw fa-address-card"></i>About the Beta Technologies OpenSimulator Grid</h1>
		  	<a href="https://betatechnologies.info" target=_blank><img src="https://betatechnologies.info/wp-content/uploads/2020/05/Beta-Technologies-Vertical-Logo-2008.svg" alt="Beta Technologies Logo" align="left" width="250"></a>
		  	<p>Sorry about (pun intended) that, but we still don't have a lot of content around here...</p>
            <a href="/">&larr; Back to Dashboard</a>
          </div>

        </div>
        <!-- /.container-fluid -->

      </div>
      <!-- End of Main Content -->

{{ template "footer.tpl" .}}
{{ end }}