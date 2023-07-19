/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./granted.html", "postbox", "semantic-ui-popup", "semantic-ui-progress"],
    function (ko, template, postbox) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.views = ko.observable(null);

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divAccessible');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Load and set initial data
            this.loadData();
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Round down progress values
                for (var i = 0; i < response.body["views"].length; i++) {

                    // Add sorted view names
                    var viewNames = response.body["views"][i].view_names.split(',')
                    response.body["views"][i].view_names = viewNames.sort((a, b) => a.localeCompare(b, navigator.languages[0] || navigator.language, {
                        numeric: true,
                        ignorePunctuation: false
                    }))

                    // Add cycle progress array
                    response.body["views"][i].scan_scope.cycle_progress = [
                        Math.round(response.body["views"][i].scan_scope.cycle_done),
                        Math.round(response.body["views"][i].scan_scope.cycle_failed),
                        Math.round(response.body["views"][i].scan_scope.cycle_active),
                    ]
                }

                // Set views
                ctx.views(response.body["views"]);

                // Update access flag to display password button, once views are granted
                userAccess(response.body["views"].length > 0)

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: 'scale',
                    reverse: 'auto', // default setting
                    duration: 200
                });
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/views/granted",
                {},
                null,
                callbackSuccess,
                null
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.initViewEntry = function (element, data) {

            // Initialize cards
            $(element).filter(".card").transition({
                animation: 'scale',
                reverse: 'auto', // default setting
                duration: 200
            });

            // Initialize flip shapes
            $(element).filter(".shape").shape({
                duration: 200,
            })

            // Initialize progress bars
            $(element).find('.ui.progress').progress({
                showActivity: false,
            });

            // Initialize tooltips
            $(element).find('[data-html]').popup();
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.toggleDatabase = function (data, event) {
            var btn = $(event.currentTarget)
            if (btn[0].classList.contains("teal")) {
                btn.removeClass("teal")
                btn.blur(); // Remove active state after click
            } else {
                btn.addClass("teal")
                btn.blur(); // Remove active state after click
            }
            $(event.currentTarget.parentElement.parentElement).find(".shape").shape("flip over")
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.toClipboard = function (data, event) {

            // Find content and copy to clipboard
            var val = event.currentTarget.getAttribute("data-value")
            var $temp = $("<textarea>");
            $("body").append($temp);
            $temp.val(val).select();
            document.execCommand("copy");
            $temp.remove();

            // Indicate success
            toast("Copied to clipboard.", "success");
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
