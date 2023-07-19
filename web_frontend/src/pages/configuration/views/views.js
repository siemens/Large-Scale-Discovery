/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./views.html", "postbox", "jquery", "semantic-ui-popup", "semantic-ui-dropdown"],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.sideNavItems = ko.observableArray([
                new NavItem("Scopes", "#configuration/scopes", ""),
                new NavItem("Views", "#configuration/views", ""),
            ]);
            this.views = ko.observable(null);
            this.actionComponent = ko.observable(null); // action form that should be shown
            this.actionArgs = ko.observable(null); // action element row to work on
            this.inactiveUsers = ko.observableArray([]);

            // Check authentication and redirect to login if necessary
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }

            // Load views data
            this.loadData()
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event, callbackCompletion) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Prepare new list
                var inactiveUsers = []

                // Extract granted users who haven't logged in for a while
                response.body["views"].forEach(function (view, i, arrayI) {
                    view.grants.forEach(function (grant, j, arrayJ) {
                        if (grant.is_user === true) {
                            if (moment(grant.user_last_login, datetimeFormatGolang).diff(moment().subtract(.5, 'years')) < 0 &&
                                moment(grant.user_created, datetimeFormatGolang).diff(moment().subtract(.5, 'years')) < 0) {
                                grant.is_inactive = true
                                if (grant.username === grant.user_company) {
                                    inactiveUsers.push(grant.username)
                                } else if (grant.user_company !== "") {
                                    inactiveUsers.push(grant.username + " (" + grant.user_company + ")")
                                } else {
                                    inactiveUsers.push(grant.username)
                                }
                            } else {
                                grant.is_inactive = false
                            }
                        }
                    })
                })

                // Update new list of inactive users
                ctx.inactiveUsers([...new Set(inactiveUsers)]);

                // Set table data
                ctx.views(response.body["views"]);

                // Execute completion callback if set
                if (callbackCompletion !== undefined) {
                    callbackCompletion()
                }
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/views",
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
