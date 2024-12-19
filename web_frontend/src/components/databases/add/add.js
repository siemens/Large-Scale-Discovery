/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./add.html", "postbox", "jquery", "semantic-ui-popup", "semantic-ui-dropdown", "semantic-ui-transition"], function (ko, template, postbox, $) {

    /////////////////////////
    // VIEWMODEL CONSTRUCTION
    /////////////////////////
    function ViewModel(params) {

        // Keep reference to PARENT view model context
        this.parent = params.parent;

        // Initialize observables
        this.databaseId = ko.observable(-1);
        this.databaseName = ko.observable("");
        this.databaseDialect = ko.observable("postgres");
        this.databaseHost = ko.observable("");
        this.databaseHostPublic = ko.observable("");
        this.databasePort = ko.observable("5432");
        this.databaseAdmin = ko.observable("postgres");
        this.databasePassword = ko.observable("");
        this.databaseArgs = ko.observable("sslmode=verify-ca");

        // Get reference to the view model's actual HTML within the DOM
        this.$domComponent = $('#divDatabasesAdd');
        this.$domForm = this.$domComponent.find("form");

        // Initialize tooltips
        this.$domComponent.find('[data-html]').popup();

        // Define custom range validation rule
        $.fn.form.settings.rules.password = function (value) {
            value = value.trim()

            // Check length
            if (value.length <= 10) {
                return false
            }

            // Test strength
            var strength = 0;
            if (value.match(/[a-z]+/)) {
                strength += 1;
            }
            if (value.match(/[A-Z]+/)) {
                strength += 1;
            }
            if (value.match(/[0-9]+/)) {
                strength += 1;
            }
            if (value.match(/[$@#&!]+/)) {
                strength += 1;
            }

            // Check strength
            if (strength < 3) {
                return false
            }

            // Return success
            return true
        };
        $.fn.form.settings.rules.isHostname = function (value) {
            return isIpOrFqdn(value.trim())
        };

        // Initialize form validators
        this.$domForm.form({
            fields: {
                inputName: ['minLength[3]'],
                inputDialect: ['minLength[3]'],
                inputHost: ['isHostname'],
                inputHostPublic: ['isHostname'],
                inputPort: ['integer[0..65535]'],
                inputAdmin: ['minLength[1]'],
                inputArgs: ['minLength[0]'],
                inputPassword: ['password'],
            },
        });

        // Fade in
        this.$domComponent.transition('fade down');

        // Scroll to form (might be outside visible area if there are long lists)
        $([document.documentElement, document.body]).animate({
            scrollTop: this.$domComponent.offset().top - 160
        }, 200);
    }

    // VIEWMODEL ACTION
    ViewModel.prototype.submitDatabase = function (data, event) {

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

        // Prepare basic request data
        var reqData = {
            name: this.databaseName(),
            dialect: this.databaseDialect(),
            host: this.databaseHost(),
            host_public: this.databaseHostPublic(),
            port: parseInt(this.databasePort(), 10),
            admin: this.databaseAdmin(),
            password: this.databasePassword(),
            args: this.databaseArgs()
        }

        // Send request
        apiCall(
            "POST",
            "/api/v1/admin/database/update",
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
        if (this.parent.actionComponent() === "databases-add") {
            this.parent.actionArgs(null);
            this.parent.actionComponent(null);
        }
    };

    return {viewModel: ViewModel, template: template};
});
