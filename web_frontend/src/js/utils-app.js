/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

var dateFormat = "DD.MM.YYYY"; // Date format to be used throughout the application
var datetimeFormat = "DD.MM.YYYY | HH:mm:ss"; // Datetime format to be used throughout the application
var datetimeFormatGolang = "YYYY-MM-DDTHH:mm:ss.SSSZ"; // Datetime format as understood by backend written in Golang
var datetimeFormatGolangNoTz = "YYYY-MM-DDTHH:mm:ss"; // Datetime format as understood by backend written in Golang (if timezone is omitted)

/**
 * Sends an ajax request including authentication header to the backend and takes care of logging connection and
 * request errors. Might also take care of popping up error messages in the frontend in the future. Callbacks for
 * a successful and a failed request can be passed. A successful ajax request with an error response will also trigger
 * the error callback. Clients of this function do not need to take care of more than their very specific error/success
 * logic.
 */
function apiCall(method, endpoint, headers, data, fnCallSuccess, fnCallError, suppressToast, suppressLoader) {

    // Build URL
    var url = endpoint;

    // Wrap success callback to add some additional debugging output
    const ajaxSuccess = function (response, textStatus, jqXHR) {

        // remove request from list of running requests
        if (suppressLoader !== true) {
            removeRunningRequest();
        }

        // Observe response for generic error flag and response message
        if (response.error === true) { // Error flag in JSON response

            // Log request error
            console.log("REQUEST FAILED: " + response.message + " (" + method + " " + url + ")");

            // Log response data
            console.log(response);

            // Emit toast message for user
            if (suppressToast !== true) {
                toast(response.message, "error");
            }

            // Execute optional request error callback
            if (fnCallError) {
                fnCallError(jqXHR, textStatus, "");
            }

        } else {

            // Log request success
            if (response.message) {
                console.log("REQUEST SUCCESS: " + response.message + " (" + method + " " + url + ")");
            } else {
                console.log("REQUEST SUCCESS (" + method + " " + url + ")");
            }

            // Log response data
            console.log(response);

            // Update JWT token if contained in response
            if (response.token != null) {

                // Update in memory
                authToken(response["token"]["access_token"]);
                authExpiry(moment(response["token"]["expire"]));

                // Update in storage
                sessionStorage.setItem("token", JSON.stringify(response["token"]));
            }

            // Execute optional request success callback
            if (fnCallSuccess) {
                fnCallSuccess(response, textStatus, jqXHR);
            }
        }
    };

    // Wrap error callback to do some standard error handling
    const ajaxError = function (jqXHR, textStatus, errorThrown) {

        // remove request from list of running requests
        if (suppressLoader !== true) {
            removeRunningRequest();
        }

        // Prepare error message
        var errMsg;
        var errIcon = "bug";

        // Observe for ajax errors
        if (jqXHR.readyState === 4) {

            // HTTP error (can be checked by xhr.status and xhr.statusText)
            console.log("PROTOCOL ERROR: " + jqXHR.status + " " + jqXHR.statusText + " (" + method + " " + url + ")");

            // In case of 401 UNAUTHORIZED re-initialize page to redirect user to login page
            if (jqXHR.status === 401) {

                // Set error message to display
                errMsg = "You are not authorized to execute this action.";
                errIcon = "universal access";

                // Clear storage and redirect to login
                sessionStorage.clear();
                window.location.replace("/");
            } else if (jqXHR.status === 503) {

                // Set error message to display
                errMsg = "Component temporary not available.";
                errIcon = "project diagram";
            } else {

                // Set error message to display
                errMsg = "Unexpected Error:&nbsp;&nbsp;" + jqXHR.status + " " + jqXHR.statusText + ".";
            }

        } else if (jqXHR.readyState === 0) {

            // Network error (i.e. connection refused, access denied due to CORS, etc.)
            console.log("PROTOCOL ERROR: Connection failed (" + method + " " + url + ")");

            // Set error message to display
            errMsg = "Service currently not reachable.";
            errIcon = "plug";

        } else {

            // something weird is happening
            console.log("PROTOCOL ERROR: Unexpected (" + method + " " + url + ")");

            // Set error message to display
            errMsg = "Unexpected protocol error.";
        }

        // Emit toast message for user
        if (errMsg) {
            if (suppressToast !== true) {
                toast(errMsg, "error", errIcon);
            }
        }

        // Execute optional request error callback
        if (fnCallError) {
            fnCallError(jqXHR, textStatus, errorThrown);
        }
    };

    // Execute ajax call (automatically adds authentication header)
    ajaxCall(method, url, headers, data, ajaxSuccess, ajaxError, suppressLoader)
}

/**
 * Execute ajax call with given parameters. In contrast to "apiCall()", this function does not handle errors or emit
 * toast messages or application state changes. This function should only be used if necessary.
 */
function ajaxCall(method, url, headers, data, ajaxSuccess, ajaxError, suppressLoader) {

    // Add authorization header if available
    if (authToken()) {
        headers["Authorization"] = "Bearer " + authToken();
    }

    // Execute ajax request
    if (data) {
        $.ajax({
            url: url,
            method: method,
            headers: headers,
            contentType: "application/json",
            dataType: "json",
            data: JSON.stringify(data),
            success: ajaxSuccess,
            error: ajaxError,
            beforeSend: suppressLoader !== true ? addRunningRequest() : null,
        });
    } else {
        $.ajax({
            url: url,
            method: method,
            headers: headers,
            contentType: "application/json",
            dataType: "json",
            success: ajaxSuccess,
            error: ajaxError,
            beforeSend: suppressLoader !== true ? addRunningRequest() : null,
        });
    }
}

/*
 * Add an item to list of current requests and cause with this visibility of loading sign.
 */
function addRunningRequest() {
    activeRequests.push(1);
}

/*
 * Removes an item from current requests and can cause invisibility of loading sign.
 */
function removeRunningRequest() {
    activeRequests.pop();
}

/*
 * Navigation item object holding relevant information to be put into e.g. a side navigation bar
 */
function NavItem(title, href, target) {
    this.title = title;
    this.href = href;
    this.target = target;
}

function toast(message, messageClass, errIcon) {
    $('body').toast({
        message: message,
        class: messageClass,
        showIcon: errIcon ? errIcon : false,
        position: 'bottom right',
        showProgress: 'bottom',
        progressUp: true
    });
}

function shake() {
    $(this).transition({
        animation: 'shake',
        duration: "700ms"
    })
}

function confirmOverlay(icon, title, messasge, fnAction, confirmWord, fnDeny) {

    // Define input HTML for confirm word
    var inputHtml = '';
    if (confirmWord) {
        inputHtml += ' \
            <div class="ui left icon mini input" style="text-align: center"> \
                <input id="confirmWord" type="text" placeholder="Please confirm name..."> \
                <i class="file signature icon"></i> \
            </div> \
        ';
    }

    // Define modal HTML
    var modalHtml = ' \
        <div class="ui tiny basic test modal front transition" style="display: block !important;"> \
            <div class="ui icon header"> \
                <i class="' + icon + ' icon"></i>' + title + ' \
            </div> \
            <div class="content" style="text-align: center"> \
                <p> \
                    ' + messasge + ' \
                </p> \
                ' + inputHtml + ' \
            </div> \
            <div class="actions"> \
                <button class="ui green basic ok inverted button"> \
                    <i class="checkmark icon"></i> \
                    Yes \
                </button> \
                <button class="ui red cancel inverted button"> \
                    <i class="remove icon"></i> \
                    No \
                </button> \
            </div> \
        </div> \
    ';

    // Define modal
    var $modal = $(modalHtml);

    // Inject modal
    $('body').append($modal);

    // Show modal
    $modal.modal({
        onDeny: function ($element) {
            // Don't do anything other than executing passed deny function
            if (fnDeny) {
                fnDeny()
            }
        },
        onApprove: function ($element) {

            // Verify confirm word, if set
            if (confirmWord) {
                var confirmInput = $("#confirmWord");
                if (confirmInput.val() !== confirmWord) {
                    confirmInput.parent().addClass("error");
                    confirmInput.parent().each(shake);
                    return false
                }
            }

            // Execute action
            fnAction();
        },
        onHidden: function () {
            $modal.remove(); // Remove modal again after use
        },
    }).modal('show');
}

function isIpV4OrSubnet(value) {
    return /^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$/igm.test(value);
}

function isIpV6OrSubnet(value) {
    return /^(?:(?:[a-fA-F\d]{1,4}:){7}(?:[a-fA-F\d]{1,4}|:)|(?:[a-fA-F\d]{1,4}:){6}(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|:[a-fA-F\d]{1,4}|:)|(?:[a-fA-F\d]{1,4}:){5}(?::(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,2}|:)|(?:[a-fA-F\d]{1,4}:){4}(?:(?::[a-fA-F\d]{1,4}){0,1}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,3}|:)|(?:[a-fA-F\d]{1,4}:){3}(?:(?::[a-fA-F\d]{1,4}){0,2}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,4}|:)|(?:[a-fA-F\d]{1,4}:){2}(?:(?::[a-fA-F\d]{1,4}){0,3}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,5}|:)|(?:[a-fA-F\d]{1,4}:){1}(?:(?::[a-fA-F\d]{1,4}){0,4}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,6}|:)|(?::(?:(?::[a-fA-F\d]{1,4}){0,5}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,7}|:)))(?:%[0-9a-zA-Z]{1,})?(\/([0-9]|[1-2][0-9]|3[0-2]))?$/igm.test(value);
}

function isFqdn(value) { // Don't be too strict, people might chose weird (sub) domain names

    // Domain must have at least one dot
    if (value.split(".").length < 2) {
        return false
    }
    // Check whether there are only valid characters
    if (!/(?=^.{4,253}$)(^((?!-)[a-zA-Z0-9-]{0,62}[a-zA-Z0-9]\.)+[a-zA-Z]{2,63}$)/gm.test(value)) {
        return false
    }
    // Return true otherwise
    return true;
}

function sanitizeTargets(targets) {

    // Clean list
    targets = targets.filter(function (x) {
        if (x === null) {
            return false
        } else if (x.input === null) {
            return false
        } else if (x.input === undefined) {
            return false
        } else if (x.input.trim() === "") {
            return false
        }
        return true
    });

    // Convert data types where necessary
    for (var i = 0; i < targets.length; i++) {

        // Sanitize values
        targets[i].timezone = parseFloat(targets[i].timezone);
        targets[i].input = targets[i].input.trim();
        if (targets[i].timezone > 12) {
            targets[i].timezone = 12
        } else if (targets[i].timezone < -12) {
            targets[i].timezone = -12
        }

        // Set timezone to something if invalid
        if (targets[i].timezone === null || isNaN(targets[i].timezone)) {
            targets[i].timezone = 0;
        }

        // Validate input before submitting
        if (!isIpV4OrSubnet(targets[i].input) && !isFqdn(targets[i].input) && targets[i].input !== "localhost") {
            return [[], "Invalid scan input '" + targets[i].input + "'!"]
        }

        // Remove scan timestamps because they will not be updated
        delete targets[i]["scan_started"]
        delete targets[i]["scan_finished"]
    }

    return [targets, ""]
}

/*
 * groups a list of items by a defined key and returns a two-dimensional array with grouped items.
 * E.g.:
 * [
 *      [1,2,3], // Group A
 *      [4,5,6], // Group B
 * ]
 *
 * If the key to group items by is not a direct key of the item, but some sub key, you can define the path to it
 * as an array of keys to follow down.
 */
function itemsByKey(items, groupKey) {

    // Return if there is no data yet
    if (items === null) {
        return null
    }

    // Get value to group by
    var getGroupName = function (item) {
        var sortVal = ""
        if (typeof groupKey === "string") {
            sortVal = item[groupKey]
        } else if (groupKey instanceof Array) {
            var targetVal = item
            for (var j = 0; j < groupKey.length; j++) {
                targetVal = targetVal[groupKey[j]]
            }
            sortVal = targetVal
        }
        return sortVal
    }

    // Sort items by group name
    items.sort((a, b) => (getGroupName(a) > getGroupName(b)) ? 1 : -1)

    // Sort items into groups
    var groups = {}
    for (var i = 0; i < items.length; i++) {

        // Get values
        var item = items[i]
        var sortVal = getGroupName(item)

        // Prepare group entry
        if (!(sortVal in groups)) {
            groups[sortVal] = []
        }

        // Add item to related group
        groups[sortVal].push(item)
    }

    // Translate dictionary into two dimensional array
    var itemsGrouped = []
    for (var k in groups) {
        itemsGrouped.push(groups[k])
    }

    // Return two-dimensional array with grouped items
    return itemsGrouped
}

/*
 * Adds an event listener to an element that only triggers once. Otherwise the event can be triggered multiple times,
 * executing sequentially.
 */
function addEventListenerOnce(element, event, fn) {
    var func = function (e) {
        element.removeEventListener(event, func);
        fn(e);
    };
    element.addEventListener(event, func);
}

/*
 * Retrieves the value of a given parameter from the current query string
 */
function getParameterByName(name) {
    name = name.replace(/[\[\]]/g, '\\$&');
    var regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)'),
        results = regex.exec(window.location.href);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, ' '));
}

function home() {
    return "home"
}

/*
 * Initialize multi select search dropdown, of given ID. The passed KnockoutJs observable referenced, is the one
 * holding the current selection. All values are converted to upper-case (or to lower-case).
 */
function initDropdown(jquerySelector, observableSelectedValues, placeholder, allowAdditions) {
    if (allowAdditions !== true) {
        allowAdditions = false
    }
    $(jquerySelector).dropdown({
        "sortSelect": true,
        "fullTextSearch": true,
        "ignore Diacritics": true,  // Treats special characters same as closest ascii equivalent
        "forceSelection": false,    // Enabling does NOT work together with KnockoutJs observables, but causes a JS error after leaving the select box!
        "allowAdditions": allowAdditions,     // Enables adding new entries
        "hideAdditions": false,     // Shows "add" button/entry in dropdown
        "placeholder": placeholder, // Text to show if nothing is selected
        // DO NOT SET "ignoreCase" to true, it is broken (Fomantic 2.8.7), preventing values from being removed again.
        "onRemove": function (value, text, $choice) {
            observableSelectedValues.remove(value) // Somehow doesn't get removed from observable again automatically
        },
    });
}

/*
 * Initialize slider element
 */
function initSlider(jquerySelector, observableValue, min, max, step) {
    $(jquerySelector).slider({
        min: min,
        max: max,
        start: observableValue(),
        step: step,
        onMove: function (firstVal, secondVal) {
            observableValue(firstVal)
        },
    });
}

/*
 * Initialize avatar image
 */
function initAvatar(jqueryElement, avatarSeed, avatarGender, avatarCircleBackground) {

    // Sanitize values
    avatarGender = avatarGender.toUpperCase()

    // Define default settings
    var style = "transparent"
    var background = ""
    var skin = ["light", "brown", "pale"]
    var tops = ["longHair", "shortHair", "eyepatch", "hat"]
    var topsChance = 95
    var hairColor = ["black", "blond", "brown", "platinum", "gray", "red"]
    var clothes = ["blazer", "sweater", "shirt", "hoodie", "overall"]
    var clothesColor = ["black", "blue", "gray", "heather", "pastel", "pink", "red", "white"]
    var beardChance = 20
    var beardColor = ["auburn", "black", "brown", "platinum", "red", "gray"]
    var accessoriesChance = 10
    var accessoriesColor = ["black", "blue", "gray", "heather", "pastel", "pink", "red", "white"]

    // Decide overall style
    if (avatarCircleBackground) {
        style = "circle"
        background = "teal"
    }

    // Decide gender specific styles
    if (avatarGender === "M") {
        skin = ["light", "brown", "darkBrown"]
        tops = ["shortHair", "eyepatch", "hat"]
        hairColor = ["black", "blond", "brown"]
        clothes = ["sweater", "shirt", "hoodie"]
        clothesColor = ["black", "blue", "gray", "heather", "red", "white"]
        beardChance = 50
        accessoriesColor = ["black", "blue", "gray", "heather", "white"]
    } else if (avatarGender === "W" || avatarGender === "F") {
        tops = ["longHair"]
        topsChance = 100
        clothes = ["blazer", "overall"]
        beardChance = 0
        accessoriesChance = 30
    }

    // Prepare avatar options
    var options = {
        // https://avatars.dicebear.com/styles/avataaars
        style: style,
        background: background,
        top: tops,
        topChance: topsChance,
        hatColor: ["black", "gray", "heather"],
        hairColor: hairColor,
        accessoriesChance: accessoriesChance,
        accessoriesColor: accessoriesColor,
        clothes: clothes,
        facialHair: ["medium", "light", "majestic", "fancy", "magnum"],
        facialHairColor: beardColor,
        facialHairChance: beardChance,
        clothesColor: clothesColor,
        eyes: ["close", "default", "dizzy", "roll", "happy", "side", "squint", "surprised", "wink", "winkWacky"],
        mouth: ["default", "eating", "grimace", "serious", "smile", "tongue", "twinkle"],
        skin: skin,
    };

    // Draw avatar
    var av = new avatar.default(avatarTiles.default, options);
    jqueryElement.innerHTML = av.create(avatarSeed + "1")
}

