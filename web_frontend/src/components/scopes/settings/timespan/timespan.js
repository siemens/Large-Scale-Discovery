/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./timespan.html", "postbox", "jquery", "semantic-ui-dropdown", "semantic-ui-calendar"],
    function (ko, template, postbox, $) {

        /////////////////////////
        // VIEWMODEL CONSTRUCTION
        /////////////////////////
        function ViewModel(params) {

            // Keep reference to PARENT view model context
            this.parent = params.parent;

            // Store parameters passed by parent
            this.id = params.id;
            this.startDay = params.startDay;
            this.startTime = params.startTime;
            this.endDay = params.endDay;
            this.endTime = params.endTime;

            // Prepare reference to DOM elements
            this.$domComponent = null
        }

        ViewModel.prototype.updateColors = function () {
            if (this.$domComponent !== null) {

                // Decide colors
                var color = ""
                var colorBg = "#e0e1e2"
                if (this.startDay === -1 || this.endDay === -1 || this.startTime === "" || this.endTime === "") {
                    color = "#c6c6c6"
                    colorBg = "whitesmoke"
                }

                // Get elements to color
                var elementStartDay = this.$domComponent.find(".start_day");
                var elementEndDay = this.$domComponent.find(".end_day");
                var elementStartTime = this.$domComponent.find(".start_time").find("input");
                var elementEndTime = this.$domComponent.find(".end_time").find("input");

                // Set elements color
                elementStartDay.css({"background-color": colorBg, "color": color});
                elementEndDay.css({"background-color": colorBg, "color": color});
                elementStartTime.css({"background-color": colorBg, "color": color});
                elementEndTime.css({"background-color": colorBg, "color": color});
            }
        }

        ViewModel.prototype.initTimespan = function (elements, ctx) {

            // Get reference to the view model's actual HTML within the DOM
            ctx.$domComponent = $(elements[1])

            // Get calendar input boxes
            var elementStartDay = ctx.$domComponent.find(".start_day");
            var elementEndDay = ctx.$domComponent.find(".end_day");
            var elementStartTime = ctx.$domComponent.find(".start_time");
            var elementEndTime = ctx.$domComponent.find(".end_time");

            // Initialize day data
            elementStartDay.dropdown({
                onChange: function (value, text, $selectedItem) {
                    ctx.startDay = value
                    ctx.parent.updateTimespan(ctx.id, "startDay", value)
                    ctx.updateColors();
                }
            });
            elementEndDay.dropdown({
                onChange: function (value, text, $selectedItem) {
                    ctx.endDay = value
                    ctx.parent.updateTimespan(ctx.id, "endDay", value)
                    ctx.updateColors();
                }
            });

            // Set default values of dropdowns
            elementStartDay.dropdown('set selected', ctx.startDay);
            elementEndDay.dropdown('set selected', ctx.endDay);

            // Initialize time picker with default values
            elementStartTime.calendar({
                type: 'time',
                formatter: {
                    time: function (date, settings) {
                        if (!date) return '';
                        const hours = String(date.getHours()).padStart(2, '0');
                        const minutes = String(date.getMinutes()).padStart(2, '0');
                        return `${hours}:${minutes}`; // e.g., "14:30"
                    }
                },
                ampm: false,
                initialDate: ctx.startTime !== "" ? moment(ctx.startTime, "HH:mm").toDate() : null,
                onChange: function (value, mode) {
                    ctx.startTime = moment(value).format('HH:mm')
                    ctx.parent.updateTimespan(ctx.id, "startTime", moment(value).format('HH:mm'))
                    ctx.updateColors();
                }
            });
            elementEndTime.calendar({
                type: 'time',
                formatter: {
                    time: function (date, settings) {
                        if (!date) return '';
                        const hours = String(date.getHours()).padStart(2, '0');
                        const minutes = String(date.getMinutes()).padStart(2, '0');
                        return `${hours}:${minutes}`; // e.g., "14:30"
                    }
                },
                ampm: false,
                initialDate: ctx.endTime !== "" ? moment(ctx.endTime, "HH:mm").toDate() : null,
                onSelect: function (value, mode) {
                    ctx.endTime = moment(value).format('HH:mm')
                    ctx.parent.updateTimespan(ctx.id, "endTime", moment(value).format('HH:mm'))
                    ctx.updateColors();
                }
            });

            // Process manual inputs
            elementStartTime.find('input').on('change blur', function () {
                var inputValue = $(this).val();
                var parsedDate = moment(inputValue, 'HH:mm', true);
                if (parsedDate.isValid()) {
                    var t = parsedDate.format('HH:mm')
                    elementStartTime.calendar('set date', t, true);
                    ctx.parent.updateTimespan(ctx.id, "startTime", t)
                    ctx.updateColors();
                } else {
                    elementStartTime.calendar('set date', "", true);
                    ctx.parent.updateTimespan(ctx.id, "startTime", "")
                    ctx.startTime = ""
                    ctx.updateColors();
                }
            });
            elementEndTime.find('input').on('change blur', function () {
                var inputValue = $(this).val();
                var parsedDate = moment(inputValue, 'HH:mm', true);
                if (parsedDate.isValid()) {
                    var t = parsedDate.format('HH:mm')
                    elementEndTime.calendar('set date', t, true);
                    ctx.parent.updateTimespan(ctx.id, "endTime", t)
                    ctx.updateColors();
                } else {
                    elementEndTime.calendar('set date', "", true);
                    ctx.parent.updateTimespan(ctx.id, "endTime", "")
                    ctx.endTime = ""
                    ctx.updateColors();
                }
            });

            // Set colors after initial load
            ctx.updateColors();
        }

        // VIEWMODEL DECONSTRUCTION
        ViewModel.prototype.dispose = function (data, event) {

        };

        return {viewModel: ViewModel, template: template};
    }
);
