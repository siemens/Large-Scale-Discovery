/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./profile.html", "postbox"],
    function (ko, template, postbox) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Check authentication and redirect to login if necessary
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divProfile');

            // Fade in table
            this.$domComponent.children("div:hidden").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                duration: 200
            });
        }

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

        // VIEWMODEL ACTION
        ViewModel.prototype.togglePresentationMode = function (data, event) {
            presentationMode(!presentationMode())
            localStorage.setItem("presentation", (presentationMode()).toString());
        }

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
