/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

/*
 * Tabulator cell mouseover callback setting a text cursor on editable cells
 */
function tabulatorCursorEditable(e, cell) {
    var colDef = cell.getColumn().getDefinition();
    if (colDef["editor"] || colDef["editable"]) {
        $(cell.getElement()).css('cursor', 'text');
    }
}

/*
 * Tabulator filter for searching formatted datetime columns
 */
function tabulatorFilterDatetime(headerValue, rowValue, rowData, filterParams) {
    var formattedDatetime = moment(rowValue).format(datetimeFormat);
    return formattedDatetime.indexOf(headerValue) >= 0;
}

/*
 * Tabulator configuration for scan scope target editor
 */
function targetsTableConfig(targetsJson, fnDataChanged, fnInputReset) {

    var fnMaintainEmptyRow = function (cell) {
        var table = cell.getTable();
        var data = table.getData();
        var lastRow = 0;
        if (data.length === 0) {
            table.addRow({enabled: true});
        } else {
            var dataLastInput = data[data.length - 1].input;
            if (dataLastInput !== undefined && dataLastInput !== null && dataLastInput !== "") {
                table.addRow({enabled: true})
                    .then(function (row) {
                        var p = table.scrollToRow(row, "bottom", true);
                    });
            }
        }
    };

    var fnToggleCheckbox = function (e, cell) {
        cell.setValue(!cell.getValue());
    };

    var fnBoolToString = function (cell, formatterParams, onRendered) {
        return cell.getValue() ? "true" : "false";
    };

    var fnStringToBool = function (value, data, type, params, column) {
        value = value.toLowerCase()
        return value === 1 || value === "1" || value === "t" || value === "true" || value === "y" || value === "yes";
    };

    var fnDatetimeToString = function (cell, formatterParams, onRendered) {
        var dtime = cell.getValue()
        if (dtime === undefined || !dtime.Valid) {
            return "-"
        }
        return moment(dtime.Time).format(datetimeFormat)
    }

    var fnCellEdited = function (cell) {
        fnMaintainEmptyRow(cell);
    };

    var fnInputCheck = function (cell, formatterParams, onRendered) {
        var val = cell.getValue();
        if (val === null || val === undefined || isIpV4OrSubnet(val) || isIpV6OrSubnet(val) || isFqdn(val) || val === "localhost") {
            cell.getElement().style.color = "";
        } else {
            if (val !== "") {
                console.log("Invalid scan target:", val)
                cell.getElement().style.color = "red";
            }
        }
        return val;
    };

    var fnPasteParser = function (clipboard) {
        var data = [],
            headerFindSuccess = true,
            columns = this.table.columnManager.columns,
            columnMap = [],
            rows = [];

        //get data from clipboard into array of columns and rows.
        clipboard = clipboard.split("\n");

        clipboard.forEach(function (row) {
            data.push(row.split("\t"));
        });

        // Continue if there is at least one row
        if (data.length && data.length > 0) {

            //check if headers are present by title
            data[0].forEach(function (value) {
                var column = columns.find(function (column) {
                    return value && column.definition.title && value.trim() && column.definition.title.trim() === value.trim();
                });

                if (column) {
                    columnMap.push(column);
                } else {
                    headerFindSuccess = false;
                }
            });

            //check if column headers are present by field
            if (!headerFindSuccess) {
                headerFindSuccess = true;
                columnMap = [];

                data[0].forEach(function (value) {
                    var column = columns.find(function (column) {
                        return value && column.field && value.trim() && column.field.trim() === value.trim();
                    });

                    if (column) {
                        columnMap.push(column);
                    } else {
                        headerFindSuccess = false;
                    }
                });

                if (!headerFindSuccess) {
                    columnMap = this.table.columnManager.columnsByIndex;
                }
            }

            //remove header row if found
            if (headerFindSuccess) {
                data.shift();
            }

            data.forEach(function (item) {
                var row = {};

                item.forEach(function (value, i) {

                    //
                    // Shift columns by 3 (because of the columns (delete, restart, row#) we injected),
                    // if there were no clear headers contained within the paste data!
                    //
                    if (!headerFindSuccess) {
                        i = i + 3
                    }

                    if (columnMap[i]) {
                        row[columnMap[i].field] = value;
                    }
                });

                rows.push(row);
            });

            return rows;
        } else {
            return false;
        }
    };

    var fnPasteAction = function (rowData) {

        // Execute table's on change function
        if (fnDataChanged) {
            fnDataChanged();
        }

        // Execute paste action
        this.table.replaceData(rowData);
    }

    var fnCopyButton = function (cell, formatterParams, onRendered, a) {

        // Create icon button
        var button = document.createElement("i");
        button.classList.add("copy");
        button.classList.add("outline");
        button.classList.add("icon");

        // Keep reference to table
        var self = this;

        // Define button click action
        var clickListener = function (e) {

            // Disable copy button in the meantime
            addRunningRequest()
            button.classList.add("disabled");

            // Execute clipboard copy asynchronously with some delay, to make sure "disabled" is set first.
            // Clipboard copy consumes all CPU, so it might block parallel DOM changes.
            setTimeout(function () {

                // Copy data to clipboard
                self.table.copyToClipboard("all");

                // Flash success message and reset button
                // Wait some time to let potentially triggered redundant button clicks pass
                setTimeout(function () {

                    // Flash success message
                    toast("Copied to clipboard.", "success");

                    // Reset button and loader
                    button.classList.remove("disabled");
                    removeRunningRequest()

                    // Re-enable click event
                    addEventListenerOnce(button, "click", clickListener)
                }, 100)
            }, 500)

            // Don't execute further click events
            e.stopPropagation()
        }

        // Add event listener to button that only fires a single time
        addEventListenerOnce(button, "click", clickListener)

        // Return button
        return button;
    };

    // Tabulator formatter function creating a clickable delete button in a given cell
    var fnDeleteButton = function (cell, formatterParams, onRendered) {

        // Get entry ID
        var row = cell.getRow();
        var table = cell.getTable();

        // Create delete button
        var button = document.createElement("i");
        button.classList.add("trash");
        button.classList.add("alternate");
        button.classList.add("outline");
        button.classList.add("icon");

        // Attach click event
        button.addEventListener("click", function (e) {
            row.delete();
            fnMaintainEmptyRow(cell);
        });

        // Set button as content
        return button;
    };

    // Tabulator formatter function creating a clickable delete button in a given cell
    var fnResetButton = function (cell, formatterParams, onRendered) {

        // Get entry ID
        var row = cell.getRow();
        var data = row.getData();

        // Return no button if the line is new without scan timestamp, or if scan timestamp is still empty (scan hasn't started)
        if (!("scan_started" in data) || data.scan_started === undefined || !data.scan_started.Valid) {
            return null
        }

        // Create icon button
        var button = document.createElement("i");
        button.classList.add("history");
        button.classList.add("icon");

        // Keep reference to table
        var self = this;

        // Define button click action
        var clickListener = function (e) {

            // Disable copy button in the meantime
            button.classList.add("disabled");

            // Execute clipboard copy asynchronously with some delay, to make sure "disabled" is set first.
            // Clipboard copy consumes all CPU, so it might block parallel DOM changes.
            setTimeout(function () {

                // Send reset request to the backend
                fnInputReset(data, function (removeButton) {

                    // Flash success message and reset button
                    // Wait some time to let potentially triggered redundant button clicks pass
                    setTimeout(function () {

                        if (removeButton) {
                            // Remove button after reset was successful
                            button.remove()

                        } else {

                            // Reset button and loader
                            button.classList.remove("disabled");

                            // Re-enable click event
                            addEventListenerOnce(button, "click", clickListener)
                        }
                    }, 100)
                });
            }, 500)

            // Don't execute further click events
            e.stopPropagation()
        }

        // Add event listener to button that only fires a single time
        addEventListenerOnce(button, "click", clickListener)

        // Return button
        return button;
    };

    // Generate configuration dictionary for Tabulator table
    return {
        history: true,
        clipboard: true,
        clipboardPasteAction: fnPasteAction,
        clipboardCopyStyled: false,
        clipboardPasteParser: fnPasteParser,
        clipboardCopyConfig: {
            formatCells: false, //show raw cell values without formatter
        },
        maxHeight: 300,
        headerSort: false,
        data: targetsJson,
        cellEdited: fnCellEdited,
        dataChanged: fnDataChanged,
        columns: [
            {
                field: 'delete',
                title: '',
                width: 10,
                titleFormatter: fnCopyButton,
                formatter: fnDeleteButton,
                clipboard: false,
            },
            {
                field: 'reset',
                title: '',
                width: 10,
                formatter: fnResetButton,
                clipboard: false,
            },
            {
                field: 'index',
                title: '#',
                width: 65,
                formatter: "rownum",
                clipboard: false,
            },
            {
                field: 'input',
                title: 'Scan Input', // Address or range to be scanned
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
                formatter: fnInputCheck,
            },
            {
                field: 'enabled',
                title: 'Enabled', // Flag whether the address/range should be scanned at all
                width: 73,
                headerFilter: "select",
                headerFilterParams: {initial: "true", values: {"": "Any", "true": "Yes", "false": "No"}},
                headerFilterPlaceholder: "Filter...",
                hozAlign: "center",
                formatter: "tickCross",
                cellClick: fnToggleCheckbox,
                mutatorClipboard: fnStringToBool,
                formatterClipboard: fnBoolToString, // Excel might have issues with boolean values
            },
            {
                field: 'priority',
                title: 'Priority', // Flag whether the address/range should be scanned first, rather than being randomly selected
                width: 71,
                headerFilter: "select",
                headerFilterParams: {values: {"": "Any", "true": "Yes", "false": "No"}},
                headerFilterPlaceholder: "Filter...",
                hozAlign: "center",
                formatter: "tickCross",
                cellClick: fnToggleCheckbox,
                mutatorClipboard: fnStringToBool,
                formatterClipboard: fnBoolToString, // Excel might have issues with boolean values
            },
            {
                field: 'timezone',
                title: 'Timezone', // Timezone of the address/range, used to decide when to scan
                width: 85,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                hozAlign: "center",
                editor: "input",
            },
            {
                field: 'lat',
                title: 'Latitude', // Geographic coordinate of the address/range
                width: 90,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                hozAlign: "center",
                editor: "input",
            },
            {
                field: 'lng',
                title: 'Longitude', // Geographic coordinate of the address/range
                width: 90,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                hozAlign: "center",
                editor: "input",
            },
            {
                field: 'postal_address',
                title: 'Postal Address', // E.g. postal address of the address/range's location
                width: 200,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_network',
                title: 'Network', // E.g. the network range the input address belongs to
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                title: 'Country', // E.g. country name or code
                field: 'input_country',
                width: 100,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_location',
                title: 'Location', // E.g. city name or tag
                width: 100,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_routing_domain',
                title: 'Routing Domain', // Internet, intranet, global
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_zone',
                title: 'Zone', // E.g. type or kind of the address/range
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_purpose',
                title: 'Purpose', // E.g. type or kind of the address/range
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_company',
                title: 'Company', // E.g. company owning the address/range
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_department',
                title: 'Department', // E.g. company's department responsible for the address/range
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_manager',
                title: 'Manager', // E.g. person responsible for the address/range
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_contact',
                title: 'Contact', // E.g. incident response contact
                width: 120,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'input_comment',
                title: 'Comment', // Anything...
                width: 200,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: "input",
            },
            {
                field: 'scan_started',
                title: 'Scan Started', // Status information
                width: 160,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: false,
                formatter: fnDatetimeToString,
                clipboard: false,
                cssClass: "col-border col-disabled",
            },
            {
                field: 'scan_finished',
                title: 'Scan Finished', // Status information
                width: 160,
                headerFilter: true, headerFilterPlaceholder: "Filter...",
                editor: false,
                formatter: fnDatetimeToString,
                clipboard: false,
                cssClass: "col-disabled",
            },
        ],
    };
}
