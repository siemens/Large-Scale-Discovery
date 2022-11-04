/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "postbox", "jquery", "moment", "semantic-ui"],
    function (ko, postbox, $, moment) {

        /*
         * Global variables and observables are available in every view model and template for data-binding by default.
         * It is only necessary to import the globals package, if a view model wants to call one of global's methods.
         */

        // Initialize global application attributes
        this.currentRoute = postbox.currentRoute;
        this.activeRequests = ko.observableArray().extend({rateLimit: 300}); // suppress change notifications for 300 ms
        this.developmentLogin = ko.observable(false);
        this.credentialsRegistration = ko.observable(false);

        // Initialize global session attributes
        this.authToken = ko.observable("");
        this.authExpiry = ko.observable(null);
        this.authenticated = ko.observable(false);

        // Initialize global user attributes
        this.userId = ko.observable(-1);
        this.userEmail = ko.observable("");
        this.userName = ko.observable("");
        this.userSurname = ko.observable("");
        this.userGender = ko.observable("");
        this.userAdmin = ko.observable(false);
        this.userOwner = ko.observable(false);
        this.userAccess = ko.observable(false);
        this.userCreated = ko.observable(null);

        // Initialize presentation mode observables used for hiding sensitive values in the frontend
        this.presentationMode = ko.observable(false)
        this.presentationClass = ko.computed(function () {
            return this.presentationMode() ? 'sensitive' : ''
        }, this);

        // Initialize global computed observables
        this.userFullName = ko.computed(function () {
            return this.userName() + " " + this.userSurname();
        }, this);

        // Constructor for global package. This package is working on global variables, so the constructor doesn't need
        // to initialize a lot.
        function Globals() {

            // Load some relevant application settings from backend
            apiCall(
                "GET",
                "/api/v1/backend/settings",
                {},
                null,
                function (response, textStatus, jqXHR) {

                    // Store returned attributes in global variables
                    self.developmentLogin(response["body"]["development_login"]);
                    self.credentialsRegistration(response["body"]["credentials_registration"]);
                },
                null
            );

            // Launch background job monitoring authentication expiry status
            setInterval(function (self) {
                var nowPretended = moment().add(2, 'seconds'); // Because our check interval might be a bit late
                if (authExpiry() && authExpiry() < nowPretended) {
                    console.log("Authentication expired.");
                    postbox.publish("redirect", "login");
                    self.discard();
                    return false;
                }
            }, 1000, this);
        }

        // Helper function to reset all global data and the persisted storage.
        Globals.prototype.discard = function () {
            sessionStorage.clear();
            userId(-1);
            userEmail("");
            userName("");
            userSurname("");
            userGender("");
            userAdmin(false);
            userOwner(false);
            userCreated(moment(null));
            authToken(null);
            authExpiry(null);
            authenticated(false);
        };

        // Helper function to set authentication data after successful authentication
        Globals.prototype.authenticationSet = function (access_token, expire) {
            authToken(access_token);
            authExpiry(moment(expire));
            authenticated(true);
        };

        // Helper function to set profile data after retrieval
        Globals.prototype.profileSet = function (id, mail, name, surname, gender, admin, owner, access, created) {
            userId(id);
            userEmail(mail);
            userName(name);
            userSurname(surname);
            userGender(gender);
            userAdmin(admin);
            userOwner(owner);
            userAccess(access);
            userCreated(moment(created));
        };

        // Helper function to check whether available authentication data has exceeded
        Globals.prototype.authenticationValid = function () {
            if (!authToken()) {
                console.log("Not authenticated.");
                this.discard();
                return false;
            }
            var nowPretended = moment().add(2, 'seconds'); // Because our check interval might be a bit late
            if (authExpiry() && authExpiry() < nowPretended) {
                console.log("Authentication expired.");
                this.discard();
                return false;
            }
            return true;
        };

        // Initialize global application values and observables
        return new Globals();

    }
);
