/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./custom.html", "postbox", "jquery", "tabulator-tables", "semantic-ui-popup", "semantic-ui-dropdown", "semantic-ui-transition"],
    function (ko, template, postbox, $, Tabulator) {

        // tabulatorResetInput is a Tabulator function plugin to apply the reset function of a certain
        // input on a given scan scope
        function tabulatorResetInput(scopeId) {
            return function (data, callback) {

                // Request approval and only proceed if action is approved
                confirmOverlay(
                    "history",
                    "Reset Scan Target",
                    "The scan target <span class=\"ui red text\">'" + data.input + "'</span> will be queued again.<br />A second result set may appear for this scan cycle, distinguishable by scan time.",
                    function () {

                        // Handle request success
                        const callbackSuccess = function (response, textStatus, jqXHR) {

                            // Show toast message for user
                            toast(response.message, "success");

                            // Execute callback after event completion
                            callback(true)
                        };

                        // Handle request error
                        const callbackError = function (response, textStatus, jqXHR) {

                            // Execute callback after event completion
                            callback(false)
                        };

                        // Prepare basic request data
                        var reqData = {
                            scope_id: scopeId,
                            input: data.input,
                        }

                        // Send request
                        apiCall(
                            "POST",
                            "/api/v1/scope/target/reset",
                            {},
                            reqData,
                            callbackSuccess,
                            callbackError
                        );
                    },
                    null,
                    function () {

                        // Execute callback after event completion
                        callback(false)
                    }
                );
            }
        }

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Check whether to enable update mode
            this.updateMode = ko.observable(params.args !== null);

            // Initialize update-mode observables
            this.scopeId = ko.observable(-1);
            this.groupName = ko.observable("");
            this.synchronizationOngoing = ko.observable(null);
            this.edited = ko.observable(false);
            this.editing = ko.observable(false);

            // Initialize create-mode observables
            this.groupsAvailable = ko.observableArray([]);
            this.groupSelected = ko.observable(-1);

            // Initialize scope type independent observables
            this.scopeName = ko.observable("");
            this.scopeCycles = ko.observable(false);
            this.scopeRetention = ko.observable("All");

            // Initialize update mode, if scope details are passed
            if (this.updateMode()) {

                // Set update-mode observables
                this.scopeId(params.args.id);
                this.groupName(params.args.group_name);

                // Set scope type independent values
                this.scopeName(params.args.name);
                this.scopeCycles(params.args.cycles);
                if (params.args.cycles_retention > 0) {
                    this.scopeRetention(params.args.cycles_retention);
                }
            }

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divScopesAddCustom');
            this.$domForm = this.$domComponent.find("form");

            // Initialize dropdown elements
            this.$domComponent.find('select.dropdown').dropdown();

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();


            // Workaround hack to focus datatable before CTRL+V paste is triggered by 'tabulator'
            // data tables. This is for usability because users would need to know that the focus
            // needs to be on the datatable in order to receive the pasted data.
            // ATTENTION: This keydown event has to be removed again when the view is closed, otherwise
            //            it would be registered multiple times, when the view is opened again.
            var ctx = this;
            var ctrlDown = false, ctrlKey = 17, cmdKey = 91, vKey = 86
            $(document).keydown(function (e) {
                if (e.keyCode === ctrlKey || e.keyCode === cmdKey) ctrlDown = true;
            }).keyup(function (e) {
                if (e.keyCode === ctrlKey || e.keyCode === cmdKey) ctrlDown = false;
            });
            $(document).keydown(function (e) {
                if (ctrlDown && (e.keyCode === vKey)) {
                    if (!ctx.editing()) {
                        document.getElementsByClassName('tabulator-tableHolder')[0].focus()
                    }
                }
            });

            // Initialize form, depending on whether update mode is desired or not
            if (this.updateMode()) {

                // Initialize form validators
                this.$domForm.form({
                    fields: {
                        inputName: ['minLength[3]'],
                    },
                });

                // Load and set current scope targets
                this.loadData();
            } else {

                // Initialize tabulator grid for scan targets
                this.grid = new Tabulator(
                    "#divDataGridTargets",
                    targetsTableConfig([{enabled: true}],)
                );

                // Initialize form validators
                this.$domForm.form({
                    fields: {
                        inputName: ['minLength[3]'],
                        selectGroup: ['empty'],
                    },
                });

                // Load and set associated groups
                this.loadGroups();
            }

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

                // Set flag indicating current synchronization
                ctx.synchronizationOngoing(response.body.synchronization)

                // Prepare scope targets and initialize table if no synchronization ongoing currently
                if (response.body.synchronization === false) {

                    // Prepare target list
                    var data = []
                    if (response.body.targets !== null) {
                        data = response.body.targets
                    }
                    data.push({enabled: true})

                    // Initialize tabulator grid for scan targets
                    ctx.grid = new Tabulator(
                        "#divDataGridTargets",
                        targetsTableConfig(
                            data,
                            function (data) {
                                ctx.editing(!ctx.editing())
                            },
                            function (data) {
                                ctx.edited(true)
                            },
                            tabulatorResetInput(ctx.scopeId())
                        )
                    );
                }
            };

            // Prepare basic request data
            var reqData = {
                id: this.scopeId()
            }

            // Send request to get groups
            apiCall(
                "POST",
                "/api/v1/scope/targets",
                {},
                reqData,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.loadGroups = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Init array of groups
                ctx.groupsAvailable(response.body["groups"]);

                // Set scope group, if there is only one
                if (response.body["groups"].length === 1) {
                    ctx.groupSelected(response.body["groups"][0].id);
                }
            };

            // Send request to get groups
            apiCall(
                "GET",
                "/api/v1/groups",
                {},
                null,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.scopeRetentionAdd = function (data, event) {
            var current = this.scopeRetention()
            if (current === "All") {
                this.scopeRetention(1)
            } else {
                this.scopeRetention(current + 1)
            }
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.scopeRetentionSub = function (data, event) {
            var current = this.scopeRetention()
            if (current > 1) {
                this.scopeRetention(current - 1)
            } else {
                this.scopeRetention("All")
            }
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.submitCustom = function (data, event) {

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

                    // Let user know that scan input population is continued in the background
                    if (ctx.edited() === true) {
                        toast("Target synchronization may take some time.", "info");
                    }

                    // Unlink component
                    ctx.dispose(data, event)
                });
            };

            // Prepare retention value as expected by the backend
            var retention = -1
            if (this.scopeRetention() > 0) {
                retention = this.scopeRetention()
            }

            // Prepare basic request data
            var reqData = {
                name: this.scopeName(),
                cycles: this.scopeCycles(),
                cycles_retention: retention,
            }

            // Append target data if changed
            if (this.updateMode() === false || this.edited() === true) {

                // Get inserted targets
                var targets = this.grid.getData();
                var sanitizationResult = sanitizeTargets(targets);
                targets = sanitizationResult[0];
                var errorMsg = sanitizationResult[1];
                if (errorMsg !== "") {
                    toast(errorMsg, "error");
                    this.$domForm.form("add prompt", "textareaTargets", errorMsg);
                    this.$domForm.each(shake);
                    return
                }

                // Add updated targets to request data
                reqData["targets"] = targets
            }

            // Send create / update request
            if (this.updateMode()) {

                // Set scope ID to indicate scope update
                reqData["scope_id"] = this.scopeId() // Required to update associated scope
            } else {

                // Set group ID to indicate new scope
                reqData["group_id"] = this.groupSelected() // Required to create new scope within
            }

            // Send request
            apiCall(
                "POST",
                "/api/v1/scope/update/custom",
                {},
                reqData,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Remove keydown listener, otherwise it would be registered multiple times
            // next time when this view is opened
            $(document).off("keydown");

            // Hide form
            this.$domComponent.transition('fade up');

            // Reset form fields
            this.$domForm.form('reset');

            // Dispose open form
            if (this.parent.actionComponent() === "scopes-add-custom") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
