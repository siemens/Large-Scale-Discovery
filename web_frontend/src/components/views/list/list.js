/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./list.html", "postbox", "jquery", "jquery-tablesort", "semantic-ui-popup", "semantic-ui-dropdown"],
    function (ko, template, postbox, $) {

        // Extends data structure with a sanitized set of values
        function viewsSanitize(views) {

            // Return if there is no data yet
            if (views === null) {
                return null
            }

            // Prepare view users grouped by company (a view may be granted to users outside ones company)
            for (var i = 0; i < views.length; i++) {

                // Reference view
                var view = views[i]

                // Convert filters map to array
                view.filter_strings = []
                for (var key of Object.keys(view.filters)) {
                    var str = view.filters[key].join(" | ")
                    if (str.length > 60) {
                        str = str.substr(0, 60) + "..."
                    }
                    view.filter_strings.push([key, str])
                }

                // Prepare list of users grouped by company
                var users = {}
                for (var j = 0; j < view.grants.length; j++) {

                    // Reference grant
                    var grant = view.grants[j]

                    // Ignore access token grant types
                    if (!grant.is_user) {
                        continue
                    }

                    // Strip company, if it isn't a real one
                    if (grant.username === grant.user_company) {
                        grant.user_company = ""
                    }

                    // Initialize company list if not existing yet
                    if (users[grant.user_company] === undefined) {
                        users[grant.user_company] = []
                    }

                    // Add user to company list
                    users[grant.user_company].push(grant)
                }

                // Convert grants to two-dimensional array
                var userGrants = []
                for (var company in users) {
                    userGrants.push(users[company])
                }

                // Inject users grouped by company
                view.users_by_company = userGrants
            }

            // Return extended data struct
            return views
        }

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Load views data from parent, parent has already requested it.
            // The data is *not* passed as a component argument, to prevent automatic
            // KnockoutJs re-rendering, which would cause fade-in animation to run again.
            this.viewsGrouped = ko.observable(null);
            this.inactiveUsers = this.parent.inactiveUsers;

            // Keep reference THIS view model context
            var ctx = this;

            // Subscribe to changes of parent view data to update component accordingly
            this.parent.views.subscribe(function (views) {
                ctx.viewsGrouped(itemsByKey(viewsSanitize(views), ["scan_scope", "group_name"]));
            });
            this.parent.inactiveUsers.subscribe(function (inactiveUsers) {
                ctx.inactiveUsers();
            });

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divViews');

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
            this.viewsGrouped(itemsByKey(viewsSanitize(this.parent.views()), ["scan_scope", "group_name"]))

            // Fade in table
            this.$domComponent.children("div:hidden").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                duration: 200
            });
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.initViewEntries = function (element, data) {

            // Initialize table sort for group
            $(element).tablesort()

            // Initialize tooltips
            $(element).find('[data-html]').popup();
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.revokeUser = function (view, grant, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "user friends",
                "Revoke User",
                "Are you sure you want to revoke the user <br /><span class=\"ui red text\">'" + grant.username + "'</span> from <span class=\"ui red text\">'" + view.scan_scope.name + "' - '" + view.name + "'</span>?",
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
                        view_id: view.id,
                        username: grant.username,
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
                }
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showViewEdit = function (data, event) {

            // Dispose open form
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);

            // Show new form
            this.parent.actionArgs(data);
            this.parent.actionComponent("views-edit");
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showViewAccess = function (data, event) {

            // Dispose open form
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);

            // Show new form
            this.parent.actionArgs(data);
            this.parent.actionComponent("views-grant");
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showTokenAdd = function (data, event) {

            // Dispose open form
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);

            // Show new form
            this.parent.actionArgs(data); // reset action args, which are not required
            this.parent.actionComponent("views-token-add");
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.deleteView = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "trash alternate outline",
                "Delete View",
                "This will remove all associated access rights. <br />Are you sure you want to delete the view <span class=\"ui red text\">'" + data.name + "'</span>?",
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
                        id: data.id,
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/view/delete",
                        {},
                        reqData,
                        callbackSuccess,
                        null
                    );
                }
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showViewAdd = function (data, event) {

            // Dispose open form
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);

            // Show new form
            this.parent.actionComponent("views-add");
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Dispose open form
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
