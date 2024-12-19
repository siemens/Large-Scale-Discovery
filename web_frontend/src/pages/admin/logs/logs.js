/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./logs.html", "postbox", "jquery", "tabulator-tables", "moment", "utils-tabulator", "semantic-ui-popup"],
    function (ko, template, postbox, $, Tabulator) {

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
            this.$domComponent = $('#divLogs');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize tabulator grid
            this.grid = new Tabulator("#divDataGridLogs", {
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
                columns: [
                    {
                        field: 'query',
                        title: 'Query',
                        width: 250,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'db_table',
                        title: 'Table',
                        width: 90,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'db_user',
                        title: 'User',
                        width: 90,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'db_name',
                        title: 'Database',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'query_timestamp',
                        title: 'Timestamp',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
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
                        field: 'query_duration_string',
                        title: 'Exec Time',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'total_duration_string',
                        title: 'Total Time',
                        width: 110,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'query_results',
                        title: 'Results',
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
                    {
                        field: 'client_name',
                        title: 'Client',
                        width: 210,
                        headerFilter: "input",
                        headerFilterPlaceholder: "Filter...",
                    },
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
                var logs = response["body"]["logs"]
                if (logs !== null) {
                    logs = logs.reverse()
                }

                // Set table data
                ctx.grid.setData(logs);

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

            // Prepare request body
            var reqData = {
                db_name: "",
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/admin/sql",
                {},
                reqData,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
