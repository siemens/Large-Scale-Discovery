/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./login.html", "postbox", 'globals', "jquery", "semantic-ui-modal", "semantic-ui-dimmer"],
    function (ko, template, postbox, globals, $) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Check authentication and redirect to login if necessary
            if (authenticated()) {
                postbox.publish("redirect", home());
                return;
            }

            // Initialize observables
            this.submitAction = ko.observable(this.submitPreAuth);
            this.loginEmail = ko.observable("");
            this.loginPassword = ko.observable("");

            // Get reference to the view model's actual HTML within the DOM
            this.$domModal = $('#modalLogin'); // Modal will be moved by Semantic UI and not within component area anymore
            this.$domForm = this.$domModal.find("form");
            this.$domEmail = $('#divEmail');
            this.$domPassword = $('#divPassword');

            // Hide password form
            this.$domPassword.transition("hide");

            // Initialize modal
            this.$domModal.modal({
                detachable: false, // Prevent Semantic-UI from moving modal into <body> where KnockoutJs looses track
                closable: false
            }).modal('show');

            // Patch modal background color to be not transparent in this case
            $('.ui.dimmer').css("background-color", "teal");

            // Initialize form with validators. keyboardShortcuts is disabled because
            // Semantic UI's Enter handler would submit the form a second time alongside
            // the browser's native submit that Knockout's submit binding already handles.
            this.$domForm.form({
                fields: {
                    inputEmail: ['notEmpty', 'email'],
                },
                keyboardShortcuts: false, // Prevent FomanticUI's own submit action handler from submitting again
            });
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadUser = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle success
            const callbackSuccess2 = function (response, textStatus, jqXHR) {

                // Save authenticated user to local storage. Values will be read from there on page reload.
                sessionStorage.setItem("user", JSON.stringify(response.body));

                // Update profile data
                globals.profileSet(
                    response["body"]["id"],
                    response["body"]["email"],
                    response["body"]["name"],
                    response["body"]["surname"],
                    response["body"]["gender"],
                    response["body"]["admin"],
                    response["body"]["owner"],
                    response["body"]["access"],
                    response["body"]["demo"],
                    response["body"]["created"]
                );

                // Reset form
                ctx.$domForm.form("reset");

                // Hide modal
                ctx.$domModal.modal('hide');

                // Redirect to originally called URL or to the user's home page
                var target = sessionStorage.getItem("redirect");
                if (target !== "" && target !== null) {

                    // Reset redirect
                    sessionStorage.setItem("redirect", "");

                    // Update redirect if it's redirecting to the user's wrong home page
                    if (target === "home" && home() !== "home") {
                        target = home()
                    }

                    // Redirect to intended URL
                    postbox.publish("redirect", target);
                } else {
                    postbox.publish("redirect", home());
                }
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/user/details",
                {},
                null,
                callbackSuccess2,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submit = function (data, event) {
            this.submitAction()(data, event, this); // Get currently set action and dispatch event
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.submitPreAuth = function (data, event, ctx) {

            // Validate form
            if (!ctx.$domForm.form('is valid')) {
                ctx.$domForm.form("validate form");
                ctx.$domEmail.each(shake);
                return;
            }

            // Handle success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Decide authentication action
                var redirect = response["body"]["entry_url"];
                if (redirect !== "") {

                    // Redirect user to authentication entry point
                    window.location.href = redirect;

                } else if (developmentLogin()) {

                    // Do development mode login without password
                    ctx.submitLogin(data, event, ctx)

                } else {

                    // Add password form validation
                    ctx.$domForm.form("add rule", "inputPassword", ['notEmpty', 'length[8]']);

                    // Show password field and login button for credentials login
                    ctx.$domEmail.transition("hide");
                    ctx.$domPassword.transition("fade left");

                    // Update submit action for second step
                    ctx.submitAction(ctx.submitLogin)
                }
            };

            // Prepare request body
            var reqData = {
                email: ctx.loginEmail().trim(),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/backend/authenticator",
                {},
                reqData,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submitLogin = function (data, event, ctx) {

            // Validate form
            if (!ctx.$domForm.form('is valid')) {
                ctx.$domForm.form("validate form");
                ctx.$domForm.each(shake);
                return;
            }

            // Handle success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Save authentication token to local storage. Values will be read from there on page reload.
                sessionStorage.setItem("token", JSON.stringify(response.token));

                // Update authentication data
                globals.authenticationSet(
                    response["token"]["access_token"],
                    response["token"]["expire"]
                );

                // Load authenticated user's data
                ctx.loadUser()
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                ctx.$domForm.form("add prompt", "inputPassword", "Invalid Password");
                ctx.$domForm.each(shake);
            };

            // Prepare request body
            var reqData = {
                email: ctx.loginEmail().trim(),
                password: ctx.loginPassword().trim()
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/auth/login",
                {},
                reqData,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.cancelLogin = function (data, event) {

            // Add password form validation
            this.$domForm.form("remove rule", "inputPassword");

            // Hide password field and login button of credentials login
            this.$domEmail.transition("fade right");
            this.$domPassword.transition("hide");

            // Reset submit action for first step
            this.submitAction(this.submitPreAuth)
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.redirectRegister = function (data, event) {
            postbox.publish("redirect", "register");
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Reset dimmer, that was changed to teal for the login modal
            $('.ui.dimmer').css("background-color", "rgba(0, 0, 0, 0.85)")
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
