/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./list.html", "postbox", "jquery", "jquery-tablesort", "semantic-ui-popup", "semantic-ui-dropdown"],
    function (ko, template, postbox, $) {

        // Extract token grants and generate simple data structure for dipslay
        function extractTokenGrants(views) {

            // Return if there is no data yet
            if (views === null) {
                return null
            }

            // Extract tokens from grants (list of grants also contains users)
            var tokens = []
            for (var i = 0; i < views.length; i++) {
                for (var j = 0; j < views[i].grants.length; j++) {
                    if (!views[i].grants[j].is_user) {
                        tokens.push({
                            group_name: views[i].scan_scope.group_name,
                            scope_name: views[i].scan_scope.name,
                            view_name: views[i].name,
                            view_id: views[i].id,
                            token: views[i].grants[j],
                        })
                    }
                }
            }

            // Return extended data struct
            return tokens
        }

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Load views data from parent, parent has already requested it.
            // The data is *not* passed as a component argument, to prevent automatic
            // KnockoutJs re-rendering, which would cause fade-in animation to run again.
            this.tokenGrouped = ko.observable(null);

            // Keep reference THIS view model context
            var ctx = this;

            // Subscribe to changes of parent view data to update component accordingly
            this.parent.views.subscribe(function (views) {
                ctx.tokenGrouped(itemsByKey(extractTokenGrants(views), "group_name"));
            });

            // Get reference to the token model's actual HTML within the DOM
            this.$domComponent = $('#divTokens');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize table sort
            this.$domComponent.find('table').tablesort();

            // Load and set initial data
            this.loadData()
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Load data from parent component
            this.tokenGrouped(itemsByKey(extractTokenGrants(this.parent.views()), "group_name"))

            // Fade in table
            this.$domComponent.children("div:hidden").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                duration: 200
            });
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.initTokenEntries = function (element, data) {

            // Initialize table sort for group
            $(element).tablesort()

            // Initialize tooltips
            $(element).find('[data-html]').popup();
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.revokeToken = function (view, data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "user friends",
                "Revoke Access Token",
                "Are you sure you want to revoke the access token <br /><span class=\"ui red text\">'" + data.token.username + "'</span> from '" + data.scope_name + " - " + data.view_name + "'?",
                function () {

                    // Handle request success
                    const callbackSuccess = function (response, textStatus, jqXHR) {

                        // Show toast message for user
                        toast(response.message, "success");

                        // Notify parent to reload updated data
                        ctx.parent.loadData();
                    };

                    // Prepare request body
                    var reqData = {
                        view_id: data.view_id,
                        username: data.token.username,
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/view/grant/revoke",
                        {},
                        reqData,
                        callbackSuccess,
                        null
                    );
                },
                data.username,
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
