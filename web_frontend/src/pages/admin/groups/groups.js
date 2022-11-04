/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./groups.html", "postbox", "jquery", "tabulator-tables", "semantic-ui-form", "moment", "utils-tabulator", "semantic-ui-popup"],
    function (ko, template, postbox, $, Tabulator) {

        // Tabulator callback sending row changes to the backend
        function tabulatorSave(cell) {

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                cell.restoreOldValue();
            };

            // This callback is called any time a cell is edited.
            var reqData = cell.getData();

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

        // Tabulator formatter function creating a clickable delete button in a given cell
        function fnConfigButton(vModel) {
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

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.sideNavItems = ko.observableArray([
                new NavItem("Users", "#admin/users", ""),
                new NavItem("Groups", "#admin/groups", ""),
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
                postbox.publish("redirect", "home");
                return;
            }

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divGroups');
            this.$componentGroupsAdd = $('#divGroupAdd');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize tabulator grid
            this.grid = new Tabulator("#divDataGridGroups", {
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
                        width: 160,
                    },
                    {
                        field: 'max_scopes',
                        title: 'Max Scopes',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "number",
                        validator: ["required", "min:-1"],
                    },
                    {
                        field: 'max_views',
                        title: 'Max Views',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "number",
                        validator: ["required", "min:-1"],
                    },
                    {
                        field: 'max_targets',
                        title: 'Max Targets',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "number",
                        validator: ["required", "min:-1"],
                    },
                    {
                        field: 'max_owners',
                        title: 'Max Owners',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "number",
                        validator: ["required", "min:-1"],
                    },
                    {
                        field: 'administrators',
                        align: "right",
                        width: 40,
                        headerSort: false,
                        formatter: fnConfigButton(this),
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

                // Sort reverse to show newest users first
                var groups = response["body"]["groups"].reverse()

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
        };

        return {viewModel: ViewModel, template: template};
    }
);