/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./access.html", "postbox", "jquery", "chartjs"],
    function (ko, template, postbox, $) {

        function getCompanyString(user) {

            // Return empty company name if it can't be retrieved anymore,
            // which may happen if the user associated with this event is deleted.
            if (user === null) {
                return ""
            }

            // Prepare default company value
            var company = user.company

            // For Siemens, split by business unit
            if (company === "SIEMENS") {
                var dep = user.department.split(" ")[0]
                company = company + " (" + dep + ")"
            }

            // Return result
            return company
        }

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.dataset = ko.observable(null);
            this.datasetAggregate = ko.observable(false)
            this.datasetOrder = ko.observable("company") // data key

            // Keep reference THIS view model context
            var ctx = this;

            // Subscribe to data or configuration changes to build updated chart
            this.datasetAggregate.subscribe(function (data) {
                ctx.buildChart();
            });
            this.datasetOrder.subscribe(function (data) {
                ctx.buildChart();
            });

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divAccess');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Load and set initial data
            this.loadData();
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Reference dataset in view
                ctx.dataset(response.body["events"])

                // Build chart with data
                ctx.buildChart();

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: "scale",
                    reverse: "false", // default setting
                    duration: 200
                });
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: "scale",
                    reverse: "auto", // default setting
                    duration: 200
                });
            };

            // Prepare request body
            var reqData = {
                event: "Database Password",
                since: moment().subtract(365, "days").format(datetimeFormatGolang),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/admin/events",
                {},
                reqData,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.toggleAggregate = function (data, event) {
            this.datasetAggregate(!this.datasetAggregate())
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.toggleOrder = function (data, event) {
            switch (this.datasetOrder()) {
                case("company"):
                    this.datasetOrder("count")
                    break;
                case("count"):
                    this.datasetOrder("company")
                    break;
                default:
                    break;
            }
        }

        // (Re-)Builds the chart based on current data
        ViewModel.prototype.buildChart = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Prepare chart values
            var labels = []
            var dataCompanyInterval = {}
            var colors = [
                "#00b5ad", "rgb(255, 99, 132)", "rgb(153, 102, 255)", "rgb(255, 205, 86)",
                "#009c95", "rgb(54, 162, 235)", "rgb(255, 159, 64)", "rgb(167,196,243)",
                "#db2828", "rgb(75, 192, 192)", "#DC8CC4", "#04646C",
                "#0577b4", "#046564", "#E04868", "#FCD06A",
                "#2E999B", "#F4742C",
            ]

            // Initialize data series for all companies
            this.dataset().forEach(function (item, index) {
                var company = getCompanyString(item.user)

                // Prepare company entry
                if (!(company in dataCompanyInterval)) {
                    dataCompanyInterval[company] = {}
                }
            });

            // Initialize the same intervals for all companies
            this.dataset().forEach(function (item, index) {

                // Define interval
                var interval = 'CW-' + moment(item.timestamp, datetimeFormatGolang).week();

                // Add interval to labels if missing
                if (labels.indexOf(interval) === -1) {
                    labels.push(interval)
                }

                // Initialize company interval with initial zero
                for (var company in dataCompanyInterval) {
                    dataCompanyInterval[company][interval] = 0
                }
            });

            // Assign events to companies and intervals (fill interval values)
            this.dataset().forEach(function (item, index) {
                var company = getCompanyString(item.user)

                // Add data
                var interval = 'CW-' + moment(item.timestamp, datetimeFormatGolang).week();
                dataCompanyInterval[company][interval]++
            });

            // Transform data into ChartJs format
            var i = 0
            var dataset = []
            var companies = Object.keys(dataCompanyInterval).sort() // Sort keys to assign same colors each time (dictionaries are random order)
            for (const company of companies) {
                if (dataCompanyInterval.hasOwnProperty(company)) {

                    // Prepare ChartJs data series
                    var d = []
                    var previous = 0
                    for (var interval in dataCompanyInterval[company]) {
                        if (dataCompanyInterval[company].hasOwnProperty(interval)) {
                            var val = dataCompanyInterval[company][interval] + previous // Aggregate data with each interval
                            d.push(val)

                            // Aggregate data values if desired
                            if (this.datasetAggregate() === true) {
                                previous = val
                            }
                        }
                    }

                    // Add ChartJs dataset
                    dataset.push({
                        label: company,
                        data: d,
                        borderColor: colors[i % colors.length],
                        backgroundColor: colors[i % colors.length],
                        borderWidth: 0,
                        fill: true,
                    })

                    // Increment counter
                    i++
                }
            }

            // Sort by company name (to have similar ones next to each other)
            if (this.datasetOrder() === "company") {
                dataset.sort((a, b) => (a.label > b.label ? 1 : -1))
            } else if (this.datasetOrder() === "count") {
                dataset.sort(
                    function (a, b) {
                        var sumA
                        var sumB
                        if (ctx.datasetAggregate() === true) {
                            sumA = a.data[a.data.length - 1]
                            sumB = b.data[b.data.length - 1]
                        } else {
                            sumA = a.data.reduce((a, b) => a + b, 0)
                            sumB = b.data.reduce((a, b) => a + b, 0)
                        }
                        if (sumA < sumB) {
                            return 1
                        } else if (sumA === sumB) {
                            return 0
                        } else {
                            return -1
                        }
                    }
                )
            }

            // Initialize chart if there is actual data
            if (dataset.length > 0) {

                // Configure chart
                var config = {
                    type: "line",
                    data: {
                        labels: labels,
                        datasets: dataset
                    },
                    options: {
                        elements: {
                            point: {
                                radius: 0
                            }
                        },
                        tooltips: {
                            callbacks: {
                                label: function (tooltipItem, data) {
                                    var label = data.datasets[tooltipItem.datasetIndex].label || '';

                                    if (label) {
                                        label += ': ';
                                    }
                                    label += isNaN(tooltipItem.yLabel) ? '0' : tooltipItem.yLabel;
                                    return label;
                                }
                            }
                        },
                        maintainAspectRatio: false,
                        responsive: true,
                        plugins: {
                            title: {
                                display: false,
                            },
                            tooltip: {
                                mode: "index",
                                filter: function (tooltipItem) {
                                    return tooltipItem.raw > 0 // Only show tooltip for event counts greater zero
                                }
                            },
                        },
                        interaction: {
                            mode: "nearest",
                            axis: "x",
                            intersect: false
                        },
                        scales: {
                            x: {
                                title: {
                                    display: true,
                                    text: "Calendar Week"
                                }
                            },
                            y: {
                                stacked: true,
                                title: {
                                    display: true,
                                    text: "Accesses"
                                }
                            }
                        }
                    }
                };

                // Destroy previous chart (if exists)
                if (this.chartAccess !== undefined) {
                    this.chartAccess.destroy();
                }

                // Initialize chart
                this.chartAccess = new Chart(
                    document.getElementById("canvasAccess"),
                    config
                );

                // Resize chart
                this.chartAccess.canvas.parentNode.style.height = "250px";
            }
        }

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
