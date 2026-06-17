/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./register.html", "postbox", "jquery", "semantic-ui-modal", "semantic-ui-dimmer"],
    function (ko, template, postbox, $) {

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
            this.registerEmail = ko.observable("");
            this.registerPassword = ko.observable("");
            this.registerName = ko.observable("");
            this.registerSurname = ko.observable("");

            // Get reference to the view model's actual HTML within the DOM
            this.$domModal = $('#modalRegister'); // Modal will be moved by Semantic UI and not within component area anymore
            this.$domForm = this.$domModal.find("form");

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
        ViewModel.prototype.submitRegister = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Validate form
            if (!this.$domForm.form('is valid')) {
                this.$domForm.form("validate form");
                this.$domForm.each(shake);
                return;
            }

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Emit toast message for user
                toast(response.message, "success");

                // Redirect to login
                postbox.publish("redirect", "login");

                // Reset form
                ctx.$domForm.form("reset");

                // Hide modal
                ctx.$domModal.modal('hide');
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                ctx.$domForm.form("add prompt", "inputEmail", "Invalid E-Mail");
                if (!developmentLogin()) {
                    ctx.$domForm.form("add prompt", "inputPassword", "Invalid Password");
                }
                ctx.$domForm.each(shake);
            };

            // Prepare request body
            var reqData = {
                email: this.registerEmail(),
                password: this.registerPassword(),
                name: this.registerName(),
                surname: this.registerSurname()
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/auth/register",
                {},
                reqData,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.redirectLogin = function (data, event) {

            // Reset form
            this.$domForm.form("reset");

            // Redirect back to login
            postbox.publish("redirect", "login");
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
