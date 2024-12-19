/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./feedback.html", "postbox", "jquery", "semantic-ui-modal"],
    function (ko, template, postbox, $) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Initialize observables
            this.feedbackSubject = ko.observable("");
            this.feedbackMessage = ko.observable("");

            // Get reference to the view model's actual HTML within the DOM
            this.$domModal = $('#modalFeedback'); // Modal will be moved by Semantic UI and not within component area anymore
            this.$domForm = this.$domModal.find("form");

            // Initialize registration form with validators and submit action
            this.$domForm.form({
                fields: {
                    inputSubject: ['empty'],
                    textareaMessage: ['empty']
                },
            });

            // Keep reference THIS view model context
            var ctx = this;

            // Prepare array of references to subscriptions in order to dispose them later
            this.subscriptions = []

            // Close feedback modal when user gets logged out, otherwise it might feel weird
            // when a user logs back in.
            this.subscriptions.push(authenticated.subscribe(function (newValue) {
                if (newValue === false) {
                    ctx.closeFeedback()
                }
            }));
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.openFeedback = function (data, event) {
            this.$domModal.modal('toggle');
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.closeFeedback = function (data, event) {
            this.feedbackSubject("");
            this.feedbackMessage("");
            this.$domModal.modal('hide');
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submitFeedback = function (data, event) {

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

                // Reset values for next time
                ctx.$domModal.modal('hide');

                // Reset login form
                ctx.feedbackSubject("");
                ctx.feedbackMessage("");

                // Show toast message for user
                toast(response.message, "success");
            };

            // Prepare request body
            var reqData = {
                subject: this.feedbackSubject(),
                message: this.feedbackMessage(),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/user/feedback",
                {},
                reqData,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Hide modal that might be open
            $('#modalFeedback').modal('hide');

            // Dispose subscriptions
            for (var k in this.subscriptions) {
                this.subscriptions[k].dispose();
            }
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
