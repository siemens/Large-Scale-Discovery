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
        g.avatarsBottts = f()
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
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M28 42C37.9411 42 46 33.9411 46 24C46 14.0589 37.9411 6 28 6C18.0589 6 10 14.0589 10 24C10 33.9411 18.0589 42 28 42Z" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M74 42C83.9411 42 92 33.9411 92 24C92 14.0589 83.9411 6 74 6C64.0589 6 56 14.0589 56 24C56 33.9411 64.0589 42 74 42Z" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M28 39C36.2843 39 43 32.2843 43 24C43 15.7157 36.2843 9 28 9C19.7157 9 13 15.7157 13 24C13 32.2843 19.7157 39 28 39Z" fill="#F1EEDA"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M74 39C82.2843 39 89 32.2843 89 24C89 15.7157 82.2843 9 74 9C65.7157 9 59 15.7157 59 24C59 32.2843 65.7157 39 74 39Z" fill="#F1EEDA"/>
    <rect x="26" y="15" width="10" height="10" rx="2" fill="black" fill-opacity="0.8"/>
    <rect x="74" y="15" width="10" height="10" rx="2" fill="black" fill-opacity="0.8"/>
`;

        }, {}],
        2: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M25 27.2L30.5 32.7C31 33.1 31.7 33.1 32.1 32.7L33.7 31.1C34.1 30.6 34.1 29.9 33.7 29.5L28.2 24L33.7 18.5C34.1 18 34.1 17.3 33.7 16.9L32.1 15.3C31.6 14.9 30.9 14.9 30.5 15.3L25 20.8L19.5 15.3C19 14.9 18.3 14.9 17.9 15.3L16.3 16.9C15.9 17.3 15.9 18 16.3 18.5L21.8 24L16.3 29.5C15.9 30 15.9 30.7 16.3 31.1L17.9 32.7C18.4 33.1 19.1 33.1 19.5 32.7L25 27.2Z" fill="black" fill-opacity="0.8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M79 27.2L84.5 32.7C85 33.1 85.7 33.1 86.1 32.7L87.7 31.1C88.1 30.6 88.1 29.9 87.7 29.5L82.2 24L87.7 18.5C88.1 18 88.1 17.3 87.7 16.9L86.1 15.3C85.6 14.9 84.9 14.9 84.5 15.3L79 20.8L73.5 15.3C73 14.9 72.3 14.9 71.9 15.3L70.3 16.9C69.9 17.3 69.9 18 70.3 18.5L75.8 24L70.3 29.5C69.9 30 69.9 30.7 70.3 31.1L71.9 32.7C72.4 33.1 73.1 33.1 73.5 32.7L79 27.2Z" fill="black" fill-opacity="0.8"/>
`;

        }, {}],
        3: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M53 0C87.7469 0 102.001 17.4742 102 31C101.999 44.5258 82.4108 48 53 48C23.9528 48 2 44.5258 2 31C2 17.4742 17.1142 0 53 0Z" fill="black" fill-opacity="0.8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M28.8179 34.654C22.2912 33.3001 17.5833 28.3121 18.3026 23.513C19.0218 18.7139 24.8959 15.9211 31.4226 17.275C37.9493 18.629 42.6572 23.617 41.9379 28.416C41.2187 33.2151 35.3446 36.0079 28.8179 34.654Z" fill="#25A6F5"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M75.4226 34.654C68.8959 36.0079 63.0218 33.2151 62.3026 28.416C61.5833 23.617 66.2912 18.629 72.8179 17.275C79.3446 15.9211 85.2187 18.7139 85.9379 23.513C86.6572 28.3121 81.9493 33.3001 75.4226 34.654Z" fill="#25A6F5"/>
`;

        }, {}],
        4: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect y="4" width="104" height="42" rx="4" fill="black" fill-opacity="0.8"/>
    <rect x="28" y="25" width="10" height="11" rx="2" fill="#8BDDFF"/>
    <rect x="66" y="25" width="10" height="11" rx="2" fill="#8BDDFF"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M21 4H29L12 46H4L21 4Z" fill="white" fill-opacity="0.4"/>
`;

        }, {}],
        5: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="8" y="10" width="88" height="36" rx="4" fill="black" fill-opacity="0.8"/>
    <rect x="28" y="21" width="10" height="17" rx="2" fill="#5EF2B8"/>
    <rect x="66" y="21" width="10" height="17" rx="2" fill="#5EF2B8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M83 10H88L76 46H71L83 10Z" fill="white" fill-opacity="0.4"/>
`;

        }, {}],
        6: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M21 45C29.2843 45 36 38.2843 36 30C36 21.7157 29.2843 15 21 15C12.7157 15 6 21.7157 6 30C6 38.2843 12.7157 45 21 45Z" fill="white" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M83 45C91.2843 45 98 38.2843 98 30C98 21.7157 91.2843 15 83 15C74.7157 15 68 21.7157 68 30C68 38.2843 74.7157 45 83 45Z" fill="white" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M21 42C27.6274 42 33 36.6274 33 30C33 23.3726 27.6274 18 21 18C14.3726 18 9 23.3726 9 30C9 36.6274 14.3726 42 21 42Z" fill="white" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M83 42C89.6274 42 95 36.6274 95 30C95 23.3726 89.6274 18 83 18C76.3726 18 71 23.3726 71 30C71 36.6274 76.3726 42 83 42Z" fill="white" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M21 36C24.3137 36 27 33.3137 27 30C27 26.6863 24.3137 24 21 24C17.6863 24 15 26.6863 15 30C15 33.3137 17.6863 36 21 36Z" fill="white" fill-opacity="0.8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M83 36C86.3137 36 89 33.3137 89 30C89 26.6863 86.3137 24 83 24C79.6863 24 77 26.6863 77 30C77 33.3137 79.6863 36 83 36Z" fill="white" fill-opacity="0.8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M21 33C22.6569 33 24 31.6569 24 30C24 28.3431 22.6569 27 21 27C19.3431 27 18 28.3431 18 30C18 31.6569 19.3431 33 21 33Z" fill="white" fill-opacity="0.8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M83 33C84.6569 33 86 31.6569 86 30C86 28.3431 84.6569 27 83 27C81.3431 27 80 28.3431 80 30C80 31.6569 81.3431 33 83 33Z" fill="white" fill-opacity="0.8"/>
`;

        }, {}],
        7: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M52 48C65.2548 48 76 37.2548 76 24C76 10.7452 65.2548 0 52 0C38.7452 0 28 10.7452 28 24C28 37.2548 38.7452 48 52 48Z" fill="white" fill-opacity="0.4"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M52 44C63.0457 44 72 35.0457 72 24C72 12.9543 63.0457 4 52 4C40.9543 4 32 12.9543 32 24C32 35.0457 40.9543 44 52 44Z" fill="black" fill-opacity="0.8"/>
    <path d="M64.5726 15.8153C64.8743 16.2779 65.4939 16.4082 65.9565 16.1064C66.419 15.8046 66.5494 15.185 66.2476 14.7225L64.5726 15.8153ZM61.5815 9.95547C61.1256 9.64384 60.5033 9.76084 60.1917 10.2168C59.88 10.6728 59.997 11.295 60.453 11.6067L61.5815 9.95547ZM56.3568 9.64222C56.8854 9.80237 57.4437 9.50373 57.6039 8.97518C57.764 8.44663 57.4654 7.88832 56.9368 7.72816L56.3568 9.64222ZM45.7206 8.19769C45.2074 8.40179 44.9569 8.98326 45.161 9.49645C45.3651 10.0096 45.9465 10.2602 46.4597 10.0561L45.7206 8.19769ZM41.7603 13.0388C42.1638 12.6617 42.1852 12.0289 41.8081 11.6254C41.431 11.2219 40.7981 11.2005 40.3947 11.5776L41.7603 13.0388ZM36.4567 17.1052C36.2325 17.6099 36.4599 18.2008 36.9646 18.425C37.4694 18.6492 38.0603 18.4218 38.2845 17.9171L36.4567 17.1052ZM66.2476 14.7225C65.0212 12.8427 63.433 11.2208 61.5815 9.95547L60.453 11.6067C62.0875 12.7238 63.49 14.1559 64.5726 15.8153L66.2476 14.7225ZM56.9368 7.72816C55.3733 7.25438 53.7155 7 52.0001 7V9C53.5169 9 54.9793 9.2248 56.3568 9.64222L56.9368 7.72816ZM52.0001 7C49.784 7 47.6646 7.42456 45.7206 8.19769L46.4597 10.0561C48.1724 9.37496 50.0413 9 52.0001 9V7ZM40.3947 11.5776C38.7378 13.1261 37.3906 15.0029 36.4567 17.1052L38.2845 17.9171C39.108 16.0633 40.2968 14.4066 41.7603 13.0388L40.3947 11.5776Z" fill="white" fill-opacity="0.9"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M52 34C57.5228 34 62 29.5228 62 24C62 18.4772 57.5228 14 52 14C46.4772 14 42 18.4772 42 24C42 29.5228 46.4772 34 52 34Z" fill="#C6080C"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M52 28C54.2091 28 56 26.2091 56 24C56 21.7909 54.2091 20 52 20C49.7909 20 48 21.7909 48 24C48 26.2091 49.7909 28 52 28Z" fill="#EE9337"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M52 25C52.5523 25 53 24.5523 53 24C53 23.4477 52.5523 23 52 23C51.4477 23 51 23.4477 51 24C51 24.5523 51.4477 25 52 25Z" fill="#F5F94F"/>
`;

        }, {}],
        8: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path d="M18 19L30 17" stroke="black" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M20 31C20 27.686 22.9098 25 27 25C30.0902 25 33 27.686 33 31" stroke="black" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M86 20L74 17" stroke="black" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M84 31C84 27.686 81.0902 25 78 25C73.9098 25 71 27.686 71 31" stroke="black" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
`;

        }, {}],
        9: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path d="M29.2691 9.67983C26.7216 9.81334 24.305 11.9225 23.0195 13.8332C21.5357 12.0676 18.9175 10.2223 16.3701 10.3558C10.8883 10.6431 7.51531 14.1586 7.74073 18.4598C8.0406 24.1816 12.6244 27.3464 17.4365 30.7046C19.1523 31.8531 22.441 34.8494 22.8627 35.5214C23.2845 36.1935 25.0034 36.1278 25.4425 35.3862C25.8817 34.6446 28.7463 31.3502 30.3355 30.0286C34.7674 26.1859 38.9981 22.5592 38.6982 16.8374C38.4728 12.5362 34.7509 9.39254 29.2691 9.67983Z" fill="#FF5353" fill-opacity="0.8"/>
    <path d="M87.6299 10.3558C85.0825 10.2223 82.4586 12.0673 80.9805 13.8332C79.6893 11.9222 77.2784 9.81331 74.7309 9.67981C69.2491 9.39252 65.5272 12.5361 65.3017 16.8374C65.0019 22.5591 69.2297 26.1857 73.6645 30.0286C75.2508 31.3501 78.2083 34.6738 78.5575 35.3862C78.9067 36.0987 80.623 36.2131 81.1373 35.5214C81.6515 34.8298 84.8449 31.8529 86.5635 30.7046C91.3728 27.3462 95.9594 24.1816 96.2593 18.4598C96.4847 14.1586 93.1117 10.6431 87.6299 10.3558Z" fill="#FF5353" fill-opacity="0.8"/>
`;

        }, {}],
        10: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const bulging_1 = __importDefault(require("./bulging"));
            const dizzy_1 = __importDefault(require("./dizzy"));
            const eva_1 = __importDefault(require("./eva"));
            const frame_1_1 = __importDefault(require("./frame-1"));
            const frame_2_1 = __importDefault(require("./frame-2"));
            const glow_1 = __importDefault(require("./glow"));
            const hal_1 = __importDefault(require("./hal"));
            const happy_1 = __importDefault(require("./happy"));
            const hearts_1 = __importDefault(require("./hearts"));
            const round_frame_01_1 = __importDefault(require("./round-frame-01"));
            const round_frame_02_1 = __importDefault(require("./round-frame-02"));
            const round_1 = __importDefault(require("./round"));
            const sensor_1 = __importDefault(require("./sensor"));
            const shade_01_1 = __importDefault(require("./shade-01"));
            exports.default = [
                bulging_1.default,
                dizzy_1.default,
                eva_1.default,
                frame_1_1.default,
                frame_2_1.default,
                glow_1.default,
                hal_1.default,
                happy_1.default,
                hearts_1.default,
                round_frame_01_1.default,
                round_frame_02_1.default,
                round_1.default,
                sensor_1.default,
                shade_01_1.default
            ];

        }, {
            "./bulging": 1,
            "./dizzy": 2,
            "./eva": 3,
            "./frame-1": 4,
            "./frame-2": 5,
            "./glow": 6,
            "./hal": 7,
            "./happy": 8,
            "./hearts": 9,
            "./round": 13,
            "./round-frame-01": 11,
            "./round-frame-02": 12,
            "./sensor": 14,
            "./shade-01": 15
        }],
        11: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect y="12" width="104" height="32" rx="16" fill="black" fill-opacity="0.8"/>
    <rect x="24" y="22" width="10" height="12" rx="2" fill="#F4F4F4"/>
    <rect x="70" y="22" width="10" height="12" rx="2" fill="#F4F4F4"/>
`;

        }, {}],
        12: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect y="11" width="104" height="34" rx="17" fill="black" fill-opacity="0.8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M29 41C36.1797 41 42 35.1797 42 28C42 20.8203 36.1797 15 29 15C21.8203 15 16 20.8203 16 28C16 35.1797 21.8203 41 29 41Z" fill="#F1EEDA"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M75 41C82.1797 41 88 35.1797 88 28C88 20.8203 82.1797 15 75 15C67.8203 15 62 20.8203 62 28C62 35.1797 67.8203 41 75 41Z" fill="#F1EEDA"/>
    <rect x="24" y="23" width="10" height="10" rx="2" fill="black" fill-opacity="0.8"/>
    <rect x="70" y="23" width="10" height="10" rx="2" fill="black" fill-opacity="0.8"/>
`;

        }, {}],
        13: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M24 36C27.3137 36 30 33.3137 30 30C30 26.6863 27.3137 24 24 24C20.6863 24 18 26.6863 18 30C18 33.3137 20.6863 36 24 36Z" fill="white"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M80 36C83.3137 36 86 33.3137 86 30C86 26.6863 83.3137 24 80 24C76.6863 24 74 26.6863 74 30C74 33.3137 76.6863 36 80 36Z" fill="white"/>
`;

        }, {}],
        14: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M28 44C38.3707 44 46.8978 36.1066 47.9012 26H89.416C90.1876 27.7659 91.9497 29 94 29C96.7614 29 99 26.7614 99 24C99 21.2386 96.7614 19 94 19C91.9497 19 90.1876 20.2341 89.416 22H47.9012C46.8978 11.8934 38.3707 4 28 4C16.9543 4 8 12.9543 8 24C8 35.0457 16.9543 44 28 44Z" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M94 26C95.1046 26 96 25.1046 96 24C96 22.8954 95.1046 22 94 22C92.8954 22 92 22.8954 92 24C92 25.1046 92.8954 26 94 26Z" fill="white"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M28 40C36.8366 40 44 32.8366 44 24C44 15.1634 36.8366 8 28 8C19.1634 8 12 15.1634 12 24C12 32.8366 19.1634 40 28 40Z" fill="black" fill-opacity="0.6"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M34 19C35.6569 19 37 17.6569 37 16C37 14.3431 35.6569 13 34 13C32.3431 13 31 14.3431 31 16C31 17.6569 32.3431 19 34 19Z" fill="white"/>
`;

        }, {}],
        15: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M0 10C0 5.58172 3.58172 2 8 2H96C100.418 2 104 5.58172 104 10V28C104 32.4183 100.418 36 96 36H82.9808C79.7112 36.1672 77.4997 37.137 76.0046 38.3473C75.5176 38.7414 75.0239 39.1634 74.5216 39.5927C72.0465 41.7082 69.3652 44 66.2779 44H52H38.6676C35.5324 44 32.732 41.8707 30.0462 39.8285L30.0461 39.8285C29.096 39.1061 28.1602 38.3945 27.229 37.792C25.7725 36.8497 23.7704 36.1407 21.0192 36H8C3.58172 36 0 32.4183 0 28V10Z" fill="black" fill-opacity="0.8"/>
    <mask id="eyesShade01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="2" width="104" height="42">
        <path fill-rule="evenodd" clip-rule="evenodd" d="M0 10C0 5.58172 3.58172 2 8 2H96C100.418 2 104 5.58172 104 10V28C104 32.4183 100.418 36 96 36H82.9808C79.7112 36.1672 77.4997 37.137 76.0046 38.3473C75.5176 38.7414 75.0239 39.1634 74.5216 39.5927C72.0465 41.7082 69.3652 44 66.2779 44H52H38.6676C35.5324 44 32.732 41.8707 30.0462 39.8285L30.0461 39.8285C29.096 39.1061 28.1602 38.3945 27.229 37.792C25.7725 36.8497 23.7704 36.1407 21.0192 36H8C3.58172 36 0 32.4183 0 28V10Z" fill="white"/>
    </mask>
    <g mask="url(#eyesShade01Mask0)">
        <path fill-rule="evenodd" clip-rule="evenodd" d="M12 19C12 16.2386 14.2386 14 17 14H87C89.7614 14 92 16.2386 92 19V21C92 23.7614 89.7614 26 87 26H74.9808C71.7112 26.1672 69.4997 27.137 68.0046 28.3473C67.5176 28.7414 67.0239 29.1634 66.5216 29.5927C64.5182 31.3051 62.3796 33.133 60 33.7674V34H58.2779H52H46.6676C43.5324 34 40.732 31.8707 38.0462 29.8285L38.0461 29.8285C37.096 29.1061 36.1602 28.3945 35.229 27.792C33.7725 26.8497 31.7704 26.1407 29.0192 26H17C14.2386 26 12 23.7614 12 21V19Z" fill="#FF3D3D"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M12 44L32 -2H28L8 44H12ZM50 -2H39L19 44H30L50 -2Z" fill="white" fill-opacity="0.2"/>
    </g>
`;

        }, {}],
        16: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const round_01_1 = __importDefault(require("./round-01"));
            const round_02_1 = __importDefault(require("./round-02"));
            const square_01_1 = __importDefault(require("./square-01"));
            const square_02_1 = __importDefault(require("./square-02"));
            const square_03_1 = __importDefault(require("./square-03"));
            const square_04_1 = __importDefault(require("./square-04"));
            exports.default = [round_01_1.default, round_02_1.default, square_01_1.default, square_02_1.default, square_03_1.default, square_04_1.default];

        }, {
            "./round-01": 17,
            "./round-02": 18,
            "./square-01": 19,
            "./square-02": 20,
            "./square-03": 21,
            "./square-04": 22
        }],
        17: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color, texture) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M66 0C124.352 0 130.001 40.6854 130 78C129.999 111.315 104.534 120 66 120C28.5387 120 0 111.315 0 78C0 40.6854 7.64843 0 66 0Z" fill="black" fill-opacity="0.8"/>
        <mask id="faceRound01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="130" height="120">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M66 0C124.352 0 130.001 40.6854 130 78C129.999 111.315 104.534 120 66 120C28.5387 120 0 111.315 0 78C0 40.6854 7.64843 0 66 0Z" fill="white"/>
        </mask>
        <g mask="url(#faceRound01Mask0)">
            <rect x="-4" y="-2" width="138" height="124" fill="${color.hex}"/>
            ${texture}
        </g>
    `;
            };

        }, {}],
        18: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color, texture) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M0 31C0 31.0205 0.0141049 30.8164 0 30C0.183375 29.5235 0.402009 28.5029 1 27C1.82671 23.944 3.7804 20.4435 7 17C16.6944 6.60017 35.1724 0 65 0C94.8276 0 113.306 6.60036 123 17C126.22 20.4435 128.173 23.944 129 27C129.598 28.5029 129.817 29.5236 130 30C129.986 30.8164 130 31.0205 130 31V70C130 69.8964 129.972 70.5012 130 71C129.739 73.1171 129.471 75.0149 129 77C127.814 82.9912 125.606 88.911 122 94C112.283 110.337 94.2553 120 65 120C35.7448 120 17.7164 110.338 8 94C4.39414 88.9108 2.1865 82.9912 1 77C0.529043 75.0149 0.261028 73.1171 0 71C0.0282767 70.5466 6.49997e-05 69.6771 0 70V31Z" fill="#E1E6E8"/>
        <mask id="faceRound02Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="130" height="120">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M0 31C0 31.0205 0.0141049 30.8164 0 30C0.183375 29.5235 0.402009 28.5029 1 27C1.82671 23.944 3.7804 20.4435 7 17C16.6944 6.60017 35.1724 0 65 0C94.8276 0 113.306 6.60036 123 17C126.22 20.4435 128.173 23.944 129 27C129.598 28.5029 129.817 29.5236 130 30C129.986 30.8164 130 31.0205 130 31V70C130 69.8964 129.972 70.5012 130 71C129.739 73.1171 129.471 75.0149 129 77C127.814 82.9912 125.606 88.911 122 94C112.283 110.337 94.2553 120 65 120C35.7448 120 17.7164 110.338 8 94C4.39414 88.9108 2.1865 82.9912 1 77C0.529043 75.0149 0.261028 73.1171 0 71C0.0282767 70.5466 6.49997e-05 69.6771 0 70V31Z" fill="white"/>
        </mask>
        <g mask="url(#faceRound02Mask0)">
            <rect x="-4" y="-2" width="138" height="124" fill="${color.hex}"/>
            ${texture}
        </g>
    `;
            };

        }, {}],
        19: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color, texture) => {
                return `
        <rect width="130" height="120" rx="18" fill="#0076DE"/>
        <mask id="faceSquare01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="130" height="120">
            <rect width="130" height="120" rx="18" fill="white"/>
        </mask>
        <g mask="url(#faceSquare01Mask0)">
            <rect x="-2" y="-2" width="134" height="124" fill="${color.hex}"/>
            ${texture}
        </g>
    `;
            };

        }, {}],
        20: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color, texture) => {
                return `
        <path d="M0 12C0 5.37259 5.37258 0 12 0H118C124.627 0 130 5.37258 130 12V82C130 102.987 112.987 120 92 120H38C17.0132 120 0 102.987 0 82V12Z" fill="#0076DE"/>
        <mask id="faceSquare01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="130" height="120">
            <path d="M0 12C0 5.37259 5.37258 0 12 0H118C124.627 0 130 5.37258 130 12V82C130 102.987 112.987 120 92 120H38C17.0132 120 0 102.987 0 82V12Z" fill="white"/>
        </mask>
        <g mask="url(#faceSquare01Mask0)">
            <rect x="-2" y="-2" width="134" height="124" fill="${color.hex}"/>
            ${texture}
        </g>
    `;
            };

        }, {}],
        21: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color, texture) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M0 18C0 8.05888 8.05888 0 18 0H112C121.941 0 130 8.05888 130 18V45.1483C130 49.6831 129.229 54.1848 127.72 58.4611L110.239 107.991C107.699 115.187 100.896 120 93.2647 120H36.7353C29.1036 120 22.3014 115.187 19.7614 107.991L2.28038 58.4611C0.771117 54.1848 0 49.6831 0 45.1483L0 18Z" fill="#0076DE"/>
        <mask id="faceSquare03Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="130" height="120">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M0 18C0 8.05888 8.05888 0 18 0H112C121.941 0 130 8.05888 130 18V45.1483C130 49.6831 129.229 54.1848 127.72 58.4611L110.239 107.991C107.699 115.187 100.896 120 93.2647 120H36.7353C29.1036 120 22.3014 115.187 19.7614 107.991L2.28038 58.4611C0.771117 54.1848 0 49.6831 0 45.1483L0 18Z" fill="white"/>
        </mask>
        <g mask="url(#faceSquare03Mask0)">
            <rect x="-2" y="-2" width="134" height="124" fill="${color.hex}"/>
            ${texture}
        </g>
    `;
            };

        }, {}],
        22: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color, texture) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M0 102V68.8517C0 64.3169 0.77112 59.8152 2.28039 55.5389L19.7614 12.0092C22.3014 4.81263 29.1036 0 36.7353 0L93.2647 0C100.896 0 107.699 4.81263 110.239 12.0092L127.72 55.5389C129.229 59.8152 130 64.3169 130 68.8517V102C130 111.941 121.941 120 112 120H18C8.05887 120 0 111.941 0 102Z" fill="#0076DE"/>
        <mask id="faceSquareMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="0" y="0" width="130" height="120">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M0 102V68.8517C0 64.3169 0.77112 59.8152 2.28039 55.5389L19.7614 12.0092C22.3014 4.81263 29.1036 0 36.7353 0L93.2647 0C100.896 0 107.699 4.81263 110.239 12.0092L127.72 55.5389C129.229 59.8152 130 64.3169 130 68.8517V102C130 111.941 121.941 120 112 120H18C8.05887 120 0 111.941 0 102Z" fill="white"/>
        </mask>
        <g mask="url(#faceSquareMask0)">
            <rect x="-2" y="-2" width="134" height="124" fill="${color.hex}"/>
            ${texture}
        </g>
    `;
            };

        }, {}],
        23: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const color_1 = __importDefault(require("@dicebear/avatars/lib/color"));
            const eyes_1 = __importDefault(require("./eyes"));
            const face_1 = __importDefault(require("./face"));
            const mouth_1 = __importDefault(require("./mouth"));
            const sides_1 = __importDefault(require("./sides"));
            const texture_1 = __importDefault(require("./texture"));
            const top_1 = __importDefault(require("./top"));
            const group = (random, content, chance, x, y) => {
                if (random.bool(chance)) {
                    return `<g transform="translate(${x}, ${y})">${content}</g>`;
                }
                return '';
            };

            function default_1(random, options = {}) {
                options = Object.assign({
                    primaryColorLevel: 600,
                    secondaryColorLevel: 400,
                    mouthChance: 100,
                    sidesChance: 100,
                    textureChance: 50,
                    topChance: 100
                }, options);
                let colorsCollection = [];
                Object.keys(color_1.default.collection).forEach((color) => {
                    if (options.colors === undefined || options.colors.length === 0 || options.colors.indexOf(color) !== -1) {
                        colorsCollection.push(color_1.default.collection[color]);
                    }
                });
                let primaryColorCollection = random.pickone(colorsCollection);
                let secondaryColorCollection = random.pickone(colorsCollection);
                let primaryColor = new color_1.default(primaryColorCollection[options.primaryColorLevel]);
                let secondaryColor = new color_1.default(primaryColorCollection[options.secondaryColorLevel]);
                if (options.colorful) {
                    secondaryColor = new color_1.default(secondaryColorCollection[options.secondaryColorLevel]);
                }
                let eyes = random.pickone(eyes_1.default);
                let face = random.pickone(face_1.default);
                let mouth = random.pickone(mouth_1.default);
                let sides = random.pickone(sides_1.default);
                let texture = random.pickone(texture_1.default);
                let top = random.pickone(top_1.default);
                // prettier-ignore
                return [
                    '<svg viewBox="0 0 180 180" xmlns="http://www.w3.org/2000/svg" fill="none">',
                    group(random, sides(secondaryColor), options.sidesChance, 0, 66),
                    group(random, top(secondaryColor), options.topChance, 41, 0),
                    group(random, face(primaryColor, random.bool(options.textureChance) ? texture() : undefined), 100, 25, 44),
                    group(random, mouth(), options.mouthChance, 52, 124),
                    group(random, eyes(), 100, 38, 76),
                    '</svg>'
                ].join('');
            }

            exports.default = default_1;

        }, {
            "./eyes": 10,
            "./face": 16,
            "./mouth": 29,
            "./sides": 38,
            "./texture": 47,
            "./top": 55,
            "@dicebear/avatars/lib/color": 79
        }],
        24: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="4" y="5" width="68" height="22" rx="5" fill="black" fill-opacity="0.2"/>
    <rect x="8" y="9" width="60" height="14" rx="2" fill="black" fill-opacity="0.6"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M20 17L26 9H14L20 17Z" fill="#E1E6E8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M32 17L38 9H26L32 17Z" fill="#E1E6E8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M44 17L50 9H38L44 17Z" fill="#E1E6E8"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M56 17L62 9H50L56 17Z" fill="#E1E6E8"/>
`;

        }, {}],
        25: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="4" y="4" width="68" height="24" rx="5" fill="black" fill-opacity="0.2"/>
    <rect x="8" y="8" width="60" height="16" rx="2" fill="black" fill-opacity="0.8"/>
    <path d="M9 17H20L22 13L25 20L29 12L31 21L34 10L37 20L40 17H55L58 13L60 20L63 17H67" stroke="#4EFAC9" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
`;

        }, {}],
        26: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="12" y="12" width="4" height="8" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="36" y="12" width="4" height="8" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="24" y="12" width="4" height="8" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="48" y="12" width="4" height="8" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="60" y="12" width="4" height="8" rx="2" fill="black" fill-opacity="0.6"/>
`;

        }, {}],
        27: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="28" y="10" width="6" height="14" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="14" y="10" width="6" height="14" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="42" y="10" width="6" height="14" rx="2" fill="black" fill-opacity="0.6"/>
    <rect x="56" y="10" width="6" height="14" rx="2" fill="black" fill-opacity="0.6"/>
`;

        }, {}],
        28: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="4" y="5" width="68" height="22" rx="5" fill="black" fill-opacity="0.2"/>
    <rect x="8" y="9" width="60" height="14" rx="2" fill="white"/>
    <rect x="18" y="9" width="4" height="14" fill="black" fill-opacity="0.1"/>
    <rect x="42" y="9" width="4" height="14" fill="black" fill-opacity="0.1"/>
    <rect x="30" y="9" width="4" height="14" fill="black" fill-opacity="0.1"/>
    <rect x="54" y="9" width="4" height="14" fill="black" fill-opacity="0.1"/>
`;

        }, {}],
        29: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const bite_1 = __importDefault(require("./bite"));
            const diagram_1 = __importDefault(require("./diagram"));
            const grill_01_1 = __importDefault(require("./grill-01"));
            const grill_02_1 = __importDefault(require("./grill-02"));
            const grill_03_1 = __importDefault(require("./grill-03"));
            const smile_01_1 = __importDefault(require("./smile-01"));
            const smile_02_1 = __importDefault(require("./smile-02"));
            const square_01_1 = __importDefault(require("./square-01"));
            const square_02_1 = __importDefault(require("./square-02"));
            exports.default = [bite_1.default, diagram_1.default, grill_01_1.default, grill_02_1.default, grill_03_1.default, smile_01_1.default, smile_02_1.default, square_01_1.default, square_02_1.default];

        }, {
            "./bite": 24,
            "./diagram": 25,
            "./grill-01": 26,
            "./grill-02": 27,
            "./grill-03": 28,
            "./smile-01": 30,
            "./smile-02": 31,
            "./square-01": 32,
            "./square-02": 33
        }],
        30: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path d="M27.0493 8.44151C26.8055 7.36419 27.4811 6.29318 28.5584 6.04935C29.6358 5.80551 30.7068 6.48119 30.9506 7.55851C31.72 10.9578 34.4016 13 37.9999 13C41.5983 13 44.2799 10.9578 45.0493 7.55851C45.2931 6.48119 46.3641 5.80551 47.4414 6.04935C48.5188 6.29318 49.1944 7.36419 48.9506 8.44151C47.7599 13.7024 43.4298 17 37.9999 17C32.5701 17 28.24 13.7024 27.0493 8.44151Z" fill="black" fill-opacity="0.6"/>
`;

        }, {}],
        31: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M18 10.2222C18 21.7845 24.4741 28 38 28C51.5182 28 58 21.6615 58 10.2222C58 9.49622 57.1739 8 55 8C39.2707 8 29.1917 8 21 8C18.949 8 18 9.38479 18 10.2222Z" fill="black" fill-opacity="0.8"/>
    <mask id="mouthSmilie02Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="18" y="8" width="40" height="20">
        <path fill-rule="evenodd" clip-rule="evenodd" d="M18 10.2222C18 21.7845 24.4741 28 38 28C51.5182 28 58 21.6615 58 10.2222C58 9.49622 57.1739 8 55 8C39.2707 8 29.1917 8 21 8C18.949 8 18 9.38479 18 10.2222Z" fill="white"/>
    </mask>
    <g mask="url(#mouthSmilie02Mask0)">
        <rect x="30" y="2" width="16" height="14" rx="2" fill="white"/>
    </g>
`;

        }, {}],
        32: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="24" y="6" width="27" height="8" rx="4" fill="black" fill-opacity="0.8"/>
`;

        }, {}],
        33: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <rect x="16" y="8" width="44" height="4" rx="2" fill="black" fill-opacity="0.8"/>
`;

        }, {}],
        34: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M13 11H11V31H12.4C10.1598 31 9.03969 31 8.18404 31.436C7.43139 31.8195 6.81947 32.4314 6.43597 33.184C6 34.0397 6 35.1598 6 37.4V38.6C6 40.8402 6 41.9603 6.43597 42.816C6.81947 43.5686 7.43139 44.1805 8.18404 44.564C9.03969 45 10.1598 45 12.4 45H18V55.6C18 57.8402 18 58.9603 18.436 59.816C18.8195 60.5686 19.4314 61.1805 20.184 61.564C21.0397 62 22.1598 62 24.4 62H47.6C49.8402 62 50.9603 62 51.816 61.564C52.5686 61.1805 53.1805 60.5686 53.564 59.816C54 58.9603 54 57.8402 54 55.6V20.4C54 18.1598 54 17.0397 53.564 16.184C53.1805 15.4314 52.5686 14.8195 51.816 14.436C50.9603 14 49.8402 14 47.6 14H24.4C22.1598 14 21.0397 14 20.184 14.436C19.4314 14.8195 18.8195 15.4314 18.436 16.184C18 17.0397 18 18.1598 18 20.4V31H13V11ZM126 34.4C126 32.1598 126 31.0397 126.436 30.184C126.819 29.4314 127.431 28.8195 128.184 28.436C129.04 28 130.16 28 132.4 28H155.6C157.84 28 158.96 28 159.816 28.436C160.569 28.8195 161.181 29.4314 161.564 30.184C162 31.0397 162 32.1598 162 34.4V45.6C162 47.8402 162 48.9603 161.564 49.816C161.181 50.5686 160.569 51.1805 159.816 51.564C158.96 52 157.84 52 155.6 52H132.4C130.16 52 129.04 52 128.184 51.564C127.431 51.1805 126.819 50.5686 126.436 49.816C126 48.9603 126 47.8402 126 45.6V34.4Z" fill="#0076DE"/>
        <mask id="sidesAntenna01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="6" y="11" width="156" height="51">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M13 11H11V31H12.4C10.1598 31 9.03969 31 8.18404 31.436C7.43139 31.8195 6.81947 32.4314 6.43597 33.184C6 34.0397 6 35.1598 6 37.4V38.6C6 40.8402 6 41.9603 6.43597 42.816C6.81947 43.5686 7.43139 44.1805 8.18404 44.564C9.03969 45 10.1598 45 12.4 45H18V55.6C18 57.8402 18 58.9603 18.436 59.816C18.8195 60.5686 19.4314 61.1805 20.184 61.564C21.0397 62 22.1598 62 24.4 62H47.6C49.8402 62 50.9603 62 51.816 61.564C52.5686 61.1805 53.1805 60.5686 53.564 59.816C54 58.9603 54 57.8402 54 55.6V20.4C54 18.1598 54 17.0397 53.564 16.184C53.1805 15.4314 52.5686 14.8195 51.816 14.436C50.9603 14 49.8402 14 47.6 14H24.4C22.1598 14 21.0397 14 20.184 14.436C19.4314 14.8195 18.8195 15.4314 18.436 16.184C18 17.0397 18 18.1598 18 20.4V31H13V11ZM126 34.4C126 32.1598 126 31.0397 126.436 30.184C126.819 29.4314 127.431 28.8195 128.184 28.436C129.04 28 130.16 28 132.4 28H155.6C157.84 28 158.96 28 159.816 28.436C160.569 28.8195 161.181 29.4314 161.564 30.184C162 31.0397 162 32.1598 162 34.4V45.6C162 47.8402 162 48.9603 161.564 49.816C161.181 50.5686 160.569 51.1805 159.816 51.564C158.96 52 157.84 52 155.6 52H132.4C130.16 52 129.04 52 128.184 51.564C127.431 51.1805 126.819 50.5686 126.436 49.816C126 48.9603 126 47.8402 126 45.6V34.4Z" fill="white"/>
        </mask>
        <g mask="url(#sidesAntenna01Mask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
            <rect y="38" width="180" height="38" fill="black" fill-opacity="0.1"/>
        </g>
        <rect x="11" y="11" width="2" height="20" fill="white" fill-opacity="0.4"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M12 12C14.2091 12 16 10.2091 16 8C16 5.79086 14.2091 4 12 4C9.79086 4 8 5.79086 8 8C8 10.2091 9.79086 12 12 12Z" fill="#FFEA8F"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M13 9C14.1046 9 15 8.10457 15 7C15 5.89543 14.1046 5 13 5C11.8954 5 11 5.89543 11 7C11 8.10457 11.8954 9 13 9Z" fill="white"/>
    `;
            };

        }, {}],
        35: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M160 9H162V28H163.6C165.84 28 166.96 28 167.816 28.436C168.569 28.8195 169.181 29.4314 169.564 30.184C170 31.0397 170 32.1598 170 34.4V53.6C170 55.8402 170 56.9603 169.564 57.816C169.181 58.5686 168.569 59.1805 167.816 59.564C166.96 60 165.84 60 163.6 60H140.4C138.16 60 137.04 60 136.184 59.564C135.431 59.1805 134.819 58.5686 134.436 57.816C134 56.9603 134 55.8402 134 53.6V34.4C134 32.1598 134 31.0397 134.436 30.184C134.819 29.4314 135.431 28.8195 136.184 28.436C137.04 28 138.16 28 140.4 28H160V9ZM10 34.4C10 32.1598 10 31.0397 10.436 30.184C10.8195 29.4314 11.4314 28.8195 12.184 28.436C13.0397 28 14.1598 28 16.4 28H39.6C41.8402 28 42.9603 28 43.816 28.436C44.5686 28.8195 45.1805 29.4314 45.564 30.184C46 31.0397 46 32.1598 46 34.4V53.6C46 55.8402 46 56.9603 45.564 57.816C45.1805 58.5686 44.5686 59.1805 43.816 59.564C42.9603 60 41.8402 60 39.6 60H16.4C14.1598 60 13.0397 60 12.184 59.564C11.4314 59.1805 10.8195 58.5686 10.436 57.816C10 56.9603 10 55.8402 10 53.6V34.4Z" fill="#0076DE"/>
        <mask id="sidesAntenna02Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="10" y="9" width="160" height="51">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M160 9H162V28H163.6C165.84 28 166.96 28 167.816 28.436C168.569 28.8195 169.181 29.4314 169.564 30.184C170 31.0397 170 32.1598 170 34.4V53.6C170 55.8402 170 56.9603 169.564 57.816C169.181 58.5686 168.569 59.1805 167.816 59.564C166.96 60 165.84 60 163.6 60H140.4C138.16 60 137.04 60 136.184 59.564C135.431 59.1805 134.819 58.5686 134.436 57.816C134 56.9603 134 55.8402 134 53.6V34.4C134 32.1598 134 31.0397 134.436 30.184C134.819 29.4314 135.431 28.8195 136.184 28.436C137.04 28 138.16 28 140.4 28H160V9ZM10 34.4C10 32.1598 10 31.0397 10.436 30.184C10.8195 29.4314 11.4314 28.8195 12.184 28.436C13.0397 28 14.1598 28 16.4 28H39.6C41.8402 28 42.9603 28 43.816 28.436C44.5686 28.8195 45.1805 29.4314 45.564 30.184C46 31.0397 46 32.1598 46 34.4V53.6C46 55.8402 46 56.9603 45.564 57.816C45.1805 58.5686 44.5686 59.1805 43.816 59.564C42.9603 60 41.8402 60 39.6 60H16.4C14.1598 60 13.0397 60 12.184 59.564C11.4314 59.1805 10.8195 58.5686 10.436 57.816C10 56.9603 10 55.8402 10 53.6V34.4Z" fill="white"/>
        </mask>
        <g mask="url(#sidesAntenna02Mask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
            <rect y="38" width="180" height="38" fill="black" fill-opacity="0.1"/>
        </g>
        <rect x="160" y="8" width="2" height="20" fill="white" fill-opacity="0.4"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M161 9C163.209 9 165 7.20914 165 5C165 2.79086 163.209 1 161 1C158.791 1 157 2.79086 157 5C157 7.20914 158.791 9 161 9Z" fill="#E1E6E8"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M162 6C163.105 6 164 5.10457 164 4C164 2.89543 163.105 2 162 2C160.895 2 160 2.89543 160 4C160 5.10457 160.895 6 162 6Z" fill="white"/>
    `;
            };

        }, {}],
        36: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <g opacity="0.9">
            <path id="Cable" d="M38 12C35.046 23.6966 18.0959 18.6663 14.6313 30.009C11.1668 41.3518 22.6565 50 32.1552 50" stroke="#2A3544" stroke-width="6"/>
            <path id="Cable_2" d="M150 55C158.394 58.4864 170.102 47.4063 166 38.5C161.898 29.5936 150 31.8056 150 19.195" stroke="#2A3544" stroke-width="4"/>
        </g>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M138 6C136.895 6 136 6.89543 136 8V22C136 23.1046 136.895 24 138 24H157C158.105 24 159 23.1046 159 22V8C159 6.89543 158.105 6 157 6H138ZM21 37C21 35.8954 21.8954 35 23 35H35C36.1046 35 37 35.8954 37 37V55C37 56.1046 36.1046 57 35 57H23C21.8954 57 21 56.1046 21 55V37ZM136 44C136 42.8954 136.895 42 138 42H157C158.105 42 159 42.8954 159 44V62C159 63.1046 158.105 64 157 64H138C136.895 64 136 63.1046 136 62V44Z" fill="#273951"/>
        <mask id="sidesCables01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="21" y="6" width="138" height="58">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M138 6C136.895 6 136 6.89543 136 8V22C136 23.1046 136.895 24 138 24H157C158.105 24 159 23.1046 159 22V8C159 6.89543 158.105 6 157 6H138ZM21 37C21 35.8954 21.8954 35 23 35H35C36.1046 35 37 35.8954 37 37V55C37 56.1046 36.1046 57 35 57H23C21.8954 57 21 56.1046 21 55V37ZM136 44C136 42.8954 136.895 42 138 42H157C158.105 42 159 42.8954 159 44V62C159 63.1046 158.105 64 157 64H138C136.895 64 136 63.1046 136 62V44Z" fill="white"/>
        </mask>
        <g mask="url(#sidesCables01Mask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
        </g>
    `;
            };

        }, {}],
        37: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <g opacity="0.9">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M32.5 41C41.6127 41 49 32.9411 49 23C49 13.0589 41.6127 5 32.5 5C23.3873 5 16 13.0589 16 23C16 32.9411 23.3873 41 32.5 41Z" stroke="#2A3544" stroke-width="6"/>
            <path d="M29.5152 36.7649C22.1017 41.0451 12.5101 38.3112 8.0918 30.6585" stroke="#2A3544" stroke-width="4"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M28 67C37.3888 67 45 60.5081 45 52.5C45 44.4919 37.3888 38 28 38C18.6112 38 12 44.4919 12 52.5C12 60.5081 18.6112 67 28 67Z" stroke="#2A3544" stroke-width="4"/>
            <path d="M168.606 60.4234C164.326 53.0099 154.653 50.5817 147 55" stroke="#2A3544" stroke-width="4"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M148 38C157.389 38 165 31.0604 165 22.5C165 13.9396 157.389 7 148 7C138.611 7 132 13.9396 132 22.5C132 31.0604 138.611 38 148 38Z" stroke="#2A3544" stroke-width="6"/>
        </g>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M145 0C143.895 0 143 0.89543 143 2V20C143 21.1046 143.895 22 145 22H157C158.105 22 159 21.1046 159 20V2C159 0.895431 158.105 0 157 0H145ZM23 27C21.8954 27 21 27.8954 21 29V47C21 48.1046 21.8954 49 23 49H35C36.1046 49 37 48.1046 37 47V29C37 27.8954 36.1046 27 35 27H23ZM24 60C22.8954 60 22 60.8954 22 62V70C22 71.1046 22.8954 72 24 72H36C37.1046 72 38 71.1046 38 70V62C38 60.8954 37.1046 60 36 60H24ZM143 44C143 42.8954 143.895 42 145 42H157C158.105 42 159 42.8954 159 44V62C159 63.1046 158.105 64 157 64H145C143.895 64 143 63.1046 143 62V44Z" fill="#273951"/>
        <mask id="sidesCables01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="21" y="0" width="138" height="72">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M145 0C143.895 0 143 0.89543 143 2V20C143 21.1046 143.895 22 145 22H157C158.105 22 159 21.1046 159 20V2C159 0.895431 158.105 0 157 0H145ZM23 27C21.8954 27 21 27.8954 21 29V47C21 48.1046 21.8954 49 23 49H35C36.1046 49 37 48.1046 37 47V29C37 27.8954 36.1046 27 35 27H23ZM24 60C22.8954 60 22 60.8954 22 62V70C22 71.1046 22.8954 72 24 72H36C37.1046 72 38 71.1046 38 70V62C38 60.8954 37.1046 60 36 60H24ZM143 44C143 42.8954 143.895 42 145 42H157C158.105 42 159 42.8954 159 44V62C159 63.1046 158.105 64 157 64H145C143.895 64 143 63.1046 143 62V44Z" fill="white"/>
        </mask>
        <g mask="url(#sidesCables01Mask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
        </g>
    `;
            };

        }, {}],
        38: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const antenna_01_1 = __importDefault(require("./antenna-01"));
            const antenna_02_1 = __importDefault(require("./antenna-02"));
            const cables_01_1 = __importDefault(require("./cables-01"));
            const cables_02_1 = __importDefault(require("./cables-02"));
            const round_1 = __importDefault(require("./round"));
            const square_assymetric_1 = __importDefault(require("./square-assymetric"));
            const square_1 = __importDefault(require("./square"));
            exports.default = [antenna_01_1.default, antenna_02_1.default, cables_01_1.default, cables_02_1.default, round_1.default, square_assymetric_1.default, square_1.default];

        }, {
            "./antenna-01": 34,
            "./antenna-02": 35,
            "./cables-01": 36,
            "./cables-02": 37,
            "./round": 39,
            "./square": 41,
            "./square-assymetric": 40
        }],
        39: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M12 39C12 50.9264 17.9543 61 28 61C38.0457 61 48 50.9264 48 39C48 26.0736 38.0457 16 28 16C17.9543 16 12 26.0736 12 39ZM168 39C168 50.9264 162.046 61 152 61C141.954 61 132 50.9264 132 39C132 26.0736 141.954 16 152 16C162.046 16 168 26.0736 168 39Z" fill="#E1E6E8"/>
        <mask id="sidesRoundMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="12" y="16" width="156" height="45">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M12 39C12 50.9264 17.9543 61 28 61C38.0457 61 48 50.9264 48 39C48 26.0736 38.0457 16 28 16C17.9543 16 12 26.0736 12 39ZM168 39C168 50.9264 162.046 61 152 61C141.954 61 132 50.9264 132 39C132 26.0736 141.954 16 152 16C162.046 16 168 26.0736 168 39Z" fill="white"/>
        </mask>
        <g mask="url(#sidesRoundMask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
            <rect x="20" width="140" height="76" fill="black" fill-opacity="0.2"/>
        </g>
    `;
            };

        }, {}],
        40: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M134.436 10.184C134 11.0397 134 12.1598 134 14.4V61.6C134 63.8402 134 64.9603 134.436 65.816C134.819 66.5686 135.431 67.1805 136.184 67.564C137.04 68 138.16 68 140.4 68H163.6C165.84 68 166.96 68 167.816 67.564C168.569 67.1805 169.181 66.5686 169.564 65.816C170 64.9603 170 63.8402 170 61.6V52.9944C171.35 52.9761 172.161 52.8979 172.816 52.564C173.569 52.1805 174.181 51.5686 174.564 50.816C175 49.9603 175 48.8402 175 46.6V29.4C175 27.1598 175 26.0397 174.564 25.184C174.181 24.4314 173.569 23.8195 172.816 23.436C172.161 23.1021 171.35 23.0239 170 23.0056V14.4C170 12.1598 170 11.0397 169.564 10.184C169.181 9.43139 168.569 8.81947 167.816 8.43597C166.96 8 165.84 8 163.6 8H140.4C138.16 8 137.04 8 136.184 8.43597C135.431 8.81947 134.819 9.43139 134.436 10.184ZM20.436 17.184C20 18.0397 20 19.1598 20 21.4V31H16.4C14.1598 31 13.0397 31 12.184 31.436C11.4314 31.8195 10.8195 32.4314 10.436 33.184C10 34.0397 10 35.1598 10 37.4V54.6C10 56.8402 10 57.9603 10.436 58.816C10.8195 59.5686 11.4314 60.1805 12.184 60.564C13.0397 61 14.1598 61 16.4 61H39.6C41.8402 61 42.9603 61 43.816 60.564C44.5686 60.1805 45.1805 59.5686 45.564 58.816C46 57.9603 46 56.8402 46 54.6V38.6V37.4V21.4C46 19.1598 46 18.0397 45.564 17.184C45.1805 16.4314 44.5686 15.8195 43.816 15.436C42.9603 15 41.8402 15 39.6 15H26.4C24.1598 15 23.0397 15 22.184 15.436C21.4314 15.8195 20.8195 16.4314 20.436 17.184Z" fill="#0076DE"/>
        <mask id="sidesSquareAssymetricMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="10" y="8" width="165" height="60">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M134.436 10.184C134 11.0397 134 12.1598 134 14.4V61.6C134 63.8402 134 64.9603 134.436 65.816C134.819 66.5686 135.431 67.1805 136.184 67.564C137.04 68 138.16 68 140.4 68H163.6C165.84 68 166.96 68 167.816 67.564C168.569 67.1805 169.181 66.5686 169.564 65.816C170 64.9603 170 63.8402 170 61.6V52.9944C171.35 52.9761 172.161 52.8979 172.816 52.564C173.569 52.1805 174.181 51.5686 174.564 50.816C175 49.9603 175 48.8402 175 46.6V29.4C175 27.1598 175 26.0397 174.564 25.184C174.181 24.4314 173.569 23.8195 172.816 23.436C172.161 23.1021 171.35 23.0239 170 23.0056V14.4C170 12.1598 170 11.0397 169.564 10.184C169.181 9.43139 168.569 8.81947 167.816 8.43597C166.96 8 165.84 8 163.6 8H140.4C138.16 8 137.04 8 136.184 8.43597C135.431 8.81947 134.819 9.43139 134.436 10.184ZM20.436 17.184C20 18.0397 20 19.1598 20 21.4V31H16.4C14.1598 31 13.0397 31 12.184 31.436C11.4314 31.8195 10.8195 32.4314 10.436 33.184C10 34.0397 10 35.1598 10 37.4V54.6C10 56.8402 10 57.9603 10.436 58.816C10.8195 59.5686 11.4314 60.1805 12.184 60.564C13.0397 61 14.1598 61 16.4 61H39.6C41.8402 61 42.9603 61 43.816 60.564C44.5686 60.1805 45.1805 59.5686 45.564 58.816C46 57.9603 46 56.8402 46 54.6V38.6V37.4V21.4C46 19.1598 46 18.0397 45.564 17.184C45.1805 16.4314 44.5686 15.8195 43.816 15.436C42.9603 15 41.8402 15 39.6 15H26.4C24.1598 15 23.0397 15 22.184 15.436C21.4314 15.8195 20.8195 16.4314 20.436 17.184Z" fill="white"/>
        </mask>
        <g mask="url(#sidesSquareAssymetricMask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
            <rect y="47" width="180" height="29" fill="black" fill-opacity="0.1"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M161 25C163.761 25 166 22.7614 166 20C166 17.2386 163.761 15 161 15C158.239 15 156 17.2386 156 20C156 22.7614 158.239 25 161 25Z" fill="white" fill-opacity="0.6"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M161 41C163.761 41 166 38.7614 166 36C166 33.2386 163.761 31 161 31C158.239 31 156 33.2386 156 36C156 38.7614 158.239 41 161 41Z" fill="white" fill-opacity="0.6"/>
        </g>
    `;
            };

        }, {}],
        41: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M14.9809 20.9141C14 22.8393 14 25.3595 14 30.4V45.6C14 50.6405 14 53.1607 14.9809 55.0859C15.8438 56.7794 17.2206 58.1562 18.9141 59.0191C20.8393 60 23.3595 60 28.4 60H35.6C40.6405 60 43.1607 60 45.0859 59.0191C46.7794 58.1562 48.1562 56.7794 49.0191 55.0859C50 53.1607 50 50.6405 50 45.6V30.4C50 25.3595 50 22.8393 49.0191 20.9141C48.1562 19.2206 46.7794 17.8438 45.0859 16.9809C43.1607 16 40.6405 16 35.6 16H28.4C23.3595 16 20.8393 16 18.9141 16.9809C17.2206 17.8438 15.8438 19.2206 14.9809 20.9141ZM130.981 20.9141C130 22.8393 130 25.3595 130 30.4V45.6C130 50.6405 130 53.1607 130.981 55.0859C131.844 56.7794 133.221 58.1562 134.914 59.0191C136.839 60 139.36 60 144.4 60H151.6C156.64 60 159.161 60 161.086 59.0191C162.779 58.1562 164.156 56.7794 165.019 55.0859C166 53.1607 166 50.6405 166 45.6V30.4C166 25.3595 166 22.8393 165.019 20.9141C164.156 19.2206 162.779 17.8438 161.086 16.9809C159.161 16 156.64 16 151.6 16H144.4C139.36 16 136.839 16 134.914 16.9809C133.221 17.8438 131.844 19.2206 130.981 20.9141Z" fill="#0076DE"/>
        <mask id="sidesSquareMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="14" y="16" width="152" height="44">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M14.9809 20.9141C14 22.8393 14 25.3595 14 30.4V45.6C14 50.6405 14 53.1607 14.9809 55.0859C15.8438 56.7794 17.2206 58.1562 18.9141 59.0191C20.8393 60 23.3595 60 28.4 60H35.6C40.6405 60 43.1607 60 45.0859 59.0191C46.7794 58.1562 48.1562 56.7794 49.0191 55.0859C50 53.1607 50 50.6405 50 45.6V30.4C50 25.3595 50 22.8393 49.0191 20.9141C48.1562 19.2206 46.7794 17.8438 45.0859 16.9809C43.1607 16 40.6405 16 35.6 16H28.4C23.3595 16 20.8393 16 18.9141 16.9809C17.2206 17.8438 15.8438 19.2206 14.9809 20.9141ZM130.981 20.9141C130 22.8393 130 25.3595 130 30.4V45.6C130 50.6405 130 53.1607 130.981 55.0859C131.844 56.7794 133.221 58.1562 134.914 59.0191C136.839 60 139.36 60 144.4 60H151.6C156.64 60 159.161 60 161.086 59.0191C162.779 58.1562 164.156 56.7794 165.019 55.0859C166 53.1607 166 50.6405 166 45.6V30.4C166 25.3595 166 22.8393 165.019 20.9141C164.156 19.2206 162.779 17.8438 161.086 16.9809C159.161 16 156.64 16 151.6 16H144.4C139.36 16 136.839 16 134.914 16.9809C133.221 17.8438 131.844 19.2206 130.981 20.9141Z" fill="white"/>
        </mask>
        <g mask="url(#sidesSquareMask0)">
            <rect width="180" height="76" fill="${color.hex}"/>
            <rect y="38" width="180" height="38" fill="black" fill-opacity="0.1"/>
        </g>
    `;
            };

        }, {}],
        42: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M141.942 1.99261C142.885 1.84335 143.725 1.51627 144.462 1.01137C144.641 0.98607 144.904 0.945398 145.233 0.894564C147.954 0.474318 155.171 -0.640471 156.477 0.495865C157.33 1.23727 156.947 3.87556 156.671 5.77254C156.566 6.49296 156.477 7.10646 156.477 7.46854C155.149 7.5314 154.142 6.06775 153.157 4.63726C151.696 2.51506 150.285 0.465817 147.954 3.58203C146.86 5.0429 147.113 6.88681 147.369 8.76079C147.639 10.7369 147.914 12.7464 146.621 14.3756C145.083 16.3133 142.629 16.5822 140.203 16.8479C137.635 17.1292 135.099 17.407 133.718 19.6598C133.321 20.3087 133.213 21.4585 133.096 22.6996C132.894 24.853 132.665 27.2809 130.854 27.8425C127.728 28.8118 127.528 26.1525 127.354 23.8314C127.26 22.5851 127.174 21.4363 126.646 20.9991C120.957 16.2901 120.42 25.3692 120.298 27.4294L120.296 27.4631C120.271 27.8889 120.348 28.2362 120.419 28.5543C120.546 29.1316 120.653 29.613 120.089 30.294C119.794 30.6507 119.391 30.5886 118.966 30.5232C118.536 30.457 118.084 30.3875 117.7 30.7457C116.192 32.1554 116.162 32.2901 115.816 33.8601L115.816 33.8608C115.791 33.9753 115.764 34.0974 115.735 34.2282C115.615 34.7621 115.545 35.2168 115.481 35.6306C115.335 36.5695 115.222 37.2985 114.644 38.2669C114.374 38.7177 114.011 39.1584 113.649 39.5982C112.448 41.0572 111.253 42.5073 113.505 44.2879C112.887 44.2948 107.505 45.7455 107.146 45.9333C106.626 46.2048 106.028 46.9103 105.493 47.5421L105.493 47.5422L105.493 47.5422C105.022 48.0979 104.6 48.5966 104.321 48.6932C101.127 49.7984 101.35 47.4119 101.579 44.9531C101.717 43.4742 101.857 41.9692 101.258 41.1821C109.541 40.9056 114.519 32.9679 115.01 26.8047C115.085 25.8635 114.714 25.0566 114.365 24.2957C114.084 23.6849 113.817 23.1038 113.805 22.5066C113.788 21.666 114.058 21.0964 114.325 20.5307C114.546 20.0635 114.766 19.599 114.822 18.9869C115.064 16.3513 113.98 15.3767 112.564 14.1039C111.851 13.463 111.054 12.7465 110.299 11.7044C107.796 8.24547 109.106 7.71356 112.894 8.33694C115.071 8.6954 117.164 9.33292 119.257 9.97051C121.819 10.751 124.381 11.5317 127.1 11.8006C130.579 12.1448 133.812 11.9759 135.458 9.14259C136.409 7.50701 136.825 3.01028 133.659 2.65738C134.452 2.53254 135.103 2.27789 135.768 2.01789C136.634 1.67892 137.524 1.33084 138.783 1.24937C139.787 1.64057 140.841 1.88849 141.942 1.99261ZM147.694 85.8747C147.695 85.8778 147.697 85.8803 147.699 85.882C147.701 85.8838 147.703 85.8847 147.705 85.8845C147.703 85.8827 147.701 85.8812 147.7 85.8797L147.694 85.8747ZM56.3794 92.4655C56.4811 92.5083 56.5948 92.5071 56.721 92.462C56.536 92.5988 56.2611 92.4881 56.3506 92.4659L56.3492 92.4666C56.3588 92.4663 56.3697 92.4661 56.3801 92.466L56.3794 92.4655ZM56.3728 92.4628C56.3659 92.461 56.3586 92.4621 56.3506 92.4659C56.3564 92.4645 56.3638 92.4634 56.3728 92.4628ZM155.437 132.487C155.1 132.461 155.106 132.452 155.449 132.426C155.509 132.149 155.604 131.857 155.696 131.577C155.913 130.913 156.106 130.323 155.73 130.19C156.861 129.97 156.65 131.987 156.529 133.145L156.529 133.146C156.501 133.417 156.477 133.642 156.477 133.779C155.417 133.784 155.298 133.193 155.437 132.487ZM155.516 131.785L155.514 131.765L155.513 131.74C155.083 131.771 155.084 131.779 155.517 131.812L155.516 131.785ZM65.6526 0.495865C66.0369 0.573975 67.1487 0.678519 67.5157 0.495865L65.6526 0.495865ZM63.0135 57.4794C61.9497 57.2283 61.5655 56.5609 61.277 56.0596C61.1157 55.7795 60.9843 55.5512 60.7808 55.4764C57.9475 54.4346 57.4941 56.3957 57.007 59.4631C56.9192 60.0161 56.8986 60.4945 56.8798 60.9308V60.9308C56.8335 62.0062 56.7983 62.8255 55.7943 63.8743C55.2238 64.4704 54.2833 64.7838 53.3522 65.0941L53.3521 65.0941C52.6756 65.3195 52.0041 65.5433 51.4829 65.8727C50.0117 66.8023 49.3589 67.6545 48.5606 68.6964C48.2439 69.1098 47.9044 69.5531 47.4817 70.0429C46.522 71.155 45.1183 72.2251 43.6615 73.3357C39.7513 76.3166 35.4585 79.5892 38.3404 84.7486C41.4364 90.2917 42.6629 85.6058 43.6093 81.9901C43.9724 80.6031 44.2942 79.3737 44.6644 78.9394C49.3258 73.474 50.3382 78.8154 50.9456 82.0196C51.0448 82.543 51.1332 83.0094 51.2249 83.3623C52.533 88.3941 54.3224 84.7674 55.7974 81.7779C56.2415 80.8778 56.657 80.0355 57.0224 79.5047C57.4606 78.868 58.0086 78.4039 58.5419 77.9521L58.542 77.952C59.2021 77.3929 59.8397 76.8528 60.2192 76.0282C60.4582 75.5086 60.3414 74.859 60.2213 74.1915C60.0232 73.0896 59.8162 71.9388 61.1869 71.2437C62.6392 70.5071 64.3071 71.4035 65.7928 72.2019C66.9191 72.8073 67.9408 73.3563 68.6844 73.095C71.0229 72.2734 68.9195 69.4984 67.2364 67.2779L67.2364 67.2778L67.2362 67.2777L67.236 67.2774L67.236 67.2774L67.2359 67.2772L67.2357 67.2771C66.4623 66.2566 65.7777 65.3534 65.6541 64.8108C65.2942 63.232 65.928 61.8627 66.5614 60.4942C67.2026 59.1089 67.8435 57.7243 67.453 56.1239C66.7978 56.2138 66.2286 56.5263 65.6648 56.8359C64.8508 57.2828 64.048 57.7236 63.0135 57.4794ZM104.606 94.0652C104.521 94.0659 104.439 94.0681 104.362 94.0737C104.444 94.0811 104.526 94.0773 104.606 94.0652ZM119.885 78.7001C120.268 76.0392 118.164 74.6787 115.986 73.27C115.297 72.8242 114.6 72.3736 113.973 71.8753C112.544 70.7421 111.577 69.7957 110.849 68.2058C110.725 67.9342 110.701 67.4777 110.677 67.0051C110.641 66.3205 110.604 65.602 110.257 65.3626C108.411 64.0887 107.376 64.9653 106.133 66.0183C105.538 66.5227 104.895 67.0677 104.092 67.436C101.371 68.6836 101.338 68.6639 99.708 67.7004L99.7079 67.7003L99.707 67.6998L99.7067 67.6996C99.3785 67.5056 98.9857 67.2733 98.4933 67.0055C98.2418 66.8687 98.0045 66.7274 97.7771 66.5919L97.777 66.5919L97.7769 66.5918C96.283 65.7019 95.2184 65.0677 93.3818 67.6662C92.6455 68.7078 92.6838 70.5384 92.7236 72.4392C92.7903 75.6242 92.8611 79.0065 89.2987 79.2054C90.8655 79.8958 90.1598 81.9519 89.4432 84.0394C88.7039 86.1932 87.953 88.3805 89.6749 89.136C93.8527 90.9689 95.0649 82.2797 95.7584 77.3077C95.987 75.6692 96.1593 74.4344 96.3628 74.1129C98.0252 71.4862 103.987 69.2534 105.703 73.1276C106.01 73.821 105.782 74.5679 105.551 75.3259C105.328 76.0601 105.101 76.8047 105.353 77.5214C105.72 78.5632 106.324 78.9442 107.023 79.3859C107.398 79.6226 107.8 79.8768 108.209 80.2596C111.103 82.9712 110.998 83.1372 109.549 85.4425L109.549 85.4426C109.301 85.8379 109.013 86.2962 108.694 86.8409C108.38 87.3761 108.133 88.2236 107.863 89.1534C107.25 91.2572 106.514 93.7826 104.606 94.0653C105.646 94.0579 106.735 94.3117 107.812 94.5625L107.812 94.5625C109.874 95.0425 111.889 95.5119 113.421 94.121C114.488 93.1515 114.306 91.3978 114.132 89.726C114.019 88.6313 113.909 87.5717 114.156 86.7903C114.603 85.3725 115.791 84.1916 116.973 83.0163C118.313 81.683 119.647 80.3568 119.885 78.7001ZM42.891 110.454C42.9281 110.934 42.9706 111.483 43.077 112.15C43.4724 114.627 44.5935 116.042 46.0326 117.859L46.0326 117.859L46.0338 117.861C46.3354 118.241 46.6509 118.64 46.9766 119.069C42.613 119.138 39.1721 123.68 43.9444 126.244C42.0129 125.206 40.5523 126.364 39.365 127.305C37.8721 128.488 36.8112 129.329 35.7897 125.031C35.7152 124.717 35.6162 124.338 35.5111 123.936C35.2841 123.067 35.0287 122.089 34.9298 121.425C34.7172 119.995 34.7543 119.817 34.9188 119.027L34.9188 119.027C34.9472 118.891 34.9793 118.737 35.0145 118.555C35.1041 118.093 35.2819 117.771 35.4461 117.473C35.7958 116.838 36.0835 116.316 35.3241 114.803C34.9285 114.014 34.4408 113.399 33.9779 112.814C33.1349 111.75 32.374 110.789 32.4006 109.08C32.408 108.6 32.6114 108.221 32.8304 107.813C33.1982 107.127 33.6103 106.359 33.2125 104.892C32.9819 104.042 32.5582 103.236 32.1341 102.429C31.625 101.461 31.1154 100.491 30.9394 99.4443C32.7406 99.5204 33.8854 100.892 34.9133 102.125C35.2744 102.557 35.6211 102.973 35.9768 103.309C36.7738 104.062 37.7345 104.595 38.6662 105.111C39.4791 105.562 40.2701 106.001 40.9112 106.563C42.712 108.142 42.775 108.955 42.891 110.454ZM48.6817 97.2964C50.1696 97.42 54.1613 94.1576 54.4184 93.2261C53.134 93.4664 51.8225 92.8698 50.5378 92.2853C49.1009 91.6316 47.6976 90.9931 46.4029 91.5578C44.1095 92.5581 46.55 97.1194 48.6817 97.2964ZM107.04 98.9245C105.406 99.2246 103.553 99.565 102.46 98.9224C101.403 98.3004 99.2975 95.6383 102.358 95.7877C103.136 96.0878 104.033 96.2816 104.918 96.4728C106.791 96.8773 108.61 97.2702 109.132 98.6335C108.558 98.6457 107.824 98.7805 107.04 98.9245ZM86.5906 74.9874C85.7581 74.9377 84.9762 74.788 84.2574 74.6504C82.0964 74.2367 80.5068 73.9324 79.831 76.7827C80.9453 76.7335 82.064 77.027 83.062 77.2888C85.1718 77.8423 86.7422 78.2542 86.5906 74.9874ZM115.566 111.439C114.85 111.078 114.176 110.738 113.642 110.7C114.325 110.592 114.447 109.933 114.563 109.307C114.661 108.781 114.754 108.279 115.172 108.144C115.337 108.186 115.533 108.232 115.748 108.282C117.513 108.693 120.631 109.42 119.864 111.101C118.934 113.138 117.148 112.237 115.566 111.439ZM47.0162 119.153C46.2809 119.191 46.7767 119.354 47.0938 119.393L47.0162 119.153ZM88.4748 71.1251C89.3016 70.5576 89.2754 67.9372 88.0176 67.8187C86.272 68.2097 87.1589 72.0285 88.4748 71.1251ZM153.602 40.4498C153.627 41.0357 156.205 43.5428 156.477 43.5344C156.477 43.4583 156.503 43.2722 156.537 43.0182C156.713 41.7316 157.125 38.7015 155.685 39.3827C155.442 39.4978 155.211 39.6592 154.979 39.8217C154.555 40.1183 154.126 40.4189 153.602 40.4498ZM45.6247 133.859C40.4841 133.859 42.3079 131.474 43.6803 132.209C43.9777 132.369 44.2343 132.645 44.4909 132.922C44.7882 133.242 45.0853 133.562 45.4455 133.7C45.5383 133.736 45.5622 133.772 45.586 133.809C45.5968 133.826 45.6076 133.843 45.6247 133.859ZM23.3002 5.12441C23.3136 4.6879 20.8753 3.932 22.3009 4.90675C22.1554 4.80729 22.2708 4.85501 22.4675 4.93637C22.7792 5.06535 23.2953 5.27884 23.3002 5.12441ZM147.37 49.7508C145.838 49.334 149.993 49.5227 149.474 50.6833C149.049 50.6259 148.68 50.4105 148.31 50.1946C148.007 50.0182 147.704 49.8416 147.37 49.7508Z" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M62.1237 32.7476C62.3573 31.9897 61.9727 31.1524 61.5867 30.3118C61.2669 29.6155 60.9462 28.917 60.9751 28.2599C61.0026 27.6354 61.4545 26.5091 61.8849 25.4362L61.8849 25.4362C62.1866 24.6841 62.4778 23.9584 62.6048 23.4502C62.6422 23.3006 62.681 23.1416 62.7218 22.9747C63.5114 19.7418 65.0135 13.5916 70.4253 16.4533C66.087 14.1593 67.8708 12.3317 69.8575 10.2963C70.6302 9.50457 71.4336 8.68143 71.9194 7.78716C72.3207 7.04849 72.643 6.71322 72.8322 6.51644C72.9437 6.40046 73.0089 6.3326 73.0168 6.2586C73.028 6.15421 72.925 6.03763 72.6767 5.75656C72.5728 5.63897 72.4435 5.49258 72.2864 5.30625C71.6938 4.60312 67.4975 3.65516 66.5006 3.67793C62.6748 3.76576 58.5247 7.38124 56.6026 10.2965C55.5712 11.8605 55.9294 12.7303 56.2708 13.5593C56.4764 14.0587 56.676 14.5433 56.5621 15.1559C56.5155 15.4068 56.5448 15.6587 56.5727 15.8983V15.8983C56.6583 16.633 56.7304 17.2521 54.5574 17.3728C51.8511 17.5233 46.786 13.0385 46.7612 10.8425C46.7531 10.109 47.5405 9.27766 48.3284 8.44581C49.7215 6.97504 51.1161 5.50268 48.1177 4.56682C45.8769 3.86763 39.0886 8.36891 38.6687 10.0767C38.4088 11.134 39.1631 12.2751 39.8246 13.2759C40.8794 14.8714 41.6983 16.1103 37.7954 16.0842C37.8093 16.0551 41.4308 18.3689 41.8033 18.8503C42.6458 19.9395 42.9689 20.7325 43.2457 21.412C43.7229 22.5835 44.0627 23.4178 46.6901 24.8508C47.7914 25.4516 48.9668 25.9715 50.1425 26.4916C51.7207 27.1898 53.2996 27.8883 54.7008 28.7827C55.4581 29.266 56.1036 29.8459 56.7338 30.412C58.2053 31.7338 59.5931 32.9805 62.1237 32.7476ZM151.366 9.50447C149.957 10.2435 150.768 10.8966 151.743 11.6822C152.717 12.4667 153.855 13.3834 153.11 14.6496C152.631 15.4628 151.5 15.6364 150.341 15.8143C149.169 15.9941 147.968 16.1784 147.385 17.0325C146.544 18.2619 146.883 19.1285 147.213 19.9694C147.597 20.9516 147.968 21.8988 146.429 23.3482C145.351 24.3645 144.333 24.6906 143.449 24.9737C142.139 25.3935 141.124 25.7186 140.644 28.0592C139.908 31.6443 141.74 34.0263 143.491 36.3018C145.472 38.8772 147.348 41.3162 145.277 45.209C145.182 45.3881 145.064 45.5936 144.935 45.8193L144.935 45.8198C143.85 47.7171 141.949 51.0423 145.558 52.1573C148.194 52.9717 148.035 52.0413 147.871 51.0752C147.747 50.3488 147.62 49.6023 148.674 49.562C148.455 48.9672 151.649 49.6599 151.254 50.5382C152.095 50.6286 152.503 51.0816 152.9 51.5211C153.471 52.1541 154.016 52.7592 155.79 52.2129C157.564 51.6668 157.847 50.0216 158.08 48.6668L158.098 48.5613C158.448 46.5304 157.51 44.7689 156.606 43.0716C155.854 41.6596 155.125 40.2921 155.182 38.851C155.191 38.6427 155.185 38.4713 155.179 38.3149C155.166 37.9454 155.157 37.6606 155.344 37.1754C155.931 36.7633 156.597 36.606 157.342 36.7034C158.495 36.9433 158.817 36.5077 158.307 35.3968C159.311 33.4728 158.932 30.2193 158.599 27.3533C158.448 26.0559 158.307 24.8379 158.307 23.8585C158.307 23.1814 158.426 22.0492 158.565 20.7237C158.93 17.2542 159.435 12.4604 158.307 11.0383C157.649 10.2091 152.498 8.91125 151.366 9.50447ZM153.582 58.3525C150.607 57.594 150.968 59.9162 151.245 61.7033V61.7034V61.7035L151.245 61.7037L151.245 61.7048L151.246 61.7055C151.326 62.2208 151.399 62.6915 151.382 63.0308C151.242 65.9859 150.899 66.7055 147.229 67.1928C146.351 67.3094 145.512 67.0967 144.684 66.8866C143.289 66.5326 141.923 66.1862 140.447 67.4391C139.617 68.1439 139.655 69.01 139.694 69.8808C139.713 70.3064 139.732 70.7331 139.649 71.1426C139.509 71.8369 138.974 72.9159 138.404 74.0659L138.404 74.066C137.354 76.185 136.184 78.5452 137.144 79.1845C137.793 79.6167 139.667 79.1235 141.802 78.562C145.475 77.5959 149.918 76.4273 150.209 79.4206C150.382 81.1953 148.049 81.4674 145.832 81.7261C144.392 81.8942 143 82.0565 142.376 82.6214C142.189 82.7908 141.826 84.0496 141.465 85.304C141.062 86.7024 140.66 88.0952 140.505 87.9673C141.16 88.5091 142.196 87.7715 143.423 86.8974C145.223 85.6155 147.435 84.0401 149.461 85.7759C146.973 81.0246 154.314 81.6998 157.449 81.9881C157.791 82.0196 158.083 82.0464 158.307 82.0611V66.2114C154.754 67.1446 154.983 65.1336 155.238 62.8959C155.457 60.9768 155.695 58.891 153.582 58.3525ZM155.511 126.554H155.511H155.512H155.512H155.512C156.413 126.604 157.324 126.654 158.307 126.641C158.307 127.459 157.569 127.819 156.828 128.18C156.187 128.493 155.544 128.806 155.373 129.418C154.886 131.155 155.954 131.752 156.935 132.3C157.643 132.697 158.307 133.067 158.307 133.823C157.47 133.823 156.529 133.864 155.552 133.907C153.483 133.997 151.253 134.093 149.512 133.823H146.581C146.496 133.443 146.411 133.062 146.326 132.681C146.2 132.622 146.056 132.555 145.898 132.481C144.683 131.913 142.618 130.948 141.712 130.317C137.199 127.172 141.253 123.9 145.492 122.596C142.565 123.497 146.582 126.388 148.001 126.784C149.094 127.088 149.739 126.934 150.494 126.754C150.944 126.646 151.435 126.529 152.083 126.494C153.286 126.43 154.391 126.491 155.511 126.554ZM37.2886 37.8454C36.6716 37.5386 36.0038 37.7321 35.3302 37.9273C34.5527 38.1525 33.7674 38.3801 33.0435 37.8433C32.6529 37.5535 32.7845 36.9667 32.906 36.4256C33.0047 35.9854 33.0967 35.5753 32.8954 35.3797C30.8973 33.4378 30.7769 33.9097 30.4413 35.2252L30.4413 35.2252C30.2905 35.8161 30.0963 36.5773 29.6689 37.3663C28.8088 38.9541 27.7062 39.3199 26.611 39.6833C25.3091 40.1152 24.0176 40.5437 23.1561 43.0171C21.3159 48.3004 17.9634 46.6329 16.7499 42.0143C16.5631 41.3031 15.9943 40.2381 15.3562 39.043C13.7646 36.0625 11.7411 32.273 14.1305 31.145C12.4739 31.9269 10.5189 32.2737 8.56377 32.6205L8.56374 32.6205C6.29134 33.0236 4.01888 33.4267 2.21494 34.513C0.300238 35.666 0.307485 35.9335 0.343384 37.259C0.350265 37.513 0.358198 37.8059 0.353855 38.1513C0.336357 39.5476 0.37875 41.2329 0.418291 42.8048C0.431446 43.3277 0.444285 43.8381 0.454498 44.3211C0.486378 45.8281 1.12304 47.103 1.77597 48.4104C2.16207 49.1836 2.55386 49.9682 2.82964 50.8188C3.40816 52.6032 3.64589 54.4792 3.88295 56.3498C4.04002 57.5893 4.1968 58.8264 4.45221 60.033C4.50527 60.2837 4.5602 60.5329 4.61482 60.7807C5.13277 63.1306 5.624 65.3593 4.24949 67.6266C4.06447 67.9317 3.60208 68.2858 3.1046 68.6668C2.22707 69.3388 1.24034 70.0944 1.47402 70.8128C1.81342 71.8567 3.51136 70.3215 4.38748 69.5293C4.57698 69.3579 4.72803 69.2214 4.81858 69.1532C5.60569 68.5598 5.64845 67.9763 5.68277 67.5078C5.72429 66.9412 5.75348 66.5427 7.0718 66.4981C8.88099 66.4372 10.7276 68.4269 11.8245 69.6088L11.825 69.6094C11.9395 69.7328 12.0458 69.8473 12.143 69.9498C12.7465 70.5859 13.3118 71.3785 13.899 72.2017C15.4223 74.3373 17.0927 76.6793 19.9585 77.0311C28.8431 78.1224 30.6236 67.8791 27.315 63.2263C26.6379 62.2742 25.671 61.3429 24.6857 60.3941C23.02 58.79 21.3021 57.1355 20.8436 55.2444C20.0843 52.1117 22.7722 49.5743 26.4608 51.373C28.2153 52.2285 28.7807 53.7196 29.3506 55.2225C29.8357 56.502 30.3241 57.79 31.5523 58.7017C34.1537 60.6331 34.4 60.0782 34.8321 59.1048C35.0633 58.5838 35.3478 57.9428 36.0753 57.499C38.1851 56.2119 41.8279 55.1525 42.3316 58.3595C42.5087 59.4875 41.5724 60.448 40.6192 61.4258C39.4839 62.5904 38.3246 63.7797 38.994 65.3059C40.2879 68.2555 41.4581 66.6601 42.4568 65.2985C42.669 65.0091 42.8735 64.7303 43.0698 64.5079C45.8557 61.3516 48.633 58.0588 47.8027 53.9147C47.2707 51.2584 45.4163 49.3654 43.5297 47.4396C42.6722 46.5642 41.808 45.682 41.0584 44.7183C40.61 44.1419 40.1833 43.0798 39.7311 41.9542C39.0348 40.221 38.2781 38.3375 37.2886 37.8454ZM86.9765 108.485C86.9506 109.737 86.9248 110.986 87.4891 112.213C88.7686 114.995 89.4705 116.66 87.7033 119.261C87.2239 119.966 86.7243 120.483 86.2421 120.981C85.4366 121.813 84.6799 122.595 84.148 124.125C83.9863 124.59 83.8708 125.02 83.763 125.422L83.763 125.422C83.3725 126.878 83.0836 127.954 81.0647 128.929C76.4187 131.172 72.9469 129.117 69.8943 126.547C69.6533 126.344 69.4236 126.154 69.204 125.971C67.6577 124.687 66.612 123.819 65.6675 122.064C65.3582 121.49 65.4067 120.988 65.4512 120.528C65.5221 119.794 65.583 119.165 64.1726 118.509C62.8043 117.872 61.4659 118.115 60.1214 118.359C58.4368 118.664 56.7428 118.971 54.9684 117.553C50.9484 114.342 51.9718 108.065 55.6858 105.339C56.7338 104.57 57.6946 104.117 58.4917 103.74C60.3362 102.87 61.3043 102.413 60.4496 99.4141C59.6872 98.8005 59.1674 98.0656 58.8902 97.2095C58.853 96.4828 59.0189 95.7887 59.3881 95.1277C59.2882 94.674 59.4408 94.2244 59.5767 93.824C59.8566 92.9992 60.0656 92.3834 57.8508 92.3717C58.4654 91.9187 58.6401 90.9249 58.8226 89.8863C59.0891 88.3696 59.3724 86.7573 61.0671 86.5948C63.9884 86.3144 64.3146 89.3042 64.5594 91.5482C64.6395 92.2823 64.7109 92.9366 64.8615 93.3705C66.0426 96.7725 67.4271 97.9421 71.2119 96.5924C72.1903 96.2434 76.7818 93.7798 77.1355 93.2416C75.4915 95.7428 78.864 98.1091 81.3843 99.8774C82.004 100.312 82.5723 100.711 83.0017 101.067C83.1751 101.21 83.3505 101.352 83.526 101.495C84.5644 102.337 85.6053 103.181 86.2431 104.298C87.0345 105.683 87.0055 107.086 86.9765 108.485ZM128.423 126.988C126.979 126.832 121.692 124.785 122.335 123.078C123.963 118.751 132.454 123.641 135.39 125.332C135.593 125.449 135.769 125.551 135.915 125.633C134.804 125.692 133.726 126.009 132.652 126.326C131.259 126.737 129.872 127.146 128.423 126.988Z" fill="black" fill-opacity="0.4"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M43.247 2.4112C43.4532 3.92187 46.4117 6.2163 48.3208 4.85377C49.5387 5.29495 47.7352 2.06265 47.4339 1.76245C45.8448 0.179932 42.8405 -0.563483 43.247 2.4112ZM76.2674 4.35599C75.5297 4.18249 74.8277 4.01736 74.2418 3.81073C73.7034 3.62087 73.67 1.56805 73.8942 1.1594C74.3321 0.360833 75.8441 0.450956 77.2231 0.533151C77.6287 0.557322 78.0227 0.580807 78.3745 0.580807L91.6674 0.580807C89.8084 1.32065 91.8167 2.07513 92.7738 2.28593C93.5498 2.06389 99.1361 6.19899 98.9811 7.92931C98.8967 8.87328 98.2338 9.56063 97.5548 10.2647C96.7407 11.1088 95.9035 11.9769 96.0126 13.3401C96.0786 14.1643 96.8804 15.5173 97.7793 17.0341C100.029 20.8301 102.886 25.6521 96.3403 25.7791C93.2883 25.8383 93.3648 24.4033 93.4543 22.7243C93.5139 21.6067 93.5792 20.3811 92.7313 19.4161C92.2407 18.858 91.4524 18.5007 90.6606 18.1418C89.6489 17.6831 88.6314 17.2219 88.2219 16.3359C88.0218 15.9029 88.1752 14.9676 88.3404 13.9609C88.7512 11.4568 89.2346 8.51013 84.5301 11.7485C84.0524 12.0775 83.7571 12.7188 83.4614 13.3612C82.8926 14.5969 82.3221 15.8362 80.4473 14.8608C78.8239 14.0162 79.722 12.3529 80.6241 10.6821C81.3088 9.41395 81.9959 8.14146 81.5844 7.21944C80.7802 5.41748 78.3781 4.85247 76.2674 4.35599ZM156.698 91.805C153.943 92.5245 152.355 89.9642 151.081 87.911C150.213 86.5132 149.492 85.3503 148.647 85.6174C146.416 83.6918 144.26 85.4524 142.475 86.9098L142.475 86.9098C141.8 87.4609 141.178 87.9686 140.626 88.2173C140.74 88.2891 135.648 87.0914 136.683 86.8898C129.862 84.5716 129.78 89.7817 129.709 94.3228C129.679 96.2818 129.65 98.1163 129.083 99.1681C128.06 101.066 126.624 101.795 124.267 102.623C123.44 102.914 122.53 103.081 121.623 103.247C119.722 103.596 117.835 103.942 116.764 105.415C116.566 105.688 116.372 106.554 116.236 107.16C116.157 107.514 116.098 107.779 116.069 107.785C116.675 107.938 120.946 109.294 121.388 109.66C121.793 109.994 121.898 110.583 121.993 111.122C122.074 111.577 122.148 111.996 122.39 112.201C123.084 112.787 123.773 113.192 124.432 113.58C125.548 114.236 126.576 114.841 127.394 116.186C127.74 116.755 127.917 117.418 128.093 118.074C128.52 119.669 128.936 121.223 131.729 121.279C134 121.324 135.345 119.558 136.044 118.052C136.359 117.374 136.275 115.49 136.181 113.369C135.951 108.195 135.659 101.614 140.932 107.711C141.449 108.309 141.647 108.714 141.813 109.054C142.087 109.616 142.274 110 143.678 110.786C144.261 111.111 144.84 111.126 145.401 111.141C146.212 111.162 146.984 111.181 147.671 112.142C148.381 113.134 148.028 114.085 147.818 114.649C147.582 115.284 147.527 115.431 149.37 114.602C153.657 112.673 153.825 108.358 153.966 104.741V104.741V104.741C153.994 104.023 154.021 103.333 154.079 102.694C154.283 100.444 154.78 99.1017 155.707 97.1137C155.762 96.9962 155.917 96.7464 156.124 96.4156C157.154 94.7624 159.444 91.0875 156.698 91.805ZM72.1973 128.551C71.417 127.459 70.4216 126.736 69.3733 125.974C68.8918 125.624 68.3992 125.266 67.9111 124.861C67.6339 124.631 67.3163 124.429 66.9979 124.226L66.9977 124.226C66.3726 123.829 65.7441 123.429 65.4111 122.808C65.1427 122.308 65.2393 121.632 65.3356 120.957C65.4463 120.181 65.5566 119.408 65.1119 118.909C63.8448 117.486 62.127 117.827 60.3241 118.184C57.9169 118.661 55.358 119.167 53.5179 115.548C52.2026 112.961 51.5582 109.234 53.2752 106.713C53.9973 105.652 55.5911 104.92 57.1567 104.2C59.1051 103.305 61.0097 102.43 61.1359 100.967C61.209 100.12 60.5708 99.5094 59.9079 98.8754C59.3271 98.32 58.7273 97.7464 58.5706 96.9793C58.4384 96.3326 58.7841 95.5579 59.1129 94.8212C59.7526 93.3878 60.3283 92.0977 57.1957 92.1718C57.2491 92.192 57.2027 92.2117 57.1565 92.2313C57.1445 92.2363 57.1326 92.2414 57.1224 92.2464C57.0761 92.2276 56.7244 92.3743 56.3061 92.5487L56.306 92.5487L56.3059 92.5487C55.7575 92.7774 55.0946 93.0538 54.8549 93.0678C54.8421 93.0651 54.7142 93.1944 54.5034 93.4076C53.5406 94.3815 50.8478 97.1049 49.4865 96.9824C46.8071 96.7415 45.7563 91.5354 47.2046 91.2673C46.5904 91.1692 46.126 90.8403 45.6471 90.501C44.8015 89.902 43.9106 89.2708 42.0693 89.8216C40.6266 90.2529 40.2698 90.8662 39.7968 91.6791C39.6421 91.945 39.475 92.2322 39.2534 92.5413C38.5608 93.5077 38.1722 94.1554 37.9227 94.5713C37.6654 95.0003 37.5559 95.1828 37.4133 95.2141C37.3076 95.2373 37.1838 95.1777 36.9682 95.0739C36.5688 94.8816 35.8546 94.5377 34.358 94.2883L34.3041 94.5924C34.0981 95.753 34.0174 96.2075 34.1729 96.5568C34.273 96.7816 34.4708 96.9629 34.796 97.2608C35.0539 97.4971 35.3919 97.8067 35.8246 98.2695C36.3723 98.8553 37.004 99.3945 37.6345 99.9327C38.0849 100.317 38.5347 100.701 38.9528 101.101C39.4753 101.601 44.3041 106.591 43.3006 106.832C44.2685 107.011 44.3667 109.313 44.4609 111.521C44.5269 113.069 44.591 114.571 44.9516 115.263C45.3105 115.952 45.8877 116.545 46.464 117.137C46.9281 117.614 47.3916 118.09 47.7398 118.616C48.6819 120.037 48.8272 121.701 48.9744 123.388C49.1156 125.005 49.2587 126.643 50.1076 128.108C50.3296 128.491 50.6928 128.808 51.0524 129.122C51.3976 129.423 51.7394 129.721 51.9495 130.072C52.0887 130.305 52.3397 130.523 52.5799 130.733L52.58 130.733C52.7214 130.856 52.8591 130.976 52.9678 131.093C53.7076 131.892 53.7946 132.311 53.8476 132.566C53.8648 132.649 53.8784 132.714 53.9095 132.77C54.0282 132.982 54.4011 133.052 56.195 133.39L56.2199 133.394C59.4959 134.01 68.8418 133.717 71.4142 131.621C71.783 131.321 72.1633 130.506 72.4497 129.893C72.641 129.483 72.7904 129.164 72.8666 129.146C72.489 128.966 72.3611 128.784 72.2331 128.601C72.2212 128.585 72.2094 128.568 72.1973 128.551ZM103.179 126.451C102.289 126.353 101.634 126.066 100.993 125.786C100.109 125.399 99.2526 125.024 97.8446 125.17C96.6429 125.294 95.7247 125.861 94.804 126.429C94.267 126.76 93.7292 127.092 93.1338 127.337C91.683 127.934 89.1726 128.192 86.4353 128.473C80.9328 129.038 74.5139 129.696 73.9436 133.394C74.8734 133.394 76.2917 133.482 77.9362 133.583C83.0135 133.896 90.2481 134.343 91.9286 132.759C92.7887 131.948 92.6708 131.267 92.5712 130.693C92.4282 129.868 92.3231 129.262 95.2044 128.802C96.1451 128.652 96.8337 128.892 97.5195 129.132C98.5692 129.498 99.6124 129.862 101.543 128.82C101.814 128.674 105.302 126.685 103.179 126.451ZM101.4 106.147C106.345 101.019 107.68 108.367 107.68 111.412C107.679 114.889 101.674 117.458 99.8128 113.716C99.1948 112.474 100.182 107.41 101.4 106.147ZM156.999 7.73732C156.337 7.26753 155.713 6.32455 155.058 5.33663C154.051 3.81592 152.973 2.18874 151.582 2.01698C147.818 1.55265 148.256 5.9924 148.612 9.60597C148.799 11.4943 148.963 13.157 148.493 13.7764C148.546 13.7798 148.773 13.9206 149.101 14.1242C150.685 15.1065 154.625 17.5496 152.646 13.0643C152.47 12.6643 151.772 12.3384 151.076 12.0131C149.788 11.4116 148.506 10.8122 150.53 9.75064C151.819 9.0751 153.251 9.7519 154.636 10.4065C155.614 10.8684 156.568 11.3193 157.432 11.2761C157.432 11.0644 157.448 10.8054 157.466 10.5237C157.53 9.49892 157.614 8.17421 156.999 7.73732ZM155.77 50.9136C156.164 50.3771 156.534 49.8744 156.847 49.6882C157.037 49.5748 157.68 66.3605 156.998 66.1514C153.698 68.1169 154.112 63.7508 154.369 61.0459C154.461 60.0737 154.533 59.316 154.405 59.1441C153.928 58.5056 152.942 58.2709 151.953 58.0356C150.588 57.7106 149.217 57.3843 149.176 55.9903C149.16 55.4012 151.278 53.7683 151.813 53.4124C152.007 53.2836 152.02 53.0621 152.034 52.8335C152.049 52.5898 152.064 52.3381 152.298 52.182C152.44 52.0875 152.785 52.181 153.171 52.2856C153.67 52.4209 154.237 52.5746 154.519 52.3623C154.939 52.0448 155.368 51.4616 155.77 50.9136ZM119.362 63.9915C118.862 63.1418 118.356 62.2801 117.215 61.775C114.184 60.4327 114.124 59.8068 114.022 58.7471C113.962 58.1223 113.888 57.3467 113.182 56.1846C111.464 53.3584 107.125 53.5541 103.684 53.7093C102.984 53.7409 102.322 53.7708 101.726 53.7732C100.963 53.7762 100.412 53.7414 99.982 53.7143C98.6213 53.6284 98.4784 53.6194 96.7152 55.1321C95.9966 55.7487 95.2913 56.4733 94.5683 57.216L94.5682 57.2161C93.7551 58.0515 92.9196 58.9099 92.0179 59.6631C90.0846 61.278 85.6015 65.3674 88.8769 67.6261C87.2754 67.9828 87.773 70.6957 88.894 71.2971C88.7591 71.3488 87.3765 74.5886 87.4481 74.7654C86.7238 74.7223 86.0335 74.544 85.3696 74.3725C83.8881 73.9897 82.538 73.641 81.2353 74.9053C80.1155 75.9918 80.48 77.5685 80.8451 79.148C81.1524 80.477 81.46 81.808 80.8844 82.8505C79.2668 85.7802 73.9719 86.0608 71.1942 84.3065C69.2919 83.1048 64.9444 79.223 63.7341 77.4322C61.6669 74.3736 62.4863 71.6361 66.8112 71.7467C67.8556 71.7735 68.6087 72.2177 69.1573 72.5412C69.979 73.0259 70.3418 73.2398 70.537 71.3745C70.6229 70.5538 69.7432 69.5018 69.0274 68.6456C68.7599 68.3257 68.5153 68.0332 68.3525 67.7903C68.2377 67.6191 68.1265 67.4543 68.019 67.2949L68.0172 67.2922C66.1704 64.5547 65.4217 63.4449 66.8142 60.1024C66.884 59.935 66.9512 59.7751 67.0156 59.6216L67.0159 59.621C68.015 57.2426 68.3602 56.4209 67.4121 53.7005C66.3213 50.5702 66.0185 49.2499 66.277 45.9414C66.2887 45.792 66.3099 45.6178 66.3332 45.4264V45.4263C66.5011 44.0492 66.7778 41.7801 64.3856 41.4357C64.7518 40.8996 66.1208 36.9896 65.7888 36.6212C67.8284 36.8603 68.1542 36.4868 68.7581 35.7947C69.141 35.3558 69.6356 34.7889 70.7497 34.1687C71.5224 33.7386 72.5039 33.6332 73.4964 33.5266C74.8831 33.3777 76.2914 33.2265 77.1818 32.184C77.9936 31.2339 78.0265 29.6933 78.0588 28.1795C78.0826 27.0651 78.106 25.9653 78.4398 25.1263C79.0531 23.584 79.872 21.6689 81.9463 20.9769C86.9458 19.3094 86.5496 23.5969 86.2713 26.6079L86.2713 26.608C86.2236 27.1238 86.1794 27.602 86.1664 28.0065C86.0016 33.1236 85.9219 38.5636 86.9243 43.5024C87.9417 48.5161 88.7918 46.5963 89.9428 43.9973C90.1636 43.4986 90.3955 42.9749 90.6418 42.4704C91.0763 41.5804 91.7933 40.9283 92.5067 40.2795C93.1722 39.6742 93.8347 39.0717 94.2622 38.2813C94.5838 37.6866 94.555 37.0692 94.5272 36.4715C94.4886 35.6447 94.4518 34.8557 95.3474 34.2174C96.0195 33.7384 96.8171 33.9001 97.5453 34.0477C98.1472 34.1697 98.7017 34.2821 99.0984 34.015C99.8696 33.4959 99.9839 32.0474 100.102 30.5529C100.27 28.4196 100.446 26.1925 102.55 26.441C104.611 26.6845 104.802 28.7766 104.961 30.5359C105.074 31.7796 105.172 32.857 105.906 32.9972C105.514 33.1021 105.041 33.2117 104.535 33.329C101.317 34.0749 96.7566 35.1323 103.055 37.2471C101.121 37.43 101.683 39.7313 101.941 40.79C101.96 40.8674 101.977 40.9383 101.992 41.0011C102.062 41.3048 102.425 41.5561 102.791 41.809C103.172 42.0729 103.556 42.3386 103.614 42.6678C103.741 43.4001 103.54 43.7653 103.341 44.1272C103.196 44.3915 103.051 44.6541 103.036 45.0567C103.017 45.5932 102.963 46.0992 102.915 46.5549C102.704 48.5634 102.595 49.5973 105.901 47.9752C106.833 47.5177 107.048 46.8894 107.217 46.3967C107.452 45.7103 107.597 45.2872 109.466 45.9569C109.537 46.1497 109.705 46.2655 109.969 46.3041C110.248 46.2233 110.7 46.5284 111.234 46.8887C112.047 47.4366 113.048 48.1122 113.916 47.7526C115.644 47.0375 114.994 45.8226 114.405 44.7227C114.187 44.3142 113.977 43.9215 113.9 43.5762C113.321 40.9664 113.693 40.7925 114.531 40.4007C115.129 40.121 115.965 39.7301 116.861 38.2618C117.601 37.0502 117.124 35.8411 116.676 34.7084C116.084 33.2109 115.545 31.847 117.942 30.7876C124.05 28.0884 125.004 34.3618 124.236 37.3479L124.146 37.6959C123.309 40.9368 122.685 43.3526 124.267 46.6947C125.351 48.9864 127.13 50.3405 129.038 51.7921L129.038 51.7922C129.998 52.523 130.991 53.2785 131.944 54.1907C135.526 57.6195 135.816 61.7061 131.727 64.9927C128.453 67.6239 124.054 68.4043 120.634 65.6467C120.065 65.1881 119.715 64.5928 119.362 63.9915ZM113.122 94.8436C108.398 94.2175 106.767 94.0971 103.199 94.9759C103.428 95.2054 103.44 95.3733 103.237 95.4797C103.838 95.7105 104.755 95.8669 105.736 96.0343C107.487 96.3331 109.443 96.6667 110.17 97.5203C110.35 97.731 110.201 98.5595 110.065 99.312L110.065 99.312L110.065 99.3123C109.947 99.9714 109.839 100.572 109.972 100.648C109.864 100.586 114.183 95.3888 114.849 94.892C114.271 94.9126 113.696 94.8966 113.122 94.8436ZM31.2317 108.277L31.2321 108.277C31.4232 108.384 31.5582 108.46 31.5872 108.471C32.0908 108.658 32.6228 108.74 33.1828 108.717C32.925 110.121 30.8707 112.723 28.6317 112.294C25.9681 111.783 26.9887 110.61 28.0705 109.368C28.6579 108.693 29.2633 107.998 29.3069 107.376C29.3188 107.208 30.5901 107.918 31.2315 108.277L31.2316 108.277L31.2317 108.277ZM30.1187 101.632C30.5158 102.186 31.2043 102.616 31.8647 103.029C32.1596 103.213 32.4489 103.394 32.7041 103.58C33.0698 103.041 32.7133 101.975 32.3622 100.925C32.0697 100.05 31.7809 99.1862 31.9163 98.6468C30.2098 98.5915 29.3721 100.591 30.1187 101.632ZM61.1522 0.69232C59.5224 1.6115 53.8064 4.8353 53.4891 1.49698C53.4632 1.19185 53.4406 0.88638 53.4215 0.580909L61.3503 0.580909C61.2936 0.612587 61.2272 0.650013 61.1522 0.69232Z" fill="white" fill-opacity="0.2"/>
`;

        }, {}],
        43: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M13.2939 131.119C10.6826 131.533 3.75571 132.631 2.50165 131.512C1.68361 130.781 2.05133 128.182 2.31572 126.314C2.41614 125.604 2.50165 125 2.50165 124.643C3.7762 124.581 4.7433 126.023 5.68851 127.432C7.09076 129.522 8.44479 131.541 10.683 128.471C11.7323 127.032 11.4901 125.216 11.244 123.37C10.9845 121.423 10.7206 119.444 11.962 117.839C13.4382 115.93 15.794 115.665 18.1219 115.404C20.5872 115.126 23.0212 114.853 24.3462 112.634C24.7278 111.994 24.8316 110.862 24.9435 109.639C25.1378 107.518 25.3568 105.126 27.0955 104.573C30.0956 103.618 30.2874 106.238 30.4548 108.524C30.5447 109.752 30.6276 110.884 31.1344 111.314C36.5946 115.953 37.11 107.01 37.227 104.98L37.2289 104.947C37.2531 104.527 37.1793 104.185 37.1116 103.872C36.9889 103.303 36.8865 102.829 37.4278 102.158C37.7112 101.807 38.0984 101.868 38.5061 101.932C38.9188 101.998 39.3525 102.066 39.7207 101.713C41.1686 100.325 41.1971 100.192 41.529 98.6453L41.5293 98.6439L41.5347 98.6183L41.5645 98.4797L41.6072 98.2826C41.7219 97.7567 41.7895 97.3089 41.8511 96.9012C41.9908 95.9763 42.0993 95.2581 42.6544 94.3042C42.913 93.8602 43.2613 93.4261 43.609 92.9928C44.7623 91.5556 45.9086 90.1271 43.7477 88.3731C44.341 88.3663 49.506 86.9372 49.8509 86.7523C50.3501 86.4848 50.9238 85.7898 51.4375 85.1674C51.8893 84.62 52.2948 84.1287 52.5626 84.0336C55.628 82.9449 55.4144 85.2958 55.1944 87.7178C55.0621 89.1747 54.9274 90.6572 55.5025 91.4326C47.5526 91.7049 42.7738 99.5241 42.3029 105.595C42.2311 106.523 42.5869 107.317 42.9224 108.067C43.1918 108.669 43.448 109.241 43.4595 109.829C43.4758 110.657 43.2171 111.219 42.9601 111.776C42.7478 112.236 42.5368 112.693 42.4831 113.297C42.2509 115.893 43.2916 116.853 44.6508 118.107C45.3351 118.738 46.1002 119.444 46.8242 120.47C49.2272 123.878 47.9694 124.402 44.3341 123.787C42.2441 123.434 40.2355 122.806 38.2267 122.178C35.7676 121.409 33.3082 120.64 30.6984 120.376C27.3593 120.036 24.2567 120.203 22.676 122.994C21.7635 124.605 21.3642 129.035 24.4031 129.382C23.6417 129.505 23.017 129.756 22.3792 130.012C21.5477 130.346 20.6938 130.689 19.4855 130.769C18.5209 130.384 17.5102 130.14 16.4528 130.037C15.5476 130.184 14.7413 130.506 14.0344 131.004C13.8625 131.029 13.6098 131.069 13.2939 131.119ZM193.244 62.0127C193.705 62.282 194.613 62.2323 195.593 62.1786C196.654 62.1206 197.799 62.058 198.548 62.3913C198.945 62.5678 199.346 63.0254 199.75 63.4868C200.12 63.9098 200.493 64.3362 200.869 64.5525C201.726 65.0451 202.652 65.2156 203.576 65.3858C204.282 65.5158 204.987 65.6456 205.66 65.9183C208.465 67.0562 210.546 69.5216 208.933 72.3852C208.485 73.18 207.456 73.7235 206.439 74.2613C205.123 74.9567 203.826 75.6424 203.825 76.849C203.825 81.3288 211.611 73.9361 212.309 72.3529C212.505 71.9074 212.576 71.3955 212.647 70.8781C212.747 70.1563 212.848 69.4238 213.294 68.8461C213.641 68.3962 214.543 67.9206 215.253 67.5469C215.73 67.2956 216.119 67.0903 216.193 66.97C217.487 64.8727 216.937 59.7113 216.514 55.7453C216.343 54.1469 216.193 52.7428 216.193 51.8116C213.995 51.8221 213.001 54.5043 212.005 57.1881C210.558 61.0913 209.11 64.9979 203.955 60.6945C202.589 59.5536 198.654 52.5843 198.654 51.2839C198.654 50.0985 199.263 49.4366 199.91 48.7333C200.222 48.3947 200.543 48.0464 200.809 47.6255C201.104 47.158 201.349 46.7952 201.555 46.4906L201.555 46.4905C202.379 45.2736 202.573 44.9868 202.786 42.6724C202.959 40.7888 203.17 40.5021 203.718 39.7591L203.718 39.7591L203.718 39.7589C203.944 39.4529 204.227 39.0695 204.587 38.4653C204.764 38.1685 205.081 37.9023 205.393 37.6394C206.107 37.0387 206.8 36.4556 205.762 35.5666C203.157 33.3362 197.518 39.1196 196.54 40.2762C193.765 43.557 189.987 48.0432 190.322 52.2016C190.397 53.1372 192.405 61.5222 193.244 62.0127ZM159.752 121.135C159.416 121.535 159.095 121.917 159.13 122.429C159.252 124.216 159.258 124.214 159.481 124.105C159.649 124.024 159.942 123.881 160.502 124.401C160.658 124.546 161.134 124.761 161.702 125.017C163.028 125.615 164.858 126.44 164.304 127.138C164.017 127.499 163.256 127.531 162.453 127.565C161.649 127.599 160.803 127.635 160.347 128.004C160 128.284 159.785 128.83 159.592 129.32C159.431 129.728 159.286 130.097 159.092 130.242C157.533 131.41 156.924 131.511 153.606 131.511C155.801 130.884 150.164 126.989 149.59 128.226C150.125 127.074 149.415 125.852 148.743 124.693C148.345 124.008 147.96 123.346 147.853 122.733C147.714 121.932 147.739 121.142 147.763 120.369C147.818 118.617 147.869 116.959 145.996 115.494C146.628 115.422 146.986 114.974 147.343 114.529C147.72 114.059 148.094 113.591 148.786 113.568C149.664 115.94 151.682 116.106 153.824 116.283C154.802 116.364 155.806 116.447 156.738 116.743C157.03 116.836 157.311 116.91 157.578 116.982C158.629 117.262 159.485 117.49 160.102 118.557C160.842 119.837 160.28 120.507 159.752 121.135ZM216.193 0.888748C216.193 -0.187622 214.213 -0.0337067 212.747 0.0802155L212.747 0.0802612C212.352 0.110962 211.994 0.138748 211.723 0.138748C212.354 0.386749 212.714 0.560104 212.964 0.679962C213.155 0.772156 213.281 0.832687 213.414 0.871185C213.667 0.944778 213.943 0.937866 214.748 0.917725H214.748H214.748H214.749H214.75C215.11 0.908676 215.576 0.897034 216.193 0.888748ZM3.25195 2.38644C3.04356 3.04115 2.85849 3.62257 3.21851 3.75348C2.13312 3.96953 2.33553 1.98274 2.45175 0.841919C2.47905 0.574051 2.50159 0.352814 2.50159 0.217697C3.9446 0.210358 3.57095 1.38426 3.25195 2.38644ZM160.759 0.138718H157.182C157.486 0.281464 157.805 0.458908 158.122 0.635178L158.122 0.635361L158.122 0.635498L158.123 0.63559C159.453 1.37607 160.746 2.09567 160.759 0.138718ZM215.304 9.1382C215.3 9.13745 215.295 9.13722 215.291 9.13749C215.731 9.05922 215.996 9.09053 216.148 9.10857H216.148H216.148H216.148C216.236 9.11903 216.287 9.12503 216.312 9.10256C216.346 9.07243 216.335 8.99108 216.309 8.80064C216.269 8.50691 216.193 7.95366 216.193 6.92846C215.166 6.94547 213.554 9.1133 215.29 9.13754C215.287 9.13777 215.284 9.13828 215.281 9.13904L215.304 9.1382ZM87.8889 131.512H89.6772C89.3083 131.435 88.2412 131.332 87.8889 131.512ZM87.9491 76.7138C88.578 76.6253 89.1243 76.3174 89.6655 76.0124C90.4468 75.5721 91.2173 75.1379 92.2102 75.3785C93.2313 75.6259 93.6 76.2833 93.877 76.7771C94.0318 77.0531 94.1579 77.2779 94.3532 77.3516C97.0727 78.3779 97.5079 76.446 97.9754 73.4244C98.0597 72.8797 98.0795 72.4084 98.0975 71.9786C98.1419 70.9193 98.1757 70.1122 99.1394 69.079C99.687 68.4919 100.59 68.1832 101.483 67.8775C102.133 67.6554 102.777 67.4349 103.278 67.1105C104.69 66.1947 105.316 65.3553 106.082 64.3289L106.083 64.3287L106.083 64.3286C106.387 63.9215 106.712 63.4849 107.118 63.0025C108.039 61.907 109.386 60.8529 110.785 59.7588C114.538 56.8224 118.658 53.5986 115.892 48.5163C112.92 43.0559 111.743 47.6719 110.835 51.2337C110.486 52.5999 110.177 53.811 109.822 54.2388C105.348 59.6227 104.376 54.3609 103.793 51.2046L103.793 51.2044C103.698 50.6889 103.613 50.2296 103.525 49.8819C102.27 44.9252 100.552 48.4977 99.1364 51.4427C98.7102 52.3293 98.3113 53.159 97.9606 53.682C97.54 54.3091 97.014 54.7664 96.5021 55.2114C95.8685 55.7622 95.2565 56.2943 94.8923 57.1066C94.6629 57.6184 94.775 58.2583 94.8902 58.9158C95.0804 60.0013 95.2791 61.1349 93.9634 61.8197C92.5695 62.5453 90.9686 61.6623 89.5426 60.8757C88.4615 60.2794 87.4809 59.7386 86.7672 59.9959C84.5226 60.8054 86.5415 63.5389 88.157 65.7262C88.8997 66.7318 89.5571 67.6219 89.6758 68.1566C90.0212 69.7118 89.4129 71.0606 88.8049 72.4087C88.1894 73.7733 87.5743 75.1372 87.9491 76.7138ZM52.523 39.3304C52.4442 39.3231 52.366 39.3268 52.2886 39.3388C52.3701 39.3381 52.4489 39.3359 52.523 39.3304ZM43.1229 46.5051C42.6932 47.9017 41.5535 49.065 40.4191 50.2228C39.1322 51.5362 37.8523 52.8425 37.6237 54.4745C37.2565 57.0957 39.2754 58.4359 41.3658 59.8235C42.0273 60.2627 42.6961 60.7066 43.2986 61.1974C44.6694 62.3138 45.5979 63.246 46.2965 64.8122C46.4157 65.0797 46.4385 65.5294 46.4621 65.995C46.4963 66.6694 46.5322 67.3771 46.8652 67.613C48.637 68.8678 49.6301 68.0043 50.8229 66.9671C51.3944 66.4701 52.0117 65.9333 52.7825 65.5705C55.3938 64.3415 55.4258 64.3609 56.99 65.3101C57.3053 65.5014 57.6827 65.7304 58.156 65.9946C58.3974 66.1293 58.6253 66.2686 58.8436 66.4021C60.2774 67.2787 61.2993 67.9034 63.0621 65.3438C63.7688 64.3176 63.7321 62.5144 63.6938 60.642C63.6298 57.5045 63.5619 54.1727 66.9811 53.9768C65.4773 53.2967 66.1547 51.2713 66.8425 49.2149C67.5521 47.0933 68.2727 44.9386 66.62 44.1944C62.6101 42.3888 61.4467 50.9484 60.7809 55.8462C60.5616 57.4602 60.3962 58.6765 60.2009 58.9932C58.6052 61.5808 52.8833 63.7803 51.236 59.9639C50.9412 59.2808 51.1597 58.5451 51.3814 57.7983C51.5961 57.0751 51.8139 56.3416 51.5715 55.6357C51.2192 54.6093 50.6402 54.234 49.9691 53.799C49.6093 53.5658 49.223 53.3154 48.831 52.9383C46.0532 50.2672 46.1534 50.1037 47.544 47.8328C47.7825 47.4433 48.059 46.9919 48.3654 46.4552C48.6663 45.928 48.9032 45.0932 49.1632 44.1773C49.7514 42.1048 50.4574 39.6171 52.2886 39.3388C51.291 39.3459 50.2452 39.096 49.2113 38.8489C47.2328 38.376 45.298 37.9136 43.8283 39.2838C42.8041 40.2388 42.9787 41.9663 43.1452 43.6132C43.2542 44.6916 43.3598 45.7354 43.1229 46.5051ZM108.509 15.9001L108.509 15.8997L108.507 15.8977L108.507 15.8976C108.218 15.5228 107.915 15.1307 107.603 14.708C111.791 14.6401 115.094 10.166 110.513 7.63995C112.367 8.66252 113.769 7.52211 114.909 6.59509C116.342 5.42937 117.36 4.6011 118.34 8.8356C118.411 9.14219 118.505 9.51212 118.605 9.90503L118.606 9.90704L118.606 9.90814L118.607 9.91057L118.607 9.91306L118.608 9.91405L118.608 9.91428L118.608 9.91471L118.608 9.91551L118.608 9.91558L118.608 9.91576L118.608 9.91611L118.608 9.91647C118.826 10.7716 119.071 11.7334 119.166 12.3872C119.37 13.7961 119.334 13.9714 119.176 14.7496C119.149 14.8836 119.118 15.0355 119.084 15.2146C118.998 15.6692 118.828 15.9871 118.67 16.2806C118.334 16.906 118.058 17.4205 118.787 18.9109C119.167 19.6873 119.635 20.2939 120.079 20.8697C120.888 21.9181 121.619 22.8646 121.593 24.5483C121.586 25.0211 121.391 25.3944 121.181 25.7965C120.828 26.4717 120.432 27.2281 120.814 28.6734C121.035 29.5109 121.442 30.305 121.849 31.0996C122.338 32.0536 122.827 33.0084 122.996 34.04C121.267 33.965 120.168 32.6134 119.181 31.3997L119.181 31.3996L119.181 31.3992C118.835 30.973 118.502 30.5639 118.161 30.233C117.396 29.4913 116.474 28.9665 115.579 28.4574C114.799 28.0132 114.04 27.581 113.425 27.0274C111.696 25.472 111.636 24.6708 111.524 23.1942V23.1941C111.489 22.7221 111.448 22.1811 111.346 21.5244C110.966 19.0841 109.89 17.6898 108.509 15.9001ZM108.153 41.8087C110.355 40.8234 108.012 36.3301 105.966 36.1558C104.538 36.034 100.707 39.2478 100.46 40.1654C101.693 39.9286 102.952 40.5163 104.185 41.0921C105.564 41.7361 106.911 42.365 108.153 41.8087ZM54.3484 34.554C55.3634 35.1668 57.3841 37.7892 54.4468 37.642C53.7002 37.3464 52.8391 37.1555 51.9893 36.9671C50.1918 36.5686 48.4456 36.1815 47.9445 34.8387C48.4958 34.8266 49.1999 34.6939 49.9525 34.552L49.9526 34.552C51.5206 34.2564 53.2995 33.9211 54.3484 34.554ZM76.0684 56.3633C74.9989 56.4118 73.9252 56.1227 72.9673 55.8648C70.9422 55.3195 69.435 54.9137 69.5804 58.1318C70.3795 58.1808 71.13 58.3282 71.8199 58.4638C73.894 58.8713 75.4198 59.1711 76.0684 56.3633ZM42.7317 24.3242C42.638 24.8422 42.5483 25.3375 42.1475 25.4698C41.9886 25.4285 41.8013 25.3836 41.5946 25.3342L41.5944 25.3341C39.9007 24.9288 36.9081 24.2127 37.6443 22.5571C38.5365 20.5505 40.2509 21.4382 41.7693 22.2244C42.4565 22.5802 43.1036 22.9153 43.6162 22.9518C42.9608 23.0587 42.8434 23.7075 42.7317 24.3242ZM134.298 51.5691C133.933 50.2797 133.983 49.265 134.316 47.987C132.126 48.0046 132.288 51.5367 134.298 51.5691ZM107.49 14.3892L107.565 14.6259C108.271 14.5877 107.795 14.4275 107.49 14.3892ZM143.056 49.3247C143.529 49.9893 143.585 50.9775 143.641 51.9834C143.726 53.4941 143.814 55.045 145.32 55.6001C146.839 56.1601 147.612 54.8952 148.444 53.5351C149.144 52.392 149.884 51.1817 151.142 50.9307C151.982 50.7632 152.839 51.1759 153.716 51.5989C154.54 51.9957 155.382 52.4015 156.246 52.345C158.754 52.1806 162.714 49.6451 162.511 47.3728C161.441 47.5488 160.565 47.3396 159.69 47.1306C158.963 46.9572 158.238 46.7839 157.402 46.8305C156.545 46.8782 155.699 47.133 154.879 47.38C154.105 47.6132 153.354 47.8393 152.639 47.8778C149.378 48.0531 149.402 47.6878 149.478 46.5166C149.511 45.9979 149.555 45.3212 149.328 44.4633C149.239 44.1272 149.189 43.8798 149.148 43.6757L149.147 43.6756C149.034 43.1191 148.986 42.8847 148.405 42.0538C148.143 41.6801 147.658 41.3973 147.163 41.1085C146.562 40.7578 145.945 40.3982 145.695 39.8561C145.6 39.6498 145.696 38.9405 145.809 38.1157C146.002 36.6964 146.242 34.9349 145.632 34.8052C144.821 35.9855 143.569 36.9038 142.318 37.8217C140.76 38.9649 139.202 40.1075 138.499 41.7551C137.276 44.6187 138.807 45.7437 140.57 47.04C141.444 47.6821 142.375 48.3663 143.056 49.3247ZM68.2108 65.1935C69.8862 64.8083 69.0349 61.0465 67.7719 61.9365C66.9784 62.4955 67.0035 65.0768 68.2108 65.1935ZM98.6068 40.9135C98.5976 40.9138 98.5871 40.914 98.5771 40.9141C98.5861 40.9197 98.5959 40.9193 98.6068 40.9135ZM3.26213 93.2051C3.49574 93.0917 3.71698 92.9328 3.93985 92.7727C4.3466 92.4805 4.75882 92.1844 5.26178 92.1539C5.23721 91.5767 2.76299 89.1071 2.50172 89.1153C2.50172 89.1903 2.47742 89.3734 2.44421 89.6233L2.44417 89.6239C2.27588 90.8913 1.87958 93.8761 3.26213 93.2051ZM109.989 1.0621C109.703 0.746719 109.418 0.431442 109.072 0.295486C108.983 0.260437 108.96 0.224121 108.938 0.18779C108.927 0.171387 108.917 0.154968 108.9 0.138672C113.834 0.138672 112.084 2.48874 110.767 1.76382C110.481 1.60667 110.235 1.33435 109.989 1.0621ZM131.287 127.166C131.427 127.264 131.316 127.217 131.127 127.137C130.828 127.01 130.333 126.8 130.328 126.952C130.315 127.382 132.655 128.127 131.287 127.166ZM10.3412 82.5545C10.6315 82.7283 10.9223 82.9023 11.2427 82.9917C12.7136 83.4024 8.72542 83.2164 9.22339 82.0731C9.63123 82.1297 9.98592 82.3419 10.3412 82.5545Z" fill="white" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M189.719 100.461L189.719 100.461L189.718 100.461C189.564 100.762 189.427 101.028 189.368 101.165C191.523 101.359 192.096 114.791 184.674 111.987C184.019 111.74 183.662 110.135 183.264 108.345C182.468 104.767 181.507 100.447 177.667 104.725C176.41 106.124 177.133 107.951 177.852 109.771C178.679 111.863 179.503 113.946 177.32 115.364C174.079 117.469 172.724 115.065 171.347 112.622C170.323 110.805 169.286 108.967 167.453 108.947C170.642 111.027 170.527 117.006 170.44 121.514C170.409 123.152 170.381 124.596 170.517 125.588C171.274 125.576 171.839 125.265 172.394 124.959C172.976 124.638 173.548 124.323 174.32 124.367C175.516 124.435 176.323 124.98 177.144 125.535C177.505 125.78 177.869 126.026 178.27 126.234C178.581 126.395 178.947 126.656 179.35 126.943L179.35 126.943C180.551 127.799 182.079 128.887 183.449 128.276C183.871 128.087 183.941 127.31 184.018 126.45C184.112 125.399 184.217 124.226 184.99 123.859C186.544 123.123 186.523 124.8 186.501 126.459C186.488 127.501 186.475 128.536 186.852 128.962C188.212 130.499 194.444 130.847 196.548 130.856C196.597 129.971 196.491 129.315 196.408 128.798C196.233 127.713 196.157 127.242 197.825 126.564C200.449 125.497 205.079 128.403 206.594 129.614C206.91 129.866 206.992 130.38 207.066 130.843C207.129 131.238 207.186 131.596 207.378 131.72C208.018 132.132 213.457 131.989 213.834 131.912C216.046 131.455 217.459 130.015 216.863 127.061C216.383 124.681 214.012 123.593 211.609 122.49C209.036 121.309 206.427 120.112 206.066 117.293C205.982 116.64 206.335 115.828 206.68 115.034C206.952 114.406 207.22 113.789 207.263 113.271C207.33 112.467 207.161 111.643 206.991 110.819C206.772 109.753 206.553 108.688 206.845 107.667C207.02 107.053 210.491 102.833 210.851 102.707L210.875 102.707C210.87 102.703 210.862 102.703 210.851 102.707C210.25 102.695 209.724 102.478 209.197 102.26C208.637 102.029 208.077 101.797 207.429 101.814C207.324 101.787 207.015 102.406 206.657 103.121C206.201 104.035 205.665 105.107 205.375 105.193C202.332 106.094 203.099 103.945 203.655 102.388C203.818 101.933 203.962 101.53 203.988 101.266C204.099 100.168 203.903 99.1199 203.706 98.0643C203.57 97.3351 203.434 96.6022 203.397 95.8461C203.369 95.2868 203.498 94.7777 203.626 94.2731C203.803 93.5734 203.978 92.8824 203.731 92.0782C203.679 91.9072 203.046 92.0556 202.453 92.1947C201.922 92.3192 201.424 92.4362 201.405 92.3087C201.297 91.5977 201.56 90.4805 201.825 89.3617C202.47 86.6276 203.118 83.8836 198.353 87.0329C195.917 88.6431 194.928 90.953 193.923 93.2986C193.057 95.3201 192.18 97.3681 190.356 99.0174C190.45 99.0393 190.045 99.8253 189.719 100.461L189.719 100.461L189.719 100.461ZM7.32916 122.637C8.68289 121.909 7.90428 121.266 6.96771 120.492C6.03242 119.719 4.93961 118.816 5.65553 117.569C6.11534 116.768 7.20139 116.597 8.31425 116.422C9.43944 116.245 10.5921 116.063 11.1526 115.222C11.9593 114.011 11.6337 113.157 11.3177 112.329C10.9486 111.361 10.5926 110.428 12.0696 109C13.1053 107.999 14.0827 107.678 14.931 107.399C16.189 106.986 17.1635 106.665 17.6248 104.36C18.3313 100.828 16.572 98.4816 14.8913 96.2401C12.9892 93.7031 11.1877 91.3004 13.176 87.4658C13.2675 87.2892 13.3804 87.0866 13.5045 86.8641C14.5461 84.9951 16.3716 81.7196 12.9065 80.6212C10.3757 79.8189 10.5278 80.7354 10.6857 81.6871C10.8044 82.4027 10.9265 83.1381 9.91397 83.1777C10.1243 83.7636 7.05772 83.0813 7.4368 82.2161C6.62999 82.127 6.23767 81.6808 5.85703 81.2479C5.30879 80.6243 4.78474 80.0283 3.08133 80.5664C1.37814 81.1043 1.10683 82.725 0.883423 84.0595L0.865982 84.1635C0.529312 86.1641 1.4303 87.8993 2.29842 89.5713C3.0206 90.9622 3.72003 92.3093 3.66527 93.7289C3.65738 93.934 3.66309 94.103 3.66829 94.2569C3.68059 94.621 3.69006 94.9015 3.5105 95.3794C2.94679 95.7854 2.30727 95.9404 1.59174 95.8444C0.484161 95.6081 0.175385 96.0372 0.665436 97.1315C-0.298599 99.0268 0.0646362 102.232 0.384613 105.055L0.384659 105.055C0.52948 106.333 0.665436 107.533 0.665436 108.498C0.665436 109.165 0.551025 110.28 0.417084 111.586C0.0664825 115.003 -0.417938 119.726 0.665436 121.126C1.2973 121.943 6.2426 123.222 7.32916 122.637ZM7.4456 71.2177C7.71201 72.9781 8.05817 75.2656 5.20143 74.5184C3.1732 73.988 3.40147 71.9333 3.61148 70.0429C3.85637 67.8386 4.07645 65.8576 0.665451 66.7768V51.1637C0.878937 51.1781 1.15738 51.2043 1.48354 51.2351L1.48625 51.2353L1.48836 51.2355L1.48927 51.2356C4.49867 51.5196 11.5477 52.1848 9.15839 47.5044C11.1045 49.2142 13.2283 47.6623 14.9563 46.3996C16.1347 45.5386 17.1291 44.8119 17.7578 45.3457C17.6093 45.2196 17.2237 46.5917 16.8367 47.9692C16.4894 49.2049 16.141 50.4449 15.9614 50.6117C15.3622 51.1682 14.0261 51.3281 12.6431 51.4937C10.5143 51.7485 8.27443 52.0166 8.44029 53.7648C8.72028 56.7134 12.9859 55.5623 16.5126 54.6106C18.5623 54.0574 20.3624 53.5716 20.9852 53.9973C21.9066 54.6271 20.7835 56.9521 19.7751 59.0395C19.2279 60.1724 18.7144 61.2353 18.5799 61.9192C18.5007 62.3227 18.5189 62.743 18.537 63.1622C18.5741 64.02 18.6109 64.8731 17.8135 65.5674C16.3962 66.8017 15.0854 66.4605 13.7453 66.1117C12.9505 65.9048 12.1454 65.6952 11.3016 65.8101C7.77838 66.2901 7.44839 66.9989 7.31386 69.91C7.29836 70.2446 7.36865 70.7092 7.4456 71.2177ZM3.3492 7.33514L3.34943 7.33517C4.42555 7.39651 5.48569 7.45693 6.64119 7.39364C7.26382 7.35954 7.73436 7.24424 8.16705 7.13821C8.89198 6.96057 9.51071 6.80896 10.5606 7.10879C11.9229 7.49803 15.7799 10.3467 12.9698 11.2336C17.0401 9.94933 20.9326 6.72562 16.5995 3.62788C15.7299 3.00636 13.7489 2.05688 12.5822 1.49776L12.582 1.49762L12.5817 1.49748L12.5813 1.4973L12.5806 1.49693C12.4281 1.42386 12.2895 1.35747 12.1693 1.29936L12.0767 0.873688L11.9243 0.173981H9.10971C7.43806 -0.0919189 5.29735 0.00344849 3.3107 0.0919495H3.31038H3.3101H3.30974H3.30951C2.37206 0.133728 1.46886 0.173981 0.665497 0.173981C0.665497 0.918732 1.30222 1.284 1.98239 1.67419C2.92442 2.21461 3.94984 2.80286 3.48267 4.51329C3.31795 5.11616 2.7005 5.42501 2.08549 5.73265C1.37387 6.08862 0.665497 6.44296 0.665497 7.24928C1.60896 7.23597 2.48425 7.28585 3.3492 7.33514ZM118.745 94.6388C118.098 94.8311 117.457 95.0217 116.864 94.7195C115.914 94.2347 115.187 92.3793 114.519 90.6719L114.519 90.6718L114.519 90.6718C114.085 89.5631 113.675 88.5169 113.244 87.9491C112.525 86.9998 111.695 86.1308 110.872 85.2685C109.06 83.3714 107.28 81.5066 106.769 78.89C105.971 74.8077 108.638 71.5641 111.313 68.4549C111.502 68.2358 111.698 67.9612 111.902 67.6761C112.861 66.3348 113.984 64.7633 115.227 67.6688C115.869 69.1722 114.756 70.3437 113.666 71.491C112.751 72.4543 111.852 73.4004 112.022 74.5116C112.506 77.6707 116.003 76.6271 118.029 75.3592C118.728 74.9221 119.001 74.2906 119.223 73.7773C119.638 72.8185 119.874 72.2719 122.372 74.1744C123.551 75.0725 124.02 76.3413 124.486 77.6017C125.033 79.0822 125.576 80.551 127.261 81.3938C130.802 83.1656 133.383 80.666 132.654 77.5801C132.214 75.7172 130.564 74.0874 128.965 72.5073C128.019 71.5726 127.091 70.6553 126.441 69.7174C123.264 65.134 124.973 55.0436 133.504 56.1186C136.256 56.4652 137.86 58.7722 139.322 60.8759C139.886 61.6869 140.429 62.4677 141.008 63.0942C141.102 63.1954 141.204 63.3084 141.314 63.4301C142.367 64.5944 144.14 66.5544 145.878 66.4944C147.143 66.4505 147.171 66.058 147.211 65.4998C147.244 65.0383 147.285 64.4635 148.041 63.879C148.128 63.8118 148.273 63.6774 148.455 63.5087L148.455 63.5085C149.296 62.7281 150.927 61.2158 151.252 62.2441C151.477 62.9518 150.529 63.6961 149.687 64.3581C149.209 64.7334 148.765 65.0822 148.588 65.3828C147.268 67.6163 147.739 69.8116 148.237 72.1264C148.289 72.3706 148.342 72.6161 148.393 72.863C148.638 74.0514 148.789 75.2698 148.939 76.4906L148.939 76.4907L148.939 76.4909L148.939 76.491L148.939 76.4912C149.167 78.3339 149.395 80.1819 149.951 81.9397C150.216 82.7776 150.592 83.5503 150.962 84.3119L150.962 84.3119L150.963 84.312L150.963 84.312L150.963 84.3121C151.589 85.6001 152.201 86.8559 152.231 88.3404C152.241 88.8148 152.253 89.316 152.266 89.8295L152.266 89.834C152.304 91.3825 152.345 93.0427 152.328 94.4181C152.324 94.7584 152.331 95.0469 152.338 95.2971C152.373 96.6028 152.38 96.8663 150.541 98.0021C148.809 99.0722 146.627 99.4693 144.445 99.8664C142.568 100.208 140.691 100.55 139.1 101.32C141.394 100.209 139.451 96.4757 137.923 93.5397L137.923 93.5396C137.31 92.3624 136.764 91.3133 136.585 90.6128C135.42 86.0631 132.201 84.4205 130.434 89.625C129.607 92.0614 128.367 92.4835 127.117 92.909C126.065 93.267 125.006 93.6273 124.18 95.1914C123.77 95.9686 123.584 96.7184 123.439 97.3006C123.117 98.5964 123.001 99.0613 121.082 97.1484C120.889 96.9556 120.977 96.5517 121.072 96.1181C121.189 95.585 121.315 95.007 120.94 94.7215C120.245 94.1927 119.491 94.4169 118.745 94.6388ZM68.6628 21.4619C69.2046 22.6706 69.1798 23.9005 69.1549 25.1345C69.1271 26.512 69.0992 27.8946 69.8591 29.2588C70.4716 30.3587 71.4711 31.1904 72.4681 32.0201C72.6366 32.1603 72.805 32.3005 72.9715 32.4419C73.3838 32.7922 73.9294 33.1849 74.5245 33.6133C76.9444 35.3553 80.1826 37.6862 78.604 40.1501C78.9436 39.6199 83.3523 37.193 84.2918 36.8493C87.9258 35.5197 89.2552 36.6719 90.3893 40.0231C90.5339 40.4505 90.6024 41.095 90.6793 41.8181V41.8182C90.9144 44.0287 91.2276 46.9738 94.0325 46.6977C95.6598 46.5375 95.9318 44.9494 96.1877 43.4553C96.3629 42.4322 96.5306 41.4532 97.1207 41.007C94.9942 40.9954 95.1948 40.3888 95.4636 39.5764C95.5941 39.182 95.7406 38.739 95.6447 38.2921C95.9992 37.641 96.1585 36.9573 96.1228 36.2413C95.8566 35.3981 95.3575 34.6741 94.6255 34.0697C93.8048 31.1155 94.7343 30.6654 96.5055 29.8079C97.2708 29.4374 98.1933 28.9907 99.1996 28.2331C102.766 25.5482 103.748 19.3649 99.8884 16.201C98.1847 14.8042 96.5581 15.1068 94.9406 15.4077C93.6497 15.6479 92.3646 15.887 91.0508 15.2601C89.6965 14.6138 89.755 13.9936 89.8231 13.2711C89.8658 12.8175 89.9124 12.3235 89.6154 11.7575C88.7085 10.0289 87.7045 9.17364 86.2197 7.90878C86.0088 7.72915 85.7882 7.54126 85.5569 7.34144C82.6258 4.81036 79.2923 2.78557 74.8313 4.99503C72.8929 5.9554 72.6155 7.01604 72.2405 8.4499L72.2404 8.45001C72.137 8.84568 72.0261 9.26975 71.8708 9.72798C71.36 11.2352 70.6335 12.0053 69.8601 12.825C69.3972 13.3157 68.9174 13.8242 68.4571 14.5193C66.7603 17.0816 67.4342 18.721 68.6628 21.4619ZM93.5336 102.141C93.163 101.313 92.7938 100.488 93.0181 99.7412C95.4478 99.5117 96.7804 100.74 98.1933 102.042C98.7984 102.6 99.4182 103.171 100.145 103.647C101.491 104.528 103.007 105.216 104.522 105.904C105.651 106.416 106.78 106.928 107.837 107.52C110.36 108.932 110.686 109.754 111.144 110.908C111.41 111.577 111.72 112.358 112.529 113.431C112.887 113.905 116.364 116.185 116.378 116.156C112.63 116.13 113.416 117.351 114.429 118.922C115.064 119.908 115.789 121.032 115.539 122.074C115.136 123.756 108.618 128.19 106.466 127.501C103.587 126.579 104.926 125.129 106.264 123.68C107.02 122.861 107.777 122.042 107.769 121.319C107.745 119.156 102.882 114.738 100.283 114.886C98.1966 115.005 98.2658 115.615 98.348 116.339C98.3748 116.575 98.4029 116.823 98.3582 117.07C98.2488 117.674 98.4404 118.151 98.6379 118.643C98.9657 119.46 99.3096 120.317 98.3193 121.857C96.4738 124.729 92.4889 128.29 88.8154 128.377C87.8583 128.399 83.829 127.466 83.26 126.773C83.1092 126.589 82.9851 126.445 82.8853 126.329C82.6465 126.052 82.5477 125.937 82.5588 125.834C82.5666 125.762 82.6292 125.695 82.736 125.581C82.9176 125.387 83.2271 125.057 83.6124 124.329C84.0789 123.448 84.8503 122.637 85.5922 121.857C87.4998 119.852 89.2126 118.052 85.047 115.792C90.2433 118.611 91.6856 112.553 92.4438 109.368L92.4442 109.367C92.4831 109.203 92.5203 109.047 92.5561 108.9C92.6781 108.399 92.9576 107.684 93.2473 106.944L93.2473 106.943C93.6606 105.887 94.0945 104.777 94.1209 104.162C94.1487 103.515 93.8407 102.827 93.5336 102.141ZM35.2049 10.7594C35.8221 9.07767 30.7452 7.06131 29.3593 6.907C27.9675 6.75203 26.6355 7.15493 25.2986 7.5593C24.267 7.87134 23.2325 8.18425 22.1654 8.24236C22.3053 8.3233 22.4745 8.42328 22.6691 8.53827C25.4886 10.2043 33.6412 15.0216 35.2049 10.7594ZM170.241 42.2852C168.149 41.0246 159.907 37.9651 158.703 40.7345C158.543 41.1022 157.147 39.1975 157.053 38.9936C156.34 37.4432 157.055 35.9806 157.75 34.5585C157.903 34.2455 158.055 33.9345 158.19 33.625C158.396 33.1544 158.634 32.701 158.869 32.251C159.452 31.1382 160.024 30.0456 160.086 28.762C160.128 27.886 159.763 27.0077 159.405 26.1459C159.024 25.2286 158.651 24.3299 158.784 23.4725C159.068 21.6517 159.876 21.3679 160.981 20.9795C161.548 20.7804 162.193 20.5538 162.886 20.0786C164.552 18.9357 164.688 18.3994 164.968 17.302C165.059 16.9428 165.166 16.5234 165.347 16.0029C165.635 15.1723 165.82 14.0545 166.012 12.8888C166.308 11.0947 166.622 9.18718 167.364 8.03935C168.275 6.63041 168.847 6.66436 169.809 6.72141C170.297 6.75039 170.886 6.78532 171.67 6.64024C172.218 6.53893 172.897 6.24195 173.607 5.93095C174.752 5.42972 175.981 4.89204 176.881 5.07823C177.247 5.15403 177.723 5.75253 178.285 6.45963C179.006 7.3683 179.871 8.45632 180.829 8.84499C181.65 9.1777 182.655 9.10434 183.621 9.03383C184.539 8.96677 185.422 8.90229 186.078 9.19206C187.345 9.75175 187.605 10.4556 187.912 11.2844C188.112 11.8269 188.333 12.423 188.868 13.0673C189.314 13.6036 189.807 13.9836 190.292 14.3571C191.025 14.9211 191.738 15.4703 192.238 16.5207C192.549 17.1718 192.522 17.7732 192.497 18.3192C192.453 19.2992 192.418 20.1009 194.354 20.6924C194.297 20.75 194.233 20.815 194.163 20.8868C192.257 22.8344 185.479 29.7598 182.845 29.4035C180.847 29.1334 180.923 27.5159 180.995 25.9985C181.008 25.7073 181.022 25.4199 181.02 25.1464C181.017 24.4787 181.061 23.8063 181.105 23.1345C181.245 20.9839 181.385 18.8401 179.981 16.8811C178.021 14.1451 175.53 13.4624 173.748 16.4625C173.361 17.1157 173.278 17.8272 173.195 18.5427C173.09 19.4396 172.985 20.3429 172.278 21.1457C171.651 21.8565 170.925 22.1522 170.18 22.4556C169.706 22.6484 169.225 22.8443 168.757 23.1518C165.365 25.3799 162.318 28.8016 163.646 32.918C164.593 35.8519 166.244 35.905 168.137 35.966C169.495 36.0097 170.978 36.0574 172.414 37.1772C173.136 37.7401 173.355 38.3049 173.565 38.8485C173.845 39.5706 174.111 40.2552 175.52 40.8478C176.722 41.3531 178.15 41.3014 179.573 41.25C180.659 41.2107 181.742 41.1715 182.718 41.3804C181.807 41.9002 179.368 46.0815 179.869 45.7363C176.311 48.7178 174.341 46.4962 172.47 44.3859C171.741 43.5643 171.027 42.7595 170.241 42.2852Z" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M152.879 130.001C152.456 128.401 148.695 127.698 147.214 128.496C142.468 131.053 153.961 134.086 152.879 130.001ZM111.143 129.625C110.945 128.137 108.105 125.876 106.273 127.219C105.104 126.784 106.835 129.968 107.124 130.264C108.649 131.823 111.533 132.555 111.143 129.625ZM81.3933 128.246C80.831 128.043 80.1572 127.88 79.4492 127.709C77.4232 127.22 75.1176 126.663 74.3457 124.888C73.9508 123.98 74.6102 122.727 75.2675 121.477C76.1333 119.831 76.9953 118.193 75.4372 117.361C73.6377 116.4 73.0901 117.621 72.5441 118.838C72.2603 119.471 71.977 120.103 71.5184 120.427C67.0029 123.617 67.4669 120.714 67.8612 118.247C68.0197 117.256 68.167 116.335 67.9749 115.908C67.5819 115.035 66.6053 114.581 65.6342 114.129C64.8742 113.775 64.1176 113.423 63.6467 112.874C62.8329 111.923 62.8956 110.716 62.9528 109.615C63.0387 107.961 63.1121 106.547 60.1827 106.606C53.8995 106.731 56.6423 111.481 58.8015 115.22C59.6643 116.714 60.4339 118.047 60.4972 118.859C60.602 120.202 59.7984 121.057 59.017 121.889C58.3653 122.582 57.729 123.259 57.648 124.189C57.4993 125.894 62.8611 129.967 63.6059 129.748C64.5246 129.956 66.4522 130.699 64.6679 131.428H77.4267C77.7644 131.428 78.1426 131.451 78.5318 131.475C79.8554 131.556 81.3066 131.645 81.727 130.858C81.9421 130.455 81.9101 128.433 81.3933 128.246ZM171.023 122.117C170.598 124.976 170.19 127.723 174.214 129.36C176.702 130.372 172.104 131.922 170.319 131.428C167.836 130.741 167.241 129.566 166.86 127.248C166.779 126.756 166.86 126.196 166.94 125.635C167.158 124.119 167.376 122.604 164.406 122.46C162.019 122.344 161.882 123.256 162.167 125.082C162.174 125.128 158.245 124.664 158.222 124.637C157.24 123.518 157.823 123.062 158.412 122.603C158.745 122.343 159.08 122.082 159.134 121.699C159.357 120.104 159.473 117.138 156.617 116.248C155.43 115.878 154.349 116.156 153.293 116.426C151.927 116.777 150.602 117.117 149.139 116.03C148.622 115.645 148.33 113.702 148.127 112.35C148.014 111.6 147.929 111.032 147.847 111.014C148.211 110.459 146.84 109.793 145.508 109.146C144.535 108.674 143.581 108.211 143.338 107.808C141.858 105.351 144.864 104.858 147.576 104.414C148.565 104.252 149.515 104.097 150.194 103.855C151.241 103.482 151.993 103.147 152.584 102.884C154.41 102.07 154.69 101.945 157.376 103.532C158.138 103.982 158.962 104.599 159.776 105.206C160.41 105.68 161.037 106.149 161.623 106.53C160.893 106.678 165.432 108.416 165.057 108.231C165.878 108.635 166.405 108.826 166.768 108.956C167.072 109.066 167.261 109.134 167.411 109.252C167.656 109.444 167.797 109.768 168.168 110.62L168.168 110.62C168.252 110.814 168.349 111.035 168.461 111.288C168.541 111.468 168.627 111.662 168.718 111.867L168.718 111.868C169.456 113.524 170.516 115.903 171.032 117.549C171.486 119 171.252 120.575 171.023 122.117ZM2.24997 41.5653C4.89432 40.8564 6.41904 43.3785 7.64178 45.4012C8.4742 46.7781 9.16667 47.9236 9.97717 47.6605C12.1186 49.5573 14.1882 47.823 15.9014 46.3874L15.9016 46.3873C16.5494 45.8444 17.1462 45.3444 17.6767 45.0994C17.5667 45.0286 22.4547 46.2084 21.4609 46.407C28.008 48.6906 28.0862 43.5583 28.1544 39.085C28.1838 37.1553 28.2114 35.3481 28.7556 34.312C29.7373 32.4425 31.1154 31.7243 33.378 30.9084C34.1715 30.6223 35.0455 30.4577 35.9162 30.2938C37.7409 29.9502 39.5518 29.6093 40.5798 28.1579C40.7699 27.8897 40.956 27.0362 41.0861 26.4392C41.1621 26.0907 41.219 25.8296 41.2466 25.8233C40.6654 25.6733 36.5656 24.3368 36.1409 23.977C35.7524 23.6475 35.6524 23.0678 35.5606 22.5361C35.4834 22.0887 35.4121 21.6753 35.1796 21.4738C34.5133 20.8964 33.8519 20.497 33.2197 20.1152C32.1488 19.4686 31.1618 18.8726 30.3769 17.5479C30.0451 16.9877 29.8746 16.3344 29.7059 15.6879C29.296 14.1168 28.8966 12.5861 26.2161 12.5313C24.036 12.4869 22.7456 14.2268 22.074 15.7102C21.772 16.3774 21.8523 18.2336 21.9427 20.3229V20.3231C22.1633 25.4196 22.4439 31.9025 17.3826 25.8966C16.886 25.3073 16.6965 24.9084 16.5375 24.5737C16.2744 24.02 16.0948 23.6418 14.7466 22.868C14.1873 22.547 13.6313 22.5323 13.0933 22.5181C12.3147 22.4975 11.5737 22.478 10.9147 21.5322C10.2332 20.5541 10.5722 19.618 10.7736 19.0619C11.0001 18.4366 11.0525 18.2919 9.28386 19.1086C5.16882 21.0088 5.00742 25.2597 4.87216 28.822C4.84531 29.5293 4.81947 30.2094 4.76395 30.8386C4.56851 33.0547 4.09152 34.3774 3.2014 36.3357C3.1489 36.4512 3 36.6965 2.80283 37.0213L2.80191 37.0228L2.80164 37.0232L2.80145 37.0235C1.81261 38.6521 -0.385406 42.2721 2.24997 41.5653ZM83.3557 5.36786C84.1046 6.44371 85.06 7.15598 86.0662 7.90614L86.0665 7.90633L86.0672 7.90685C86.529 8.25117 87.0015 8.60348 87.4697 9.00262C87.7358 9.22952 88.0406 9.42847 88.3463 9.62797C88.9464 10.0195 89.5496 10.4132 89.8692 11.0245C90.1268 11.5171 90.0341 12.1838 89.9417 12.8485C89.8355 13.6125 89.7296 14.374 90.1564 14.8657C91.3726 16.2671 93.0214 15.932 94.7519 15.5802C97.0624 15.1105 99.5184 14.6112 101.285 18.1766C102.547 20.7251 103.166 24.3959 101.518 26.88C100.824 27.925 99.2947 28.6462 97.792 29.3548L97.792 29.3548C95.9219 30.2365 94.0938 31.0985 93.9727 32.5398C93.9025 33.3746 94.5151 33.9758 95.1513 34.6003C95.7088 35.1475 96.2845 35.7125 96.4349 36.4681C96.5618 37.1052 96.23 37.8683 95.9144 38.594C95.3004 40.0061 94.7478 41.2769 97.7546 41.2039C97.7033 41.184 97.7478 41.1646 97.7922 41.1453C97.8036 41.1403 97.8151 41.1354 97.8249 41.1304C97.8694 41.1489 98.2069 41.0045 98.6084 40.8326C99.1348 40.6074 99.7712 40.335 100.001 40.3213C100.014 40.3239 100.136 40.1967 100.338 39.987L100.338 39.9868L100.338 39.9867L100.339 39.9865C101.263 39.0273 103.847 36.3444 105.154 36.4651C107.726 36.7024 108.734 41.8308 107.344 42.0948C107.934 42.1915 108.379 42.5156 108.839 42.8498C109.651 43.4398 110.506 44.0616 112.273 43.5191C113.658 43.0941 114 42.49 114.454 41.6892C114.603 41.4273 114.763 41.1444 114.976 40.8399C115.641 39.8879 116.014 39.2499 116.253 38.8402L116.253 38.8401C116.501 38.4163 116.606 38.2367 116.743 38.2067C116.845 38.1846 116.963 38.2433 117.169 38.3451C117.553 38.5346 118.238 38.8733 119.675 39.1189L119.726 38.8194L119.726 38.8193L119.726 38.8191C119.925 37.6701 120.002 37.2238 119.85 36.8789C119.753 36.66 119.564 36.4819 119.254 36.1908C119.007 35.9581 118.682 35.6531 118.267 35.1972C117.741 34.6201 117.135 34.089 116.53 33.5589C116.097 33.1802 115.666 32.802 115.265 32.408C114.763 31.9155 110.128 26.9996 111.091 26.7628C110.162 26.5864 110.068 24.319 109.978 22.1438C109.914 20.6187 109.853 19.1389 109.507 18.457C109.162 17.7782 108.608 17.194 108.055 16.6109C107.61 16.1412 107.165 15.6722 106.831 15.1549C105.926 13.7549 105.787 12.1157 105.645 10.4541C105.51 8.86094 105.373 7.24724 104.558 5.80413C104.345 5.42674 103.996 5.11449 103.651 4.80542L103.651 4.80535C103.32 4.50864 102.992 4.21487 102.79 3.86917C102.656 3.64023 102.415 3.42476 102.185 3.21864C102.049 3.09723 101.917 2.97905 101.813 2.86328C101.103 2.07613 101.019 1.66389 100.968 1.4127C100.951 1.33003 100.938 1.26482 100.907 1.20941C100.792 1.00226 100.43 0.932388 98.715 0.601303L98.6912 0.59671C95.5467 -0.0103149 86.5764 0.278534 84.1074 2.34348C83.7534 2.63943 83.3884 3.44151 83.1134 4.04561C82.9298 4.44908 82.7864 4.76424 82.7133 4.7813C83.0757 4.95875 83.1985 5.13831 83.3213 5.31792C83.3327 5.33456 83.3441 5.35121 83.3557 5.36786ZM211.556 115.329C211.036 117.559 212.657 119.405 214.74 120.401C215.81 120.912 215.718 121.852 215.608 122.972C215.572 123.337 215.534 123.721 215.534 124.116C213.06 124.106 208.124 121.279 206.57 119.292C204.935 117.201 205.439 116.188 206.205 114.65C206.435 114.188 206.688 113.679 206.914 113.079C207.352 111.919 207.517 110.746 207.668 109.669C208.064 106.855 208.368 104.697 213.198 105.128C213.19 105.675 212.87 106.002 212.533 106.346C212.079 106.81 211.596 107.303 211.809 108.408C211.944 109.109 212.435 109.497 212.921 109.881C213.612 110.427 214.295 110.967 213.938 112.392C213.8 112.941 213.296 113.365 212.784 113.797C212.247 114.249 211.7 114.709 211.556 115.329ZM53.6192 7.43633C54.4734 7.53318 55.1021 7.81554 55.7173 8.09186C56.5651 8.47263 57.3874 8.84192 58.7389 8.69836C59.8923 8.57599 60.7736 8.01775 61.6573 7.45801C62.1727 7.13155 62.6889 6.80458 63.2604 6.56329C64.6529 5.97527 67.0625 5.72145 69.6898 5.4447C74.9711 4.88839 81.1322 4.23942 81.6796 0.59671C80.7875 0.59671 79.427 0.510544 77.8495 0.410629H77.8493H77.8491H77.8488H77.8486H77.8483H77.848H77.8477H77.8476H77.8474C72.9741 0.102158 66.0301 -0.337387 64.4172 1.22287C63.5916 2.02158 63.7048 2.69177 63.8004 3.25757C63.9377 4.06986 64.0385 4.66698 61.273 5.12029C60.3701 5.2683 59.7092 5.03156 59.0509 4.79576C58.0434 4.43488 57.0421 4.07622 55.1891 5.10268C54.9289 5.24676 51.5806 7.20537 53.6192 7.43633ZM205.373 37.422C205.947 35.9763 206.859 33.682 208.312 33.5851C208.829 34.0427 209.49 35.0123 210.078 35.8753C210.585 36.6192 211.039 37.284 211.299 37.4734C212.515 38.3579 212.854 38.3194 213.805 38.2112C214.222 38.1638 214.757 38.103 215.534 38.1006C215.534 40.3344 214.919 41.4736 214.214 42.7802C213.747 43.6446 213.241 44.5823 212.847 45.9587C212.807 46.0999 212.773 46.3723 212.731 46.7164C212.521 48.4149 212.096 51.8589 209.574 49.8777C208.954 49.3902 208.934 47.9012 208.912 46.1463C208.871 42.9162 208.818 38.7853 204.981 38.3393C205.09 38.1322 205.22 37.8072 205.373 37.422ZM145.332 4.08382C145.332 7.08354 146.613 14.3215 151.36 9.26977C152.529 8.0257 153.476 3.03751 152.883 1.8139C151.097 -1.87186 145.333 0.658051 145.332 4.08382ZM3.82394 126.743C3.19591 125.77 2.59644 124.841 1.96149 124.378C1.37094 123.948 1.4511 122.643 1.51311 121.633C1.53015 121.356 1.54582 121.101 1.54582 120.892C2.37483 120.85 3.29062 121.294 4.22887 121.749C5.55847 122.394 6.93317 123.06 8.16989 122.395C10.1133 121.349 8.88205 120.759 7.64636 120.166C6.9781 119.846 6.30855 119.525 6.13914 119.131C4.23973 114.712 8.02139 117.119 9.54181 118.087C9.85698 118.287 10.075 118.426 10.1251 118.429C9.67448 119.039 9.83195 120.677 10.0108 122.537C10.353 126.097 10.7735 130.471 7.16086 130.013C5.8251 129.844 4.79068 128.241 3.82394 126.743ZM3.14038 81.8464L3.14047 81.8463C3.52681 81.3065 3.93803 80.7319 4.34192 80.4192C4.61201 80.2101 5.15619 80.3615 5.63535 80.4948C6.00546 80.5978 6.33678 80.69 6.47301 80.5968C6.69826 80.4431 6.71269 80.1951 6.72667 79.9551C6.73978 79.7299 6.75249 79.5117 6.93848 79.3848C7.45216 79.0342 9.48564 77.4257 9.46944 76.8453C9.43092 75.4722 8.11501 75.1508 6.804 74.8306C5.85498 74.5988 4.90854 74.3676 4.45105 73.7386C4.32793 73.5693 4.39693 72.823 4.48547 71.8652C4.73184 69.2007 5.12949 64.8998 1.96188 66.8359C1.30722 66.6299 1.92499 83.1651 2.10753 83.0534C2.40776 82.87 2.76201 82.3751 3.14015 81.8468L3.14024 81.8466L3.14038 81.8464ZM183.318 126.525C183.639 124.472 183.942 122.533 186.682 123.266C186.892 123.322 186.591 124.115 186.229 125.066L186.229 125.066C185.684 126.5 185.002 128.295 185.73 128.467C185.391 128.527 185.354 128.955 185.319 129.35C185.302 129.548 185.286 129.737 185.233 129.867C184.536 131.58 184.367 131.571 182.785 131.483C182.339 131.459 181.779 131.428 181.063 131.428C181.161 131.385 180.55 130.611 180.411 130.487C182.708 130.437 183.021 128.432 183.318 126.525ZM180.387 130.487L180.411 130.487C180.387 130.465 180.377 130.463 180.387 130.487ZM169.506 3.81471C169.602 3.02972 169.577 2.24457 169.431 1.45924C169.293 1.16638 169.141 0.878891 168.976 0.596603C169.076 0.596603 169.2 0.59549 169.343 0.594208H169.344C171.334 0.57634 176.996 0.525513 173.854 2.97186C172.885 3.72615 170.795 3.83333 169.506 3.81471ZM182.092 1.19322C182.104 1.20111 185.402 3.03207 185.525 2.60217C186.184 0.29248 184.067 0.427383 182.198 0.546448H182.197H182.197H182.197H182.196H182.196L182.196 0.546494L182.193 0.546677C181.795 0.572006 181.409 0.596603 181.063 0.596603C181.479 0.68634 181.822 0.885101 182.092 1.19322ZM38.0861 68.9636C38.5655 69.8006 39.0517 70.6494 40.1469 71.147C43.0563 72.4692 43.1138 73.0858 43.2113 74.1297C43.2688 74.7452 43.3401 75.5092 44.018 76.654C45.6666 79.438 49.8316 79.2452 53.1337 79.0923C53.8055 79.0612 54.4417 79.0318 55.0136 79.0294C55.7454 79.0264 56.2747 79.0607 56.6873 79.0874C57.9934 79.1721 58.1305 79.1809 59.8229 77.6907C60.5126 77.0834 61.1896 76.3695 61.8836 75.6379L61.8838 75.6376C62.6642 74.8148 63.4661 73.9693 64.3315 73.2274C66.1871 71.6366 70.4901 67.6082 67.3462 65.3832C68.8834 65.0318 68.4058 62.3595 67.3298 61.7671C67.4593 61.7161 68.7864 58.5247 68.7176 58.3505C69.4128 58.3929 70.0754 58.5686 70.7126 58.7376C72.1346 59.1146 73.4304 59.4582 74.6808 58.2127C75.7556 57.1424 75.4058 55.5892 75.0553 54.0333C74.7604 52.7241 74.4651 51.413 75.0176 50.3861C76.5702 47.5001 81.6524 47.2237 84.3185 48.9518C86.1444 50.1355 90.3172 53.9595 91.4789 55.7235C93.463 58.7365 92.6765 61.4331 88.5254 61.3241C87.5229 61.2978 86.8001 60.8602 86.2735 60.5415C85.4848 60.0641 85.1366 59.8533 84.9493 61.6908C84.8669 62.4992 85.7111 63.5355 86.3982 64.3789L86.3983 64.379C86.655 64.6941 86.8898 64.9822 87.046 65.2215C87.1568 65.3911 87.2641 65.5544 87.3678 65.7122C89.1404 68.4088 89.859 69.502 88.5225 72.7946C88.4555 72.9597 88.3909 73.1175 88.3289 73.2689C87.37 75.6117 87.0387 76.4212 87.9487 79.101C88.9956 82.1846 89.2863 83.4852 89.0381 86.7443C89.0269 86.8914 89.0065 87.0631 88.9841 87.2517C88.823 88.6082 88.5575 90.8435 90.8535 91.1828C90.5021 91.7108 89.1881 95.5624 89.5067 95.9254C87.5491 95.6898 87.2363 96.0578 86.6567 96.7396C86.2892 97.1719 85.8144 97.7304 84.7451 98.3413C84.0034 98.7649 83.0614 98.8688 82.1088 98.9737C80.7778 99.1205 79.4261 99.2694 78.5714 100.296C77.7922 101.232 77.7607 102.75 77.7297 104.241C77.7069 105.339 77.6844 106.422 77.364 107.249C76.7754 108.768 75.9893 110.654 73.9984 111.336C69.1997 112.979 69.5801 108.755 69.8472 105.789L69.8472 105.788C69.893 105.281 69.9354 104.81 69.9479 104.411C70.106 99.3708 70.1826 94.0119 69.2204 89.1469C68.2439 84.208 67.4279 86.0991 66.3232 88.6594C66.1113 89.1506 65.8887 89.6665 65.6523 90.1635C65.2352 91.0402 64.5471 91.6825 63.8623 92.3217C63.2235 92.9179 62.5876 93.5114 62.1773 94.29C61.8686 94.8759 61.8962 95.4841 61.923 96.0728C61.96 96.8874 61.9953 97.6645 61.1357 98.2933C60.4906 98.7652 59.7251 98.6059 59.0262 98.4605C58.4484 98.3403 57.9162 98.2296 57.5354 98.4927C56.7952 99.004 56.6855 100.431 56.5723 101.903C56.4108 104.005 56.2421 106.198 54.2227 105.954C52.244 105.714 52.0615 103.653 51.908 101.92C51.7995 100.695 51.7055 99.6334 51.0015 99.4953C51.3779 99.392 51.8318 99.284 52.3176 99.1684C55.4056 98.4337 59.7832 97.3921 53.7375 95.3088C55.5941 95.1286 55.055 92.8618 54.807 91.8189L54.8067 91.8178C54.7887 91.7419 54.7721 91.6725 54.7582 91.6108C54.6908 91.3117 54.3423 91.0642 53.9914 90.815C53.6254 90.5551 53.2568 90.2933 53.2017 89.9691C53.0795 89.2477 53.2722 88.888 53.4632 88.5314C53.6027 88.271 53.7412 88.0123 53.7556 87.6158C53.7747 87.0875 53.8259 86.5892 53.872 86.1404L53.872 86.1401V86.1399C54.0752 84.1614 54.1797 83.143 51.0066 84.7409C50.1116 85.1915 49.9051 85.8104 49.7432 86.2958C49.5176 86.972 49.3785 87.3887 47.5842 86.729C47.5162 86.5391 47.3553 86.4251 47.1015 86.387C46.834 86.4666 46.3999 86.1661 45.8871 85.8112C45.1075 85.2714 44.1461 84.6059 43.3128 84.9601C41.6548 85.6646 42.2787 86.8613 42.8436 87.9448C43.0534 88.3472 43.255 88.7341 43.3286 89.0742C43.8844 91.645 43.5274 91.8164 42.7232 92.2023C42.1489 92.4779 41.3466 92.8629 40.486 94.3093C39.776 95.5028 40.2345 96.6938 40.6641 97.8096C41.232 99.2848 41.7492 100.628 39.4485 101.672C33.5859 104.331 32.6706 98.1511 33.4078 95.2095L33.494 94.8668C34.2974 91.6742 34.8962 89.2945 33.3785 86.0022C32.3378 83.7448 30.6298 82.4108 28.7989 80.9809C27.8772 80.261 26.9243 79.5168 26.0096 78.6181C22.5714 75.2405 22.2929 71.2149 26.2179 67.9774C29.3604 65.3854 33.5826 64.6167 36.865 67.3331C37.4109 67.7849 37.7468 68.3713 38.0861 68.9636ZM129.865 21.9299C130.306 22.1423 130.895 22.9289 131.504 23.742C132.063 24.4887 132.639 25.2577 133.132 25.6251C133.738 26.0755 134.385 26.4183 135.02 26.7544C136.075 27.3123 137.094 27.8515 137.823 28.8331C140.874 32.9387 138.627 35.6727 135.947 38.935L135.856 39.0448C132.954 42.5778 131.378 53.0557 134.517 56.4581C135.017 56.9999 135.802 57.4103 136.622 57.8391C137.526 58.3118 138.472 58.8069 139.127 59.5248C139.507 59.9416 139.754 60.4395 140 60.9382C140.23 61.403 140.46 61.8686 140.8 62.2699C142.02 63.7121 144.822 66.5803 146.939 66.7005C147.376 64.2453 148.209 63.3885 150.661 62.0445C153.952 63.1612 155.536 56.0061 155.377 55.1279C155.057 53.3621 154.012 50.9436 151.482 50.8797C149.171 50.8214 148.719 52.0326 148.267 53.2417C147.975 54.0209 147.685 54.7992 146.897 55.2364C144.439 56.6007 144.054 55.5035 143.441 53.7574L143.441 53.757C143.343 53.4788 143.239 53.1842 143.121 52.8804C142.883 52.2718 142.789 51.7668 142.702 51.3042C142.55 50.4937 142.422 49.8134 141.594 48.9347C141.143 48.4562 140.47 48.1659 139.792 47.8737C138.948 47.5098 138.097 47.1429 137.66 46.4059C135.588 42.9139 138.752 40.4144 141.647 38.1277C142.94 37.1059 144.18 36.1266 144.874 35.1202C144.697 35.0718 151.318 33.1256 151.905 33.0241C152.284 32.9585 152.767 32.9031 153.289 32.8431C155.627 32.5748 158.771 32.2139 157.1 30.4281C156.47 29.7541 155.451 29.5391 154.39 29.3154C153.169 29.0578 151.893 28.7885 151.09 27.7939C150.061 26.5192 150.324 25.0782 150.582 23.6646C150.791 22.5215 150.997 21.3964 150.513 20.3918C149.09 20.4196 147.744 20.9393 146.419 21.4507C144.068 22.3582 141.785 23.2397 139.26 21.2988C137.536 19.9742 137.693 18.2417 137.849 16.5154C137.975 15.1301 138.1 13.7487 137.252 12.5854C136.861 12.0484 136.721 11.7852 136.659 11.6676C136.627 11.6086 136.615 11.5863 136.601 11.5846C136.593 11.5836 136.583 11.5901 136.568 11.6005C136.489 11.6537 136.26 11.8095 135.231 11.5892C132.196 10.9396 132.078 11.3798 131.784 12.4698C131.689 12.8242 131.575 13.2474 131.336 13.7242C130.871 14.6536 128.819 21.4264 129.865 21.9299ZM53.5993 38.4416C50.1751 39.3073 48.6092 39.1887 44.0752 38.572C43.5246 38.5198 42.9723 38.504 42.418 38.5243C43.0569 38.0351 47.1988 32.9202 47.0992 32.8546C47.2262 32.9296 47.1227 33.5212 47.0091 34.1703C46.8793 34.9115 46.7364 35.7276 46.9088 35.9352C47.6068 36.7761 49.4835 37.1047 51.1645 37.3991C52.1059 37.5639 52.9859 37.718 53.5632 37.9453C53.3681 38.0502 53.38 38.2155 53.5993 38.4416ZM122.334 25.1478C122.362 25.1592 122.492 25.2337 122.675 25.3391C123.291 25.6923 124.511 26.3926 124.523 26.2262C124.565 25.614 125.146 24.929 125.71 24.2645C126.748 23.0405 127.727 21.8858 125.171 21.3824C123.022 20.9592 121.05 23.5227 120.803 24.9056C121.34 24.8833 121.851 24.964 122.334 25.1478ZM161.273 75.9077C161.092 75.609 158.782 73.3479 158.5 73.2317C155.914 72.1639 160.963 66.6254 162.798 64.613C163.026 64.3629 163.204 64.1673 163.312 64.0415C163.89 63.3681 164.98 61.119 165.965 59.0879C166.894 57.1705 167.73 55.4474 167.951 55.4271C168.568 55.3705 170.003 65.1496 170.76 70.3098L170.76 70.31L170.76 70.3103L170.76 70.3104C170.979 71.7997 171.141 72.9042 171.21 73.2761C171.242 73.4474 171.276 73.6205 171.309 73.7949L171.31 73.796C171.753 76.098 172.244 78.6454 170.466 80.5707C168.906 82.2585 165.009 83.618 162.907 81.9278C161.811 81.0456 161.794 79.8136 161.777 78.576C161.765 77.6403 161.752 76.7013 161.273 75.9077ZM202.653 21.3326C202.869 22.2556 204.427 24.6171 205.593 24.6495C205.65 24.2228 206.375 22.8764 207.101 21.5306L207.101 21.5305L207.101 21.5304C207.907 20.034 208.713 18.5382 208.599 18.308C207.454 15.9995 202.025 18.6371 202.653 21.3326ZM122.068 30.509C122.702 30.9154 123.363 31.3392 123.744 31.8846C124.46 32.9102 123.656 34.88 122.018 34.8255C122.148 34.2942 121.871 33.4434 121.59 32.5817C121.253 31.5473 120.911 30.4973 121.262 29.9656C121.507 30.1495 121.785 30.3275 122.068 30.509ZM179.95 79.3278L179.95 79.3278C180.485 80.7586 180.999 82.133 181.273 82.5831C181.673 83.2405 182.141 83.887 182.61 84.536L182.61 84.536C183.355 85.5653 184.104 86.6008 184.594 87.6958C185.793 90.3765 185.304 92.4241 184.799 94.5368C184.431 96.0774 184.055 97.6525 184.319 99.533C184.997 104.361 188.435 102.341 189.99 99.1133C188.625 98.93 191.936 92.4361 193.581 91.0966C193.893 90.8425 194.347 90.4007 194.89 89.8713C197.019 87.7969 200.526 84.3786 202.302 85.6541C203.923 86.818 203.007 88.1729 202.102 89.512C201.468 90.4497 200.839 91.3797 201.09 92.2311C201.143 92.4108 201.715 92.2402 202.164 92.1062C202.452 92.0203 202.69 91.9494 202.708 91.9954C202.723 92.0326 202.825 92.0498 202.966 92.0735C203.294 92.1289 203.834 92.2199 203.97 92.6852C204.064 93.0098 203.896 93.3879 203.737 93.7471C203.598 94.0582 203.467 94.3551 203.517 94.5908C203.88 96.2921 203.655 97.859 203.42 99.4984C203.287 100.425 203.151 101.375 203.116 102.386C202.917 108.092 204.559 104.988 205.725 102.784C206.16 101.962 206.529 101.265 206.711 101.196C204.681 99.5831 205.311 93.8031 208.467 93.1645C209.773 92.9002 210.8 93.4841 211.758 94.0292C212.982 94.7251 214.094 95.3576 215.534 94.0812C216.086 93.5914 215.632 86.4284 215.324 86.1726C214.857 85.7833 213.826 85.6463 212.879 85.5204C212.302 85.4438 211.756 85.3712 211.389 85.2482C210.066 84.8059 204.779 83.0221 208.607 82.2534C209.5 82.0741 210.455 82.5798 211.413 83.0867C212.495 83.6593 213.58 84.2334 214.581 83.824C216.522 83.0297 216.01 75.1484 215.695 70.2909C215.607 68.9366 215.534 67.8174 215.534 67.1523C213.54 67.1976 211.889 70.0247 211.34 71.4041C209.781 75.3182 209.148 76.453 204.321 78.8259C204.208 78.7866 200.993 80.4477 200.766 80.565L200.765 80.5651L200.765 80.5652L200.753 80.5715C200.453 80.7524 199.988 80.7424 199.547 80.7329C199.081 80.7229 198.643 80.7135 198.453 80.9291C198.202 81.2136 198.358 81.8469 198.534 82.5649C198.768 83.5176 199.039 84.6193 198.445 85.2523C196.411 87.4216 195.053 86.0277 193.495 84.4291L193.495 84.4291C193.295 84.2244 193.092 84.0163 192.885 83.8119C192.665 83.5962 192.388 83.3376 192.084 83.0537L192.084 83.0536C191.061 82.0994 189.732 80.8591 189.249 79.9949C187.589 77.0262 189.312 76.5181 191.268 75.9414C192.823 75.4828 194.526 74.9809 194.79 73.1636C194.983 71.8379 193.256 70.9775 191.509 70.1072L191.509 70.1072C190.555 69.6322 189.595 69.1543 188.939 68.5961C188.492 68.216 188.173 67.75 187.853 67.2843C187.292 66.4651 186.731 65.6467 185.477 65.2975C180.613 63.9432 174.464 73.7562 178.231 75.3651C178.513 75.4855 179.249 77.455 179.95 79.3278ZM101.312 130.525C101.008 127.237 95.5213 130.413 93.957 131.318C93.885 131.36 93.8213 131.397 93.7668 131.428L101.377 131.428C101.359 131.127 101.337 130.826 101.312 130.525ZM160.726 130.89C161.251 130.725 163.045 128.219 163.28 127.697C163.202 127.869 163.169 128.128 163.132 128.424C163.005 129.418 162.823 130.843 160.726 130.89Z" fill="black" fill-opacity="0.1"/>
`;

        }, {}],
        44: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M15 0H11V18H15V0ZM23 0H19V30V34H23H72.416C73.1876 35.7659 74.9497 37 77 37C79.7614 37 82 34.7614 82 32C82 29.2386 79.7614 27 77 27C74.9497 27 73.1876 28.2341 72.416 30H23V0Z" fill="black" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M122 34.584C123.766 33.8124 125 32.0503 125 30C125 27.2386 122.761 25 120 25C117.239 25 115 27.2386 115 30C115 32.0503 116.234 33.8124 118 34.584V60V64H122H141V60H122V34.584Z" fill="white" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M114 46H110V68V72H114H141V68H114V46Z" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M27 103.584C25.2341 102.812 24 101.05 24 99C24 96.2386 26.2386 94 29 94C31.7614 94 34 96.2386 34 99C34 101.05 32.7659 102.812 31 103.584V129V133H27H8V129H27V103.584Z" fill="black" fill-opacity="0.2"/>
`;

        }, {}],
        45: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M0.880859 128.092C5.24177 97.4016 13.9384 123.534 34.866 118.124C44.1119 115.734 36.6414 105.296 43.5373 100.772C49.3622 96.9503 57.3002 100.029 63.9726 97.5579C73.6018 93.9913 73.4753 86.7354 82.4474 85.651C90.9322 84.6255 99.8208 88.005 108.378 86.7354C124.529 84.3395 117.907 52.7111 130.018 47.1252C142.129 41.5393 162.068 51.1858 164.678 67C167.222 82.4167 150.22 139.78 150.22 139.78H0.880859" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M45.9268 128C83.6043 128 87.5268 113.25 106.065 106.493C124.603 99.7354 140.846 117.214 147.078 100.233C153.31 83.2528 153.276 136.181 153.276 136.181H45.9268L45.9268 128Z" fill="black" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M7.99964 10.8673C19.7328 9.74652 15.9616 22.8061 20.597 22.8061C22.4461 22.8061 24.6249 14.3489 27.9829 14.3489C30.4376 14.3489 29.9749 19.8425 35.3046 19.2617C40.3073 18.7166 39.375 15.2687 41.9023 15.2687C47.345 15.2687 45.2869 35.4636 49.6437 35.4636C54.0005 35.4636 55.3734 20.668 57.2911 14.8164C59.7177 7.4119 74.2632 3.26163 63.8919 0.320129H7.71289" fill="black" fill-opacity="0.4"/>
`;

        }, {}],
        46: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = () => `
    <path fill-rule="evenodd" clip-rule="evenodd" d="M120 130.649C116.812 113.308 110.453 128.073 95.1523 125.017C88.3923 123.666 93.8542 117.769 88.8124 115.213C84.5536 113.054 78.7499 114.793 73.8715 113.397C66.8312 111.382 66.9237 107.282 60.3639 106.669C54.1604 106.09 47.6616 107.999 41.4049 107.282C29.5967 105.928 34.4383 88.0585 25.5835 84.9025C16.7287 81.7465 2.15043 87.1967 0.242157 96.1317C-1.61816 104.842 10.8129 137.252 10.8129 137.252H120" fill="black" fill-opacity="0.2"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M156 122.767C104.757 122.767 99.4219 96.9746 74.2092 85.1582C48.9964 73.3418 26.9059 103.905 18.4298 74.2122C9.9537 44.5195 10.0005 137.072 10.0005 137.072H156V122.767Z" fill="black" fill-opacity="0.1"/>
    <path fill-rule="evenodd" clip-rule="evenodd" d="M161.575 9.17587C149.64 8.20079 156.546 19.9843 151.83 19.9843C144.863 19.9843 149.727 12.0062 141.247 12.2048C139.028 12.2568 136.303 14.1443 136.135 16.709C135.773 22.2193 139.852 31.7253 134.297 31.9953C129.117 32.247 133.193 22.3562 132.652 15.7593C132.343 11.9871 128.47 9.00745 125.684 9.00745C118.113 9.00745 123.998 24.7147 118.549 24.7147C113.1 24.7147 117.341 15.6639 113.467 10.8715C108.964 5.29973 94.1677 2.55906 104.718 0H161.867" fill="black" fill-opacity="0.4"/>
`;

        }, {}],
        47: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const camo_01_1 = __importDefault(require("./camo-01"));
            const camo_02_1 = __importDefault(require("./camo-02"));
            const circuits_1 = __importDefault(require("./circuits"));
            const dirty_01_1 = __importDefault(require("./dirty-01"));
            const dirty_02_1 = __importDefault(require("./dirty-02"));
            exports.default = [camo_01_1.default, camo_02_1.default, circuits_1.default, dirty_01_1.default, dirty_02_1.default];

        }, {"./camo-01": 42, "./camo-02": 43, "./circuits": 44, "./dirty-01": 45, "./dirty-02": 46}],
        48: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M53.5683 39L55.5437 34.3851L49.3528 23.7098L52.2483 13.0836L49.3539 12.2949L46.1288 24.1305L52.179 34.5631L50.0881 39H38V52H62V39H53.5683Z" fill="#E6E6E6"/>
        <mask id="topAntennaCrookedMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="38" y="12" width="24" height="40">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M53.5683 39L55.5437 34.3851L49.3528 23.7098L52.2483 13.0836L49.3539 12.2949L46.1288 24.1305L52.179 34.5631L50.0881 39H38V52H62V39H53.5683Z" fill="white"/>
        </mask>
        <g mask="url(#topAntennaCrookedMask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <rect x="38" y="39" width="24" height="13" fill="white" fill-opacity="0.2"/>
        </g>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M50 16C54.4183 16 58 12.4183 58 8C58 3.58172 54.4183 0 50 0C45.5817 0 42 3.58172 42 8C42 12.4183 45.5817 16 50 16Z" fill="#FFE65C"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M53 8C54.6569 8 56 6.65685 56 5C56 3.34315 54.6569 2 53 2C51.3431 2 50 3.34315 50 5C50 6.65685 51.3431 8 53 8Z" fill="white"/>
    `;
            };

        }, {}],
        49: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M52 5H48V36H40C38.8954 36 38 36.8954 38 38V52H62V38C62 36.8954 61.1046 36 60 36H52V5Z" fill="#E1E6E8"/>
        <mask id="topAntennaMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="38" y="5" width="24" height="47">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M52 5H48V36H40C38.8954 36 38 36.8954 38 38V52H62V38C62 36.8954 61.1046 36 60 36H52V5Z" fill="white"/>
        </mask>
        <g mask="url(#topAntennaMask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <rect x="38" y="36" width="24" height="16" fill="white" fill-opacity="0.2"/>
        </g>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M50 16C54.4183 16 58 12.4183 58 8C58 3.58172 54.4183 0 50 0C45.5817 0 42 3.58172 42 8C42 12.4183 45.5817 16 50 16Z" fill="#FFE65C"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M53 8C54.6569 8 56 6.65685 56 5C56 3.34315 54.6569 2 53 2C51.3431 2 50 3.34315 50 5C50 6.65685 51.3431 8 53 8Z" fill="white"/>
    `;
            };

        }, {}],
        50: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M48 0C39.1634 0 32 7.16344 32 16V32C32 36.4183 35.5817 40 40 40H23C22.4477 40 22 40.4477 22 41V51C22 51.5523 22.4477 52 23 52H77C77.5523 52 78 51.5523 78 51V41C78 40.4477 77.5523 40 77 40H60C64.4183 40 68 36.4183 68 32V16C68 7.16344 60.8366 0 52 0H48Z" fill="#59C4FF"/>
        <mask id="topBulb011Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="22" y="0" width="56" height="52">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M48 0C39.1634 0 32 7.16344 32 16V32C32 36.4183 35.5817 40 40 40H23C22.4477 40 22 40.4477 22 41V51C22 51.5523 22.4477 52 23 52H77C77.5523 52 78 51.5523 78 51V41C78 40.4477 77.5523 40 77 40H60C64.4183 40 68 36.4183 68 32V16C68 7.16344 60.8366 0 52 0H48Z" fill="white"/>
        </mask>
        <g mask="url(#topBulb011Mask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <rect x="20" y="-3" width="60" height="43" fill="white" fill-opacity="0.4"/>
            <path d="M49 3.5C53.9315 3.5 58.366 5.62814 61.4352 9.01616" stroke="white" stroke-width="2" stroke-linecap="round"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M49.8284 26L40.8284 17L38 19.8284L48 29.8284V40H52V29.9706L62.1421 19.8284L59.3137 17L50.3137 26H49.8284Z" fill="white" fill-opacity="0.8"/>
        </g>
    `;
            };

        }, {}],
        51: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M50 13C38.9543 13 30 21.9543 30 33V36H21C20.4477 36 20 36.4477 20 37V51C20 51.5523 20.4477 52 21 52H79C79.5523 52 80 51.5523 80 51V37C80 36.4477 79.5523 36 79 36H70V33C70 21.9543 61.0457 13 50 13Z" fill="#59C4FF"/>
        <mask id="topBulb01Mask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="20" y="13" width="60" height="39">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M50 13C38.9543 13 30 21.9543 30 33V36H21C20.4477 36 20 36.4477 20 37V51C20 51.5523 20.4477 52 21 52H79C79.5523 52 80 51.5523 80 51V37C80 36.4477 79.5523 36 79 36H70V33C70 21.9543 61.0457 13 50 13Z" fill="white"/>
        </mask>
        <g mask="url(#topBulb01Mask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M50 36C52.2091 36 54 35.028 54 31.7143C54 28.4006 52.2091 24 50 24C47.7909 24 46 28.4006 46 31.7143C46 35.028 47.7909 36 50 36Z" fill="white" fill-opacity="0.6"/>
            <rect x="20" y="13" width="60" height="23" fill="white" fill-opacity="0.4"/>
            <path d="M50 14.5C49.4477 14.5 49 14.9477 49 15.5C49 16.0523 49.4477 16.5 50 16.5V14.5ZM61.6941 21.6875C62.0649 22.0968 62.6973 22.1281 63.1066 21.7573C63.5159 21.3865 63.5471 20.7541 63.1763 20.3448L61.6941 21.6875ZM65.7595 24.0473C65.5035 23.5579 64.8993 23.3686 64.4099 23.6246C63.9205 23.8806 63.7313 24.4848 63.9873 24.9742L65.7595 24.0473ZM65.4248 28.9559C65.5404 29.4959 66.0719 29.84 66.6119 29.7244C67.152 29.6088 67.4961 29.0773 67.3805 28.5373L65.4248 28.9559ZM50 16.5C54.6375 16.5 58.8065 18.4999 61.6941 21.6875L63.1763 20.3448C59.9256 16.7563 55.2256 14.5 50 14.5V16.5ZM63.9873 24.9742C64.6357 26.2139 65.1239 27.5501 65.4248 28.9559L67.3805 28.5373C67.0411 26.9518 66.4904 25.4448 65.7595 24.0473L63.9873 24.9742Z" fill="white"/>
        </g>
    `;
            };

        }, {}],
        52: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <g filter="url(#topGlowingBulb01Filter0)">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M32 24C32 15.1634 39.1634 8 48 8H52C60.8366 8 68 15.1634 68 24V32C68 36.4183 64.4183 40 60 40H40C35.5817 40 32 36.4183 32 32V24Z" fill="white" fill-opacity="0.3"/>
        </g>
        <path d="M49 11.5C53.9315 11.5 58.366 13.6281 61.4352 17.0162" stroke="white" stroke-width="2" stroke-linecap="round"/>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M49.8284 29L40.8284 20L38 22.8284L48 32.8284V40H52V32.9706L62.1421 22.8284L59.3137 20L50.3137 29H49.8284Z" fill="white" fill-opacity="0.8"/>
        <rect x="22" y="40" width="56" height="12" rx="1" fill="${color.hex}"/>
        <defs>
            <filter id="topGlowingBulb01Filter0" x="24" y="0" width="52" height="48" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
                <feFlood flood-opacity="0" result="BackgroundImageFix"/>
                <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"/>
                <feOffset/>
                <feGaussianBlur stdDeviation="4"/>
                <feColorMatrix type="matrix" values="0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 0 0.5 0"/>
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow"/>
                <feBlend mode="normal" in="SourceGraphic" in2="effect1_dropShadow" result="shape"/>
                <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha"/>
                <feOffset/>
                <feGaussianBlur stdDeviation="2"/>
                <feComposite in2="hardAlpha" operator="arithmetic" k2="-1" k3="1"/>
                <feColorMatrix type="matrix" values="0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 0 0.5 0"/>
                <feBlend mode="normal" in2="shape" result="effect2_innerShadow"/>
            </filter>
        </defs>
    `;
            };

        }, {}],
        53: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <g filter="url(#topGlowingBulb02Filter0)">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M30 33C30 21.9543 38.9543 13 50 13V13C61.0457 13 70 21.9543 70 33V44H30V33Z" fill="white" fill-opacity="0.3"/>
        </g>
        <path fill-rule="evenodd" clip-rule="evenodd" d="M50 36C52.2091 36 54 35.028 54 31.7143C54 28.4006 52.2091 24 50 24C47.7909 24 46 28.4006 46 31.7143C46 35.028 47.7909 36 50 36Z" fill="white" fill-opacity="0.6"/>
        <path d="M50 14.5C49.4477 14.5 49 14.9477 49 15.5C49 16.0523 49.4477 16.5 50 16.5V14.5ZM61.6941 21.6875C62.0649 22.0968 62.6973 22.1281 63.1066 21.7573C63.5159 21.3865 63.5471 20.7541 63.1763 20.3448L61.6941 21.6875ZM65.7595 24.0473C65.5035 23.5579 64.8993 23.3686 64.4099 23.6246C63.9205 23.8806 63.7313 24.4848 63.9873 24.9742L65.7595 24.0473ZM65.4248 28.9559C65.5404 29.4959 66.0719 29.84 66.6119 29.7244C67.152 29.6088 67.4961 29.0773 67.3805 28.5373L65.4248 28.9559ZM50 16.5C54.6375 16.5 58.8065 18.4999 61.6941 21.6875L63.1763 20.3448C59.9256 16.7563 55.2256 14.5 50 14.5V16.5ZM63.9873 24.9742C64.6357 26.2139 65.1239 27.5501 65.4248 28.9559L67.3805 28.5373C67.0411 26.9518 66.4904 25.4448 65.7595 24.0473L63.9873 24.9742Z" fill="white"/>
        <rect x="20" y="36" width="60" height="16" rx="1" fill="${color.hex}"/>
        <defs>
            <filter id="topGlowingBulb02Filter0" x="22" y="5" width="56" height="47" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
                <feFlood flood-opacity="0" result="BackgroundImageFix"/>
                <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"/>
                <feOffset/>
                <feGaussianBlur stdDeviation="4"/>
                <feColorMatrix type="matrix" values="0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 0 0.5 0"/>
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow"/>
                <feBlend mode="normal" in="SourceGraphic" in2="effect1_dropShadow" result="shape"/>
                <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha"/>
                <feOffset/>
                <feGaussianBlur stdDeviation="2"/>
                <feComposite in2="hardAlpha" operator="arithmetic" k2="-1" k3="1"/>
                <feColorMatrix type="matrix" values="0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 0 0.5 0"/>
                <feBlend mode="normal" in2="shape" result="effect2_innerShadow"/>
            </filter>
        </defs>
    `;
            };

        }, {}],
        54: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M71.2104 40C78.8499 33.2931 84.6313 20.6882 84 14C83.8635 12.5535 85.9998 12.2993 87 14C91.418 21.5124 89.7172 36.0672 89.1535 40H92V52H66V40H71.2104ZM16.521 13.7408C16.521 21.2733 21.4918 33.445 29.2618 40H34V52H8V40H11.2251C10.6299 36.4414 8.52929 21.6012 13.4337 14.1009C14.3353 12.7219 16.521 12.6807 16.521 13.7408Z" fill="#E1E6E8"/>
        <mask id="topHornsMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="8" y="12" width="84" height="40">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M71.2104 40C78.8499 33.2931 84.6313 20.6882 84 14C83.8635 12.5535 85.9998 12.2993 87 14C91.418 21.5124 89.7172 36.0672 89.1535 40H92V52H66V40H71.2104ZM16.521 13.7408C16.521 21.2733 21.4918 33.445 29.2618 40H34V52H8V40H11.2251C10.6299 36.4414 8.52929 21.6012 13.4337 14.1009C14.3353 12.7219 16.521 12.6807 16.521 13.7408Z" fill="white"/>
        </mask>
        <g mask="url(#topHornsMask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <rect y="40" width="100" height="12" fill="black" fill-opacity="0.4"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M15.4558 13H31.5689V40H20.8201C13.3712 32.1499 15.4558 13 15.4558 13Z" fill="white" fill-opacity="0.4"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M84.8203 13H92.5691V40H81.8203C87.5713 32.1946 84.8203 13 84.8203 13Z" fill="white" fill-opacity="0.4"/>
        </g>
    `;
            };

        }, {}],
        55: [function (require, module, exports) {
            "use strict";
            var __importDefault = (this && this.__importDefault) || function (mod) {
                return (mod && mod.__esModule) ? mod : {"default": mod};
            };
            Object.defineProperty(exports, "__esModule", {value: true});
            const antenna_crooked_1 = __importDefault(require("./antenna-crooked"));
            const antenna_1 = __importDefault(require("./antenna"));
            const bulb_01_1_1 = __importDefault(require("./bulb-01-1"));
            const bulb_01_1 = __importDefault(require("./bulb-01"));
            const glowing_bulb_01_1 = __importDefault(require("./glowing-bulb-01"));
            const glowing_bulb_02_1 = __importDefault(require("./glowing-bulb-02"));
            const horns_1 = __importDefault(require("./horns"));
            const lights_1 = __importDefault(require("./lights"));
            const pyramid_1 = __importDefault(require("./pyramid"));
            const radar_1 = __importDefault(require("./radar"));
            exports.default = [antenna_crooked_1.default, antenna_1.default, bulb_01_1_1.default, bulb_01_1.default, glowing_bulb_01_1.default, glowing_bulb_02_1.default, horns_1.default, lights_1.default, pyramid_1.default, radar_1.default];

        }, {
            "./antenna": 49,
            "./antenna-crooked": 48,
            "./bulb-01": 51,
            "./bulb-01-1": 50,
            "./glowing-bulb-01": 52,
            "./glowing-bulb-02": 53,
            "./horns": 54,
            "./lights": 56,
            "./pyramid": 57,
            "./radar": 58
        }],
        56: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M23 22C20.2386 22 18 24.2386 18 27V40H12C10.8954 40 10 40.8954 10 42V52H18H34H42H58H66H82H90V42C90 40.8954 89.1046 40 88 40H82V27C82 24.2386 79.7614 22 77 22H71C68.2386 22 66 24.2386 66 27V40H58V27C58 24.2386 55.7614 22 53 22H47C44.2386 22 42 24.2386 42 27V40H34V27C34 24.2386 31.7614 22 29 22H23Z" fill="#E1E6E8"/>
        <mask id="topLightsMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="10" y="22" width="80" height="30">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M23 22C20.2386 22 18 24.2386 18 27V40H12C10.8954 40 10 40.8954 10 42V52H18H34H42H58H66H82H90V42C90 40.8954 89.1046 40 88 40H82V27C82 24.2386 79.7614 22 77 22H71C68.2386 22 66 24.2386 66 27V40H58V27C58 24.2386 55.7614 22 53 22H47C44.2386 22 42 24.2386 42 27V40H34V27C34 24.2386 31.7614 22 29 22H23Z" fill="white"/>
        </mask>
        <g mask="url(#topLightsMask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <rect width="100" height="40" fill="white" fill-opacity="0.6"/>
            <rect x="24" y="28" width="4" height="8" rx="2" fill="white" fill-opacity="0.6"/>
            <rect x="48" y="28" width="4" height="8" rx="2" fill="white" fill-opacity="0.6"/>
            <rect x="72" y="28" width="4" height="8" rx="2" fill="white" fill-opacity="0.6"/>
        </g>
    `;
            };

        }, {}],
        57: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M50 8L82 52H18L50 8Z" fill="#E1E6E8"/>
        <mask id="topPyramidMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="18" y="8" width="64" height="44">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M50 8L82 52H18L50 8Z" fill="white"/>
        </mask>
        <g mask="url(#topPyramidMask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <rect x="50" y="4" width="30" height="48" fill="white" fill-opacity="0.8"/>
        </g>
    `;
            };

        }, {}],
        58: [function (require, module, exports) {
            "use strict";
            Object.defineProperty(exports, "__esModule", {value: true});
            exports.default = (color) => {
                return `
        <path fill-rule="evenodd" clip-rule="evenodd" d="M43.7993 28.3969C35.9888 20.5865 35.9888 7.92316 43.7993 0.112671L57.2343 13.5477L63.6874 7.09463C62.7814 5.56072 62.9874 3.55192 64.3054 2.23399C65.8675 0.671894 68.4001 0.671894 69.9622 2.23399C71.5243 3.79609 71.5243 6.32875 69.9622 7.89085C68.6443 9.20878 66.6355 9.41478 65.1016 8.50884L58.6485 14.9619L72.0835 28.3969C66.6332 33.8472 58.8199 35.4942 51.9414 33.3379V52.1127H47.9414V31.581C46.4606 30.7252 45.0661 29.6638 43.7993 28.3969Z" fill="#E1E6E8"/>
        <mask id="topRadarMask0" mask-type="alpha" maskUnits="userSpaceOnUse" x="37" y="0" width="36" height="53">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M43.7993 28.3969C35.9888 20.5865 35.9888 7.92316 43.7993 0.112671L57.2343 13.5477L63.6874 7.09463C62.7814 5.56072 62.9874 3.55192 64.3054 2.23399C65.8675 0.671894 68.4001 0.671894 69.9622 2.23399C71.5243 3.79609 71.5243 6.32875 69.9622 7.89085C68.6443 9.20878 66.6355 9.41478 65.1016 8.50884L58.6485 14.9619L72.0835 28.3969C66.6332 33.8472 58.8199 35.4942 51.9414 33.3379V52.1127H47.9414V31.581C46.4606 30.7252 45.0661 29.6638 43.7993 28.3969Z" fill="white"/>
        </mask>
        <g mask="url(#topRadarMask0)">
            <rect width="100" height="52" fill="${color.hex}"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M43.7988 0.112671C35.9883 7.92316 35.9883 20.5865 43.7988 28.3969C51.6093 36.2074 64.2726 36.2074 72.0831 28.3969" fill="white" fill-opacity="0.2"/>
            <path fill-rule="evenodd" clip-rule="evenodd" d="M64.3054 7.89092C65.8675 9.45302 68.4001 9.45302 69.9622 7.89092C71.5243 6.32882 71.5243 3.79616 69.9622 2.23407C68.4001 0.67197 65.8675 0.67197 64.3054 2.23407C62.7433 3.79616 62.7433 6.32882 64.3054 7.89092Z" fill="white" fill-opacity="0.8"/>
        </g>
    `;
            };

        }, {}],
        59: [function (require, module, exports) {
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

        }, {}],
        60: [function (require, module, exports) {
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

        }, {}],
        61: [function (require, module, exports) {
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

        }, {}],
        62: [function (require, module, exports) {
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

        }, {}],
        63: [function (require, module, exports) {
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

        }, {}],
        64: [function (require, module, exports) {
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

        }, {}],
        65: [function (require, module, exports) {
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

        }, {}],
        66: [function (require, module, exports) {
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

        }, {}],
        67: [function (require, module, exports) {
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

        }, {}],
        68: [function (require, module, exports) {
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
            "./amber": 59,
            "./blue": 60,
            "./blueGrey": 61,
            "./brown": 62,
            "./cyan": 63,
            "./deepOrange": 64,
            "./deepPurple": 65,
            "./green": 66,
            "./grey": 67,
            "./indigo": 69,
            "./lightBlue": 70,
            "./lightGreen": 71,
            "./lime": 72,
            "./orange": 73,
            "./pink": 74,
            "./purple": 75,
            "./red": 76,
            "./teal": 77,
            "./yellow": 78
        }],
        69: [function (require, module, exports) {
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

        }, {}],
        70: [function (require, module, exports) {
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

        }, {}],
        71: [function (require, module, exports) {
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

        }, {}],
        72: [function (require, module, exports) {
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

        }, {}],
        73: [function (require, module, exports) {
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

        }, {}],
        74: [function (require, module, exports) {
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

        }, {}],
        75: [function (require, module, exports) {
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

        }, {}],
        76: [function (require, module, exports) {
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

        }, {}],
        77: [function (require, module, exports) {
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

        }, {}],
        78: [function (require, module, exports) {
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

        }, {}],
        79: [function (require, module, exports) {
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
            "./collection": 68,
            "pure-color/convert/hsv2rgb": 80,
            "pure-color/convert/rgb2hex": 81,
            "pure-color/convert/rgb2hsv": 82,
            "pure-color/parse/hex": 83
        }],
        80: [function (require, module, exports) {
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
        }, {}],
        81: [function (require, module, exports) {
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
        }, {"../util/clamp": 84}],
        82: [function (require, module, exports) {
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
        }, {}],
        83: [function (require, module, exports) {
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
        }, {}],
        84: [function (require, module, exports) {
            function clamp(val, min, max) {
                return Math.min(Math.max(val, min), max);
            }

            module.exports = clamp;
        }, {}]
    }, {}, [23])(23)
});
