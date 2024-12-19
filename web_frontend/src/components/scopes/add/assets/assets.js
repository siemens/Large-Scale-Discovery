/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./assets.html", "postbox", "jquery", "semantic-ui-popup", "semantic-ui-dropdown", "semantic-ui-transition"], function (ko, template, postbox, $) {

    /////////////////////////
    // VIEWMODEL CONSTRUCTION
    /////////////////////////
    function ViewModel(params) {

        // Keep reference to PARENT view model context
        this.parent = params.parent;

        // Check whether to enable update mode
        this.type = ko.observable(params.type);
        this.updateMode = ko.observable(params.args !== null);

        // Initialize update-mode observables
        this.scopeId = ko.observable(-1);
        this.groupName = ko.observable("");

        // Initialize create-mode observables
        this.groupsAvailable = ko.observableArray([]);
        this.groupSelected = ko.observable(-1);

        // Initialize scope type independent observables
        this.scopeName = ko.observable("");
        this.scopeCycles = ko.observable(false);
        this.scopeRetention = ko.observable("All");

        // Initialize scope type specific observables
        this.scopeSync = ko.observable(false);
        this.scopeAssetType = ko.observableArray(["Any", "Server", "Network", "Client"]);
        this.scopeAssetTypeSelected = ko.observable();
        this.scopeAssetCompanies = ko.observableArray([]);
        this.scopeAssetCompaniesPossible = ko.observableArray([]); // For update mode
        this.scopeAssetDepartments = ko.observableArray([]);
        this.scopeAssetDepartmentsPossible = ko.observableArray([]); // For update mode
        this.scopeAssetCountries = ko.observableArray([]);
        this.scopeAssetCountriesPossible = ko.observableArray([]); // For update mode
        this.scopeAssetLocations = ko.observableArray([]);
        this.scopeAssetLocationsPossible = ko.observableArray([]); // For update mode
        this.scopeAssetContacts = ko.observableArray([]);
        this.scopeAssetContactsPossible = ko.observableArray([]); // For update mode
        this.scopeAssetCritical = ko.observableArray(["Any", "Yes", "No"]);
        this.scopeAssetCriticalSelected = ko.observable();

        // Initialize update mode, if scope details are passed
        if (this.updateMode()) {

            // Set update-mode observables
            this.scopeId(params.args.id)
            this.groupName(params.args.group_name)

            // Set scope type independent values
            this.scopeName(params.args.name);
            this.scopeCycles(params.args.cycles);
            if (params.args.cycles_retention > 0) {
                this.scopeRetention(params.args.cycles_retention);
            }

            // Set scope type specific values
            this.scopeSync(params.args.attributes.sync);
            this.scopeAssetTypeSelected(params.args.attributes.asset_type);
            if (params.args.attributes.asset_companies != null) {
                this.scopeAssetCompanies(params.args.attributes.asset_companies.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
                this.scopeAssetCompaniesPossible(params.args.attributes.asset_companies.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
            }
            if (params.args.attributes.asset_departments != null) {
                this.scopeAssetDepartments(params.args.attributes.asset_departments.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
                this.scopeAssetDepartmentsPossible(params.args.attributes.asset_departments.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
            }
            if (params.args.attributes.asset_countries != null) {
                this.scopeAssetCountries(params.args.attributes.asset_countries.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
                this.scopeAssetCountriesPossible(params.args.attributes.asset_countries.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
            }
            if (params.args.attributes.asset_locations != null) {
                this.scopeAssetLocations(params.args.attributes.asset_locations.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
                this.scopeAssetLocationsPossible(params.args.attributes.asset_locations.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
            }
            if (params.args.attributes.asset_contacts != null) {
                this.scopeAssetContacts(params.args.attributes.asset_contacts.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
                this.scopeAssetContactsPossible(params.args.attributes.asset_contacts.slice(0)); // Copy array, or dispose will wipe parent data (array reference)
            }
            this.scopeAssetCriticalSelected(params.args.attributes.asset_critical);
        }

        // Get reference to the view model's actual HTML within the DOM
        this.$domComponent = $('#divScopesAddAsset');
        this.$domForm = this.$domComponent.find("form");

        // Initialize multi select dropdown elements with upper case observables
        initDropdown("#selectCompanies", this.scopeAssetCompanies, "*", true)
        initDropdown("#selectDepartments", this.scopeAssetDepartments, "*", true)
        initDropdown("#selectCountries", this.scopeAssetCountries, "*", true)
        initDropdown("#selectLocations", this.scopeAssetLocations, "*", true)
        initDropdown("#selectContacts", this.scopeAssetContacts, "*", true)

        // Initialize dropdown elements
        this.$domComponent.find('select.dropdown').dropdown();

        // Initialize tooltips
        this.$domComponent.find('[data-html]').popup();

        // Initialize form, depending on whether update mode is desired or not
        if (this.updateMode()) {

            // Initialize form validators
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                },
            });
        } else {

            // Initialize form validators
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                    selectGroup: ['empty'],
                },
            });

            // Load and set initial data
            this.loadData();
        }

        // Fade in
        this.$domComponent.transition('fade down');

        // Scroll to form (might be outside of visible area if there are long lists)
        $([document.documentElement, document.body]).animate({
            scrollTop: this.$domComponent.offset().top - 160
        }, 200);
    }

    // VIEWMODEL ACTION
    ViewModel.prototype.loadData = function (data, event) {

        // Keep reference THIS view model context
        var ctx = this;

        // Handle request success
        const callbackSuccess = function (response, textStatus, jqXHR) {

            // Init array of groups
            ctx.groupsAvailable(response.body["groups"]);

            // Set scope group, if there is only one
            if (response.body["groups"].length === 1) {
                ctx.groupSelected(response.body["groups"][0].id);
            }
        };

        // Send request to get groups
        apiCall(
            "GET",
            "/api/v1/groups",
            {},
            null,
            callbackSuccess,
            null
        );
    };

    // VIEWMODEL ACTION
    ViewModel.prototype.scopeRetentionAdd = function (data, event) {
        var current = this.scopeRetention()
        if (current === "All") {
            this.scopeRetention(1)
        } else {
            this.scopeRetention(current + 1)
        }
    }

    // VIEWMODEL ACTION
    ViewModel.prototype.scopeRetentionSub = function (data, event) {
        var current = this.scopeRetention()
        if (current > 1) {
            this.scopeRetention(current - 1)
        } else {
            this.scopeRetention("All")
        }
    }

    // VIEWMODEL ACTION
    ViewModel.prototype.submitAssets = function (data, event) {

        // Keep reference THIS view model context
        var ctx = this;

        // Validate form
        if (!this.$domForm.form('is valid')) {
            this.$domForm.form("validate form");
            this.$domForm.each(shake);
            return;
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

        // Prepare retention value as expected by the backend
        var retention = -1
        if (this.scopeRetention() > 0) {
            retention = this.scopeRetention()
        }

        // Prepare basic request data
        var reqData = {
            type: this.type(),
            name: this.scopeName(),
            cycles: this.scopeCycles(),
            cycles_retention: retention,
            sync: this.scopeSync(),
            asset_type: this.scopeAssetTypeSelected(),
            asset_companies: this.scopeAssetCompanies(),
            asset_departments: this.scopeAssetDepartments(),
            asset_countries: this.scopeAssetCountries(),
            asset_locations: this.scopeAssetLocations(),
            asset_contacts: this.scopeAssetContacts(),
            asset_critical: this.scopeAssetCriticalSelected()
        }

        // Send create / update request
        if (this.updateMode()) {

            // Set scope ID to indicate scope update
            reqData["scope_id"] = this.scopeId() // Required to update associated scope
        } else {

            // Set group ID to indicate new scope
            reqData["group_id"] = this.groupSelected() // Required to create new scope within
        }

        // Send request
        apiCall(
            "POST",
            "/api/v1/scope/update/assets",
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
        if (this.parent.actionComponent() === "scopes-add-assets") {
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);
        }
    };

    return {viewModel: ViewModel, template: template};
});
