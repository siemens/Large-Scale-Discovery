/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./grant.html", "postbox", "jquery", "semantic-ui-popup", "semantic-ui-dropdown"],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Store parameters passed by parent
            this.args = params.args;

            // Initialize observables
            this.viewGrantsPossible = ko.observableArray([]);
            this.viewGrants = ko.observableArray([]);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divViewsAccess');
            this.$domForm = this.$domComponent.find('form');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize multi select dropdown elements
            initDropdown("#selectGrants", this.viewGrants, "-", true)

            // Initialize form validators
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                    inputGroup: ['empty'],
                },
            });

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

                // Fill possible users based on request
                if (response.body["users"].length === 0) {
                    var grantsPossible = []
                    for (var i = 0; i < ctx.args.grants.length; i++) {
                        if (ctx.args.grants[i].is_user) { // Ignore access token grant types
                            grantsPossible.push(ctx.args.grants[i].username)
                        }
                    }
                    ctx.viewGrantsPossible(grantsPossible)
                } else {

                    // Set possible users field
                    for (var j = 0; j < response.body["users"].length; j++) {
                        ctx.viewGrantsPossible.push(response.body["users"][j]["email"])
                    }

                    // Additionally push those entries that are in current but were not returned by the server (otherwise
                    // the dropdown box won't work!
                    for (var k = 0; k < ctx.args.grants.length; k++) {
                        if (ctx.args.grants[k].is_user) { // Ignore access token grant types
                            if (ctx.viewGrantsPossible().indexOf(ctx.args.grants[k].username) === -1) {
                                ctx.viewGrantsPossible.push(ctx.args.grants[k].username)
                            }
                        }
                    }
                }

                // Extract list of currently granted users
                var users = []
                for (var l = 0; l < ctx.args.grants.length; l++) {
                    if (ctx.args.grants[l].is_user) { // Ignore access token grant types
                        users.push(ctx.args.grants[l].username)
                    }
                }

                // Set users field (already known via passed arguments)
                ctx.viewGrants(users);
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/users",
                {},
                null,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submitEdit = function (data, event) {

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

                    // Show toast message for user (but only after parent has reloaded)
                    toast(response.message, "success");

                    // Unlink component (but only after parent has reloaded)
                    ctx.dispose(data, event)
                });
            };

            // Prepare request body
            var reqData = {
                view_id: this.args.id,
                users: this.viewGrants(),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/view/grant/users",
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
            if (this.parent.actionComponent() === "views-grant") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
