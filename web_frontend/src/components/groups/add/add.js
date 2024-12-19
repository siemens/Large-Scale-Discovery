/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./add.html", "postbox", "jquery", "semantic-ui-popup", "semantic-ui-dropdown"],
    function (ko, template, postbox, $) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Initialize observables
            this.groupName = ko.observable("");
            this.groupMaxScopes = ko.observable("*");
            this.groupMaxViews = ko.observable("*");
            this.groupMaxTargets = ko.observable("*");
            this.groupMaxOwners = ko.observable("*");
            this.allowCustom = ko.observable(true);
            this.allowNetwork = ko.observable(false);
            this.allowAsset = ko.observable(false);
            this.databasesAvailable = ko.observableArray([]);
            this.databaseSelected = ko.observable(-1);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divGroupAdd');
            this.$domForm = this.$domComponent.find("form");

            // Initialize dropdown elements
            this.$domComponent.find('select.dropdown').dropdown();

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Define custom range validation rule
            $.fn.form.settings.rules.numberOrUnlimited = function (value) {
                value = value.trim();
                if (value === "" || value === "*") {
                    return true
                }
                return parseInt(value, 10) >= -1;
            };

            // Initialize Create Group form validators
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                    selectDatabase: ['empty'],
                    inputMaxScopes: ['numberOrUnlimited'],
                    inputMaxViews: ['numberOrUnlimited'],
                    inputMaxTargets: ['numberOrUnlimited'],
                    inputMaxOwners: ['numberOrUnlimited'],
                },
            });

            // Fade in
            this.$domComponent.transition('fade down');

            // Scroll to form (might be outside visible area if there are long lists)
            $([document.documentElement, document.body]).animate({
                scrollTop: this.$domComponent.offset().top - 160
            }, 200);

            // Load databases to map DB IDs and DB names
            this.loadDatabases();
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadDatabases = function (data, event, callbackCompletion) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Init array of databases
                ctx.databasesAvailable(response.body["databases"]);

                // Set scope group, if there is only one
                if (response.body["databases"].length === 1) {
                    ctx.databaseSelected(response.body["databases"][0].id);
                }
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/admin/databases",
                {},
                null,
                callbackSuccess,
            );

        }

        // VIEWMODEL ACTION
        ViewModel.prototype.submitGroup = function (data, event) {

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

            // Initialize with unlimited value, unless something was entered
            var maxScopes = -1;
            var maxViews = -1;
            var maxTargets = -1;
            var maxOwners = -1;
            if (this.groupMaxScopes() !== "" && this.groupMaxScopes() !== "*") {
                maxScopes = parseInt(this.groupMaxScopes(), 10);
            }
            if (this.groupMaxViews() !== "" && this.groupMaxViews() !== "*") {
                maxViews = parseInt(this.groupMaxViews(), 10);
            }
            if (this.groupMaxTargets() !== "" && this.groupMaxTargets() !== "*") {
                maxTargets = parseInt(this.groupMaxTargets(), 10);
            }
            if (this.groupMaxOwners() !== "" && this.groupMaxOwners() !== "*") {
                maxOwners = parseInt(this.groupMaxOwners(), 10);
            }

            // Prepare request body
            var reqData = {
                name: this.groupName(),
                db_server_id: this.databaseSelected(),
                max_scopes: maxScopes,
                max_views: maxViews,
                max_targets: maxTargets,
                max_owners: maxOwners,
                allow_custom: this.allowCustom(),
                allow_network: this.allowNetwork(),
                allow_asset: this.allowAsset()
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/admin/group/create",
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
            if (this.parent.actionComponent() === "groups-add") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
