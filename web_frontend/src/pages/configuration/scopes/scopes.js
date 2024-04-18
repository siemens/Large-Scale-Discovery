/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./scopes.html", "postbox", "jquery", "semantic-ui-popup", "semantic-ui-dropdown", 'semantic-ui-transition'],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.sideNavItems = ko.observableArray([
                new NavItem("Scopes", "#configuration/scopes", ""),
                new NavItem("Views", "#configuration/views", ""),
            ]);
            this.scopes = ko.observable(null);

            this.allowCustom = ko.observable(false);
            this.allowAsset = ko.observable(false);
            this.allowNetwork = ko.observable(false);

            this.actionArgs = ko.observable(null); // action element row to work on
            this.actionName = ko.observable("")
            this.actionComponent = ko.observable(null); // action form that should be shown
            this.actionComponentRecent = ko.observable("custom"); // action form that should be shown

            // Check authentication and redirect to login if necessary
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divConfigurationScopes');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Load and set initial data
            this.loadData();
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event, callbackCompletion) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Round down progress values
                for (var i = 0; i < response.body["scopes"].length; i++) {
                    response.body["scopes"][i].cycle_progress = [
                        Math.round(response.body["scopes"][i].cycle_done),
                        Math.round(response.body["scopes"][i].cycle_failed),
                        Math.round(response.body["scopes"][i].cycle_active),
                    ]
                }

                // Set table data
                ctx.scopes(response.body["scopes"]);
                ctx.allowCustom(response.body["allow_custom"]);
                ctx.allowNetwork(response.body["allow_network"]);
                ctx.allowAsset(response.body["allow_asset"]);

                // Execute completion callback if set
                if (callbackCompletion !== undefined) {
                    callbackCompletion()
                }
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/scopes",
                {},
                null,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Dispose open form
            this.actionArgs(null);
            this.actionComponent(null);
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
