/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./add.html", "postbox", "jquery", "semantic-ui-dropdown"],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Store parameters passed by parent
            this.args = params.args;

            // Initialize observables
            this.scopeId = ko.observable(params.args.id);
            this.tokenDescription = ko.observable(params.args.description);
            this.tokenExpiryDays = ko.observable(params.args.expiry_days || 7);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divTokensAdd');
            this.$domForm = this.$domComponent.find("form");

            // Initialize Create Group form validators
            this.$domForm.form({
                fields: {
                    inputDescription: ['minLength[5]'],
                },
            });

            // Initialize slider
            initSlider("#sliderExpiry", this.tokenExpiryDays, 1, 365, 1);

            // Fade in
            this.$domComponent.transition('fade down');

            // Scroll to form (might be outside of visible area if there are long lists)
            $([document.documentElement, document.body]).animate({
                scrollTop: this.$domComponent.offset().top - 160
            }, 200);
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.submitGenerate = function (data, event) {

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

                // Show toast message for user
                toast(response.message, "success");

                // Notify parent to reload updated data
                ctx.parent.loadData();

                // Unlink component
                ctx.dispose(data, event)
            };

            // Prepare request body
            var reqData = {
                view_id: this.scopeId(),
                description: this.tokenDescription(),
                expiry_days: this.tokenExpiryDays(),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/view/grant/token",
                {},
                reqData,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Hide form
            this.$domComponent.transition('fade up');

            // Reset form fields
            this.$domForm.form('reset');

            // Dispose open form
            if (this.parent.actionComponent() === "tokens-add") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
