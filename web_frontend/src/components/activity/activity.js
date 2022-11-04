/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./activity.html", "postbox", "jquery"],
    function (ko, template, postbox, $) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.activities = ko.observable(null);
            this.suppressLoadIndicators = false

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divActivity');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            // Load and set initial data
            this.loadData();

            // Keep reference THIS view model context
            var ctx = this;

            // Schedule regular update. LoadData() should only update the respective data observable(s), if the
            // new data is different to the previous one.
            this.Refresher = setInterval(function () {
                ctx.loadData();
            }, 10000); // Reload occasionally
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Get list of users
                var data = response.body["users"]

                // Translate last seen into words and attach
                for (var i = 0; i < data.length; i++) {

                    // Prepare department description
                    var desc = ""
                    if (data[i]["company"] && data[i]["department"]) {
                        desc = data[i]["company"] + ' (' + data[i]["department"] + ')'
                    } else if (data[i]["company"]) {
                        desc = data[i]["company"]
                    } else if (data[i]["department"]) {
                        desc = data[i]["department"]
                    }
                    data[i]["description"] = desc

                    // Attach parsed timestamp
                    data[i]["last_login"] = moment(data[i].last_login, datetimeFormatGolang)

                    // Calculate display color
                    var timestamp = moment(data[i].last_login, datetimeFormatGolang).set({
                        hour: 0,
                        minute: 0,
                        second: 0,
                        millisecond: 0
                    })
                    var daysAgo = moment.duration(moment().diff(timestamp)).asDays()

                    if (daysAgo >= 62) {
                        data[i]["last_login_color"] = "#dddddd"
                        data[i]["last_login_text"] = "long ago"
                    } else if (daysAgo >= 31) {
                        data[i]["last_login_color"] = "#bbbbbb"
                        data[i]["last_login_text"] = "a month ago"
                    } else if (daysAgo >= 14) {
                        data[i]["last_login_color"] = "#bbbbbb"
                        data[i]["last_login_text"] = "weeks ago"
                    } else if (daysAgo >= 7) {
                        data[i]["last_login_color"] = "#888888"
                        data[i]["last_login_text"] = "a week ago"
                    } else if (daysAgo >= 2) {
                        data[i]["last_login_color"] = "#666666"
                        data[i]["last_login_text"] = Math.round(daysAgo) + " days ago"
                    } else if (daysAgo > 1) {
                        data[i]["last_login_color"] = "#444444"
                        data[i]["last_login_text"] = "yesterday"
                    } else {
                        data[i]["last_login_color"] = "#000000"
                        data[i]["last_login_text"] = "today"
                    }
                }

                // Sort users by last login timestamp
                data.sort((a, b) => b.last_login.valueOf() - a.last_login.valueOf())

                // Reduce to the first n entries
                data = data.slice(0, 16);

                // Drop current user
                data = data.filter(data => data.email !== userEmail());

                // Update data observable, if new data is different
                var dataOld = []
                if (ctx.activities() !== null) {
                    dataOld = ctx.activities()
                }
                var fnCompare = function (value, index) {
                    return value.email === dataOld[index].email
                }
                if (data.length !== dataOld.length || !data.every(fnCompare)) {

                    // Update observable with new data
                    ctx.activities(data);
                    ctx.suppressLoadIndicators = true // Subsequent calls load should happen silently in the background
                }

                // If there was no data, set agentStats to empty list
                if (ctx.activities() === null) {
                    ctx.activities([])
                }

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: 'scale',
                    reverse: 'false', // default setting
                    duration: 200
                });
            };

            // Handle request error
            const callbackError = function (response, textStatus, jqXHR) {

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
                "/api/v1/users",
                {},
                null,
                callbackSuccess,
                callbackError,
                this.suppressLoadIndicators,
                this.suppressLoadIndicators
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.initActivityEntry = function (element, data) {

            // Initialize description tooltips
            $(element).find('[data-tooltip]').popup();

            // Generate and render scan agent avatar
            $(element).find(".image").each(function (index) {
                initAvatar(this, data.email, data.gender, true);
            })
        };

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

            // Clear scheduled update intervals or they will pile up
            clearInterval(this.Refresher);
        };

        // Initialize page with view model and according template
        return {viewModel: ViewModel, template: template};
    }
);
