/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./agents.html", "postbox", "jquery", "avatars-bottts", "semantic-ui-popup"],
    function (ko, template, postbox, $, avatarBottts) {

        // VIEWMODEL CONSTRUCTION
        function ViewModel(params) {

            // Initialize observables
            this.agentStats = ko.observable(null);
            this.suppressLoadIndicators = false

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divAgents');

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
            }, 19549); // Reload occasionally. Uneven to lower probability of running in parallel with activity update
        }

        // VIEWMODEL ACTION
        ViewModel.prototype.loadData = function (data, event) {

            // Keep reference THIS view model context
            var ctx = this;

            // Handle request success
            const callbackSuccess = function (response, textStatus, jqXHR) {

                // Get list of users
                var data = response.body["scope_agents"]

                // Sanitize empty list
                if (!data || data.length === 0) {
                    data = []
                    ctx.agentStats([])
                }

                // Sanitize values
                data.forEach(function (value, index, array) {

                    // Sanitize empty list
                    if (!value["agents"]) {
                        value["agents"] = []
                    }

                    // Sort list by name naturally
                    value["agents"].sort((a, b) => a.host.localeCompare(b.host, navigator.languages[0] || navigator.language, {
                        numeric: true,
                        ignorePunctuation: false
                    }));

                    // Translate last seen into words and attach
                    for (var j = 0; j < value["agents"].length; j++) {

                        // Strip sensitive characters which might be injected by malicious scan agents
                        var platform = value["agents"][j]["platform"].replace(/[^a-z0-9.:; ]/gi, '')
                        var platform_family = value["agents"][j]["platform_family"].replace(/[^a-z0-9.:; ]/gi, '')
                        var platform_version = value["agents"][j]["platform_version"].replace(/[^a-z0-9.:; ]/gi, '')
                        value["agents"][j]["platform"] = platform.toTitleCase()
                        value["agents"][j]["platform_family"] = platform_family.toTitleCase()
                        value["agents"][j]["platform_version"] = platform_version.toTitleCase()

                        // Calculate time since last seen
                        var now = moment()
                        var last = moment(value["agents"][j].last_seen, datetimeFormatGolang)
                        var minutes = moment.duration(now.diff(last)).asMinutes();
                        var hours = minutes / 60
                        var days = hours / 24
                        var weeks = days / 7;

                        // Generate text for timespan
                        var show_delete = false
                        var last_seen_text = "just now"
                        var last_seen_color = "teal" // seems all good
                        if (days >= 14) {
                            last_seen_text = "weeks ago"
                            last_seen_color = "#9d9d9d" // seems disabled
                            show_delete = true
                        } else if (days >= 7) {
                            last_seen_text = "a week ago"
                            last_seen_color = "#db2828" // seems critical
                            show_delete = true
                        } else if (days >= 2) {
                            last_seen_text = Math.floor(days) + " days ago"
                            last_seen_color = "#db2828" // seems critical
                            show_delete = true
                        } else if (!now.isSame(last, 'day')) {
                            last_seen_text = "yesterday"
                            if (minutes > 30) {
                                last_seen_color = "#fbbd08" // seems strange
                            } else {
                                last_seen_color = "teal" // still ok, maybe just midnight
                            }
                        } else if (hours < 2 && hours >= 1) {
                            last_seen_text = "an hour ago"
                            last_seen_color = "#fbbd08" // seems strange
                        } else if (hours >= 1) {
                            last_seen_text = Math.floor(hours) + " hours ago"
                            last_seen_color = "#fbbd08" // seems strange
                        } else if (minutes > 30) {
                            last_seen_text = Math.floor(minutes) + " min ago"
                            last_seen_color = "#fbbd08" // seems all good
                        } else if (minutes >= 5) {
                            last_seen_text = Math.floor(minutes) + " min ago"
                            last_seen_color = "teal" // seems all good
                        }

                        // Attach timespan text
                        value["agents"][j]["last_seen_text"] = last_seen_text
                        value["agents"][j]["last_seen_color"] = last_seen_color
                        value["agents"][j]["show_delete"] = show_delete
                    }

                    // Move disabled scan agents (those dead for a long time) to the rear of the list. All
                    // remaining scan agents are sorted alphabetically by their name.
                    value["agents"].sort(function (x, y) {
                        return x.last_seen_color !== '#9d9d9d' ? -1 : y.last_seen_color !== '#9d9d9d' ? 1 : 0;
                    });

                });

                // Prepare sequential list of scan agent data to compare whether there were updates
                var oldAgentsOrder = []
                var newAgentsOrder = []
                if (ctx.agentStats() !== null) {
                    ctx.agentStats().forEach(function (value, index, array) {
                        oldAgentsOrder = oldAgentsOrder.concat(value.agents);
                    })
                }
                data.forEach(function (value, index, array) {
                    newAgentsOrder = newAgentsOrder.concat(value.agents);
                })

                // Update data observable, if new data is different
                var fnCompare = function (value, index) {
                    return value.last_seen === oldAgentsOrder[index].last_seen &&
                        value.cpu_rate === oldAgentsOrder[index].cpu_rate &&
                        value.memory_rate === oldAgentsOrder[index].memory_rate
                }
                if (newAgentsOrder.length !== oldAgentsOrder.length || !newAgentsOrder.every(fnCompare)) {

                    // Update observable with new data
                    ctx.agentStats(data);
                    ctx.suppressLoadIndicators = true // Subsequent calls load should happen silently in the background
                }

                // Fade in table
                ctx.$domComponent.children("div:hidden").transition({
                    animation: 'scale',
                    reverse: 'false', // default setting
                    duration: 200
                });
            };

            // Send request
            apiCall(
                "GET",
                "/api/v1/agents",
                {},
                null,
                callbackSuccess,
                null,
                this.suppressLoadIndicators,
                this.suppressLoadIndicators
            );
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.initAgentEntry = function (element, data) {

            // Initialize agent hostname/IP popup
            $(element).find('[data-html]').popup({
                hoverable: true,
            });

            // Initialize tasks popup
            if (data.tasks) {
                $(element).find(".image").popup({
                    inline: true,
                    position: 'left center',
                    forcePosition: true,
                });
            }

            // Generate and render scan agent avatar
            $(element).find(".image").each(function (index) {
                var options = {
                    // https://avatars.dicebear.com/styles/bottts
                    colors: ["amber", "blue", "blueGrey", "brown", "cyan", "deepOrange",
                        "deepPurple", "green", "indigo", "lightBlue", "lightGreen", "lime",
                        "purple", "teal", "yellow"],
                };

                var av = new avatar.default(avatarBottts.default, options);
                this.innerHTML = av.create(data.name)
            })
        };

        // VIEWMODEL ACTION
        ViewModel.prototype.deleteAgent = function (data, event, scopeId, scopeName) {

            // Abort with error if necessary values are not available. This should not happen.
            if (!scopeId || !scopeName) {
                toast("Could not delete scan agent stats.", "error");
                return
            }

            // Request approval and only proceed if action is approved
            confirmOverlay(
                "mask",
                "Delete Scan Agent",
                "This will delete this scan agent's stats until new activity is observed. <br />Are you sure you want to delete scan agent stats of </br><span class=\"ui red text\">'" + data.name + "'</span></br> from</br>  <span class=\"ui red text\">'" + scopeName + "'</span>?",
                function () {

                    // Handle request success
                    const callbackSuccess = function (response, textStatus, jqXHR) {

                        // Show toast message for user
                        toast(response.message, "success");

                        // Hide element without re-organizing list just yet
                        $(event.currentTarget.parentElement.parentElement.parentElement).css("visibility", "hidden")
                    };

                    // Send request
                    apiCall(
                        "POST",
                        "/api/v1/agent/delete",
                        {},
                        {
                            "id": data.id,
                        },
                        callbackSuccess,
                        null
                    );
                }
            );
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
