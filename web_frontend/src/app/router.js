/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "postbox", "crossroads", "hasher"],
    function (ko, postbox, crossroads, hasher) {

        // Initialize crossroads routing
        function activateCrossroads() {
            function parseHash(newHash, oldHash) {
                crossroads.parse(newHash);
            }

            crossroads.normalizeFn = crossroads.NORM_AS_OBJECT;
            hasher.initialized.add(parseHash);
            hasher.changed.add(parseHash);
            hasher.init();
        }

        // Router item constructor
        function RouterItem(url, componentName, componentGroup) {
            this.url = url;
            this.componentName = componentName;
            this.componentGroup = componentGroup;
        }

        // Router constructor
        function Router(config) {
            var currentRoute = this.currentRoute = ko.observable({});
            ko.utils.arrayForEach(config.routes, function (navItem) {
                crossroads.addRoute(navItem.url, function (requestParams) {
                    currentRoute(ko.utils.extend(requestParams, {
                        componentName: navItem.componentName,
                        componentGroup: navItem.componentGroup
                    }));
                });
            });
            activateCrossroads();
        }

        // Redirect channel listening for redirect requests
        postbox.subscribe('redirect', function (url) {

            // Remember originally URL if user got redirected to the login page
            if (url === "login" && hasher.getHash() !== "") {
                sessionStorage.setItem("redirect", hasher.getHash())
            }

            // Redirect to new URL
            hasher.setHash(url);
        });

        // Initialize router
        return new Router({
            routes: [
                new RouterItem('', 'home', 'home'),
                new RouterItem('login', 'login', 'auth'),
                new RouterItem('register', 'register', 'auth'),
                new RouterItem('home', 'home', 'home'),
                new RouterItem('configuration/scopes', 'configuration-scopes', 'configuration'),
                new RouterItem('configuration/views', 'configuration-views', 'configuration'),
                new RouterItem('configuration/tokens', 'configuration-tokens', 'configuration'),
                new RouterItem('admin/users', 'admin-users', 'admin'),
                new RouterItem('admin/groups', 'admin-groups', 'admin'),
                new RouterItem('installation', 'installation', 'installation'),
                new RouterItem('profile', 'profile', 'profile'),
            ]
        });
    }
);
