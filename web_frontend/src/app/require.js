var require = {
    baseUrl: "../src",
    paths: {
        "utils-app": "./js/utils-app",
        "utils-tabulator": "./js/utils-tabulator",
        "globals": "./app/globals",
        "crossroads": "../node_modules/crossroads/dist/crossroads.min",
        "hasher": "../node_modules/hasher/dist/js/hasher.min",
        "knockout": "../node_modules/knockout/build/output/knockout-latest",
        "postbox": "../node_modules/knockout-postbox/build/knockout-postbox.min",
        "text": "../node_modules/requirejs-text/text",
        "signals": "../node_modules/signals/dist/signals.min",
        "chartjs": "../node_modules/chart.js/dist/chart",
        "jquery": "../node_modules/jquery/dist/jquery.min",
        "jquery-tablesort": "../node_modules/jquery-tablesort/jquery.tablesort.min",
        "semantic-ui": "../node_modules/fomantic-ui-css/semantic.min",
        "semantic-ui-visibility": "../node_modules/fomantic-ui-css/components/visibility.min",
        "semantic-ui-dropdown": "../node_modules/fomantic-ui-css/components/dropdown.min",
        "semantic-ui-sticky": "../node_modules/fomantic-ui-css/components/sticky.min",
        "semantic-ui-modal": "../node_modules/fomantic-ui-css/components/modal.min",
        "semantic-ui-dimmer": "../node_modules/fomantic-ui-css/components/dimmer.min",
        "semantic-ui-transition": "../node_modules/fomantic-ui-css/components/transition.min",
        "semantic-ui-form": "../node_modules/fomantic-ui-css/components/form",
        "semantic-ui-popup": "../node_modules/fomantic-ui-css/components/popup",
        "semantic-ui-progress": "../node_modules/fomantic-ui-css/components/progress",
        "semantic-ui-accordion": "../node_modules/fomantic-ui-css/components/accordion",
        "semantic-ui-calendar": "../node_modules/fomantic-ui-css/components/calendar",
        "tabulator-tables": "../node_modules/tabulator-tables/dist/js/tabulator.min",
        "moment": "../node_modules/moment/min/moment.min",
        "avatars": "./js/avatars",
        "avatars-bottts": "./js/avatars-bottts",
        "avatars-avataaars": "./js/avatars-avataaars"
    },

    shim: {
        "globals": {
            deps: ["utils-app", "utils-tabulator"]
        },
        "crossroads": {
            deps: ["signals"]
        },
        "hasher": {
            deps: ["signals"]
        },
        "semantic-ui": {
            deps: ["jquery"]
        },
        "semantic-ui-visibility": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-dropdown": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-sticky": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-modal": {
            deps: ["jquery", "semantic-ui", "semantic-ui-dimmer", "semantic-ui-transition"]
        },
        "semantic-ui-dimmer": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-transition": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-form": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-popup": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-progress": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-accordion": {
            deps: ["jquery", "semantic-ui"]
        },
        "semantic-ui-calendar": {
            deps: ["jquery", "semantic-ui"]
        },
        "jquery-tablesort": {
            deps: ["jquery"]
        },
        "tabulator-tables": {
            deps: ["moment"]
        },
        "hash": {
            deps: ["hash-core"]
        },
        "avatars-bottts": {
            deps: ["avatars"]
        },
        "avatars-avataaars": {
            deps: ["avatars"]
        }
    }
};
