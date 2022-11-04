/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./register.html", "postbox", "jquery", "semantic-ui-modal", "semantic-ui-dimmer"],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Check authentication and redirect to login if necessary
            if (authenticated()) {
                postbox.publish("redirect", "home");
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

            // Initialize registration form with validators and submit action
            this.$domForm.form({
                fields: {
                    inputEmail: ['empty', 'email'],
                },
            });
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.submitRegister = function (data, event) {

            // Keep reference THIS view model context
            var parent = this;

            // Validate form
            if (!this.$domForm.form('is valid')) {
                this.$domForm.form("validate form");
                this.$domForm.each(shake);
                return;
            }

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                parent.$domForm.form('add prompt', 'inputEmail');
                if (!developmentLogin()) {
                    parent.$domForm.form('add prompt', 'inputPassword');
                }
                parent.$domForm.each(shake);
            };

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Emit toast message for user
                toast(response.message, "success");

                // Redirect to login
                postbox.publish("redirect", "login");

                // Reset form
                parent.$domForm.form("reset");

                // Hide modal
                parent.$domModal.modal('hide');
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
