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

        // Type component map
        var components = {
            "custom": "scopes-add-custom",
            "assets": "scopes-add-assets",
            "networks": "scopes-add-networks",

            // Some company specific types
            "caremore": "scopes-add-assets",
            "itam": "scopes-add-assets",
            "snic": "scopes-add-networks",
        }

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Load scopes data from parent, parent has already requested it.
            // The data is *not* passed as a component argument, to prevent automatic
            // KnockoutJs re-rendering, which would cause fade-in animation to run again.
            this.scopesGrouped = ko.observable(null);

            // Keep reference THIS view model context
            var ctx = this;

            // Subscribe to changes of parent scope data to update component accordingly
            this.parent.scopes.subscribe(function (scopes) {
                ctx.scopesGrouped(itemsByKey(scopes, "group_name"))
            });

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divScopes');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize table sort
            this.$domComponent.find('table').tablesort();

            // Initialize scopes add button
            this.$domComponent.find('.ui.dropdown').dropdown({
                action: 'combo'
            });

            // Load and set initial data
            this.loadData()
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Load data from parent component
            this.scopesGrouped(itemsByKey(this.parent.scopes(), "group_name"))

            // Fade in table
            this.$domComponent.children("div:hidden").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                duration: 200
            });
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.initScopeEntries = function (element, data) {

            // Initialize table sort for group
            $(element).tablesort()

            // Initialize progress bars for group scopes
            $(element).find('.ui.progress').progress({
                showActivity: false,
            });

            // Initialize tooltips
            $(element).find('[data-html]').popup();
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.pauseScope = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Show toast message for user
                toast(response.message, "success");

                // Notify parent to reload updated data
                ctx.parent.loadData();
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/scope/pause",
                {},
                {"id": data.id},
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showScopeTargets = function (data, event) {

            // Dispose open form
            this.parent.actionName(null);
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);

            // Set arguments to pre-fill form with
            this.parent.actionArgs(data);

            // Retrieve component to load
            var component = components[data.type]

            // Show form
            this.parent.actionName(data.type);
            this.parent.actionComponent(component);
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showScopeSettings = function (data, event) {

            // Dispose open form
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);

            // Show new form
            this.parent.actionArgs(data);
            this.parent.actionComponent("scopes-settings");
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.newScanCycle = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "redo",
                "New Scan Cycle",
                "This will initialize a new scan cycle. All scan progress will be reset. Results of running scans will be dropped, but already stored ones will be kept!<br />Are you sure you want to initialize a new scan cycle for <span class=\"ui red text\">'" + data.name + "'</span>?",
                function () {

                    // Handle request success
                    const callbackSuccess = function (response, textStatus, jqXHR) {

                        // Show toast message for user
                        toast(response.message, "success");

                        // Notify parent to reload updated data
                        ctx.parent.loadData();
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/scope/cycle",
                        {},
                        {"id": data.id},
                        callbackSuccess,
                        null
                    );
                },
                data.name
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.resetScopeSecret = function (data, event) {

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "mask",
                "Reset Scope Secret",
                "This will lock out all scan agents using the old secret. <br />Are you sure you want to reset the scope secret of <span class=\"ui red text\">'" + data.name + "'</span>?",
                function () {

                    // Handle request success
                    const callbackSuccess = function (response, textStatus, jqXHR) {

                        // Show toast message for user
                        toast(response.message, "success");
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/scope/secret",
                        {},
                        {"id": data.id},
                        callbackSuccess,
                        null
                    );
                }
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.deleteScope = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "trash alternate outline",
                "Delete Scope",
                "This will delete all scan progress, associated views and access rights. <br />Are you sure you want to delete the scan scope <span class=\"ui red text\">'" + data.name + "'</span>?",
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
                        "id": data.id,
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/scope/delete",
                        {},
                        reqData,
                        callbackSuccess,
                        null
                    );
                },
                data.name // This optional argument adds a confirmation input to the modal
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showScopeAdd = function (data, event, type) {

            // Dispose open form
            this.parent.actionName(null);
            this.parent.actionComponent(null);

            // Clear form data
            this.parent.actionArgs(null);

            // Retrieve component to load
            var component = components[type]

            // Show new form
            this.parent.actionName(type);
            this.parent.actionComponent(component);
            this.parent.actionComponentRecent(type);
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showScopeAddRecent = function (data, event) {
            this.showScopeAdd(data, event, this.parent.actionComponentRecent())
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Dispose open form
            this.parent.actionName(null);
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
