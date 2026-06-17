/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./databases.html", "postbox", "jquery", "tabulator-tables", "moment", "utils-tabulator", "semantic-ui-popup"],
    function (ko, template, postbox, $, Tabulator) {

        // Tabulator formatter function replacing unknown passwords with * placeholder
        var passwordPlaceholder = "**********"

        // Tabulator callback sending row changes to the backend
        function tabulatorSave(cell) {

            // Abort second request that gets triggered automatically after cell.setValue() is called below
            if (cell.getField() === "password" && (cell.getValue() === undefined || cell.getValue() === "")) {
                return
            }

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Discard password from cell data
                if (cell.getField() === "password") {
                    cell.setValue(undefined)
                }
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                cell.restoreOldValue();
            };

            // Prepare request body
            var reqData = cell.getData();
            if (typeof reqData.port === "string") {
                reqData.port = parseInt(reqData.port.valueOf())
            }

            // Send request
            apiCall(
                "POST",
                "/api/v1/admin/database/update",
                {},
                reqData,
                callbackSuccess,
                callbackError
            );
        }

        function fnPasswordFormatter(cell, formatterParams, onRendered) {
            if (cell.value === undefined || cell.value === "") {
                return passwordPlaceholder
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
                    "Delete Database",
                    "This will remove a database server if no scan scopes are currently stored on it. <br />Are you sure you want to delete the database server <span class=\"ui red text\">'" + row.getData().name + "'</span>?",
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
                            "/api/v1/admin/database/remove",
                            {},
                            {id: row.getIndex()},
                            callbackSuccess,
                            null
                        );
                    }
                );
            });

            // Set button as content
            return button;
        }

        var fnIsHostnameValidator = function (cell, value) {
            //cell - the cell component for the edited cell
            //value - the new input value of the cell
            //parameters - the parameters passed in with the validator
            return isIpOrFqdn(value); //don't allow values divisible by divisor ;
        }

        var fnIsPortValidator = function (cell, value) {
            //cell - the cell component for the edited cell
            //value - the new input value of the cell
            //parameters - the parameters passed in with the validator
            if (typeof value !== "number") {
                value = parseInt(value, 10)
            }
            return value >= 0 && value <= 65535
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

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divDatabases');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize tabulator grid
            this.grid = new Tabulator("#divDataGridDatabases", {
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
                        title: 'DB Name',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                    },
                    {
                        field: 'dialect',
                        title: 'Dialect',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                    },
                    {
                        field: 'host',
                        title: 'Host',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                        validator: [
                            {
                                type: fnIsHostnameValidator
                            }
                        ],
                    },
                    {
                        field: 'host_public',
                        title: 'Host Public',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                        validator: [
                            {
                                type: fnIsHostnameValidator
                            }
                        ],
                    },
                    {
                        field: 'port',
                        title: 'Port',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                        validator: [
                            {
                                type: fnIsPortValidator
                            }
                        ],
                    },
                    {
                        field: 'args',
                        title: 'Args',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                    },
                    {
                        field: 'admin',
                        title: 'Admin',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                    },
                    {
                        field: 'password',
                        title: 'Password',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        formatter: fnPasswordFormatter,
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                        headerSort: false,
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
            this.loadData();
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event, callbackCompletion) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Sort reverse to show newest databases first
                var databases = response["body"]["databases"].reverse()

                // Set table data
                ctx.grid.setData(databases);

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
                "/api/v1/admin/databases",
                {},
                null,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.showDatabaseAdd = function (data, event) {

            // Dispose open form
            this.actionArgs(null);
            this.actionComponent(null);

            // Show new form
            this.actionComponent("databases-add");
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
