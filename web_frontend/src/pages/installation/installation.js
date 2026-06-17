/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./installation.html", "postbox", "semantic-ui-accordion"],
    function (ko, template, postbox) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Initialize observables
            this.supportData = ko.observableArray([
                {
                    "title": "Discovery Module",
                    "subEntry": false,
                    "linuxSupport": 1,   // Scale: -1 = N/A, 0 = Not supported, 1 = Supported, 2 = Partially supported
                    "windowsSupport": 1, // Scale: -1 = N/A, 0 = Not supported, 1 = Supported, 2 = Partially supported
                    "domainSupport": 1,  // Scale: -1 = N/A, 0 = Not supported, 1 = Supported, 2 = Partially supported
                    "comment": "The discovery module executes devices discovery, port scanning and service detection and is required to run. Furthermore, it expands scan results with information from various repositories.",
                },
                {
                    "title": "Device Discovery",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "OT Discovery",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 2,
                    "windowsSupportComment": "Only network-based protocols (NDP, mDNS/SSDP) are supported. L2 protocols (PROFINET DCP, EtherCAT, LLDP) require raw Ethernet sockets which are not available on Windows.",
                    "domainSupport": 2,
                    "domainSupportComment": "Only network-based protocols (NDP, mDNS/SSDP) are supported. L2 protocols (PROFINET DCP, EtherCAT, LLDP) require raw Ethernet sockets which are not available on Windows.",
                    "comment": "",
                },
                {
                    "title": "Port Enumeration",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Service Detection",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Hostname Discovery",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Extraction Of Remote Interfaces",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "domainSupportComment": "Via implicit authentication.",
                    "comment": "",
                },
                {
                    "title": "Extraction Of Admin/Rdp Users",
                    "subEntry": true,
                    "linuxSupport": 0,
                    "windowsSupport": 0,
                    "domainSupport": 1,
                    "domainSupportComment": "Via implicit authentication.",
                    "comment": "",
                },
                {
                    "title": "Active Directory Lookup",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "linuxSupportComment": "You can configure explicit credentials. Cross-domain authentication is possible via GSSAPI. By default the discovered target's domain is queried, but you can define a static LDAP server.",
                    "windowsSupport": 1,
                    "windowsSupportComment": "You can configure explicit credentials. Cross-domain authentication is possible via GSSAPI. By default the discovered target's domain is queried, but you can define a static LDAP server.",
                    "domainSupport": 1,
                    "domainSupportComment": "Implicit authentication without credentials can be used. You can configure explicit credentials. Cross-domain authentication is possible via GSSAPI. By default the discovered target's domain is queried, but you can define a static LDAP server.",
                    "comment": "",
                },
                {
                    "title": "Asset Inventory Lookup",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Banner Grabbing",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The banner grabbing module connects to discovered ports and extracts data returned by the remote service.",
                },
                {
                    "title": "SMB Crawling",
                    "subEntry": false,
                    "linuxSupport": 0,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The SMB crawling module iterates shares and folders to discover accessible and/or writable files.",
                },
                {
                    "title": "Mime Type Detection",
                    "subEntry": true,
                    "linuxSupport": -1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Microsoft Information Protection",
                    "subEntry": true,
                    "linuxSupport": -1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "NFS Crawling",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The NFS crawling module mounts NFS shares and iterates folders to discover accessible and/or writable files.",
                },
                {
                    "title": "NFSv3",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "NFSv4",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 0,
                    "domainSupport": 0,
                    "comment": "",
                },
                {
                    "title": "Unix ACL Flags",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Mime Type Detection",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Microsoft Information Protection",
                    "subEntry": true,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "",
                },
                {
                    "title": "Nuclei Scanning",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The Nuclei module leverages the large list of community-curated templates at ProjectDiscovery to find security vulnerabilities.",
                },
                {
                    "title": "Web Crawling",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The web crawling module crawls web services to extract links, response headers and HTML contents.",
                },
                {
                    "title": "Web Enumeration",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The web enumeration module guesses common URLs in order to discover common sensitive hidden resources.",
                },
                {
                    "title": "SSL Enumeration",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The SSL module enumerates existing SSL protocols and ciphers in order to detect deployed configurations and vulnerabilities.",
                },
                {
                    "title": "SSH Enumeration",
                    "subEntry": false,
                    "linuxSupport": 1,
                    "windowsSupport": 1,
                    "domainSupport": 1,
                    "comment": "The SSH module enumerates existing SSH protocols and ciphers in order to detect deployed configurations and vulnerabilities.",
                },
            ]);

            // Check authentication and redirect to login if necessary
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }

            // Check privileges and redirect to home if necessary
            if (userAdmin() === false && userOwner() === false) {
                postbox.publish("redirect", home());
                return;
            }

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divTutorial');

            // Initialize mouseover cursor
            this.$domComponent.find('.step').css('cursor', 'pointer');

            // Initialize accordion
            this.$domComponent.find('.ui.accordion').accordion();

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize message close button
            $('.message .close').on('click', function () {
                $(this)
                    .closest('.message')
                    .transition('fade')
                ;
            });

            // Fade in table
            this.$domComponent.children("div:hidden").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                duration: 200
            });

            // Reference content elements
            this.support = $("#divSupportMatrix");
            this.nmap = $("#divInstallNmap");
            this.sslyze = $("#divInstallSslyze");
            this.agent = $("#divInstallAgent");
            this.launch = $("#divLaunchAgent");
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.initSupportEntry = function (element, data) {

            // Initialize tooltips
            $(element).find('[data-html]').popup();
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.switchStep = function (data, event) {

            // Get referenced clicked element
            var currentElement = $(event.currentTarget);

            // Get referenced content element
            var targetElement = $("#" + event.currentTarget.attributes["target-id"].value);

            // Toggle active step element
            currentElement.parent().find(".step").removeClass("active");
            currentElement.addClass("active");

            // Hide all content elements
            this.support.transition("hide");
            this.nmap.transition("hide");
            this.sslyze.transition("hide");
            this.agent.transition("hide");
            this.launch.transition("hide");

            // Fade in requested content element
            targetElement.transition("fade left");
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.toClipboard = function (data, event) {

            // Find content and copy to clipboard
            var $temp = $("<textarea>");
            $("body").append($temp);
            $temp.val(event.currentTarget.nextElementSibling.innerText).select();
            document.execCommand("copy");
            $temp.remove();

            // Indicate success
            toast("Copied to clipboard.", "success");
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
