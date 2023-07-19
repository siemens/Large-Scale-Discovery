/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "../../globals", "text!./nav-top.html", "postbox", "jquery", "semantic-ui-visibility", "semantic-ui-form", "semantic-ui-dropdown"],
    function (ko, globals, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.currentRoute = params.currentRoute
            this.avatarInput = ko.computed(function () {
                return [userEmail(), userGender()];
            }, this);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divNavTop');

            // Keep reference THIS view model context
            var ctx = this;

            // Initialize top menu and make it fixed to top on scroll
            this.$domComponent.visibility({
                type: 'fixed',
                offset: 40,
                includeMargin: true,
                onFixed: function (a, b, c) {
                    ctx.$domComponent.css('background-color', 'teal');
                },
                onUnfixed: function (a, b, c) {
                    ctx.$domComponent.css('background-color', 'transparent');
                },
            });

            // Nav-top is already shown in the background with cached data. It needs to be updated if new data
            // arrived after user login (if there was no active session).
            initAvatar(this.$domComponent.find(".image")[0], userEmail(), userGender(), false);
            this.avatarInput.subscribe(function (avatarInput) {
                initAvatar(ctx.$domComponent.find(".image")[0], avatarInput[0], avatarInput[1], false);
            });
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.btnColor = function (route1, route2) {
            var currentRoute = this.currentRoute().componentGroup
            if (currentRoute !== route1 && currentRoute !== route2) {
                return ''
            }
            return 'teal'
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.btnBackground = function (route1, route2) {
            var currentRoute = this.currentRoute().componentGroup
            if (currentRoute !== route1 && currentRoute !== route2) {
                return ''
            }
            return 'white'
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.submitLogout = function (data, event) {

            // Send request
            apiCall(
                "POST",
                "/api/v1/backend/logout",
                {},
                null,
                null,
                null,
                true
            );

            // Reset observed user data
            globals.discard();

            // Redirect to login
            postbox.publish("redirect", "login");

            // Log successful logout
            console.log("Logout successful.");
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.generateDatabasePassword = function (data, event) {

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "key",
                "Generate Database Password",
                "A temporary database password will be generated, valid for a limited time frame.<br />Active sessions are not affected and can be continued.",
                function () {

                    // Handle request success
                    const callbackSuccess = function (response, textStatus, jqXHR) {

                        // Show toast message for user
                        toast(response.message, "success");
                    };

                    // Prepare request body
                    var reqData = {};

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/user/reset",
                        {},
                        null,
                        callbackSuccess,
                        null
                    );
                }
            );
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        return {viewModel: ViewModel, template: template};
    }
);
