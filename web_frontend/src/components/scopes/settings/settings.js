/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./settings.html", "postbox", "jquery", "tabulator-tables", "semantic-ui-dropdown", "semantic-ui-calendar"],
    function (ko, template, postbox, $, Tabulator, mod1, mod2, str) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Store parameters passed by parent
            this.args = ko.observable(params.args); // Args is already form data, so it should be an observable

            // Prepare timespans data
            var timespansCopy = null
            if (params.args["scan_settings"]["discovery_timespans"] !== null) {

                // Create copy, otherwise array data would be manipulated (passed by reference)
                timespansCopy = params.args["scan_settings"]["discovery_timespans"].slice()
            }

            // Prepare active timespans of scan scope
            this.timespans = ko.observableArray(timespansCopy)

            // Patch IDs into timespans that can be referenced by timespan component to callback parent
            var cnt = 0
            this.timespans().forEach(timespan => {
                timespan.id = cnt
                cnt += 1
            })

            // Add dummy
            this.addTimespan() // Add new empty timespan

            // Module instances settings observables
            this.maxInstancesDiscovery = ko.observable(params.args["scan_settings"]["max_instances_discovery"]);
            this.maxInstancesBanner = ko.observable(params.args["scan_settings"]["max_instances_banner"]);
            this.maxInstancesNfs = ko.observable(params.args["scan_settings"]["max_instances_nfs"]);
            this.maxInstancesNuclei = ko.observable(params.args["scan_settings"]["max_instances_nuclei"]);
            this.maxInstancesSmb = ko.observable(params.args["scan_settings"]["max_instances_smb"]);
            this.maxInstancesSsh = ko.observable(params.args["scan_settings"]["max_instances_ssh"]);
            this.maxInstancesSsl = ko.observable(params.args["scan_settings"]["max_instances_ssl"]);
            this.maxInstancesWebcrawler = ko.observable(params.args["scan_settings"]["max_instances_webcrawler"]);
            this.maxInstancesWebenum = ko.observable(params.args["scan_settings"]["max_instances_webenum"]);

            // Prepare flag for agent limit notification
            this.agentLimits = ko.computed(function () {
                var limits = false
                if (params.args["scan_agents"] !== null) {
                    params.args["scan_agents"].forEach(function (obj) {
                        if (obj.limits === true) {
                            limits = true
                        }
                    });
                }
                return limits
            }, this);

            // Prepare webcrawler settings observables
            this.followQs = ko.observable(params.args["scan_settings"]["webcrawler_follow_query_strings"]);
            this.storeRoot = ko.observable(params.args["scan_settings"]["webcrawler_always_store_root"]);
            this.probeRobots = ko.observable(params.args["scan_settings"]["webenum_probe_robots"]);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divScopesEdit');
            this.$domForm = this.$domComponent.find("form");

            // Initialize dropdown elements
            this.$domComponent.find('select.dropdown').dropdown();

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize form with validators. keyboardShortcuts is disabled because
            // Semantic UI's Enter handler would submit the form a second time alongside
            // the browser's native submit that Knockout's submit binding already handles.
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                    inputGroup: ['notEmpty'],
                    inputWorkingHoursStart: ['notEmpty'],
                    inputWorkingHoursEnd: ['notEmpty'],
                    inputInstancesDiscovery: ['integer[0..150]'],
                    inputInstancesBanner: ['integer[0..150]'],
                    inputInstancesNfs: ['integer[0..150]'],
                    inputInstancesNuclei: ['integer[0..150]'],
                    inputInstancesSmb: ['integer[0..150]'],
                    inputInstancesSsh: ['integer[0..150]'],
                    inputInstancesSsl: ['integer[0..150]'],
                    inputInstancesWebcrawler: ['integer[0..150]'],
                    inputInstancesWebenum: ['integer[0..150]'],
                    inputNmapArgs: ['minLength[10]'],
                    inputNetworkTimeout: ['integer[1..30]'],
                    inputSmbExcludeFileSize: ['integer[-1..1099511627776]'],
                    inputSmbExcludeLastModifiedBelow: ['notEmpty'],
                    inputSmbDepth: ['integer[-1..150]'],
                    inputSmbThreads: ['integer[1..30]'],
                    inputWebcrawlerDepth: ['integer[-1..150]'],
                    inputWebcrawlerThreads: ['integer[1..30]'],
                    inputWebcrawlerFollowTypes: ['minLength[10]'],
                },
                keyboardShortcuts: false, // Prevent FomanticUI's own submit action handler from submitting again
            });

            // Get calendar input boxes
            this.smbExcludeLastModifiedBelow = $("#calendarSmbExcludeLastModifiedBelow");

            // Initialize sliders
            initSlider("#sliderDiscovery", this.maxInstancesDiscovery, 0, 150, 1);
            initSlider("#sliderNFS", this.maxInstancesNfs, 0, 150, 1);
            initSlider("#sliderNuclei", this.maxInstancesNuclei, 0, 5, 1);
            initSlider("#sliderSMB", this.maxInstancesSmb, 0, 150, 1);
            initSlider("#sliderBanner", this.maxInstancesBanner, 0, 150, 1);
            initSlider("#sliderSSH", this.maxInstancesSsh, 0, 150, 1);
            initSlider("#sliderSSL", this.maxInstancesSsl, 0, 150, 1);
            initSlider("#sliderWebcrawler", this.maxInstancesWebcrawler, 0, 150, 1);
            initSlider("#sliderWebenum", this.maxInstancesWebenum, 0, 150, 1);

            // Initialize date range
            this.smbExcludeLastModifiedBelow.calendar({
                type: 'datetime',
                firstDayOfWeek: 1,
                monthFirst: false,
                ampm: false,
                initialDate: moment(this.args()["scan_settings"]["smb_exclude_last_modified_below"], datetimeFormatGolangNoTz).toDate(),
            });

            // Fade in
            this.$domComponent.transition('fade down');

            // Scroll to form (might be outside visible area if there are long lists)
            $([document.documentElement, document.body]).animate({
                scrollTop: this.$domComponent.offset().top - 160
            }, 200);
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.addTimespan = function () { // Adds an empty new timespan if there is none

            // Prepare working variables
            var addTimespan = true
            var maxTimespanId = 0

            // Iterate timespans and verify them
            this.timespans().forEach(timespan => {

                // Remember largest ID
                if (timespan.id > maxTimespanId) {
                    maxTimespanId = timespan.id
                }

                // Dummy timespan found
                if (timespan.startDay === -1 || timespan.startTime === "" || timespan.endDay === -1 || timespan.endTime === "") {
                    addTimespan = false
                }
            })

            // Add new dummy timespan if necessary
            if (addTimespan === true) {
                this.timespans.push({id: maxTimespanId + 1, startDay: -1, startTime: "", endDay: -1, endTime: ""})
            }
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.updateTimespan = function (id, type, value) {
            this.timespans().forEach(timespan => {
                if (timespan.id === id) {
                    timespan[type] = value
                }
            })

            // Check if new dummy timespan should be added
            this.addTimespan()
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.removeTimespan = function (ctx, event) {

            // Prepare list of remaining timespans
            var newTimespans = []
            this.timespans().forEach(timespan => {

                // Append to new list of timespans if it's not the one to remove
                if (timespan.id !== ctx.id) {
                    newTimespans.push(timespan);
                }
            })

            // Prepare updated list of timespans
            this.timespans(newTimespans)

            // Check if new dummy timespan should be added
            this.addTimespan()
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submitEdit = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Validate form
            if (!this.$domForm.form('is valid')) {
                this.$domForm.form("validate form");
                this.$domForm.each(shake);
                return;
            }

            // Convert comma separated string of ports back to slice of integers if necessary
            var sensitivePorts = [];
            var sensitivePortsInput = this.args()["scan_settings"]["sensitive_ports"];
            if (typeof sensitivePortsInput === "string" && sensitivePortsInput.length > 0) {
                sensitivePortsInput = sensitivePortsInput.trim().split(",");
                for (var i = 0; i < sensitivePortsInput.length; i++) {
                    var val = sensitivePortsInput[i].trim();
                    if (val.length > 0) {
                        var valInt = parseInt(val, 10);
                        if (isNaN(valInt) || valInt < 0 || valInt > 65535) { // Validate port values
                            this.$domForm.form("add prompt", "inputSensitivePorts", "Invalid list of sensitive ports!");
                            this.$domForm.each(shake);
                            return
                        }
                        sensitivePorts.push(valInt);
                    }
                }
            } else {
                sensitivePorts = sensitivePortsInput
            }

            // Read timespans and transform into appropriate data structure
            var timespans = []
            this.timespans().forEach(timespan => {
                if (timespan.startDay !== -1 && timespan.startTime !== "" && timespan.endDay !== -1 && timespan.endTime !== "") {
                    timespans.push(timespan);
                }
            })

            // Sort timespans
            timespans.sort(function (x, y) {

                // Parse strings to integers for comparison
                xStart = parseInt(x.startDay, 10)
                xEnd = parseInt(x.endDay, 10)
                yStart = parseInt(y.startDay, 10)
                yEnd = parseInt(y.endDay, 10)

                // Translate integer values. By default, Sunday is 0 and Monday is 1,
                // but we want Sunday to be 7 at the end of the list.
                if (xStart === 0) {
                    xStart = 7
                } else {
                    xStart -= 1
                }
                if (xEnd === 0) {
                    xEnd = 7
                } else {
                    xEnd -= 1
                }
                if (yStart === 0) {
                    yStart = 7
                } else {
                    yStart -= 1
                }
                if (yEnd === 0) {
                    yEnd = 7
                } else {
                    yEnd -= 1
                }

                // Decide based on first factor start day
                if (xStart < yStart) {
                    return -1
                } else if (xStart > yStart) {
                    return 1
                } else {

                    // Decide based on second factor start time
                    var xStart = moment(x.startTime, "HH:mm")
                    var yStart = moment(y.startTime, "HH:mm")
                    if (xStart.isBefore(yStart)) {
                        return -1
                    } else if (xStart.isAfter(yStart)) {
                        return 1
                    } else {

                        // Decide based on third factor end day
                        if (xEnd < yEnd) {
                            return -1
                        } else if (xEnd > yEnd) {
                            return 1
                        } else {

                            // Decide based on fourth factor end time
                            var xEnd = moment(x.endTime, "HH:mm")
                            var yEnd = moment(y.endTime, "HH:mm")
                            if (xEnd.isBefore(yEnd)) {
                                return -1
                            } else if (xEnd.isAfter(yEnd)) {
                                return 1
                            } else {
                                return 0
                            }
                        }
                    }
                }
            });

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Reload parent table, because data got updated
                ctx.parent.loadData(null, null, function () {

                    // Show toast message for user (but only after parent has reloaded)
                    toast(response.message, "success");

                    // Unlink component (but only after parent has reloaded)
                    ctx.dispose(data, event)
                });
            };

            // Sanitize scan settings data
            var scanSettings = {
                "max_instances_banner": parseInt(this.maxInstancesBanner(), 10),
                "max_instances_discovery": parseInt(this.maxInstancesDiscovery(), 10),
                "max_instances_nfs": parseInt(this.maxInstancesNfs(), 10),
                "max_instances_nuclei": parseInt(this.maxInstancesNuclei(), 10),
                "max_instances_smb": parseInt(this.maxInstancesSmb(), 10),
                "max_instances_ssh": parseInt(this.maxInstancesSsh(), 10),
                "max_instances_ssl": parseInt(this.maxInstancesSsl(), 10),
                "max_instances_webcrawler": parseInt(this.maxInstancesWebcrawler(), 10),
                "max_instances_webenum": parseInt(this.maxInstancesWebenum(), 10),

                "sensitive_ports": sensitivePorts,
                "network_timeout_seconds": parseInt(this.args()["scan_settings"]["network_timeout_seconds"], 10),

                "discovery_timespans": timespans,
                "discovery_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["discovery_scan_timeout_minutes"], 10),
                "discovery_nmap_args": this.args()["scan_settings"]["discovery_nmap_args"],
                "discovery_nmap_args_prescan": this.args()["scan_settings"]["discovery_nmap_args_prescan"],
                "discovery_exclude_domains": this.args()["scan_settings"]["discovery_exclude_domains"],

                "nfs_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["nfs_scan_timeout_minutes"], 10),
                "nfs_depth": parseInt(this.args()["scan_settings"]["nfs_depth"], 10),
                "nfs_threads": parseInt(this.args()["scan_settings"]["nfs_threads"], 10),
                "nfs_exclude_shares": this.args()["scan_settings"]["nfs_exclude_shares"],
                "nfs_exclude_folders": this.args()["scan_settings"]["nfs_exclude_folders"],
                "nfs_exclude_extensions": this.args()["scan_settings"]["nfs_exclude_extensions"],
                "nfs_exclude_file_size_below": parseInt(this.args()["scan_settings"]["nfs_exclude_file_size_below"], 10),
                "nfs_exclude_last_modified_below": this.args()["scan_settings"]["nfs_exclude_last_modified_below"],
                "nfs_accessible_only": parseInt(this.args()["scan_settings"]["nfs_accessible_only"], 10),

                "smb_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["smb_scan_timeout_minutes"], 10),
                "smb_depth": parseInt(this.args()["scan_settings"]["smb_depth"], 10),
                "smb_threads": parseInt(this.args()["scan_settings"]["smb_threads"], 10),
                "smb_forced_shares": this.args()["scan_settings"]["smb_forced_shares"],
                "smb_exclude_shares": this.args()["scan_settings"]["smb_exclude_shares"],
                "smb_exclude_folders": this.args()["scan_settings"]["smb_exclude_folders"],
                "smb_exclude_extensions": this.args()["scan_settings"]["smb_exclude_extensions"],
                "smb_exclude_file_size_below": parseInt(this.args()["scan_settings"]["smb_exclude_file_size_below"], 10),
                "smb_exclude_last_modified_below": moment(this.smbExcludeLastModifiedBelow.calendar("get date"), datetimeFormat).format(datetimeFormatGolang),
                "smb_accessible_only": parseInt(this.args()["scan_settings"]["smb_accessible_only"], 10),

                "nuclei_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["nuclei_scan_timeout_minutes"], 10),
                "nuclei_include_severities": this.args()["scan_settings"]["nuclei_include_severities"],
                "nuclei_exclude_severities": this.args()["scan_settings"]["nuclei_exclude_severities"],
                "nuclei_include_tags": this.args()["scan_settings"]["nuclei_include_tags"],
                "nuclei_exclude_tags": this.args()["scan_settings"]["nuclei_exclude_tags"],
                "nuclei_include_ids": this.args()["scan_settings"]["nuclei_include_ids"],
                "nuclei_exclude_ids": this.args()["scan_settings"]["nuclei_exclude_ids"],
                "nuclei_include_protocols": this.args()["scan_settings"]["nuclei_include_protocols"],
                "nuclei_exclude_protocols": this.args()["scan_settings"]["nuclei_exclude_protocols"],


                "ssh_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["ssh_scan_timeout_minutes"], 10),

                "ssl_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["ssl_scan_timeout_minutes"], 10),

                "http_user_agent": this.args()["scan_settings"]["http_user_agent"],

                "webcrawler_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["webcrawler_scan_timeout_minutes"], 10),
                "webcrawler_depth": parseInt(this.args()["scan_settings"]["webcrawler_depth"], 10),
                "webcrawler_max_threads": parseInt(this.args()["scan_settings"]["webcrawler_max_threads"], 10),
                "webcrawler_follow_query_strings": this.followQs(),
                "webcrawler_always_store_root": this.storeRoot(),
                "webcrawler_follow_types": this.args()["scan_settings"]["webcrawler_follow_types"],

                "webenum_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["webenum_scan_timeout_minutes"], 10),
                "webenum_probe_robots": this.probeRobots(),
            };

            // Prepare request body
            var reqData = {
                "id": this.args()["id"],
                "scan_settings": scanSettings,
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/scope/update/settings",
                {},
                reqData,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Hide form
            this.$domComponent.transition('fade up');

            // Reset form fields
            this.$domForm.form('reset');

            // Dispose open form
            if (this.parent.actionComponent() === "scopes-settings") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
