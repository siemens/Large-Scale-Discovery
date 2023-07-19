/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./owners.html", "postbox", "jquery", "semantic-ui-modal"],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Store parameters passed by parent
            this.args = ko.observable(params.args); // Args is already form data, so it should be an observable

            // Initialize observables
            this.groupOwners = ko.observableArray([]);
            this.groupOwnersPossible = ko.observableArray([]);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divGroupOwners');
            this.$domForm = this.$domComponent.find("form");

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize multi select dropdown elements
            initDropdown("#selectOwners", this.groupOwners, "-", true)

            // Initialize Create Group form validators
            this.$domForm.form({
                fields: {},
            });

            // Load and set initial data
            this.loadData();

            // Fade in
            this.$domComponent.transition('fade down');

            // Scroll to form (might be outside of visible area if there are long lists)
            $([document.documentElement, document.body]).animate({
                scrollTop: this.$domComponent.offset().top - 160
            }, 200);
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Set possible owners filed
                for (var j = 0; j < response.body["users"].length; j++) {
                    ctx.groupOwnersPossible.push(response.body["users"][j]["email"])
                }

                // Set owners field (already known via passed arguments)
                for (var i = 0; i < ctx.args()["ownerships"].length; i++) {
                    ctx.groupOwners.push(ctx.args()["ownerships"][i]["user"]["email"])
                }
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                ctx.dispose(null, null) // pass event != undefined, to make full dispose function execute
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/users",
                {},
                null,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submitOwners = function (data, event) {

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

                // Reload parent table, because data got updated
                ctx.parent.loadData(null, null, function () {

                    // Show toast message for user (but only after table has reloaded)
                    toast(response.message, "success");

                    // Unlink component (but only after table has reloaded)
                    ctx.dispose(data, event)
                });
            };

            // Prepare request body
            var reqData = {
                id: this.args().id,
                owners: this.groupOwners(),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/admin/group/assign",
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
            if (this.parent.actionComponent() === "groups-owners") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
