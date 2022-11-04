/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./settings.html", "postbox", "jquery", "tabulator-tables", "semantic-ui-dropdown", "semantic-ui-calendar"],
    function (ko, template, postbox, $, Tabulator, mod1, mod2, str) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Store parameters passed by parent
            this.args = ko.observable(params.args); // Args is already form data, so it should be an observable

            // Prepare array of disabled week days
            var arr = new Array(7);
            for (var i = 0; i < 7; i++) {
                arr[i] = [i + 1, false]
            }

            // Enable week days according to current configuration
            this.skipDays = ko.observable(new Map(arr));
            if (params.args["scan_settings"]["discovery_skip_days"]) {
                for (var j = 0; j < params.args["scan_settings"]["discovery_skip_days"].length; j++) {
                    this.skipDays()[params.args["scan_settings"]["discovery_skip_days"][j]] = true
                }
            }

            // Module instances observables
            this.maxInstancesDiscovery = ko.observable(params.args["scan_settings"]["max_instances_discovery"]);
            this.maxInstancesNfs = ko.observable(params.args["scan_settings"]["max_instances_nfs"]);
            this.maxInstancesSmb = ko.observable(params.args["scan_settings"]["max_instances_smb"]);
            this.maxInstancesBanner = ko.observable(params.args["scan_settings"]["max_instances_banner"]);
            this.maxInstancesSsh = ko.observable(params.args["scan_settings"]["max_instances_ssh"]);
            this.maxInstancesSsl = ko.observable(params.args["scan_settings"]["max_instances_ssl"]);
            this.maxInstancesWebcrawler = ko.observable(params.args["scan_settings"]["max_instances_webcrawler"]);
            this.maxInstancesWebenum = ko.observable(params.args["scan_settings"]["max_instances_webenum"]);

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

            // Initialize form validators
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                    inputGroup: ['empty'],
                    inputWorkingHoursStart: ['empty'],
                    inputWorkingHoursEnd: ['empty'],
                    inputInstancesDiscovery: ['integer[0..150]'],
                    inputInstancesBanner: ['integer[0..150]'],
                    inputInstancesNfs: ['integer[0..150]'],
                    inputInstancesSmb: ['integer[0..150]'],
                    inputInstancesSsh: ['integer[0..150]'],
                    inputInstancesSsl: ['integer[0..150]'],
                    inputInstancesWebcrawler: ['integer[0..150]'],
                    inputInstancesWebenum: ['integer[0..150]'],
                    inputNmapArgs: ['minLength[10]'],
                    inputNetworkTimeout: ['integer[1..30]'],
                    inputSmbExcludeFileSize: ['integer[-1..1099511627776]'],
                    inputSmbExcludeModified: ['empty'],
                    inputSmbDepth: ['integer[-1..150]'],
                    inputSmbThreads: ['integer[1..30]'],
                    inputWebcrawlerDepth: ['integer[-1..150]'],
                    inputWebcrawlerThreads: ['integer[1..30]'],
                    inputWebcrawlerFollowTypes: ['minLength[10]'],
                },
            });

            // Get calendar input boxes
            this.startInput = $("#calendarStart");
            this.endInput = $("#calendarEnd");
            this.oldInput = $("#calendarExclude");

            // Initialize sliders
            initSlider("#sliderDiscovery", this.maxInstancesDiscovery, 0, 150, 1);
            initSlider("#sliderNFS", this.maxInstancesNfs, 0, 150, 1);
            initSlider("#sliderSMB", this.maxInstancesSmb, 0, 150, 1);
            initSlider("#sliderBanner", this.maxInstancesBanner, 0, 150, 1);
            initSlider("#sliderSSH", this.maxInstancesSsh, 0, 150, 1);
            initSlider("#sliderSSL", this.maxInstancesSsl, 0, 150, 1);
            initSlider("#sliderWebcrawler", this.maxInstancesWebcrawler, 0, 150, 1);
            initSlider("#sliderWebenum", this.maxInstancesWebenum, 0, 150, 1);

            // Initialize date range
            this.startInput.calendar({
                type: 'time',
                endCalendar: this.endInput,
                firstDayOfWeek: 1,
                ampm: false,
                initialDate: moment(this.args()["scan_settings"]["discovery_time_earliest"], "HH:mm").toDate(),
            });
            this.endInput.calendar({
                type: 'time',
                startCalendar: this.startInput,
                firstDayOfWeek: 1,
                ampm: false,
                initialDate: moment(this.args()["scan_settings"]["discovery_time_latest"], "HH:mm").toDate(),
            });
            this.oldInput.calendar({
                type: 'datetime',
                firstDayOfWeek: 1,
                monthFirst: false,
                ampm: false,
                initialDate: moment(this.args()["scan_settings"]["smb_exclude_last_modified_below"], datetimeFormatGolangNoTz).toDate(),
            });

            // Fade in
            this.$domComponent.transition('fade down');

            // Scroll to form (might be outside of visible area if there are long lists)
            $([document.documentElement, document.body]).animate({
                scrollTop: this.$domComponent.offset().top - 160
            }, 200);
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.updateSkipDays = function (data, event) {
            var skipDays = this.skipDays();
            skipDays[data] = !skipDays[data];
            this.skipDays(skipDays);
            return true;
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

            // Prepare list of weekdays to skip (0=Sunday,...,6=Saturday)
            var skipDays = [];
            var skipMap = this.skipDays();
            for (i = 0; i <= 6; i++) {
                if (skipMap[i] === true) {
                    skipDays.push(i);
                }
            }

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
                "max_instances_smb": parseInt(this.maxInstancesSmb(), 10),
                "max_instances_ssh": parseInt(this.maxInstancesSsh(), 10),
                "max_instances_ssl": parseInt(this.maxInstancesSsl(), 10),
                "max_instances_webcrawler": parseInt(this.maxInstancesWebcrawler(), 10),
                "max_instances_webenum": parseInt(this.maxInstancesWebenum(), 10),

                "sensitive_ports": sensitivePorts,

                "network_timeout_seconds": parseInt(this.args()["scan_settings"]["network_timeout_seconds"], 10),

                "http_user_agent": this.args()["scan_settings"]["http_user_agent"],

                "discovery_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["discovery_scan_timeout_minutes"], 10),
                "discovery_skip_days": skipDays,
                "discovery_time_earliest": moment(this.startInput.calendar("get date"), datetimeFormat).format("HH:mm"),
                "discovery_time_latest": moment(this.endInput.calendar("get date"), datetimeFormat).format("HH:mm"),
                "discovery_nmap_args": this.args()["scan_settings"]["discovery_nmap_args"],

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
                "smb_exclude_shares": this.args()["scan_settings"]["smb_exclude_shares"],
                "smb_exclude_folders": this.args()["scan_settings"]["smb_exclude_folders"],
                "smb_exclude_extensions": this.args()["scan_settings"]["smb_exclude_extensions"],
                "smb_exclude_file_size_below": parseInt(this.args()["scan_settings"]["smb_exclude_file_size_below"], 10),
                "smb_exclude_last_modified_below": moment(this.oldInput.calendar("get date"), datetimeFormat).format(datetimeFormatGolang),
                "smb_accessible_only": parseInt(this.args()["scan_settings"]["smb_accessible_only"], 10),

                "ssh_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["ssh_scan_timeout_minutes"], 10),

                "ssl_scan_timeout_minutes": parseInt(this.args()["scan_settings"]["ssl_scan_timeout_minutes"], 10),

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
