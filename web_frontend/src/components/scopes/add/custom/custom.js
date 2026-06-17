/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
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

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
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
            this.scopeOt = ko.observable(false);
            this.scopeRetention = ko.observable("All");

            // Coordinate Tabulator initialization with the fade-in transition. Initializing the grid
            // while the container is still being animated leads to Tabulator's adjustTableSize/redraw
            // infinite recursion. fadeComplete is set by the transition's onComplete callback below,
            // pendingGridInit holds a deferred init function if the grid was ready before the animation.
            this.fadeComplete = false;
            this.pendingGridInit = null;

            // Initialize update mode, if scope details are passed
            if (this.updateMode()) {

                // Set update-mode observables
                this.scopeId(params.args.id);
                this.groupName(params.args.group_name);

                // Set scope type independent values
                this.scopeName(params.args.name);
                this.scopeCycles(params.args.cycles);
                this.scopeOt(params.args.scan_settings.ot);
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

            // Keep reference THIS view model context
            var ctx = this;

            // Workaround hack to focus datatable before CTRL+V paste is triggered by 'tabulator'
            // data tables. This is for usability because users would need to know that the focus
            // needs to be on the datatable in order to receive the pasted data.
            // ATTENTION: This keydown event has to be removed again when the view is closed, otherwise
            //            it would be registered multiple times, when the view is opened again.
            var ctrlDown = false, ctrlKey = 17, cmdKey = 91, vKey = 86
            $(document).keydown(function (e) {
                if (e.keyCode === ctrlKey || e.keyCode === cmdKey) ctrlDown = true;
            }).keyup(function (e) {
                if (e.keyCode === ctrlKey || e.keyCode === cmdKey) ctrlDown = false;
            });
            $(document).keydown(function (e) {

                // Set paste focus if
                if (ctrlDown && (e.keyCode === vKey)) {

                    // Don't set paste focus if user is editing filter field, user might try to paste a string
                    if (document.activeElement.tagName === "INPUT" && document.activeElement.type === "search") {
                        return
                    }

                    // Don't set paste focus if user is editing cell data, user might try to paste a string
                    if (ctx.editing()) {
                        return
                    }

                    // Set paste focus in order to allow insert from clipboard
                    document.getElementsByClassName('tabulator-tableHolder')[0].focus()
                }
            });

            // Initialize form, depending on whether update mode is desired or not
            if (this.updateMode()) {

                // Initialize form with validators. keyboardShortcuts is disabled because
                // Semantic UI's Enter handler would submit the form a second time alongside
                // the browser's native submit that Knockout's submit binding already handles.
                this.$domForm.form({
                    fields: {
                        inputName: ['minLength[3]'],
                    },
                    keyboardShortcuts: false, // Prevent FomanticUI's own submit action handler from submitting again
                });

                // Load and set current scope targets
                this.loadData();
            } else {

                // Defer Tabulator initialization until the fade-down transition has completed,
                // otherwise Tabulator measures a still-animating container and falls into a
                // adjustTableSize/redraw infinite recursion.
                var initGrid = function () {
                    if (!document.getElementById("divDataGridTargets")) {
                        return; // Component disposed in the meantime
                    }

                    // Initialize tabulator grid for scan targets
                    ctx.grid = new Tabulator(
                        "#divDataGridTargets",
                        targetsTableConfig(
                            [{enabled: true}],
                            undefined,
                            undefined,
                            undefined,
                            function () {
                                return ctx.scopeOt();
                            }
                        )
                    );

                    // Redraw grid when OT toggle changes so scan input column re-renders and collapse
                    // to a single row (preserving the first row's attributes) when OT is enabled.
                    ctx.scopeOtSubscription = ctx.scopeOt.subscribe(function (enabled) {
                        if (ctx.grid) {
                            if (enabled) {
                                var data = ctx.grid.getData();
                                var row = data.length > 0 ? data[0] : {};
                                row.input = "";
                                row.enabled = true;
                                ctx.grid.replaceData([row]);
                            }
                            ctx.grid.redraw(true);
                        }
                    });
                };
                if (this.fadeComplete) {
                    initGrid();
                } else {
                    this.pendingGridInit = initGrid;
                }

                // Initialize form with validators. keyboardShortcuts is disabled because
                // Semantic UI's Enter handler would submit the form a second time alongside
                // the browser's native submit that Knockout's submit binding already handles.
                this.$domForm.form({
                    fields: {
                        inputName: ['minLength[3]'],
                        selectGroup: ['notEmpty'],
                    },
                    keyboardShortcuts: false, // Prevent FomanticUI's own submit action handler from submitting again
                });

                // Load and set associated groups
                this.loadGroups();
            }

            // Fade in. Use the verbose form so we get an onComplete callback to coordinate with grid init.
            var ctxFade = this;
            this.$domComponent.transition({
                animation: 'fade down',
                onComplete: function () {
                    ctxFade.fadeComplete = true;
                    if (ctxFade.pendingGridInit) {
                        var fn = ctxFade.pendingGridInit;
                        ctxFade.pendingGridInit = null;
                        fn();
                    }
                }
            });

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

                    // Add additional line for users to add items, if OT discovery is not enabled
                    if (!ctx.scopeOt()) {
                        data.push({enabled: true})
                    }

                    // Defer Tabulator initialization until the fade-down transition has completed,
                    // otherwise Tabulator measures a still-animating container and falls into a
                    // adjustTableSize/redraw infinite recursion.
                    var initGrid = function () {
                        if (!document.getElementById("divDataGridTargets")) {
                            return; // Component disposed in the meantime
                        }

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
                                tabulatorResetInput(ctx.scopeId()),
                                function () {
                                    return ctx.scopeOt();
                                }
                            )
                        );

                        // Redraw grid when OT toggle changes so Scan Input column re-renders
                        if (ctx.scopeOtSubscription) {
                            ctx.scopeOtSubscription.dispose();
                        }
                        ctx.scopeOtSubscription = ctx.scopeOt.subscribe(function () {
                            if (ctx.grid) {
                                ctx.grid.redraw(true);
                            }
                        });
                    };
                    if (ctx.fadeComplete) {
                        initGrid();
                    } else {
                        ctx.pendingGridInit = initGrid;
                    }
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

                    // Show toast messages for warnings
                    if (response.body !== null && response.body.warnings !== null && response.body.warnings.length > 0) {
                        response.body.warnings.forEach(function (item, index) {
                            toast(item, "warning", false, 6000);
                        })
                    }

                    // Let user know that scan input population is continued in the background
                    if (ctx.edited() === true) {
                        toast("Target synchronization may take some time.", "teal");
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
                ot: this.scopeOt(),
                cycles_retention: retention,
            }

            // Helper variable for more natural understanding
            var create = !this.updateMode()

            // Prepare list of targets depending on scan scope mode (OT vs normal)
            if (this.scopeOt()) { // OT scanning (create & update)

                // OT scopes always send exactly one target row
                var otRow = this.grid.getData()[0] || {enabled: true};
                otRow.input = otRow.input || "";

                // Sanitize data columns of target
                sanitizeTarget(otRow);

                // Add updated target to request data
                reqData["targets"] = [otRow];

            } else { // Normal scanning

                // Append target data if changed
                if (create || this.edited() === true) {

                    // Sanitize data columns of targets
                    var targets = sanitizeTargets(this.grid.getData());

                    // Validate input column of targets
                    var errorMsg = validateInputs(targets);
                    if (errorMsg !== "") {
                        toast(errorMsg, "error");
                        this.$domForm.form("add prompt", "textareaTargets", errorMsg);
                        this.$domForm.each(shake);
                        return
                    }

                    // Add updated targets to request data
                    reqData["targets"] = targets
                }
            }

            // Send create / update request
            if (create) {

                // Set group ID to indicate new scope
                reqData["group_id"] = this.groupSelected() // Required to create new scope within
            } else {

                // Set scope ID to indicate scope update
                reqData["scope_id"] = this.scopeId() // Required to update associated scope
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

            // Dispose OT-toggle subscription that triggers grid redraw
            if (this.scopeOtSubscription) {
                this.scopeOtSubscription.dispose();
                this.scopeOtSubscription = null;
            }

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
