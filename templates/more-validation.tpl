{{- define "more-validation.tpl" -}}
<!-- this should be called at the end of the page -->
<!--
<script type="text/javascript">
$(document).ready(function () {
	$("#contact_form")
		.bootstrapValidator({
			// To use feedback icons, ensure that you use Bootstrap v3.1.0 or later
			feedbackIcons: {
				valid: "fas fa-thumbs-up",
				invalid: "fas fa-thumbs-down",
				validating: "fas fa-sync fa-spin",
			},
			fields: {
				first_name: {
					validators: {
						stringLength: {
							min: 2,
						},
						notEmpty: {
							message: "Please enter your First Name",
						},
					},
				},
				last_name: {
					validators: {
						stringLength: {
							min: 2,
						},
						notEmpty: {
							message: "Please enter your Last Name",
						},
					},
				},
				user_name: {
					validators: {
						stringLength: {
							min: 8,
						},
						notEmpty: {
							message: "Please enter your Username",
						},
					},
				},
				user_password: {
					validators: {
						stringLength: {
							min: 8,
						},
						notEmpty: {
							message: "Please enter your Password",
						},
					},
				},
				confirm_password: {
					validators: {
						stringLength: {
							min: 8,
						},
						notEmpty: {
							message: "Please confirm your Password",
						},
					},
				},
				email: {
					validators: {
						notEmpty: {
							message: "Please enter your Email Address",
						},
						emailAddress: {
							message: "Please enter a valid Email Address",
						},
					},
				},
				contact_no: {
					validators: {
						stringLength: {
							min: 12,
							max: 12,
							notEmpty: {
								message: "Please enter your Contact No.",
							},
						},
					},
					department: {
						validators: {
							notEmpty: {
								message: "Please select your Department/Office",
							},
						},
					},
				},
			},
		})
		.on("success.form.bv", function (e) {
			$("#success_message").slideDown({ opacity: "show" }, "slow"); // Do something ...
			$("#contact_form").data("bootstrapValidator").resetForm();

			// Prevent form submission
			e.preventDefault();

			// Get the form instance
			var $form = $(e.target);

			// Get the BootstrapValidator instance
			var bv = $form.data("bootstrapValidator");

			// Use Ajax to submit form data
			$.post(
				$form.attr("action"),
				$form.serialize(),
				function (result) {
					console.log(result);
				},
				"json"
			);
		});
});
</script>
<noscript>With JavaScript, this would look so much nicer.</noscript>-->
{{- end -}}