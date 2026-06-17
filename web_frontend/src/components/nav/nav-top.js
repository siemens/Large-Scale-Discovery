/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "../../globals", "text!./nav-top.html", "postbox", "jquery", "semantic-ui-visibility", "semantic-ui-form", "semantic-ui-dropdown"],
    function (ko, globals, template, postbox, $) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Initialize observables
            this.currentRoute = params.currentRoute

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divNavTop');

            // Keep reference THIS view model context
            var ctx = this;

            // Prepare array of references to subscriptions in order to dispose them later
            this.subscriptions = []

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
        ViewModel.prototype.togglePresentationMode = function (data, event) {
            presentationMode(!presentationMode())
            localStorage.setItem("presentation", (presentationMode()).toString());
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

            // Clear user data
            globals.discard();

            // Navigate to login
            postbox.publish("redirect", "login");

            // Discard cached redirect in case of logout
            sessionStorage.setItem("redirect", "")

            // Log success
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

                        // Get password if returned by the backend.
                        // It's only returned if it couldn't be sent out via encrypted e-mail.
                        var password = response.body["password"]

                        // Show toast message for successful modal.
                        // If password is empty, it was sent out by e-mail by the backend.
                        if (password === "") {
                            toast(response.message, "success");
                        } else {
                            infoOverlay(
                                "key",
                                "Generated Database Password",
                                'Please note the following database password, it will disappear shortly.</br>\n' +
                                '<div class="ui sixteen column centered grid">\n' +
                                '  <div class="six wide column">\n' +
                                '       <table class="ui centered inverted black table">\n' +
                                '         <tbody>\n' +
                                '           <tr class="center aligned">\n' +
                                '             <td>' + password + '</td>\n' +
                                '           </tr>\n' +
                                '         </tbody>\n' +
                                '       </table>\n' +
                                '  </div>\n' +
                                '</div>\n',
                                function () {

                                    // Clear password after dialog close
                                    password = ""

                                    // Bug fix, manually reset right margin to zero. It was changed by first
                                    // modal dimmer to mitigate jumping content when scroll bar disappears.
                                    // However, it fails to automatically reset if a new modal is opened before
                                    // the previous one was completely terminated.
                                    $('body').css("margin-right", "0px")
                                },
                                10000, // Safety timeout for modal, in case it's showing sensitive data
                            )
                        }
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/user/password",
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

            // Dispose subscriptions
            for (var k in this.subscriptions) {
                this.subscriptions[k].dispose();
            }
        };

        return {viewModel: ViewModel, template: template};
    }
);
