{{- define "profile.tpl" -}}
{{- template "header.tpl" . -}}
{{- template "navigation.tpl" . -}}
	<!-- Content Wrapper -->
	<div id="content-wrapper" class="d-flex flex-column">
		<!-- Main Content -->
		<div id="content">
{{- template "topbar.tpl" . -}}
		<!-- Begin Page Content -->
		<div class="container-fluid">
			<!-- Content Row -->
			<div class="row">
				<div class="col-12 mb-4">
					{{- if .BoxTitle -}}
					{{ template "infobox.tpl" .}}
					{{- end -}}
					<form class="well form-horizontal" action="/user/profile" method="post"  id="profileForm">
						<fieldset>
							<!-- Your Profile: UserName -->
							<legend class="text-center"><h2>{{- if .Username -}}{{- .Username -}}'s{{- else -}}Your{{- end -}}&nbsp;Profile</h2></legend><br />
							<!-- About -->
							<div class="form-group">
								<label for="AboutText" id="labelAboutText" class="control-label">About you</label>
								<div class="input-group">
									<div class="image pr-2">
										<!-- this will be the user's mugshot -->
										{{ if .ProfileURL}}
										<a href="{{- .ProfileURL -}}" target="_blank">
											{{ end -}}
											{{- if .ProfileImage }}
											<img class="rounded shadow" src="{{- .ProfileImage -}}" alt="{{- .Username }} ({{- .UserUUID -}})" height="256" width="256" srcset="{{- .ProfileImage }} 1x, {{ .ProfileRetinaImage }} 2x">
											{{ else }}
											<img class="rounded shadow" src="{{- .Libravatar -}}" alt="{{ .Username }} ({{- .UserUUID -}}" height="256" width="256">
											{{ end -}}
											{{- if .ProfileURL}}
										</a>
										{{ end }}
									</div>
    								<textarea class="form-control" id="AboutText" rows="10" aria-describedby="labelAboutText">{{- .ProfileAboutText -}}</textarea>
								</div>
  							</div>
							<!-- ProfileURL -->
							<div class="form-group">
								<label for="ProfileURL" id="labelProfileURL" class="control-label">Profile URL</label>
								<div class="input-group">
									<div class="input-group-prepend">
										<span class="input-group-text bg-primary border-right-0 border-primary"><a href="{{- .ProfileURL -}}" target="_blank"><i class="fas fa-globe fa-fw text-light"></i></a></span>
									</div>
									<input id="ProfileURL" name="ProfileURL" value="{{- .ProfileURL -}}" class="form-control" type="url" aria-describedby="labelProfileURL">
								</div>
							</div>
							<!-- Partner -->
							<div class="form-group">
								<label for="ProfilePartner" id="labelProfilePartner" class="control-label">Partner</label>
								<div class="input-group">
									<div class="input-group-prepend">
										<span class="input-group-text bg-primary border-right-0 border-primary"><i class="fas fa-user fa-fw text-light"></i></span>
									</div>
									<input id="ProfilePartner" name="ProfilePartner" value="{{- .ProfilePartner -}}" class="form-control" type="text" aria-describedby="labelProfilePartner">
								</div>
							</div>
							<!-- Checkboxes for Publishing & Mature -->
							<div class="form-row">
								<div class="form-group col-3 mt-2 mb-2">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="ProfileAllowPublish" {{ if ne .ProfileAllowPublish 0 -}}checked{{- end -}}>
										<label class="form-check-label" for="ProfileAllowPublish">Allow publishing?</label>
									</div>
								</div>
								<div class="form-group col-3 mt-2 mb-2">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="ProfileMaturePublish" {{ if ne .ProfileMaturePublish 0 -}}checked{{- end -}}>
										<label class="form-check-label" for="ProfileMaturePublish">Mature profile?</label>
									</div>
								</div>
							</div>
							<hr>
							<!-- Want to... -->
							<label for="WantToLeft" id="labelProfileWantToText">I Want To:</label>
							<div class="form-row">
								<div class="form-group col-3 mt-2 mb-2" id="WantToLeft">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Build" id="WantToBuild" {{ if (bitTest .ProfileWantToMask 1) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToBuild">Build</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Meet" id="WantToMeet" {{ if (bitTest .ProfileWantToMask 4) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToMeet">Meet</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Group"  id="WantToGroup" {{ if (bitTest .ProfileWantToMask 8) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToMeet">Group</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Sell" id="WantToSell" {{ if (bitTest .ProfileWantToMask 32) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToSell">Sell</label>
									</div>
								</div>
								<div class="form-group col-3 mt-2 mb-2" id="WantToRight">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Explore" id="WantToExplore" {{ if (bitTest .ProfileWantToMask 2) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToExplore">Explore</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="BeHired" id="WantToBeHired" {{ if (bitTest .ProfileWantToMask 64) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToBeHired">Be Hired</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Buy" id="WantToBuy" {{ if (bitTest .ProfileWantToMask 16) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToBuy">Buy</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" name="WantTo[]" value="Hire" id="WantToHire" {{ if (bitTest .ProfileWantToMask 128) -}}checked{{- end -}}>
										<label class="form-check-label" for="WantToHire">Hire</label>
									</div>
								</div>
							</div>
							<div class="form-row">
								<div class="form-group col-6 mb-1" id="inputProfileWantToText">
									<div class="input-group">
										<div class="input-group-prepend">
											<span class="input-group-text bg-primary border-right-0 border-primary"><i class="fas fa-hand-holding fa-fw text-light"></i></span>
										</div>
										<input id="ProfileWantToText" name="ProfileWantToText" value="{{- .ProfileWantToText -}}" class="form-control" type="text" aria-describedby="labelProfileWantToText">
									</div>
								</div>
							</div>
							<hr>
							<!-- Skills-->
							<label for="SkillsLeft" id="labelProfileSkillsText">Skills:</label>
							<div class="form-row">
								<div class="form-group col-3 mt-2 mb-2" id="SkillsLeft">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="SkillsTextures" {{ if (bitTest .ProfileSkillsMask 1) -}}checked{{- end -}}>
										<label class="form-check-label" for="SkillsTextures">Textures</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="SkillsModeling" {{ if (bitTest .ProfileSkillsMask 8) -}}checked{{- end -}}>
										<label class="form-check-label" for="SkillsModeling">Modeling</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="SkillsScripting" {{ if (bitTest .ProfileSkillsMask 16) -}}checked{{- end -}}>
										<label class="form-check-label" for="SkillsScripting">Scripting</label>
									</div>
								</div>
								<div class="form-group col-3 mt-2 mb-2" id="SkillsRight">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="SkillsArchitecture" {{ if (bitTest .ProfileSkillsMask 2) -}}checked{{- end -}}>
										<label class="form-check-label" for="SkillsArchitecture">Architecture</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="SkillsEventPlanning" {{ if (bitTest .ProfileSkillsMask 4) -}}checked{{- end -}}>
										<label class="form-check-label" for="SkillsEventPlanning">Event Planning</label>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="SkillsCustomCharacters" {{ if (bitTest .ProfileSkillsMask 32) -}}checked{{- end -}}>
										<label class="form-check-label" for="SkillsCustomCharacters">Custom Characters</label>
									</div>
								</div>
							</div>
							<div class="form-row">
								<div class="form-group col-6 mb-1">
									<div class="input-group">
										<div class="input-group-prepend">
											<span class="input-group-text bg-primary border-right-0 border-primary"><i class="fas fa-toolbox fa-fw text-light"></i></span>
										</div>
										<input id="ProfileSkillsText" name="ProfileSkillsText" value="{{- .ProfileSkillsText -}}" class="form-control" type="text" aria-describedby="labelProfileSkillsText">
									</div>
								</div>
							</div>
							<hr>
							<!-- Languages -->
							<div class="form-group">
								<label for="ProfileLanguages" id="labelProfileLanguages" class="control-label">Languages spoken</label>
								<div class="input-group">
									<div class="input-group-prepend">
										<span class="input-group-text bg-primary border-right-0 border-primary"><i class="fas fa-language fa-fw text-light"></i></span>
									</div>
									<input id="ProfileLanguages" name="ProfileLanguages" value="{{- .ProfileLanguages -}}" class="form-control" type="text" aria-describedby="labelProfileLanguages">
								</div>
							</div>
							<!-- Text for First Life and associated image -->
							<div class="form-group">
								<label for="ProfileFirstText" id="labelProfileFirstText" class="control-label">About your real life</label>
								<div class="input-group">
									<div class="image pr-1">
										{{- if .ProfileFirstImage -}}
										<img class="rounded shadow-sm" src="{{- .ProfileFirstImage -}}" alt="Real Life Image for {{- .Username -}}" height="128" width="128" srcset="{{- .ProfileFirstImage }} 1x, {{ .ProfileRetinaFirstImage }} 2x">
										{{- else -}}
										<img class="rounded shadow-sm" src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mOMa1zzHwAFXAKM3m3GvQAAAABJRU5ErkJggg==" alt="No image for {{- .Username -}}" height="128" width="128">
										{{- end }}
									</div>
									<textarea class="form-control" id="ProfileFirstText" rows="3" aria-describedby="labelProfileFirstText">{{- .ProfileFirstText -}}</textarea>
								</div>
							</div>

							<!-- Success message -->
							<div class="alert alert-success invisible" role="alert" id="success_message">Success&nbsp;<i class="fas fa-thumbs-up"></i>Success!</div>
							<!-- Submit Button -->
							<div class="form-group mx-auto text-center mb-4">
								<button type="submit" class="btn btn-primary shadow-sm" value="Submit">Submit&nbsp;<i class="fas fa-paper-plane"></i></button>
							</div>
						</fieldset>
					</form>
					{{- if .ProfileData }}
					<div class="invisible">Raw data: {{- .ProfileData -}}</div>
					{{ end -}}
				</div> <!-- /.col -->
			</div> <!-- /.row -->
			<div class="row">
				{{ if .usersOnline }}
				<div class="col mb-4">
					<!-- DataTables for online Friends -->
					<div class="card shadow mb-4">
						<a href="#onlineUsersCard" class="card-header py-3" data-toggle="collapse" role="button" aria-expanded="true" aria-controls="onlineUsersCard">
							<h6 class="m-0 font-weight-bold text-primary">Your online friends</h6>
						</a>
						<div class="collapse show" id="onlineUsersCard">
							<div class="card-body">
								<div class="table-responsive">
									<table class="table table-bordered table-squeezed" id="usersOnline"
										data-order='[]' data-page-length='25'>
										<thead>
											<tr>
												<th>Avatar Name</th>
											</tr>
										</thead>
									</table>
								</div>
							</div>
							<!-- ./card-body -->
						</div>
						<!-- ./collapse -->
					</div>
				</div>
				{{ end }}
			</div> <!-- /.row -->
			{{ if .Debug }}
			<div class="row">
				<div class="col">
					<h2>Debug info:</h2>
					<p><b>usersOnline:</b>&nbsp;{{ .usersOnline -}}</p>
				</div>
			</div>
			{{ end }}
			{{ template "back.tpl"}}
		</div>
		<!-- /.container-fluid -->
	</div>
	<!-- End of Main Content -->
{{ template "footer.tpl" . -}}
{{- end -}}