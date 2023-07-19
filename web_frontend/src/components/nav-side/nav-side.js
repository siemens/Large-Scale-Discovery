/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./nav-side.html", "jquery", 'semantic-ui-sticky'],
    function (ko, template, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Store parameters passed by parent
            this.sideNavItems = params.sideNavItems;

            // Get reference to the view model's actual HTML within the DOM
            this.$page = $('#divPage');
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.initSticky = function (element, index, data) {

            // Flip in cards
            $(element).filter(".card").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                interval: 80
            });

            // Initialize progress bars
            $(element).find('.ui.progress').progress({
                showActivity: false,
            });

            // Initialize tooltips
            $(element).find('[data-html]').popup();
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
