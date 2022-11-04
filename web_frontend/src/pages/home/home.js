/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./home.html", "postbox", "semantic-ui-popup", "semantic-ui-progress"],
    function (ko, template, postbox) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Check authentication and redirect to login if necessary
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }
        }

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
