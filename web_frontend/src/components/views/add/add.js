/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./add.html", "postbox", "jquery", 'semantic-ui-dropdown'],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Initialize observables
            this.scopesAvailable = ko.observableArray([]);
            this.scopeSelected = ko.observable(-1);
            this.viewName = ko.observable("");

            this.inputTargetsPossible = ko.observableArray([]);
            this.inputTargets = ko.observableArray([]);
            this.inputCountries = ko.observableArray([]);
            this.inputLocations = ko.observableArray([]);
            this.inputRoutingDomains = ko.observableArray([]);
            this.inputZones = ko.observableArray([]);
            this.inputPurposes = ko.observableArray([]);

            this.inputCompanies = ko.observableArray([]);
            this.inputDepartments = ko.observableArray([]);
            this.inputManagers = ko.observableArray([]);
            this.inputContacts = ko.observableArray([]);
            this.inputComments = ko.observableArray([]);

            // Initialize other params used by this component
            this.scopesDict = {};
            this.scopeSelectedSubscription = null;

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divViewsAdd');
            this.$domForm = this.$domComponent.find("form");

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Initialize multi select dropdown elements
            initDropdown("#selectTargets", this.inputTargets, "-", true);
            initDropdown("#selectCountries", this.inputCountries, "-", true);
            initDropdown("#selectLocations", this.inputLocations, "-", true);
            initDropdown("#selectRoutingDomains", this.inputRoutingDomains, "-", true);
            initDropdown("#selectZones", this.inputZones, "-", true);
            initDropdown("#selectPurposes", this.inputPurposes, "-", true);
            initDropdown("#selectCompanies", this.inputCompanies, "-", true);
            initDropdown("#selectDepartments", this.inputDepartments, "-", true);
            initDropdown("#selectManagers", this.inputManagers, "-", true);
            initDropdown("#selectContacts", this.inputContacts, "-", true);
            initDropdown("#selectComments", this.inputComments, "-", true);

            // Initialize dropdown elements
            this.$domComponent.find('select.dropdown').dropdown();

            // Initialize form validators
            this.$domForm.form({
                fields: {
                    inputName: ['minLength[3]'],
                    selectScope: ['empty'],
                },
            });

            // Load and set initial data
            this.loadData();

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

                // Init array of scopes
                ctx.scopesAvailable(response.body["scopes"]);

                // Set scan scope, if there is only one
                if (response.body["scopes"].length === 1) {
                    ctx.scopeSelected(response.body["scopes"][0].id);
                }

                // Prepare lookup list of scopes
                for (var i = 0; i < response.body["scopes"].length; i++) {
                    ctx.scopesDict[response.body["scopes"][i].id] = response.body["scopes"][i];
                }

                // Observe scope select box to update scan targets on changes
                ctx.scopeSelectedSubscription = ctx.scopeSelected.subscribe(function (newValue) {
                    if (ctx.scopeSelected() !== -1) {

                        // Clear selection
                        ctx.inputTargetsPossible([]);
                        ctx.inputTargets([]);
                        ctx.$domComponent.find('#selectTargets').dropdown('clear')

                        // Set new possible values
                        var targets = ctx.scopesDict[newValue].attributes["targets"];
                        if (targets) {
                            targets.forEach(function (arrayItem) {
                                ctx.inputTargetsPossible.push(arrayItem.input);
                            });
                        }
                    }
                    ctx.scopeSelected(newValue);
                });
            };

            // Send request to get scopes
            apiCall(
                "GET",
                "/api/v1/scopes",
                {},
                null,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.submitView = function (data, event) {

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

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {
                if (response.responseJSON.message.indexOf('Duplicate view name') !== -1) {
                    var inputName = $("#inputName");
                    inputName.parent().addClass("error");
                    inputName.each(shake);
                }
            };

            // Prepare request body
            var reqData = {
                scope_id: this.scopeSelected(),
                view_name: this.viewName(),
                filter_input_targets: this.inputTargets(),
                filter_input_countries: this.inputCountries(),
                filter_input_locations: this.inputLocations(),
                filter_input_routing_domains: this.inputRoutingDomains(),
                filter_input_zones: this.inputZones(),
                filter_input_purposes: this.inputPurposes(),
                filter_input_companies: this.inputCompanies(),
                filter_input_departments: this.inputDepartments(),
                filter_input_managers: this.inputManagers(),
                filter_input_contacts: this.inputContacts(),
                filter_input_comments: this.inputComments(),
            };

            // Send request
            apiCall(
                "POST",
                "/api/v1/view/create",
                {},
                reqData,
                callbackSuccess,
                callbackError
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Stop subscription of scope selection input, otherwise this will be triggered again on dispose
            this.scopeSelectedSubscription.dispose();

            // Hide form
            this.$domComponent.transition('fade up');

            // Reset form fields
            this.$domForm.form('reset');

            // Dispose open form
            if (this.parent.actionComponent() === "views-add") {
                this.parent.actionArgs(null);
                this.parent.actionComponent(null);
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
