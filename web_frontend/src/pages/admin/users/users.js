/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./users.html", "postbox", "jquery", "tabulator-tables", "moment", "utils-tabulator", "semantic-ui-popup"],
    function (ko, template, postbox, $, Tabulator) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.sideNavItems = ko.observableArray([
                new NavItem("Users", "#admin/users", ""),
                new NavItem("Groups", "#admin/groups", ""),
            ]);

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
            this.$domComponent = $('#divUsers');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize tabulator grid
            this.grid = new Tabulator("#divDataGridUsers", {
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
                        field: 'email',
                        title: 'E-Mail',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editorParams: {
                            search: true,
                        },
                    },
                    {
                        field: 'active',
                        title: 'Active',
                        headerFilter: "tickCross",
                        headerFilterParams: {"tristate": true},
                        formatter: "tickCross",
                        align: "center",
                        width: 60,
                        editor: "tickCross",
                    },
                    {
                        field: 'admin',
                        title: 'Admin',
                        headerFilter: "tickCross",
                        headerFilterParams: {"tristate": true},
                        formatter: "tickCross",
                        align: "center",
                        width: 60,
                        editor: "tickCross",
                    },
                    {
                        field: 'name',
                        title: 'Name',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                        editorParams: {
                            search: true,
                        },
                    },
                    {
                        field: 'surname',
                        title: 'Surname',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                    },
                    {
                        field: 'company',
                        title: 'Company',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                        editor: "input",
                    },
                    {
                        field: 'last_login',
                        title: 'Last Login',
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
        ViewModel.prototype.loadData = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Sort reverse to show newest users first
                var users = response["body"]["users"].reverse()

                // Set table data
                ctx.grid.setData(users);

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: 'scale',
                    reverse: 'auto', // default setting
                    duration: 200
                });
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
                "/api/v1/users",
                {},
                null,
                callbackSuccess,
                callbackError
            );
        };

        // Tabulator callback sending row changes to the backend
        function tabulatorSave(cell) {

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                cell.restoreOldValue();
            };

            // Prepare request body
            var reqData = cell.getData();

            // Send request
            apiCall(
                "POST",
                "/api/v1/admin/user/update",
                {},
                reqData,
                null,
                callbackError
            );
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
                    "Delete User",
                    "This will remove all access rights and privileges. <br />Are you sure you want to delete the user <span class=\"ui red text\">'" + row.getData().email + "'</span>?",
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
                            "/api/v1/admin/user/delete",
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

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
