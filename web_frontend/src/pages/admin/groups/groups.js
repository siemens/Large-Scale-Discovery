/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./groups.html", "postbox", "jquery", "tabulator-tables", "semantic-ui-form", "moment", "utils-tabulator", "semantic-ui-popup"],
    function (ko, template, postbox, $, Tabulator) {

        // Tabulator callback sending row changes to the backend
        function tabulatorSave(cell) {

            // Prepare function to execute API request
            var fnUpdateGroup = function () {

                // Handle request error
                const callbackError = function (response, textStatus, jqXHR) {
                    cell.restoreOldValue();
                };

                // This callback is called any time a cell is edited.
                var reqData = cell.getData();

                // Convert back database ID from string to integer. I was converted to string onload to
                // mitigate a bug in Tabulator.
                reqData.db_server_id = parseInt(reqData.db_server_id, 10)

                // Send request
                apiCall(
                    "POST",
                    "/api/v1/admin/group/update",
                    {},
                    reqData,
                    null,
                    callbackError
                );
            }

            // Execute API request either after modal or without modal, depending on changed key
            if (cell.getField() === "db_server_id") {

                // Request approval and only proceed if action is approved
                var proceed = false
                var groupName = cell.getRow().getCell("name").getValue()
                confirmOverlay(
                    "server icon",
                    "Change Database",
                    "This will change the databse server for <u><b>new</b></u> scan scopes of group <span class=\"ui red text\">'" + groupName + "'</span>. <br />Existing scan scopes will <b>remain on the current</b> database server. <br />Are you sure you want to continue?",
                    function () {
                        fnUpdateGroup()
                    },
                    null,
                    function () {
                        cell.restoreOldValue();
                    }
                );
            } else {
                fnUpdateGroup()
            }
        }

        // Tabulator formatter function creating a clickable delete button in a given cell
        function fnEditButton(vModel) {
            return function (cell, formatterParams, onRendered) {

                // Get entry ID
                var row = cell.getRow();

                // Create icon button
                var button = document.createElement("i");
                button.classList.add("users");
                button.classList.add("cog");
                button.classList.add("icon");

                // Attach click event
                button.addEventListener("click", function (e) {

                    // Dispose open form
                    vModel.actionArgs(null);
                    vModel.actionComponent(null);

                    // Set entry ID the opened form should operate on. The value is set in this view and can be accessed
                    // by the action component
                    vModel.actionArgs(row.getData());

                    // Show new form
                    vModel.actionComponent("groups-owners");
                });

                // Set button as content
                return button;
            }
        }

        // Tabulator formatter function creating a clickable delete button in a given cell
        function fnDeleteButton(cell, formatterParams, onRendered) {

            // Get entry ID
            var row = cell.getRow();

            // Create icon button
            var button = document.createElement("i");
            button.classList.add("trash");
            button.classList.add("alternate");
            button.classList.add("outline");
            button.classList.add("icon");

            // Attach click event
            button.addEventListener("click", function (e) {

                // Request approval and only proceed if action is approved
                confirmOverlay(
                    "trash alternate outline",
                    "Delete Group",
                    "This will remove all associated SCAN SCOPES, VIEWS and ACCESS RIGHTS. <br />Are you sure you want to delete the group <span class=\"ui red text\">'" + row.getData().name + "'</span>?",
                    function () {

                        // Handle request success
                        const callbackSuccess = function (response, textStatus, jqXHR) {

                            // Show toast message for user
                            toast(response.message, "success");

                            // Delete row from table
                            row.delete();
                        };

                        // Send request
                        apiCall(
                            "POST",
                            "/api/v1/admin/group/delete",
                            {},
                            {id: row.getIndex()},
                            callbackSuccess,
                            null
                        );
                    },
                    row.getData().name // This optional argument adds a confirmation input to the modal
                );
            });

            // Set button as content
            return button;
        }

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Initialize observables
            this.sideNavItems = ko.observableArray([
                new NavItem("Users", "#admin/users", ""),
                new NavItem("Groups", "#admin/groups", ""),
                new NavItem("Databases", "#admin/databases", ""),
                new NavItem("Query Logs", "#admin/logs", ""),
            ]);
            this.actionComponent = ko.observable(null); // action form that should be shown
            this.actionArgs = ko.observable(null); // action element row to work on
            this.databases = ko.observable({});

            // Check authentication and redirect to login if necessary
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }

            // Check privileges and redirect to home if necessary
            if (userAdmin() === false) {
                postbox.publish("redirect", home());
                return;
            }

            // Prepare array of references to subscriptions in order to dispose them later
            this.subscriptions = []

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divGroups');
            this.$componentGroupsAdd = $('#divGroupAdd');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Keep reference THIS view model context
            var ctx = this;

            // Define formatter for database names.
            // Define it within ViewModel to make sure databse observable is accessible.
            function fnDatabaseFormatter(cell, formatterParams, onRendered) {
                return ctx.databases()[cell.getValue()]
            }

            // Subscribe to data or configuration changes to build updated chart
            this.subscriptions.push(this.databases.subscribe(function (data) {

                // Initialize tabulator grid
                ctx.grid = new Tabulator("#divDataGridGroups", {
                    index: "id",
                    layout: "fitColumns",
                    movableRows: true,
                    movableColumns: true,
                    tooltips: true,
                    tooltipsHeader: true,
                    pagination: "local",
                    paginationSize: 10,
                    paginationSizeSelector: [5, 10, 20, 50, 100, 1000, 10000],
                    paginationButtonCount: 10,
                    cellMouseOver: tabulatorCursorEditable,
                    cellEdited: tabulatorSave,
                    columns: [
                        {
                            field: 'name',
                            title: 'Name',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            width: 160,
                            editor: "input",
                            editorParams: {
                                search: true,
                            },
                            validator: ["unique"],
                        },
                        {
                            field: 'created',
                            title: 'Created',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            headerFilterFunc: tabulatorFilterDatetime,
                            sorter: "datetime",
                            sorterParams: {
                                format: datetimeFormat
                            },
                            formatter: "datetime",
                            formatterParams: {
                                outputFormat: datetimeFormat,
                                invalidPlaceholder: " "
                            },
                            width: 110,
                        },
                        {
                            field: 'max_scopes',
                            title: 'Scopes',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            editor: "number",
                            validator: ["required", "min:-1"],
                        },
                        {
                            field: 'max_views',
                            title: 'Views',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            editor: "number",
                            validator: ["required", "min:-1"],
                        },
                        {
                            field: 'max_targets',
                            title: 'Targets',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            editor: "number",
                            validator: ["required", "min:-1"],
                        },
                        {
                            field: 'max_owners',
                            title: 'Owners',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            editor: "number",
                            validator: ["required", "min:-1"],
                        },
                        {
                            field: 'allow_custom',
                            title: 'Custom Scopes',
                            headerFilter: "tickCross",
                            headerFilterParams: {"tristate": true},
                            formatter: "tickCross",
                            align: "center",
                            width: 60,
                            editor: "tickCross",
                        },
                        {
                            field: 'allow_network',
                            title: 'Network Imports',
                            headerFilter: "tickCross",
                            headerFilterParams: {"tristate": true},
                            formatter: "tickCross",
                            align: "center",
                            width: 60,
                            editor: "tickCross",
                        },
                        {
                            field: 'allow_asset',
                            title: 'Asset Imports',
                            headerFilter: "tickCross",
                            headerFilterParams: {"tristate": true},
                            formatter: "tickCross",
                            align: "center",
                            width: 60,
                            editor: "tickCross",
                        },
                        {
                            field: 'db_server_id',
                            title: 'Database',
                            headerFilter: "input",
                            headerFilterPlaceholder: "Filter...",
                            formatter: "lookup",
                            formatterParams: ctx.databases(),
                            align: "center",
                            width: 160,
                            editor: "select",
                            editorParams: {
                                values: ctx.databases(),
                            },
                        },
                        {
                            field: 'administrators',
                            align: "right",
                            width: 40,
                            headerSort: false,
                            formatter: fnEditButton(ctx),
                        },
                        {
                            field: 'delete',
                            align: "right",
                            width: 40,
                            headerSort: false,
                            formatter: fnDeleteButton,
                        }
                    ],
                });

                // Load and set initial data
                ctx.loadData();
            }));

            // Load databases to map DB IDs and DB names
            this.loadDatabases();
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadDatabases = function (data, event, callbackCompletion) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Prepare database map
                var dbs = {}
                response["body"]["databases"].forEach(function (item, index) {
                    dbs[item.id] = item.name
                })

                // Set database map
                ctx.databases(dbs);
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
        ViewModel.prototype.loadData = function (data, event, callbackCompletion) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Sort reverse to show newest users first
                var groups = response["body"]["groups"].reverse()

                // Convert Server IDs to string, since there is a bug in Tabulator, causing troubles
                // looking up integer values in the lookup select box:
                // https://github.com/olifolkerd/tabulator/issues/1648
                groups.forEach(function (item, index) {
                    item.db_server_id = "" + item.db_server_id
                })

                // Set table data
                ctx.grid.setData(groups);

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: 'scale',
                    reverse: 'auto', // default setting
                    duration: 200
                });

                // Execute completion callback if set
                if (callbackCompletion !== undefined) {
                    callbackCompletion()
                }
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: 'scale',
                    reverse: 'auto', // default setting
                    duration: 200
                });
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/groups",
                {},
                null,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showGroupAdd = function (data, event) {

            // Dispose open form
            this.actionArgs(null);
            this.actionComponent(null);

            // Show new form
            this.actionComponent("groups-add");
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Dispose open form
            this.actionArgs(null);
            this.actionComponent(null);

            // Dispose subscriptions
            for (var k in this.subscriptions) {
                this.subscriptions[k].dispose();
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);