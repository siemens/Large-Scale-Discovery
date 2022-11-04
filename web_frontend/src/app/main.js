/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "globals", './router', "postbox", "jquery", "moment", "avatars", "avatars-avataaars", "utils-app"],
    function (ko, globals, router, postbox, $, Moment, avatar, avatarTiles) {

        // Bug fix because Tabulator.js is looking for moments globally
        window.moment = Moment;

        // Register avatar library globally for initAvatar() to pick up
        window.avatar = avatar;
        window.avatarTiles = avatarTiles;

        // Register some custom global functions that might come in handy sometimes
        String.prototype.toTitleCase = function () {
            return this.replace(/(?:^|\s)\S/g, function (a) {
                return a.toUpperCase();
            });
        };

        // Register pages. Pages represent the logical hierarchy and structure of the web application and embed components.
        ko.components.register("register", {require: "pages/register/register"});
        ko.components.register("login", {require: "pages/login/login"});
        ko.components.register("home", {require: "pages/home/home"});
        ko.components.register("configuration-scopes", {require: "pages/configuration/scopes/scopes"});
        ko.components.register("configuration-views", {require: "pages/configuration/views/views"});
        ko.components.register("admin-users", {require: "pages/admin/users/users"});
        ko.components.register("admin-groups", {require: "pages/admin/groups/groups"});
        ko.components.register("installation", {require: "pages/installation/installation"});
        ko.components.register("profile", {require: "pages/profile/profile"});

        // Register regular components. While technically equal to pages, components are functional units can be
        // included everywhere. They usually focus on functionality and (e.g.) don't take care of authentication or
        // application structure.
        ko.components.register("activity", {require: "components/activity/activity"});
        ko.components.register("agents", {require: "components/agents/agents"});
        ko.components.register("charts-access", {require: "components/charts/access/access"});
        ko.components.register("feedback", {require: "components/feedback/feedback"});
        ko.components.register("footer", {require: "components/footer/footer"});
        ko.components.register("groups-add", {require: "components/groups/add/add"});
        ko.components.register("groups-owners", {require: "components/groups/owners/owners"});
        ko.components.register("nav-top", {require: "components/nav-top/nav-top"});
        ko.components.register("nav-side", {require: "components/nav-side/nav-side"});
        ko.components.register("scopes-list", {require: "components/scopes/list/list"});
        ko.components.register("scopes-add-custom", {require: "components/scopes/add/custom/custom"});
        ko.components.register("scopes-add-assets", {require: "components/scopes/add/assets/assets"});
        ko.components.register("scopes-add-networks", {require: "components/scopes/add/networks/networks"});
        ko.components.register("scopes-settings", {require: "components/scopes/settings/settings"});
        ko.components.register("views-list", {require: "components/views/list/list"});
        ko.components.register("views-add", {require: "components/views/add/add"});
        ko.components.register("views-edit", {require: "components/views/edit/edit"});
        ko.components.register("views-grant", {require: "components/views/grant/grant"});
        ko.components.register("views-granted", {require: "components/views/granted/granted"});
        ko.components.register("views-token-list", {require: "components/views/token/list/list"});
        ko.components.register("views-token-add", {require: "components/views/token/add/add"});

        // Load cached authentication data from local storage. It might not be valid anymore, though!

        // Load presentation mode setting from cache
        var currentMode = localStorage.getItem("presentation")
        if (currentMode === "true") {
            presentationMode(true);
        }

        // Load profile data from cache
        var currentUser = JSON.parse(sessionStorage.getItem("user"));
        if (currentUser != null) {
            globals.profileSet(
                currentUser["id"],
                currentUser["email"],
                currentUser["name"],
                currentUser["surname"],
                currentUser["gender"],
                currentUser["admin"],
                currentUser["owner"],
                currentUser["access"],
                currentUser["created"]
            );
        }

        // Load authentication data from cache
        var currentToken = JSON.parse(sessionStorage.getItem("token"));
        if (currentToken != null) {
            globals.authenticationSet(currentToken["access_token"], currentToken["expire"]);
        }

        // Check current authentication status.
        var authValid = globals.authenticationValid();

        // Double check with backend, whether the user data might have changed
        if (authValid === true) {

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Save updated user data to local storage. Values will be read from there on page reload.
                sessionStorage.setItem("user", JSON.stringify(response.body));

                // Update profile data
                globals.profileSet(
                    response["body"]["id"],
                    response["body"]["email"],
                    response["body"]["name"],
                    response["body"]["surname"],
                    response["body"]["gender"],
                    response["body"]["admin"],
                    response["body"]["owner"],
                    response["body"]["access"],
                    response["body"]["created"]
                );
            };

            // Update application globals with server-side values
            apiCall("GET", "/api/v1/user/details", {}, null, callbackSuccess, null);
        }

        // Redirect to login if authentication is not valid
        if (authValid === false) {

            // Check if there is an error defined as a URL parameter to be shown (might be there after a redirect)
            var initialError = getParameterByName("error")
            if (initialError != null && initialError.length > 0) {

                // Flash message
                if (initialError === "Unauthorized") {
                    toast("Unauthorized.", "error", "universal access");
                } else if (initialError === "Temporary") {
                    toast("Component temporary not available.", "error", "project diagram");
                } else {
                    toast("Unexpected Error.", "error", "bug");
                }

                // Remove message from address bar to prevent message showing on page reload
                history.replaceState('', '', window.location.href.replace(window.location.search, ''))
            }

            if (router.currentRoute().componentGroup !== "auth") {
                postbox.publish("redirect", "login");
            }
        }

        // Define a custom binding handler writing rounded (floored) decimal places
        ko.bindingHandlers.floatFloor = {
            update: function (element, valueAccessor, allBindingsAccessor) {
                var value = ko.utils.unwrapObservable(valueAccessor());
                var decimalPlaces = ko.utils.unwrapObservable(allBindingsAccessor().decimalPlaces) || ko.bindingHandlers.floatFloor.defaultDecimalPlaces;
                var factor = Math.pow(10, decimalPlaces);
                var valueFloored = decimalPlaces === 0 ? Math.floor(value) : Math.floor(value * factor) / factor;

                ko.bindingHandlers.text.update(element, function () {
                    return valueFloored.toFixed(decimalPlaces).toString();
                });
            },
            defaultDecimalPlaces: 0
        };

        // Define a custom binding handler making the applied element sticky
        ko.bindingHandlers.initSticky = {
            init: function (element, valueAccessor, allBindings, vModel, bindingContext) {
                $(element).sticky({ // Initialize sticky menu
                    context: vModel.$page,
                    offset: 150,
                    observeChanges: true, // e.g. observes for height changes of the context element
                });
            }
        };

        // Start the application
        ko.applyBindings({route: router.currentRoute});

    }
);