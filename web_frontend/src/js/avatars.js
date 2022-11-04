/**
 MIT License

 Copyright (c) 2020 Florian Körner
 https://github.com/DiceBear/avatars

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
 */

(function (f) {
    if (typeof exports === "object" && typeof module !== "undefined") {
        module.exports = f()
    } else if (typeof define === "function" && define.amd) {
        define([], f)
    } else {
        var g;
        if (typeof window !== "undefined") {
            g = window
        } else if (typeof global !== "undefined") {
            g = global
        } else if (typeof self !== "undefined") {
            g = self
        } else {
            g = this
        }
        g.avatars = f()
    }
})(function () {
    var define, module, exports;
    return (function () {
        function r(e, n, t) {
            function o(i, f) {
                if (!n[i]) {
                    if (!e[i]) {
                        var c = "function" == typeof require && require;
                        if (!f && c) return c(i, !0);
                        if (u) return u(i, !0);
                        var a = new Error("Cannot find module '" + i + "'");
                        throw a.code = "MODULE_NOT_FOUND", a
                    }
                    var p = n[i] = {exports: {}};
                    e[i][0].call(p.exports, function (r) {
                        var n = e[i][1][r];
                        return o(n || r)
                    }, p, p.exports, r, e, n, t)
                }
                return n[i].exports
            }

            for (var u = "function" == typeof require && require, i = 0; i < t.length; i++) o(t[i]);
            return o
        }

        return r
    })()({
        1: [function (require, module, exports) {

        }, {}], 2: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FFF8E1',
                100: '#FFECB3',
                200: '#FFE082',
                300: '#FFB74D',
                400: '#FFCA28',
                500: '#FFC107',
                600: '#FFB300',
                700: '#FFA000',
                800: '#FF8F00',
                900: '#FF6F00'
            };
            exports.default = color;

        }, {}], 3: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#E3F2FD',
                100: '#BBDEFB',
                200: '#90CAF9',
                300: '#64B5F6',
                400: '#42A5F5',
                500: '#2196F3',
                600: '#1E88E5',
                700: '#1976D2',
                800: '#1565C0',
                900: '#0D47A1'
            };
            exports.default = color;

        }, {}], 4: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#ECEFF1',
                100: '#CFD8DC',
                200: '#B0BEC5',
                300: '#90A4AE',
                400: '#78909C',
                500: '#607D8B',
                600: '#546E7A',
                700: '#455A64',
                800: '#37474F',
                900: '#263238'
            };
            exports.default = color;

        }, {}], 5: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#EFEBE9',
                100: '#D7CCC8',
                200: '#BCAAA4',
                300: '#A1887F',
                400: '#8D6E63',
                500: '#795548',
                600: '#6D4C41',
                700: '#5D4037',
                800: '#4E342E',
                900: '#3E2723'
            };
            exports.default = color;

        }, {}], 6: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#E0F7FA',
                100: '#B2EBF2',
                200: '#80DEEA',
                300: '#4DD0E1',
                400: '#26C6DA',
                500: '#00BCD4',
                600: '#00ACC1',
                700: '#0097A7',
                800: '#00838F',
                900: '#006064'
            };
            exports.default = color;

        }, {}], 7: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FBE9E7',
                100: '#FFCCBC',
                200: '#FFAB91',
                300: '#A1887F',
                400: '#FF7043',
                500: '#FF5722',
                600: '#F4511E',
                700: '#E64A19',
                800: '#D84315',
                900: '#BF360C'
            };
            exports.default = color;

        }, {}], 8: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#EDE7F6',
                100: '#D1C4E9',
                200: '#B39DDB',
                300: '#9575CD',
                400: '#7E57C2',
                500: '#673AB7',
                600: '#5E35B1',
                700: '#512DA8',
                800: '#4527A0',
                900: '#311B92'
            };
            exports.default = color;

        }, {}], 9: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#E8F5E9',
                100: '#C8E6C9',
                200: '#A5D6A7',
                300: '#81C784',
                400: '#66BB6A',
                500: '#4CAF50',
                600: '#43A047',
                700: '#388E3C',
                800: '#2E7D32',
                900: '#1B5E20'
            };
            exports.default = color;

        }, {}], 10: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FAFAFA',
                100: '#F5F5F5',
                200: '#EEEEEE',
                300: '#E0E0E0',
                400: '#BDBDBD',
                500: '#9E9E9E',
                600: '#757575',
                700: '#616161',
                800: '#424242',
                900: '#212121'
            };
            exports.default = color;

        }, {}], 11: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var amber_1 = require("./amber");
            var blue_1 = require("./blue");
            var blueGrey_1 = require("./blueGrey");
            var brown_1 = require("./brown");
            var cyan_1 = require("./cyan");
            var deepOrange_1 = require("./deepOrange");
            var deepPurple_1 = require("./deepPurple");
            var green_1 = require("./green");
            var grey_1 = require("./grey");
            var indigo_1 = require("./indigo");
            var lightBlue_1 = require("./lightBlue");
            var lightGreen_1 = require("./lightGreen");
            var lime_1 = require("./lime");
            var orange_1 = require("./orange");
            var pink_1 = require("./pink");
            var purple_1 = require("./purple");
            var red_1 = require("./red");
            var teal_1 = require("./teal");
            var yellow_1 = require("./yellow");
            var collection = {
                amber: amber_1.default,
                blue: blue_1.default,
                blueGrey: blueGrey_1.default,
                brown: brown_1.default,
                cyan: cyan_1.default,
                deepOrange: deepOrange_1.default,
                deepPurple: deepPurple_1.default,
                green: green_1.default,
                grey: grey_1.default,
                indigo: indigo_1.default,
                lightBlue: lightBlue_1.default,
                lightGreen: lightGreen_1.default,
                lime: lime_1.default,
                orange: orange_1.default,
                pink: pink_1.default,
                purple: purple_1.default,
                red: red_1.default,
                teal: teal_1.default,
                yellow: yellow_1.default,
            };
            exports.default = collection;

        }, {
            "./amber": 2,
            "./blue": 3,
            "./blueGrey": 4,
            "./brown": 5,
            "./cyan": 6,
            "./deepOrange": 7,
            "./deepPurple": 8,
            "./green": 9,
            "./grey": 10,
            "./indigo": 12,
            "./lightBlue": 13,
            "./lightGreen": 14,
            "./lime": 15,
            "./orange": 16,
            "./pink": 17,
            "./purple": 18,
            "./red": 19,
            "./teal": 20,
            "./yellow": 21
        }], 12: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#E8EAF6',
                100: '#C5CAE9',
                200: '#9FA8DA',
                300: '#7986CB',
                400: '#5C6BC0',
                500: '#3F51B5',
                600: '#3949AB',
                700: '#303F9F',
                800: '#283593',
                900: '#1A237E'
            };
            exports.default = color;

        }, {}], 13: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#E1F5FE',
                100: '#B3E5FC',
                200: '#81D4FA',
                300: '#4FC3F7',
                400: '#29B6F6',
                500: '#03A9F4',
                600: '#039BE5',
                700: '#0288D1',
                800: '#0277BD',
                900: '#01579B'
            };
            exports.default = color;

        }, {}], 14: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#F1F8E9',
                100: '#DCEDC8',
                200: '#C5E1A5',
                300: '#AED581',
                400: '#9CCC65',
                500: '#8BC34A',
                600: '#7CB342',
                700: '#689F38',
                800: '#558B2F',
                900: '#33691E'
            };
            exports.default = color;

        }, {}], 15: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#F9FBE7',
                100: '#F0F4C3',
                200: '#E6EE9C',
                300: '#DCE775',
                400: '#D4E157',
                500: '#CDDC39',
                600: '#C0CA33',
                700: '#AFB42B',
                800: '#9E9D24',
                900: '#827717'
            };
            exports.default = color;

        }, {}], 16: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FFF3E0',
                100: '#FFE0B2',
                200: '#FFCC80',
                300: '#FF8A65',
                400: '#FFA726',
                500: '#FF9800',
                600: '#FB8C00',
                700: '#F57C00',
                800: '#EF6C00',
                900: '#E65100'
            };
            exports.default = color;

        }, {}], 17: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FCE4EC',
                100: '#F8BBD0',
                200: '#F48FB1',
                300: '#F06292',
                400: '#EC407A',
                500: '#E91E63',
                600: '#D81B60',
                700: '#C2185B',
                800: '#AD1457',
                900: '#880E4F'
            };
            exports.default = color;

        }, {}], 18: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#F3E5F5',
                100: '#E1BEE7',
                200: '#CE93D8',
                300: '#BA68C8',
                400: '#AB47BC',
                500: '#9C27B0',
                600: '#8E24AA',
                700: '#7B1FA2',
                800: '#6A1B9A',
                900: '#4A148C'
            };
            exports.default = color;

        }, {}], 19: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FFEBEE',
                100: '#FFCDD2',
                200: '#EF9A9A',
                300: '#E57373',
                400: '#EF5350',
                500: '#F44336',
                600: '#E53935',
                700: '#D32F2F',
                800: '#C62828',
                900: '#B71C1C'
            };
            exports.default = color;

        }, {}], 20: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#E0F2F1',
                100: '#B2DFDB',
                200: '#80CBC4',
                300: '#4DB6AC',
                400: '#26A69A',
                500: '#009688',
                600: '#00897B',
                700: '#00796B',
                800: '#00695C',
                900: '#004D40'
            };
            exports.default = color;

        }, {}], 21: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var color = {
                50: '#FFFDE7',
                100: '#FFF9C4',
                200: '#FFF59D',
                300: '#FFF176',
                400: '#FFEE58',
                500: '#FFEB3B',
                600: '#FDD835',
                700: '#FBC02D',
                800: '#F9A825',
                900: '#F57F17'
            };
            exports.default = color;

        }, {}], 22: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var hexToRgb = require("pure-color/parse/hex");
            var rgbToHsv = require("pure-color/convert/rgb2hsv");
            var rgbToHex = require("pure-color/convert/rgb2hex");
            var hsvToRgb = require("pure-color/convert/hsv2rgb");
            var collection_1 = require("./collection");
            var Color = /** @class */ (function () {
                function Color(color) {
                    if (color === void 0) {
                        color = '#000';
                    }
                    this.alpha = 1;
                    if (color[0] == '#') {
                        this.hex = color;
                    } else {
                        var match = /(.*)\((.*)\)/.exec(color);
                        if (match) {
                            var values = match[2].split(',').map(function (val) {
                                return parseInt(val.trim());
                            });
                            switch (match[1].trim()) {
                                case 'rgb':
                                    this.rgb = values;
                                    break;
                                case 'rgba':
                                    this.rgba = values;
                                    break;
                                case 'hsv':
                                    this.hsv = values;
                                    break;
                                default:
                                    throw new Error('Unsupported color format: ' + color);
                            }
                        } else {
                            throw new Error('Unknown color format: ' + color);
                        }
                    }
                }

                Color.prototype.clone = function () {
                    return new Color('rgb(' + this.rgb.join(',') + ')');
                };
                Object.defineProperty(Color.prototype, "rgb", {
                    get: function () {
                        return (this.color.rgb = this.color.rgb || (this.color.hex ? this.hexToRgb(this.hex) : this.hsvToRgb(this.hsv)));
                    },
                    set: function (rgb) {
                        if (rgb.length != 3) {
                            throw new Error('An array with a length of 3 is expected.');
                        }
                        this.alpha = 1;
                        this.color = {
                            rgb: rgb
                        };
                    },
                    enumerable: false,
                    configurable: true
                });
                Object.defineProperty(Color.prototype, "rgba", {
                    get: function () {
                        return [this.rgb[0], this.rgb[1], this.rgb[2], this.alpha];
                    },
                    set: function (rgba) {
                        if (rgba.length != 4) {
                            throw new Error('An array with a length of 3 is expected.');
                        }
                        this.rgb = [rgba[0], rgba[1], rgba[2]];
                        this.alpha = rgba[3];
                    },
                    enumerable: false,
                    configurable: true
                });
                Object.defineProperty(Color.prototype, "hsv", {
                    get: function () {
                        // Slice array to return copy
                        return (this.color.hsv = this.color.hsv || this.rgbToHsv(this.rgb)).slice(0);
                    },
                    set: function (hsv) {
                        if (hsv.length != 3) {
                            throw new Error('An array with a length of 3 is expected.');
                        }
                        this.alpha = 1;
                        this.color = {
                            hsv: hsv
                        };
                    },
                    enumerable: false,
                    configurable: true
                });
                Object.defineProperty(Color.prototype, "hex", {
                    get: function () {
                        // Slice array to return copy
                        return (this.color.hex = this.color.hex || this.rgbToHex(this.rgb)).slice(0);
                    },
                    set: function (hex) {
                        this.alpha = 1;
                        this.color = {
                            hex: hex
                        };
                    },
                    enumerable: false,
                    configurable: true
                });
                Color.prototype.brighterThan = function (color, difference) {
                    var primaryColorHsv = this.hsv;
                    var secondaryColorHsv = color.hsv;
                    if (primaryColorHsv[2] >= secondaryColorHsv[2] + difference) {
                        return this;
                    }
                    primaryColorHsv[2] = secondaryColorHsv[2] + difference;
                    if (primaryColorHsv[2] > 360) {
                        primaryColorHsv[2] = 360;
                    }
                    this.hsv = primaryColorHsv;
                    return this;
                };
                Color.prototype.darkerThan = function (color, difference) {
                    var primaryColorHsv = this.hsv;
                    var secondaryColorHsv = color.hsv;
                    if (primaryColorHsv[2] <= secondaryColorHsv[2] - difference) {
                        return this;
                    }
                    primaryColorHsv[2] = secondaryColorHsv[2] - difference;
                    if (primaryColorHsv[2] < 0) {
                        primaryColorHsv[2] = 0;
                    }
                    this.hsv = primaryColorHsv;
                    return this;
                };
                Color.prototype.brighterOrDarkerThan = function (color, difference) {
                    var primaryColorHsv = this.hsv;
                    var secondaryColorHsv = color.hsv;
                    if (primaryColorHsv[2] <= secondaryColorHsv[2]) {
                        return this.darkerThan(color, difference);
                    } else {
                        return this.brighterThan(color, difference);
                    }
                };
                Color.prototype.rgbToHex = function (rgb) {
                    return rgbToHex(rgb);
                };
                Color.prototype.hexToRgb = function (hex) {
                    return hexToRgb(hex).map(function (val) {
                        return Math.round(val);
                    });
                };
                Color.prototype.rgbToHsv = function (rgb) {
                    return rgbToHsv(rgb).map(function (val) {
                        return Math.round(val);
                    });
                };
                Color.prototype.hsvToRgb = function (hsv) {
                    return hsvToRgb(hsv).map(function (val) {
                        return Math.round(val);
                    });
                };
                Color.collection = collection_1.default;
                return Color;
            }());
            exports.default = Color;

        }, {
            "./collection": 11,
            "pure-color/convert/hsv2rgb": 26,
            "pure-color/convert/rgb2hex": 27,
            "pure-color/convert/rgb2hsv": 28,
            "pure-color/parse/hex": 29
        }], 23: [function (require, module, exports) {
            "use strict";
            var __assign = (this && this.__assign) || function () {
                __assign = Object.assign || function (t) {
                    for (var s, i = 1, n = arguments.length; i < n; i++) {
                        s = arguments[i];
                        for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
                            t[p] = s[p];
                    }
                    return t;
                };
                return __assign.apply(this, arguments);
            };
            var __spreadArrays = (this && this.__spreadArrays) || function () {
                for (var s = 0, i = 0, il = arguments.length; i < il; i++) s += arguments[i].length;
                for (var r = Array(s), k = 0, i = 0; i < il; i++)
                    for (var a = arguments[i], j = 0, jl = a.length; j < jl; j++, k++)
                        r[k] = a[j];
                return r;
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            var random_1 = require("./random");
            var color_1 = require("./color");
            var parser_1 = require("./parser");
            var Avatars = /** @class */ (function () {
                /**
                 * @param spriteCollection
                 */
                function Avatars(spriteCollection, defaultOptions) {
                    this.spriteCollection = spriteCollection;
                    this.defaultOptions = __assign({userAgent: typeof window !== 'undefined' && window.navigator && window.navigator.userAgent}, defaultOptions);
                }

                /**
                 * Creates an avatar
                 *
                 * @param seed
                 */
                Avatars.prototype.create = function (seed, options) {
                    var _this = this;
                    options = __assign(__assign({}, this.defaultOptions), options);
                    // Apply alias options
                    options = __assign({
                        radius: options.r,
                        width: options.w,
                        height: options.h,
                        margin: options.m,
                        background: options.b
                    }, options);
                    var svg = this.spriteCollection(new random_1.default(seed), options);
                    if (options) {
                        svg = parser_1.default.parse(svg);
                        var viewBox = svg.attributes['viewBox'].split(' ');
                        var viewBoxX = parseInt(viewBox[0]);
                        var viewBoxY = parseInt(viewBox[1]);
                        var viewBoxWidth = parseInt(viewBox[2]);
                        var viewBoxHeight = parseInt(viewBox[3]);
                        if (options.width) {
                            svg.attributes['width'] = options.width.toString();
                        }
                        if (options.height) {
                            svg.attributes['height'] = options.height.toString();
                        }
                        if (options.margin) {
                            var groupable_1 = [];
                            svg.children = svg.children.filter(function (child) {
                                if (_this.isGroupable(child)) {
                                    groupable_1.push(child);
                                    return false;
                                }
                                return true;
                            });
                            svg.children.push({
                                name: 'g',
                                type: 'element',
                                value: '',
                                children: [
                                    {
                                        name: 'g',
                                        type: 'element',
                                        value: '',
                                        children: __spreadArrays([
                                            {
                                                name: 'rect',
                                                type: 'element',
                                                value: '',
                                                children: [],
                                                attributes: {
                                                    fill: 'none',
                                                    width: viewBoxWidth.toString(),
                                                    height: viewBoxHeight.toString(),
                                                    x: viewBoxX.toString(),
                                                    y: viewBoxY.toString(),
                                                },
                                            }
                                        ], groupable_1),
                                        attributes: {
                                            transform: "scale(" + (1 - (options.margin * 2) / 100) + ")",
                                        },
                                    },
                                ],
                                attributes: {
                                    // prettier-ignore
                                    transform: "translate(" + viewBoxWidth * options.margin / 100 + ", " + viewBoxHeight * options.margin / 100 + ")"
                                },
                            });
                        }
                        if (options.background) {
                            svg.children.unshift({
                                name: 'rect',
                                type: 'element',
                                value: '',
                                children: [],
                                attributes: {
                                    fill: options.background,
                                    width: viewBoxWidth.toString(),
                                    height: viewBoxHeight.toString(),
                                    x: viewBoxX.toString(),
                                    y: viewBoxY.toString(),
                                },
                            });
                        }
                        if (options.radius) {
                            var groupable_2 = [];
                            svg.children = svg.children.filter(function (child) {
                                if (_this.isGroupable(child)) {
                                    groupable_2.push(child);
                                    return false;
                                }
                                return true;
                            });
                            svg.children.push({
                                name: 'mask',
                                type: 'element',
                                value: '',
                                children: [
                                    {
                                        name: 'rect',
                                        type: 'element',
                                        value: '',
                                        children: [],
                                        attributes: {
                                            width: viewBoxWidth.toString(),
                                            height: viewBoxHeight.toString(),
                                            rx: ((viewBoxWidth * options.radius) / 100).toString(),
                                            ry: ((viewBoxHeight * options.radius) / 100).toString(),
                                            fill: '#fff',
                                            x: viewBoxX.toString(),
                                            y: viewBoxY.toString(),
                                        },
                                    },
                                ],
                                attributes: {
                                    id: 'avatarsRadiusMask',
                                },
                            }, {
                                name: 'g',
                                type: 'element',
                                value: '',
                                children: groupable_2,
                                attributes: {
                                    mask: "url(#avatarsRadiusMask)",
                                },
                            });
                        }
                    }
                    svg = parser_1.default.stringify(svg);
                    return options && options.base64 ? "data:image/svg+xml;base64," + this.base64EncodeUnicode(svg) : svg;
                };
                Avatars.prototype.isGroupable = function (element) {
                    return element.type === 'element' && ['title', 'desc', 'defs', 'metadata'].indexOf(element.name) === -1;
                };
                Avatars.prototype.base64EncodeUnicode = function (value) {
                    // @see https://www.base64encoder.io/javascript/
                    var utf8Bytes = encodeURIComponent(value).replace(/%([0-9A-F]{2})/g, function (match, p1) {
                        return String.fromCharCode(parseInt("0x" + p1));
                    });
                    return btoa(utf8Bytes);
                };
                Avatars.random = random_1.default;
                Avatars.color = color_1.default;
                return Avatars;
            }());
            exports.default = Avatars;

        }, {"./color": 22, "./parser": 24, "./random": 25}], 24: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var svgson_1 = require("svgson");
            var Parser = /** @class */ (function () {
                function Parser() {
                }

                Parser.parse = function (svg) {
                    return typeof svg === 'string' ? svgson_1.parseSync(svg) : svg;
                };
                Parser.stringify = function (svg) {
                    return typeof svg === 'string' ? svg : svgson_1.stringify(svg);
                };
                return Parser;
            }());
            exports.default = Parser;

        }, {"svgson": 32}], 25: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            var seedrandom = require("seedrandom/seedrandom");
            var Random = /** @class */ (function () {
                function Random(seed) {
                    this.seed = seed;
                    this.prng = seedrandom(seed);
                }

                Random.prototype.bool = function (likelihood) {
                    if (likelihood === void 0) {
                        likelihood = 50;
                    }
                    return this.prng() * 100 < likelihood;
                };
                Random.prototype.integer = function (min, max) {
                    return Math.floor(this.prng() * (max - min + 1) + min);
                };
                Random.prototype.pickone = function (arr) {
                    return arr[this.integer(0, arr.length - 1)];
                };
                return Random;
            }());
            exports.default = Random;

        }, {"seedrandom/seedrandom": 31}], 26: [function (require, module, exports) {
            function hsv2rgb(hsv) {
                var h = hsv[0] / 60,
                    s = hsv[1] / 100,
                    v = hsv[2] / 100,
                    hi = Math.floor(h) % 6;

                var f = h - Math.floor(h),
                    p = 255 * v * (1 - s),
                    q = 255 * v * (1 - (s * f)),
                    t = 255 * v * (1 - (s * (1 - f))),
                    v = 255 * v;

                switch (hi) {
                    case 0:
                        return [v, t, p];
                    case 1:
                        return [q, v, p];
                    case 2:
                        return [p, v, t];
                    case 3:
                        return [p, q, v];
                    case 4:
                        return [t, p, v];
                    case 5:
                        return [v, p, q];
                }
            }

            module.exports = hsv2rgb;
        }, {}], 27: [function (require, module, exports) {
            var clamp = require("../util/clamp");

            function componentToHex(c) {
                var value = Math.round(clamp(c, 0, 255));
                var hex = value.toString(16);

                return hex.length == 1 ? "0" + hex : hex;
            }

            function rgb2hex(rgb) {
                var alpha = rgb.length === 4 ? componentToHex(rgb[3] * 255) : "";

                return "#" + componentToHex(rgb[0]) + componentToHex(rgb[1]) + componentToHex(rgb[2]) + alpha;
            }

            module.exports = rgb2hex;
        }, {"../util/clamp": 30}], 28: [function (require, module, exports) {
            function rgb2hsv(rgb) {
                var r = rgb[0],
                    g = rgb[1],
                    b = rgb[2],
                    min = Math.min(r, g, b),
                    max = Math.max(r, g, b),
                    delta = max - min,
                    h, s, v;

                if (max == 0)
                    s = 0;
                else
                    s = (delta / max * 1000) / 10;

                if (max == min)
                    h = 0;
                else if (r == max)
                    h = (g - b) / delta;
                else if (g == max)
                    h = 2 + (b - r) / delta;
                else if (b == max)
                    h = 4 + (r - g) / delta;

                h = Math.min(h * 60, 360);

                if (h < 0)
                    h += 360;

                v = ((max / 255) * 1000) / 10;

                return [h, s, v];
            }

            module.exports = rgb2hsv;
        }, {}], 29: [function (require, module, exports) {
            function expand(hex) {
                var result = "#";

                for (var i = 1; i < hex.length; i++) {
                    var val = hex.charAt(i);
                    result += val + val;
                }

                return result;
            }

            function hex(hex) {
                // #RGB or #RGBA
                if (hex.length === 4 || hex.length === 5) {
                    hex = expand(hex);
                }

                var rgb = [
                    parseInt(hex.substring(1, 3), 16),
                    parseInt(hex.substring(3, 5), 16),
                    parseInt(hex.substring(5, 7), 16)
                ];

                // #RRGGBBAA
                if (hex.length === 9) {
                    var alpha = parseFloat((parseInt(hex.substring(7, 9), 16) / 255).toFixed(2));
                    rgb.push(alpha);
                }

                return rgb;
            }

            module.exports = hex;
        }, {}], 30: [function (require, module, exports) {
            function clamp(val, min, max) {
                return Math.min(Math.max(val, min), max);
            }

            module.exports = clamp;
        }, {}], 31: [function (require, module, exports) {
            /*
            Copyright 2019 David Bau.

            Permission is hereby granted, free of charge, to any person obtaining
            a copy of this software and associated documentation files (the
            "Software"), to deal in the Software without restriction, including
            without limitation the rights to use, copy, modify, merge, publish,
            distribute, sublicense, and/or sell copies of the Software, and to
            permit persons to whom the Software is furnished to do so, subject to
            the following conditions:

            The above copyright notice and this permission notice shall be
            included in all copies or substantial portions of the Software.

            THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
            EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
            MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
            IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
            CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
            TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
            SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

            */

            (function (global, pool, math) {
//
// The following constants are related to IEEE 754 limits.
//

                var width = 256,        // each RC4 output is 0 <= x < 256
                    chunks = 6,         // at least six RC4 outputs for each double
                    digits = 52,        // there are 52 significant digits in a double
                    rngname = 'random', // rngname: name for Math.random and Math.seedrandom
                    startdenom = math.pow(width, chunks),
                    significance = math.pow(2, digits),
                    overflow = significance * 2,
                    mask = width - 1,
                    nodecrypto;         // node.js crypto module, initialized at the bottom.

//
// seedrandom()
// This is the seedrandom function described above.
//
                function seedrandom(seed, options, callback) {
                    var key = [];
                    options = (options == true) ? {entropy: true} : (options || {});

                    // Flatten the seed string or build one from local entropy if needed.
                    var shortseed = mixkey(flatten(
                        options.entropy ? [seed, tostring(pool)] :
                            (seed == null) ? autoseed() : seed, 3), key);

                    // Use the seed to initialize an ARC4 generator.
                    var arc4 = new ARC4(key);

                    // This function returns a random double in [0, 1) that contains
                    // randomness in every bit of the mantissa of the IEEE 754 value.
                    var prng = function () {
                        var n = arc4.g(chunks),             // Start with a numerator n < 2 ^ 48
                            d = startdenom,                 //   and denominator d = 2 ^ 48.
                            x = 0;                          //   and no 'extra last byte'.
                        while (n < significance) {          // Fill up all significant digits by
                            n = (n + x) * width;              //   shifting numerator and
                            d *= width;                       //   denominator and generating a
                            x = arc4.g(1);                    //   new least-significant-byte.
                        }
                        while (n >= overflow) {             // To avoid rounding up, before adding
                            n /= 2;                           //   last byte, shift everything
                            d /= 2;                           //   right using integer math until
                            x >>>= 1;                         //   we have exactly the desired bits.
                        }
                        return (n + x) / d;                 // Form the number within [0, 1).
                    };

                    prng.int32 = function () {
                        return arc4.g(4) | 0;
                    }
                    prng.quick = function () {
                        return arc4.g(4) / 0x100000000;
                    }
                    prng.double = prng;

                    // Mix the randomness into accumulated entropy.
                    mixkey(tostring(arc4.S), pool);

                    // Calling convention: what to return as a function of prng, seed, is_math.
                    return (options.pass || callback ||
                        function (prng, seed, is_math_call, state) {
                            if (state) {
                                // Load the arc4 state from the given state if it has an S array.
                                if (state.S) {
                                    copy(state, arc4);
                                }
                                // Only provide the .state method if requested via options.state.
                                prng.state = function () {
                                    return copy(arc4, {});
                                }
                            }

                            // If called as a method of Math (Math.seedrandom()), mutate
                            // Math.random because that is how seedrandom.js has worked since v1.0.
                            if (is_math_call) {
                                math[rngname] = prng;
                                return seed;
                            }

                                // Otherwise, it is a newer calling convention, so return the
                            // prng directly.
                            else return prng;
                        })(
                        prng,
                        shortseed,
                        'global' in options ? options.global : (this == math),
                        options.state);
                }

//
// ARC4
//
// An ARC4 implementation.  The constructor takes a key in the form of
// an array of at most (width) integers that should be 0 <= x < (width).
//
// The g(count) method returns a pseudorandom integer that concatenates
// the next (count) outputs from ARC4.  Its return value is a number x
// that is in the range 0 <= x < (width ^ count).
//
                function ARC4(key) {
                    var t, keylen = key.length,
                        me = this, i = 0, j = me.i = me.j = 0, s = me.S = [];

                    // The empty key [] is treated as [0].
                    if (!keylen) {
                        key = [keylen++];
                    }

                    // Set up S using the standard key scheduling algorithm.
                    while (i < width) {
                        s[i] = i++;
                    }
                    for (i = 0; i < width; i++) {
                        s[i] = s[j = mask & (j + key[i % keylen] + (t = s[i]))];
                        s[j] = t;
                    }

                    // The "g" method returns the next (count) outputs as one number.
                    (me.g = function (count) {
                        // Using instance members instead of closure state nearly doubles speed.
                        var t, r = 0,
                            i = me.i, j = me.j, s = me.S;
                        while (count--) {
                            t = s[i = mask & (i + 1)];
                            r = r * width + s[mask & ((s[i] = s[j = mask & (j + t)]) + (s[j] = t))];
                        }
                        me.i = i;
                        me.j = j;
                        return r;
                        // For robust unpredictability, the function call below automatically
                        // discards an initial batch of values.  This is called RC4-drop[256].
                        // See http://google.com/search?q=rsa+fluhrer+response&btnI
                    })(width);
                }

//
// copy()
// Copies internal state of ARC4 to or from a plain object.
//
                function copy(f, t) {
                    t.i = f.i;
                    t.j = f.j;
                    t.S = f.S.slice();
                    return t;
                };

//
// flatten()
// Converts an object tree to nested arrays of strings.
//
                function flatten(obj, depth) {
                    var result = [], typ = (typeof obj), prop;
                    if (depth && typ == 'object') {
                        for (prop in obj) {
                            try {
                                result.push(flatten(obj[prop], depth - 1));
                            } catch (e) {
                            }
                        }
                    }
                    return (result.length ? result : typ == 'string' ? obj : obj + '\0');
                }

//
// mixkey()
// Mixes a string seed into a key that is an array of integers, and
// returns a shortened string seed that is equivalent to the result key.
//
                function mixkey(seed, key) {
                    var stringseed = seed + '', smear, j = 0;
                    while (j < stringseed.length) {
                        key[mask & j] =
                            mask & ((smear ^= key[mask & j] * 19) + stringseed.charCodeAt(j++));
                    }
                    return tostring(key);
                }

//
// autoseed()
// Returns an object for autoseeding, using window.crypto and Node crypto
// module if available.
//
                function autoseed() {
                    try {
                        var out;
                        if (nodecrypto && (out = nodecrypto.randomBytes)) {
                            // The use of 'out' to remember randomBytes makes tight minified code.
                            out = out(width);
                        } else {
                            out = new Uint8Array(width);
                            (global.crypto || global.msCrypto).getRandomValues(out);
                        }
                        return tostring(out);
                    } catch (e) {
                        var browser = global.navigator,
                            plugins = browser && browser.plugins;
                        return [+new Date, global, plugins, global.screen, tostring(pool)];
                    }
                }

//
// tostring()
// Converts an array of charcodes to a string
//
                function tostring(a) {
                    return String.fromCharCode.apply(0, a);
                }

//
// When seedrandom.js is loaded, we immediately mix a few bits
// from the built-in RNG into the entropy pool.  Because we do
// not want to interfere with deterministic PRNG state later,
// seedrandom will not call math.random on its own again after
// initialization.
//
                mixkey(math.random(), pool);

//
// Nodejs and AMD support: export the implementation as a module using
// either convention.
//
                if ((typeof module) == 'object' && module.exports) {
                    module.exports = seedrandom;
                    // When in node.js, try using crypto package for autoseeding.
                    try {
                        nodecrypto = require('crypto');
                    } catch (ex) {
                    }
                } else if ((typeof define) == 'function' && define.amd) {
                    define(function () {
                        return seedrandom;
                    });
                } else {
                    // When included as a plain script, set up Math.seedrandom global.
                    math['seed' + rngname] = seedrandom;
                }


// End anonymous scope, and pass initial values.
            })(
                // global: `self` in browsers (including strict mode and web workers),
                // otherwise `this` in Node and other environments
                (typeof self !== 'undefined') ? self : this,
                [],     // pool: entropy pool starts empty
                Math    // math: package containing random, pow, and seedrandom
            );

        }, {"crypto": 1}], 32: [function (require, module, exports) {
            (function (global, factory) {
                typeof exports === 'object' && typeof module !== 'undefined' ? module.exports = factory() :
                    typeof define === 'function' && define.amd ? define(factory) :
                        (global = global || self, global.svgson = factory());
            }(this, (function () {
                'use strict';

                function _defineProperty(obj, key, value) {
                    if (key in obj) {
                        Object.defineProperty(obj, key, {
                            value: value,
                            enumerable: true,
                            configurable: true,
                            writable: true
                        });
                    } else {
                        obj[key] = value;
                    }

                    return obj;
                }

                function ownKeys(object, enumerableOnly) {
                    var keys = Object.keys(object);

                    if (Object.getOwnPropertySymbols) {
                        var symbols = Object.getOwnPropertySymbols(object);
                        if (enumerableOnly) symbols = symbols.filter(function (sym) {
                            return Object.getOwnPropertyDescriptor(object, sym).enumerable;
                        });
                        keys.push.apply(keys, symbols);
                    }

                    return keys;
                }

                function _objectSpread2(target) {
                    for (var i = 1; i < arguments.length; i++) {
                        var source = arguments[i] != null ? arguments[i] : {};

                        if (i % 2) {
                            ownKeys(Object(source), true).forEach(function (key) {
                                _defineProperty(target, key, source[key]);
                            });
                        } else if (Object.getOwnPropertyDescriptors) {
                            Object.defineProperties(target, Object.getOwnPropertyDescriptors(source));
                        } else {
                            ownKeys(Object(source)).forEach(function (key) {
                                Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key));
                            });
                        }
                    }

                    return target;
                }

                /*!
                 * isobject <https://github.com/jonschlinkert/isobject>
                 *
                 * Copyright (c) 2014-2017, Jon Schlinkert.
                 * Released under the MIT License.
                 */

                var isobject = function isObject(val) {
                    return val != null && typeof val === 'object' && Array.isArray(val) === false;
                };

                function isObjectObject(o) {
                    return isobject(o) === true
                        && Object.prototype.toString.call(o) === '[object Object]';
                }

                var isPlainObject = function isPlainObject(o) {
                    var ctor, prot;

                    if (isObjectObject(o) === false) return false;

                    // If has modified constructor
                    ctor = o.constructor;
                    if (typeof ctor !== 'function') return false;

                    // If has modified prototype
                    prot = ctor.prototype;
                    if (isObjectObject(prot) === false) return false;

                    // If constructor does not have an Object-specific method
                    if (prot.hasOwnProperty('isPrototypeOf') === false) {
                        return false;
                    }

                    // Most likely a plain Object
                    return true;
                };

                var toString = {}.toString;

                var isarray = Array.isArray || function (arr) {
                    return toString.call(arr) == '[object Array]';
                };

                var isobject$1 = function isObject(val) {
                    return val != null && typeof val === 'object' && isarray(val) === false;
                };

                /*!
                 * has-values <https://github.com/jonschlinkert/has-values>
                 *
                 * Copyright (c) 2014-2015, Jon Schlinkert.
                 * Licensed under the MIT License.
                 */

                var hasValues = function hasValue(o, noZero) {
                    if (o === null || o === undefined) {
                        return false;
                    }

                    if (typeof o === 'boolean') {
                        return true;
                    }

                    if (typeof o === 'number') {
                        if (o === 0 && noZero === true) {
                            return false;
                        }
                        return true;
                    }

                    if (o.length !== undefined) {
                        return o.length !== 0;
                    }

                    for (var key in o) {
                        if (o.hasOwnProperty(key)) {
                            return true;
                        }
                    }
                    return false;
                };

                /*!
                 * get-value <https://github.com/jonschlinkert/get-value>
                 *
                 * Copyright (c) 2014-2015, Jon Schlinkert.
                 * Licensed under the MIT License.
                 */

                var getValue = function (obj, prop, a, b, c) {
                    if (!isObject(obj) || !prop) {
                        return obj;
                    }

                    prop = toString$1(prop);

                    // allowing for multiple properties to be passed as
                    // a string or array, but much faster (3-4x) than doing
                    // `[].slice.call(arguments)`
                    if (a) prop += '.' + toString$1(a);
                    if (b) prop += '.' + toString$1(b);
                    if (c) prop += '.' + toString$1(c);

                    if (prop in obj) {
                        return obj[prop];
                    }

                    var segs = prop.split('.');
                    var len = segs.length;
                    var i = -1;

                    while (obj && (++i < len)) {
                        var key = segs[i];
                        while (key[key.length - 1] === '\\') {
                            key = key.slice(0, -1) + '.' + segs[++i];
                        }
                        obj = obj[key];
                    }
                    return obj;
                };

                function isObject(val) {
                    return val !== null && (typeof val === 'object' || typeof val === 'function');
                }

                function toString$1(val) {
                    if (!val) return '';
                    if (Array.isArray(val)) {
                        return val.join('.');
                    }
                    return val;
                }

                var hasValue = function (obj, prop, noZero) {
                    if (isobject$1(obj)) {
                        return hasValues(getValue(obj, prop), noZero);
                    }
                    return hasValues(obj, prop);
                };

                var unsetValue = function unset(obj, prop) {
                    if (!isobject(obj)) {
                        throw new TypeError('expected an object.');
                    }
                    if (obj.hasOwnProperty(prop)) {
                        delete obj[prop];
                        return true;
                    }

                    if (hasValue(obj, prop)) {
                        var segs = prop.split('.');
                        var last = segs.pop();
                        while (segs.length && segs[segs.length - 1].slice(-1) === '\\') {
                            last = segs.pop().slice(0, -1) + '.' + last;
                        }
                        while (segs.length) obj = obj[prop = segs.shift()];
                        return (delete obj[last]);
                    }
                    return true;
                };

                var omitDeep = function omitDeep(value, keys) {
                    if (typeof value === 'undefined') {
                        return {};
                    }

                    if (Array.isArray(value)) {
                        for (var i = 0; i < value.length; i++) {
                            value[i] = omitDeep(value[i], keys);
                        }
                        return value;
                    }

                    if (!isPlainObject(value)) {
                        return value;
                    }

                    if (typeof keys === 'string') {
                        keys = [keys];
                    }

                    if (!Array.isArray(keys)) {
                        return value;
                    }

                    for (var j = 0; j < keys.length; j++) {
                        unsetValue(value, keys[j]);
                    }

                    for (var key in value) {
                        if (value.hasOwnProperty(key)) {
                            value[key] = omitDeep(value[key], keys);
                        }
                    }

                    return value;
                };

                /*!
                 * Determine if an object is a Buffer
                 *
                 * @author   Feross Aboukhadijeh <https://feross.org>
                 * @license  MIT
                 */

                // The _isBuffer check is for Safari 5-7 support, because it's missing
                // Object.prototype.constructor. Remove this eventually
                var isBuffer_1 = function (obj) {
                    return obj != null && (isBuffer(obj) || isSlowBuffer(obj) || !!obj._isBuffer)
                };

                function isBuffer(obj) {
                    return !!obj.constructor && typeof obj.constructor.isBuffer === 'function' && obj.constructor.isBuffer(obj)
                }

                // For Node v0.10 support. Remove this eventually.
                function isSlowBuffer(obj) {
                    return typeof obj.readFloatLE === 'function' && typeof obj.slice === 'function' && isBuffer(obj.slice(0, 0))
                }

                var toString$2 = Object.prototype.toString;

                /**
                 * Get the native `typeof` a value.
                 *
                 * @param  {*} `val`
                 * @return {*} Native javascript type
                 */

                var kindOf = function kindOf(val) {
                    // primitivies
                    if (typeof val === 'undefined') {
                        return 'undefined';
                    }
                    if (val === null) {
                        return 'null';
                    }
                    if (val === true || val === false || val instanceof Boolean) {
                        return 'boolean';
                    }
                    if (typeof val === 'string' || val instanceof String) {
                        return 'string';
                    }
                    if (typeof val === 'number' || val instanceof Number) {
                        return 'number';
                    }

                    // functions
                    if (typeof val === 'function' || val instanceof Function) {
                        return 'function';
                    }

                    // array
                    if (typeof Array.isArray !== 'undefined' && Array.isArray(val)) {
                        return 'array';
                    }

                    // check for instances of RegExp and Date before calling `toString`
                    if (val instanceof RegExp) {
                        return 'regexp';
                    }
                    if (val instanceof Date) {
                        return 'date';
                    }

                    // other objects
                    var type = toString$2.call(val);

                    if (type === '[object RegExp]') {
                        return 'regexp';
                    }
                    if (type === '[object Date]') {
                        return 'date';
                    }
                    if (type === '[object Arguments]') {
                        return 'arguments';
                    }
                    if (type === '[object Error]') {
                        return 'error';
                    }

                    // buffer
                    if (isBuffer_1(val)) {
                        return 'buffer';
                    }

                    // es6: Map, WeakMap, Set, WeakSet
                    if (type === '[object Set]') {
                        return 'set';
                    }
                    if (type === '[object WeakSet]') {
                        return 'weakset';
                    }
                    if (type === '[object Map]') {
                        return 'map';
                    }
                    if (type === '[object WeakMap]') {
                        return 'weakmap';
                    }
                    if (type === '[object Symbol]') {
                        return 'symbol';
                    }

                    // typed arrays
                    if (type === '[object Int8Array]') {
                        return 'int8array';
                    }
                    if (type === '[object Uint8Array]') {
                        return 'uint8array';
                    }
                    if (type === '[object Uint8ClampedArray]') {
                        return 'uint8clampedarray';
                    }
                    if (type === '[object Int16Array]') {
                        return 'int16array';
                    }
                    if (type === '[object Uint16Array]') {
                        return 'uint16array';
                    }
                    if (type === '[object Int32Array]') {
                        return 'int32array';
                    }
                    if (type === '[object Uint32Array]') {
                        return 'uint32array';
                    }
                    if (type === '[object Float32Array]') {
                        return 'float32array';
                    }
                    if (type === '[object Float64Array]') {
                        return 'float64array';
                    }

                    // must be a plain object
                    return 'object';
                };

                function createCommonjsModule(fn, module) {
                    return module = {exports: {}}, fn(module, module.exports), module.exports;
                }

                var renameKeys = createCommonjsModule(function (module) {
                    (function () {

                        function rename(obj, fn) {
                            if (typeof fn !== 'function') {
                                return obj;
                            }

                            var res = {};
                            for (var key in obj) {
                                if (Object.prototype.hasOwnProperty.call(obj, key)) {
                                    res[fn(key, obj[key]) || key] = obj[key];
                                }
                            }
                            return res;
                        }

                        if (module.exports) {
                            module.exports = rename;
                        } else {
                            {
                                window.rename = rename;
                            }
                        }
                    })();
                });

                /**
                 * Expose `renameDeep`
                 */

                var deepRenameKeys = function renameDeep(obj, cb) {
                    var type = kindOf(obj);

                    if (type !== 'object' && type !== 'array') {
                        throw new Error('expected an object');
                    }

                    var res = [];
                    if (type === 'object') {
                        obj = renameKeys(obj, cb);
                        res = {};
                    }

                    for (var key in obj) {
                        if (obj.hasOwnProperty(key)) {
                            var val = obj[key];
                            if (kindOf(val) === 'object' || kindOf(val) === 'array') {
                                res[key] = renameDeep(val, cb);
                            } else {
                                res[key] = val;
                            }
                        }
                    }
                    return res;
                };

                var eventemitter3 = createCommonjsModule(function (module) {

                    var has = Object.prototype.hasOwnProperty
                        , prefix = '~';

                    /**
                     * Constructor to create a storage for our `EE` objects.
                     * An `Events` instance is a plain object whose properties are event names.
                     *
                     * @constructor
                     * @api private
                     */
                    function Events() {
                    }

                    //
                    // We try to not inherit from `Object.prototype`. In some engines creating an
                    // instance in this way is faster than calling `Object.create(null)` directly.
                    // If `Object.create(null)` is not supported we prefix the event names with a
                    // character to make sure that the built-in object properties are not
                    // overridden or used as an attack vector.
                    //
                    if (Object.create) {
                        Events.prototype = Object.create(null);

                        //
                        // This hack is needed because the `__proto__` property is still inherited in
                        // some old browsers like Android 4, iPhone 5.1, Opera 11 and Safari 5.
                        //
                        if (!new Events().__proto__) prefix = false;
                    }

                    /**
                     * Representation of a single event listener.
                     *
                     * @param {Function} fn The listener function.
                     * @param {Mixed} context The context to invoke the listener with.
                     * @param {Boolean} [once=false] Specify if the listener is a one-time listener.
                     * @constructor
                     * @api private
                     */
                    function EE(fn, context, once) {
                        this.fn = fn;
                        this.context = context;
                        this.once = once || false;
                    }

                    /**
                     * Minimal `EventEmitter` interface that is molded against the Node.js
                     * `EventEmitter` interface.
                     *
                     * @constructor
                     * @api public
                     */
                    function EventEmitter() {
                        this._events = new Events();
                        this._eventsCount = 0;
                    }

                    /**
                     * Return an array listing the events for which the emitter has registered
                     * listeners.
                     *
                     * @returns {Array}
                     * @api public
                     */
                    EventEmitter.prototype.eventNames = function eventNames() {
                        var names = []
                            , events
                            , name;

                        if (this._eventsCount === 0) return names;

                        for (name in (events = this._events)) {
                            if (has.call(events, name)) names.push(prefix ? name.slice(1) : name);
                        }

                        if (Object.getOwnPropertySymbols) {
                            return names.concat(Object.getOwnPropertySymbols(events));
                        }

                        return names;
                    };

                    /**
                     * Return the listeners registered for a given event.
                     *
                     * @param {String|Symbol} event The event name.
                     * @param {Boolean} exists Only check if there are listeners.
                     * @returns {Array|Boolean}
                     * @api public
                     */
                    EventEmitter.prototype.listeners = function listeners(event, exists) {
                        var evt = prefix ? prefix + event : event
                            , available = this._events[evt];

                        if (exists) return !!available;
                        if (!available) return [];
                        if (available.fn) return [available.fn];

                        for (var i = 0, l = available.length, ee = new Array(l); i < l; i++) {
                            ee[i] = available[i].fn;
                        }

                        return ee;
                    };

                    /**
                     * Calls each of the listeners registered for a given event.
                     *
                     * @param {String|Symbol} event The event name.
                     * @returns {Boolean} `true` if the event had listeners, else `false`.
                     * @api public
                     */
                    EventEmitter.prototype.emit = function emit(event, a1, a2, a3, a4, a5) {
                        var evt = prefix ? prefix + event : event;

                        if (!this._events[evt]) return false;

                        var listeners = this._events[evt]
                            , len = arguments.length
                            , args
                            , i;

                        if (listeners.fn) {
                            if (listeners.once) this.removeListener(event, listeners.fn, undefined, true);

                            switch (len) {
                                case 1:
                                    return listeners.fn.call(listeners.context), true;
                                case 2:
                                    return listeners.fn.call(listeners.context, a1), true;
                                case 3:
                                    return listeners.fn.call(listeners.context, a1, a2), true;
                                case 4:
                                    return listeners.fn.call(listeners.context, a1, a2, a3), true;
                                case 5:
                                    return listeners.fn.call(listeners.context, a1, a2, a3, a4), true;
                                case 6:
                                    return listeners.fn.call(listeners.context, a1, a2, a3, a4, a5), true;
                            }

                            for (i = 1, args = new Array(len - 1); i < len; i++) {
                                args[i - 1] = arguments[i];
                            }

                            listeners.fn.apply(listeners.context, args);
                        } else {
                            var length = listeners.length
                                , j;

                            for (i = 0; i < length; i++) {
                                if (listeners[i].once) this.removeListener(event, listeners[i].fn, undefined, true);

                                switch (len) {
                                    case 1:
                                        listeners[i].fn.call(listeners[i].context);
                                        break;
                                    case 2:
                                        listeners[i].fn.call(listeners[i].context, a1);
                                        break;
                                    case 3:
                                        listeners[i].fn.call(listeners[i].context, a1, a2);
                                        break;
                                    case 4:
                                        listeners[i].fn.call(listeners[i].context, a1, a2, a3);
                                        break;
                                    default:
                                        if (!args) for (j = 1, args = new Array(len - 1); j < len; j++) {
                                            args[j - 1] = arguments[j];
                                        }

                                        listeners[i].fn.apply(listeners[i].context, args);
                                }
                            }
                        }

                        return true;
                    };

                    /**
                     * Add a listener for a given event.
                     *
                     * @param {String|Symbol} event The event name.
                     * @param {Function} fn The listener function.
                     * @param {Mixed} [context=this] The context to invoke the listener with.
                     * @returns {EventEmitter} `this`.
                     * @api public
                     */
                    EventEmitter.prototype.on = function on(event, fn, context) {
                        var listener = new EE(fn, context || this)
                            , evt = prefix ? prefix + event : event;

                        if (!this._events[evt]) this._events[evt] = listener, this._eventsCount++;
                        else if (!this._events[evt].fn) this._events[evt].push(listener);
                        else this._events[evt] = [this._events[evt], listener];

                        return this;
                    };

                    /**
                     * Add a one-time listener for a given event.
                     *
                     * @param {String|Symbol} event The event name.
                     * @param {Function} fn The listener function.
                     * @param {Mixed} [context=this] The context to invoke the listener with.
                     * @returns {EventEmitter} `this`.
                     * @api public
                     */
                    EventEmitter.prototype.once = function once(event, fn, context) {
                        var listener = new EE(fn, context || this, true)
                            , evt = prefix ? prefix + event : event;

                        if (!this._events[evt]) this._events[evt] = listener, this._eventsCount++;
                        else if (!this._events[evt].fn) this._events[evt].push(listener);
                        else this._events[evt] = [this._events[evt], listener];

                        return this;
                    };

                    /**
                     * Remove the listeners of a given event.
                     *
                     * @param {String|Symbol} event The event name.
                     * @param {Function} fn Only remove the listeners that match this function.
                     * @param {Mixed} context Only remove the listeners that have this context.
                     * @param {Boolean} once Only remove one-time listeners.
                     * @returns {EventEmitter} `this`.
                     * @api public
                     */
                    EventEmitter.prototype.removeListener = function removeListener(event, fn, context, once) {
                        var evt = prefix ? prefix + event : event;

                        if (!this._events[evt]) return this;
                        if (!fn) {
                            if (--this._eventsCount === 0) this._events = new Events();
                            else delete this._events[evt];
                            return this;
                        }

                        var listeners = this._events[evt];

                        if (listeners.fn) {
                            if (
                                listeners.fn === fn
                                && (!once || listeners.once)
                                && (!context || listeners.context === context)
                            ) {
                                if (--this._eventsCount === 0) this._events = new Events();
                                else delete this._events[evt];
                            }
                        } else {
                            for (var i = 0, events = [], length = listeners.length; i < length; i++) {
                                if (
                                    listeners[i].fn !== fn
                                    || (once && !listeners[i].once)
                                    || (context && listeners[i].context !== context)
                                ) {
                                    events.push(listeners[i]);
                                }
                            }

                            //
                            // Reset the array, or remove it completely if we have no more listeners.
                            //
                            if (events.length) this._events[evt] = events.length === 1 ? events[0] : events;
                            else if (--this._eventsCount === 0) this._events = new Events();
                            else delete this._events[evt];
                        }

                        return this;
                    };

                    /**
                     * Remove all listeners, or those of the specified event.
                     *
                     * @param {String|Symbol} [event] The event name.
                     * @returns {EventEmitter} `this`.
                     * @api public
                     */
                    EventEmitter.prototype.removeAllListeners = function removeAllListeners(event) {
                        var evt;

                        if (event) {
                            evt = prefix ? prefix + event : event;
                            if (this._events[evt]) {
                                if (--this._eventsCount === 0) this._events = new Events();
                                else delete this._events[evt];
                            }
                        } else {
                            this._events = new Events();
                            this._eventsCount = 0;
                        }

                        return this;
                    };

                    //
                    // Alias methods names because people roll like that.
                    //
                    EventEmitter.prototype.off = EventEmitter.prototype.removeListener;
                    EventEmitter.prototype.addListener = EventEmitter.prototype.on;

                    //
                    // This function doesn't apply anymore.
                    //
                    EventEmitter.prototype.setMaxListeners = function setMaxListeners() {
                        return this;
                    };

                    //
                    // Expose the prefix.
                    //
                    EventEmitter.prefixed = prefix;

                    //
                    // Allow `EventEmitter` to be imported as module namespace.
                    //
                    EventEmitter.EventEmitter = EventEmitter;

                    //
                    // Expose the module.
                    //
                    {
                        module.exports = EventEmitter;
                    }
                });

                function _defineProperty$1(obj, key, value) {
                    if (key in obj) {
                        Object.defineProperty(obj, key, {
                            value: value,
                            enumerable: true,
                            configurable: true,
                            writable: true
                        });
                    } else {
                        obj[key] = value;
                    }
                    return obj;
                }


                var noop = function noop() {
                };

                var State = {
                    data: 'state-data',
                    cdata: 'state-cdata',
                    tagBegin: 'state-tag-begin',
                    tagName: 'state-tag-name',
                    tagEnd: 'state-tag-end',
                    attributeNameStart: 'state-attribute-name-start',
                    attributeName: 'state-attribute-name',
                    attributeNameEnd: 'state-attribute-name-end',
                    attributeValueBegin: 'state-attribute-value-begin',
                    attributeValue: 'state-attribute-value'
                };

                var Action = {
                    lt: 'action-lt',
                    gt: 'action-gt',
                    space: 'action-space',
                    equal: 'action-equal',
                    quote: 'action-quote',
                    slash: 'action-slash',
                    char: 'action-char',
                    error: 'action-error'
                };

                var Type = {
                    text: 'text',
                    openTag: 'open-tag',
                    closeTag: 'close-tag',
                    attributeName: 'attribute-name',
                    attributeValue: 'attribute-value'
                };

                var charToAction = {
                    ' ': Action.space,
                    '\t': Action.space,
                    '\n': Action.space,
                    '\r': Action.space,
                    '<': Action.lt,
                    '>': Action.gt,
                    '"': Action.quote,
                    "'": Action.quote,
                    '=': Action.equal,
                    '/': Action.slash
                };

                var getAction = function getAction(char) {
                    return charToAction[char] || Action.char;
                };

                /**
                 * @param  {Object} options
                 * @param  {Boolean} options.debug
                 * @return {Object}
                 */
                var create = function create(options) {
                    var _State$data, _State$tagBegin, _State$tagName, _State$tagEnd, _State$attributeNameS,
                        _State$attributeName, _State$attributeNameE, _State$attributeValue, _State$attributeValue2,
                        _lexer$stateMachine;

                    options = Object.assign({debug: false}, options);
                    var lexer = new eventemitter3();
                    var state = State.data;
                    var data = '';
                    var tagName = '';
                    var attrName = '';
                    var attrValue = '';
                    var isClosing = '';
                    var openingQuote = '';

                    var emit = function emit(type, value) {
                        // for now, ignore tags like: '?xml', '!DOCTYPE' or comments
                        if (tagName[0] === '?' || tagName[0] === '!') {
                            return;
                        }
                        var event = {type: type, value: value};
                        if (options.debug) {
                            console.log('emit:', event);
                        }
                        lexer.emit('data', event);
                    };

                    lexer.stateMachine = (_lexer$stateMachine = {}, _defineProperty$1(_lexer$stateMachine, State.data, (_State$data = {}, _defineProperty$1(_State$data, Action.lt, function () {
                        if (data.trim()) {
                            emit(Type.text, data);
                        }
                        tagName = '';
                        isClosing = false;
                        state = State.tagBegin;
                    }), _defineProperty$1(_State$data, Action.char, function (char) {
                        data += char;
                    }), _State$data)), _defineProperty$1(_lexer$stateMachine, State.cdata, _defineProperty$1({}, Action.char, function (char) {
                        data += char;
                        if (data.substr(-3) === ']]>') {
                            emit(Type.text, data.slice(0, -3));
                            data = '';
                            state = State.data;
                        }
                    })), _defineProperty$1(_lexer$stateMachine, State.tagBegin, (_State$tagBegin = {}, _defineProperty$1(_State$tagBegin, Action.space, noop), _defineProperty$1(_State$tagBegin, Action.char, function (char) {
                        tagName = char;
                        state = State.tagName;
                    }), _defineProperty$1(_State$tagBegin, Action.slash, function () {
                        tagName = '';
                        isClosing = true;
                    }), _State$tagBegin)), _defineProperty$1(_lexer$stateMachine, State.tagName, (_State$tagName = {}, _defineProperty$1(_State$tagName, Action.space, function () {
                        if (isClosing) {
                            state = State.tagEnd;
                        } else {
                            state = State.attributeNameStart;
                            emit(Type.openTag, tagName);
                        }
                    }), _defineProperty$1(_State$tagName, Action.gt, function () {
                        if (isClosing) {
                            emit(Type.closeTag, tagName);
                        } else {
                            emit(Type.openTag, tagName);
                        }
                        data = '';
                        state = State.data;
                    }), _defineProperty$1(_State$tagName, Action.slash, function () {
                        state = State.tagEnd;
                        emit(Type.openTag, tagName);
                    }), _defineProperty$1(_State$tagName, Action.char, function (char) {
                        tagName += char;
                        if (tagName === '![CDATA[') {
                            state = State.cdata;
                            data = '';
                            tagName = '';
                        }
                    }), _State$tagName)), _defineProperty$1(_lexer$stateMachine, State.tagEnd, (_State$tagEnd = {}, _defineProperty$1(_State$tagEnd, Action.gt, function () {
                        emit(Type.closeTag, tagName);
                        data = '';
                        state = State.data;
                    }), _defineProperty$1(_State$tagEnd, Action.char, noop), _State$tagEnd)), _defineProperty$1(_lexer$stateMachine, State.attributeNameStart, (_State$attributeNameS = {}, _defineProperty$1(_State$attributeNameS, Action.char, function (char) {
                        attrName = char;
                        state = State.attributeName;
                    }), _defineProperty$1(_State$attributeNameS, Action.gt, function () {
                        data = '';
                        state = State.data;
                    }), _defineProperty$1(_State$attributeNameS, Action.space, noop), _defineProperty$1(_State$attributeNameS, Action.slash, function () {
                        isClosing = true;
                        state = State.tagEnd;
                    }), _State$attributeNameS)), _defineProperty$1(_lexer$stateMachine, State.attributeName, (_State$attributeName = {}, _defineProperty$1(_State$attributeName, Action.space, function () {
                        state = State.attributeNameEnd;
                    }), _defineProperty$1(_State$attributeName, Action.equal, function () {
                        emit(Type.attributeName, attrName);
                        state = State.attributeValueBegin;
                    }), _defineProperty$1(_State$attributeName, Action.gt, function () {
                        attrValue = '';
                        emit(Type.attributeName, attrName);
                        emit(Type.attributeValue, attrValue);
                        data = '';
                        state = State.data;
                    }), _defineProperty$1(_State$attributeName, Action.slash, function () {
                        isClosing = true;
                        attrValue = '';
                        emit(Type.attributeName, attrName);
                        emit(Type.attributeValue, attrValue);
                        state = State.tagEnd;
                    }), _defineProperty$1(_State$attributeName, Action.char, function (char) {
                        attrName += char;
                    }), _State$attributeName)), _defineProperty$1(_lexer$stateMachine, State.attributeNameEnd, (_State$attributeNameE = {}, _defineProperty$1(_State$attributeNameE, Action.space, noop), _defineProperty$1(_State$attributeNameE, Action.equal, function () {
                        emit(Type.attributeName, attrName);
                        state = State.attributeValueBegin;
                    }), _defineProperty$1(_State$attributeNameE, Action.gt, function () {
                        attrValue = '';
                        emit(Type.attributeName, attrName);
                        emit(Type.attributeValue, attrValue);
                        data = '';
                        state = State.data;
                    }), _defineProperty$1(_State$attributeNameE, Action.char, function (char) {
                        attrValue = '';
                        emit(Type.attributeName, attrName);
                        emit(Type.attributeValue, attrValue);
                        attrName = char;
                        state = State.attributeName;
                    }), _State$attributeNameE)), _defineProperty$1(_lexer$stateMachine, State.attributeValueBegin, (_State$attributeValue = {}, _defineProperty$1(_State$attributeValue, Action.space, noop), _defineProperty$1(_State$attributeValue, Action.quote, function (char) {
                        openingQuote = char;
                        attrValue = '';
                        state = State.attributeValue;
                    }), _defineProperty$1(_State$attributeValue, Action.gt, function () {
                        attrValue = '';
                        emit(Type.attributeValue, attrValue);
                        data = '';
                        state = State.data;
                    }), _defineProperty$1(_State$attributeValue, Action.char, function (char) {
                        openingQuote = '';
                        attrValue = char;
                        state = State.attributeValue;
                    }), _State$attributeValue)), _defineProperty$1(_lexer$stateMachine, State.attributeValue, (_State$attributeValue2 = {}, _defineProperty$1(_State$attributeValue2, Action.space, function (char) {
                        if (openingQuote) {
                            attrValue += char;
                        } else {
                            emit(Type.attributeValue, attrValue);
                            state = State.attributeNameStart;
                        }
                    }), _defineProperty$1(_State$attributeValue2, Action.quote, function (char) {
                        if (openingQuote === char) {
                            emit(Type.attributeValue, attrValue);
                            state = State.attributeNameStart;
                        } else {
                            attrValue += char;
                        }
                    }), _defineProperty$1(_State$attributeValue2, Action.gt, function (char) {
                        if (openingQuote) {
                            attrValue += char;
                        } else {
                            emit(Type.attributeValue, attrValue);
                            data = '';
                            state = State.data;
                        }
                    }), _defineProperty$1(_State$attributeValue2, Action.slash, function (char) {
                        if (openingQuote) {
                            attrValue += char;
                        } else {
                            emit(Type.attributeValue, attrValue);
                            isClosing = true;
                            state = State.tagEnd;
                        }
                    }), _defineProperty$1(_State$attributeValue2, Action.char, function (char) {
                        attrValue += char;
                    }), _State$attributeValue2)), _lexer$stateMachine);

                    var step = function step(char) {
                        if (options.debug) {
                            console.log(state, char);
                        }
                        var actions = lexer.stateMachine[state];
                        var action = actions[getAction(char)] || actions[Action.error] || actions[Action.char];
                        action(char);
                    };

                    lexer.write = function (str) {
                        var len = str.length;
                        for (var i = 0; i < len; i++) {
                            step(str[i]);
                        }
                    };

                    return lexer;
                };

                var lexer = {
                    State: State,
                    Action: Action,
                    Type: Type,
                    create: create
                };

                var Type$1 = lexer.Type;

                var NodeType = {
                    element: 'element',
                    text: 'text'
                };

                var createNode = function createNode(params) {
                    return Object.assign({
                        name: '',
                        type: NodeType.element,
                        value: '',
                        parent: null,
                        attributes: {},
                        children: []
                    }, params);
                };

                var create$1 = function create(options) {
                    options = Object.assign({
                        stream: false,
                        parentNodes: true,
                        doneEvent: 'done',
                        tagPrefix: 'tag:',
                        emitTopLevelOnly: false,
                        debug: false
                    }, options);

                    var lexer$1 = void 0,
                        rootNode = void 0,
                        current = void 0,
                        attrName = void 0;

                    var reader = new eventemitter3();

                    var handleLexerData = function handleLexerData(data) {
                        switch (data.type) {

                            case Type$1.openTag:
                                if (current === null) {
                                    current = rootNode;
                                    current.name = data.value;
                                } else {
                                    var node = createNode({
                                        name: data.value,
                                        parent: current
                                    });
                                    current.children.push(node);
                                    current = node;
                                }
                                break;

                            case Type$1.closeTag:
                                var parent = current.parent;
                                if (!options.parentNodes) {
                                    current.parent = null;
                                }
                                if (current.name !== data.value) {
                                    // ignore unexpected closing tag
                                    break;
                                }
                                if (options.stream && parent === rootNode) {
                                    rootNode.children = [];
                                    // do not expose parent node in top level nodes
                                    current.parent = null;
                                }
                                if (!options.emitTopLevelOnly || parent === rootNode) {
                                    reader.emit(options.tagPrefix + current.name, current);
                                    reader.emit('tag', current.name, current);
                                }
                                if (current === rootNode) {
                                    // end of document, stop listening
                                    lexer$1.removeAllListeners('data');
                                    reader.emit(options.doneEvent, current);
                                    rootNode = null;
                                }
                                current = parent;
                                break;

                            case Type$1.text:
                                if (current) {
                                    current.children.push(createNode({
                                        type: NodeType.text,
                                        value: data.value,
                                        parent: options.parentNodes ? current : null
                                    }));
                                }
                                break;

                            case Type$1.attributeName:
                                attrName = data.value;
                                current.attributes[attrName] = '';
                                break;

                            case Type$1.attributeValue:
                                current.attributes[attrName] = data.value;
                                break;
                        }
                    };

                    reader.reset = function () {
                        lexer$1 = lexer.create({debug: options.debug});
                        lexer$1.on('data', handleLexerData);
                        rootNode = createNode();
                        current = null;
                        attrName = '';
                        reader.parse = lexer$1.write;
                    };

                    reader.reset();
                    return reader;
                };

                var parseSync = function parseSync(xml, options) {
                    options = Object.assign({}, options, {stream: false, tagPrefix: ':'});
                    var reader = create$1(options);
                    var res = void 0;
                    reader.on('done', function (ast) {
                        res = ast;
                    });
                    reader.parse(xml);
                    return res;
                };

                var reader = {
                    parseSync: parseSync,
                    create: create$1,
                    NodeType: NodeType
                };
                var reader_1 = reader.parseSync;

                var parseInput = function parseInput(input) {
                    var parsed = reader_1(input, {
                        parentNodes: false
                    });
                    var hasMoreChildren = parsed.name === 'root' && parsed.children.length > 1;
                    var isValid = hasMoreChildren ? parsed.children.reduce(function (acc, _ref) {
                        var name = _ref.name;
                        return !acc ? name === 'svg' : true;
                    }, false) : parsed.children[0].name === 'svg';

                    if (isValid) {
                        return hasMoreChildren ? parsed : parsed.children[0];
                    } else {
                        throw Error('nothing to parse');
                    }
                };
                var removeDoctype = function removeDoctype(input) {
                    return input.replace(/<[\/]{0,1}(\!?DOCTYPE|\??xml)[^><]*>/gi, '');
                };
                var wrapInput = function wrapInput(input) {
                    return "<root>".concat(input, "</root>");
                };
                var removeAttrs = function removeAttrs(obj) {
                    return omitDeep(obj, ['parent']);
                };
                var camelize = function camelize(node) {
                    return deepRenameKeys(node, function (key) {
                        if (!notCamelcase(key)) {
                            return toCamelCase(key);
                        }

                        return key;
                    });
                };
                var toCamelCase = function toCamelCase(prop) {
                    return prop.replace(/[-|:]([a-z])/gi, function (all, letter) {
                        return letter.toUpperCase();
                    });
                };

                var notCamelcase = function notCamelcase(prop) {
                    return /^(data|aria)(-\w+)/.test(prop);
                };

                var escapeText = function escapeText(text) {
                    if (text) {
                        var str = String(text);
                        return /[&<>]/.test(str) ? "<![CDATA[".concat(str.replace(/]]>/, ']]]]><![CDATA[>'), "]]>") : str;
                    }

                    return '';
                };
                var escapeAttr = function escapeAttr(attr) {
                    return String(attr).replace(/&/g, '&amp;').replace(/'/g, '&apos;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                };

                var svgsonSync = function svgsonSync(input) {
                    var _ref = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {},
                        _ref$transformNode = _ref.transformNode,
                        transformNode = _ref$transformNode === void 0 ? function (node) {
                            return node;
                        } : _ref$transformNode,
                        _ref$camelcase = _ref.camelcase,
                        camelcase = _ref$camelcase === void 0 ? false : _ref$camelcase;

                    var wrap = function wrap(input) {
                        var cleanInput = removeDoctype(input);
                        return wrapInput(cleanInput);
                    };

                    var unwrap = function unwrap(res) {
                        return res.name === 'root' ? res.children : res;
                    };

                    var applyFilters = function applyFilters(input) {
                        var applyTransformNode = function applyTransformNode(node) {
                            var children = node.children;
                            return node.name === 'root' ? children.map(applyTransformNode) : _objectSpread2(_objectSpread2({}, transformNode(node)), children && children.length > 0 ? {
                                children: children.map(applyTransformNode)
                            } : {});
                        };

                        var n;
                        n = removeAttrs(input);
                        n = applyTransformNode(n);

                        if (camelcase) {
                            n = camelize(n);
                        }

                        return n;
                    };

                    return unwrap(applyFilters(parseInput(wrap(input))));
                };

                function svgson() {
                    for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
                        args[_key] = arguments[_key];
                    }

                    return new Promise(function (resolve, reject) {
                        try {
                            var res = svgsonSync.apply(void 0, args);
                            resolve(res);
                        } catch (e) {
                            reject(e);
                        }
                    });
                }

                var stringify = function stringify(ast) {
                    var _ref = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {},
                        _ref$transformAttr = _ref.transformAttr,
                        transformAttr = _ref$transformAttr === void 0 ? function (key, value, escape) {
                            return "".concat(key, "=\"").concat(escape(value), "\"");
                        } : _ref$transformAttr,
                        _ref$selfClose = _ref.selfClose,
                        selfClose = _ref$selfClose === void 0 ? true : _ref$selfClose;

                    if (Array.isArray(ast)) {
                        return ast.map(function (ast) {
                            return stringify(ast, {
                                transformAttr: transformAttr,
                                selfClose: selfClose
                            });
                        }).join('');
                    }

                    if (ast.type === 'text') {
                        return escapeText(ast.value);
                    }

                    var attributes = '';

                    for (var attr in ast.attributes) {
                        var attrStr = transformAttr(attr, ast.attributes[attr], escapeAttr, ast.name);
                        attributes += attrStr ? " ".concat(attrStr) : '';
                    }

                    return ast.children.length || !selfClose ? "<".concat(ast.name).concat(attributes, ">").concat(stringify(ast.children, {
                        transformAttr: transformAttr,
                        selfClose: selfClose
                    }), "</").concat(ast.name, ">") : "<".concat(ast.name).concat(attributes, "/>");
                };

                var indexUmd = Object.assign({}, {
                    parse: svgson,
                    parseSync: svgsonSync,
                    stringify: stringify
                });

                return indexUmd;

            })));

        }, {}]
    }, {}, [23])(23)
});
