// Node modules
var fs = require('fs');
var vm = require('vm');
var merge = require('deeply');
var chalk = require('chalk');
var es = require('event-stream');

// Gulp and plugins
var gulp = require('gulp');
var reqjs = require('requirejs');
var concat = require('gulp-concat');
var clean = require('gulp-clean');
var replace = require('gulp-replace');
var uglify = require('gulp-uglify');
var htmlreplace = require('gulp-html-replace');
var webserver = require('gulp-webserver');
var rename = require("gulp-rename");
var execSync = require('child_process').execSync;
var babel = require("@babel/core");

// Config
var requireJsRuntimeConfig = vm.runInNewContext(fs.readFileSync('src/app/require.js') + '; require;');
requireJsOptimizerConfig = merge(requireJsRuntimeConfig, {
    out: './dist/scripts.js',
    baseUrl: './src',
    name: 'app/main',
    paths: {
        requireLib: '../node_modules/requirejs/require'
    },
    include: [
        'requireLib',
        'pages/register/register',
        'pages/login/login',
        'pages/home/home',
        'pages/configuration/scopes/scopes',
        'pages/configuration/views/views',
        'pages/admin/users/users',
        'pages/admin/groups/groups',
        'pages/installation/installation',
        'pages/profile/profile',
        'components/activity/activity',
        'components/agents/agents',
        'components/charts/access/access',
        'components/feedback/feedback',
        'components/footer/footer',
        'components/groups/add/add',
        'components/groups/owners/owners',
        'components/nav-side/nav-side',
        'components/nav-top/nav-top',
        'components/scopes/list/list',
        'components/scopes/add/custom/custom',
        'components/scopes/add/assets/assets',
        'components/scopes/add/networks/networks',
        'components/scopes/settings/settings',
        'components/views/list/list',
        'components/views/add/add',
        'components/views/edit/edit',
        'components/views/grant/grant',
        'components/views/granted/granted',
        'components/views/token/add/add',
        'components/views/token/list/list',
    ],
    babel: [ // RequireJs modules that need to be translated into older JS syntax by babel
        'chartjs',
    ],
    insertRequire: ['app/main'],
    bundles: {
        // If you want parts of the site to load on demand, remove them from the 'include' list
        // above, and group them into bundles here.
        // 'bundle-name': [ 'some/module', 'another/module' ],
        // 'another-bundle-name': [ 'yet-another-module' ]
        // 'about-page': [ 'pages/about/about' ]
    },
    onBuildRead: function (name, path, contents) { // Some JS libraries use new JS syntax gulp cannot understand. First translate all input into legacy JS code.
        if (this.babel.indexOf(name) > -1) {
            return babel.transform(contents, {
                "presets": ["@babel/preset-env"]
            }).code
        } else {
            return contents;
        }
    },
});

// Moves the Semantic-ui font images and files to the dist-folder
gulp.task('fonts', function () {
    return gulp.src([
        './node_modules/fomantic-ui-css/themes/default/assets/fonts/*',
        './node_modules/fomantic-ui-css/themes/default/assets/images/*'
    ])
        .pipe(gulp.dest('./dist/fonts/'));
});

// Discovers all AMD dependencies, concatenates together all required .js files, minifies them
gulp.task('js', function () {
    return new Promise(function (resolve, reject) {
        return reqjs.optimize(
            requireJsOptimizerConfig,
            function () {
                resolve();
            },
            function (error) {
                console.error('packing required JS files failed:', error)
                reject();
            }
        )
    });
});

// Moves the images to the dist-folder
gulp.task('img', function () {
    return gulp.src([
        './src/img/header.jpg',
        './src/img/header2.jpg'
    ])
        .pipe(gulp.dest('./dist/img/'));
});

// Concatenates CSS files, rewrites relative paths,...
gulp.task('css', function () {
    //Array of all CSS files needed
    var appCss = gulp.src([
        './node_modules/fomantic-ui-css/semantic.css',
        './node_modules/tabulator-tables/dist/css/semantic-ui/tabulator_semantic-ui.css',
        './src/css/*.css'
    ])
        .pipe(replace(/url\((")?\.\.\/fonts\//g, 'url($1fonts/'))
        .pipe(replace(/url\((")?\.\/themes\/default\/assets\/fonts\//g, 'url($1fonts/'))
        .pipe(replace(/url\((")?\.\/themes\/default\/assets\/images\//g, 'url($1fonts/'));
    var combinedCss = es.concat(appCss).pipe(concat('css.css'));
    return es.concat(combinedCss)
        .pipe(gulp.dest('./dist/'));
});

// Moves the tabulator CSS map file to the dist-folder
gulp.task('maps', function () {
    return gulp.src([
        './node_modules/tabulator-tables/dist/css/semantic-ui/tabulator_semantic-ui.min.css.map',
        './node_modules/moment/min/moment.min.js.map',
        './src/favicon.ico'
    ])
        .pipe(rename(function (path) {
            // Fixes some bug in moments.js appearing after joining JavaScript files, where an ";" is added to
            // the moment.min.js.map request
            if (path.basename === "moment.min.js" && path.extname === ".map") {
                path.extname = path.extname + ";"
            }
        }))
        .pipe(gulp.dest('./dist/'));
});

// Copies index.html, replacing <script> and <link> tags to reference production URLs
gulp.task('html', function () {
    return gulp.src('./src/index.html')
        .pipe(htmlreplace({
            'css': 'css.css?' + Date.now(),
            'js': 'scripts.js?' + Date.now(),
        }))
        .pipe(gulp.dest('./dist/'));
});

gulp.task('audit', function () {
    return new Promise(function (resolve, reject) {
        console.log(execSync('npm audit --production').toString());
        resolve();
    });
});


// Removes all files from ./dist/
gulp.task('clean', function () {
    return gulp.src('./dist/**/*', {read: false})
        .pipe(clean());
});

// Runs a default set of tasks
gulp.task('default', gulp.series('fonts', 'js', 'img', 'css', 'maps', 'html', 'audit'));

// Sets up a webserver with live reload for development
gulp.task('webserver', function () {
    gulp.src('')
        .pipe(webserver({
            livereload: true,
            port: 8050,
            directoryListing: true,
            open: 'http://localhost:8050/src/index.html'
        }));
});

