/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

define(["knockout", "text!./api.html", "postbox", "jquery", "semantic-ui-popup"],
    function (ko, template, postbox, $) {

        function ViewModel(params) {
            var self = this;

            // Validate user permissions
            if (!authenticated()) {
                postbox.publish("redirect", "login");
                return;
            }

            // Get reference to the view model's actual HTML within the DOM
            this.$domComponent = $('#divApiDoc');

            // Initialize tooltips
            this.$domComponent.find('[data-html]').popup();

            self.title = ko.observable("");
            self.version = ko.observable("");
            self.description = ko.observable("");
            self.basePath = ko.observable("");
            self.groups = ko.observableArray([]);
            self.loading = ko.observable(true);
            self.error = ko.observable("");

            self.definitions = {};

            // Extract a $ref string from a schema, handling both direct $ref and swaggo's
            // allOf:[{$ref:...}] wrapper (used for optional pointer types).
            self.extractRef = function (schema) {
                if (!schema) return null;
                if (schema.$ref) return schema.$ref;
                if (schema.allOf && schema.allOf.length > 0 && schema.allOf[0].$ref) {
                    return schema.allOf[0].$ref;
                }
                return null;
            };

            // Resolve a schema $ref (or allOf wrapper), or return the schema itself.
            self.resolveRef = function (schema) {
                var ref = self.extractRef(schema);
                if (ref) {
                    var name = ref.replace("#/definitions/", "");
                    return self.definitions[name] || {};
                }
                return schema || {};
            };

            // Strip the Go package prefix from a definition name.
            self.stripPkg = function (name) {
                var dot = name.lastIndexOf(".");
                return dot >= 0 ? name.substring(dot + 1) : name;
            };

            // Return property keys in report-matching order for a schema.
            self.orderedKeys = function (schema, resolved) {
                if (!resolved || !resolved.properties) return [];
                var allKeys = Object.keys(resolved.properties);
                var ref = self.extractRef(schema);
                if (!ref) return allKeys;
                var typeName = self.stripPkg(ref.split("/").pop());

                var hidden = (typeof apiFieldHidden !== "undefined") && apiFieldHidden[typeName];
                if (hidden) {
                    allKeys = allKeys.filter(function (k) {
                        return hidden.indexOf(k) < 0;
                    });
                }

                var order = (typeof apiFieldOrder !== "undefined") && apiFieldOrder[typeName];
                if (!order) return allKeys;
                var ordered = [];
                order.forEach(function (k) {
                    if (allKeys.indexOf(k) >= 0) ordered.push(k);
                });
                allKeys.forEach(function (k) {
                    if (ordered.indexOf(k) < 0) ordered.push(k);
                });
                return ordered;
            };

            // Produce a human-readable type label for a property schema.
            self.typeLabel = function (prop) {
                if (!prop) return "any";
                var ref = self.extractRef(prop);
                if (ref) {
                    return self.stripPkg(ref.split("/").pop());
                }
                if (prop.type === "object" && prop.additionalProperties) {
                    var valRef = self.extractRef(prop.additionalProperties);
                    var valType = valRef
                        ? self.stripPkg(valRef.split("/").pop())
                        : (prop.additionalProperties.type || "any");
                    return "map[string]" + valType;
                }
                if (prop.type === "array") {
                    var itemType;
                    var itemRef = prop.items ? self.extractRef(prop.items) : null;
                    if (itemRef) {
                        itemType = self.stripPkg(itemRef.split("/").pop());
                    } else {
                        itemType = (prop.items && prop.items.type) || "string";
                    }
                    return itemType + "[]";
                }
                var label = prop.type || "any";
                if (prop.format) label += " (" + prop.format + ")";
                return label;
            };

            // Extract a usable default value from a property's example.
            self.initialValueFor = function (paramType, example) {
                if (example === undefined || example === null) {
                    if (paramType === "boolean") return "false";
                    return "";
                }
                if (paramType === "boolean") return example ? "true" : "false";
                if (paramType === "array") {
                    if (Array.isArray(example)) return example.join(", ");
                    return String(example);
                }
                return String(example);
            };

            // Flatten a JSON-schema object into a list of form fields.
            self.flattenSchema = function (schema, parentPath, depth, location) {
                var resolved = self.resolveRef(schema);
                if (!resolved || resolved.type !== "object" || !resolved.properties) return [];

                var requiredSet = {};
                (resolved.required || []).forEach(function (r) {
                    requiredSet[r] = true;
                });

                var fields = [];
                self.orderedKeys(schema, resolved).forEach(function (propName) {
                    var prop = resolved.properties[propName];
                    var path = parentPath ? parentPath + "." + propName : propName;

                    var resolvedProp = self.resolveRef(prop);
                    var isObj = resolvedProp && resolvedProp.type === "object";

                    var isArrayOfObjects = false;
                    var arrayItemFields = [];
                    if (prop.type === "array" && prop.items) {
                        var itemResolved = self.resolveRef(prop.items);
                        if (itemResolved && itemResolved.type === "object" && itemResolved.properties) {
                            isArrayOfObjects = true;
                            arrayItemFields = self.flattenResponseFields(prop.items, 0);
                        }
                    }

                    var paramType;
                    if (isObj) {
                        paramType = "object";
                    } else if (prop.type === "array") {
                        paramType = "array";
                    } else {
                        paramType = prop.type || "string";
                    }

                    var initial = isObj ? null : self.initialValueFor(paramType, prop.example);
                    var placeholder;
                    if (prop.example !== undefined && prop.example !== null) {
                        placeholder = Array.isArray(prop.example)
                            ? prop.example.join(", ")
                            : String(prop.example);
                    } else {
                        placeholder = paramType;
                    }

                    if (isArrayOfObjects) {
                        initial = JSON.stringify([self.buildExampleValue(prop.items)], null, 2);
                        placeholder = "JSON array";
                    }

                    fields.push({
                        name: propName,
                        path: path,
                        label: propName,
                        location: location,
                        paramType: paramType,
                        typeLabel: self.typeLabel(prop),
                        required: !!requiredSet[propName],
                        description: prop.description || "",
                        example: prop.example,
                        placeholder: placeholder,
                        depth: depth,
                        isObject: isObj,
                        isArrayOfObjects: isArrayOfObjects,
                        arrayItemFields: arrayItemFields,
                        value: ko.observable(initial)
                    });

                    if (isObj) {
                        fields = fields.concat(
                            self.flattenSchema(prop, path, depth + 1, location)
                        );
                    }
                });

                return fields;
            };

            // Build the list of editable input fields for a swagger operation.
            self.buildInputFields = function (operation) {
                var fields = [];
                (operation.parameters || []).forEach(function (p) {
                    if (p.in === "body" && p.schema) {
                        fields = fields.concat(self.flattenSchema(p.schema, "", 0, "body"));
                    } else {
                        var paramType = p.type || "string";
                        var initial = self.initialValueFor(paramType, p.example);
                        var placeholder;
                        if (p.example !== undefined && p.example !== null) {
                            placeholder = Array.isArray(p.example) ? p.example.join(", ") : String(p.example);
                        } else {
                            placeholder = paramType;
                        }
                        fields.push({
                            name: p.name,
                            path: p.name,
                            label: p.name,
                            location: p.in,
                            paramType: paramType,
                            typeLabel: self.typeLabel(p),
                            required: !!p.required,
                            description: p.description || "",
                            example: p.example,
                            placeholder: placeholder,
                            depth: 0,
                            isObject: false,
                            value: ko.observable(initial)
                        });
                    }
                });
                return fields;
            };

            // Set a value at a dotted path inside an object, creating intermediate objects as needed.
            self.setNested = function (obj, path, value) {
                var parts = path.split(".");
                var cur = obj;
                for (var i = 0; i < parts.length - 1; i++) {
                    if (typeof cur[parts[i]] !== "object" || cur[parts[i]] === null) {
                        cur[parts[i]] = {};
                    }
                    cur = cur[parts[i]];
                }
                cur[parts[parts.length - 1]] = value;
            };

            // Parse an input field's raw value into its JSON-typed representation.
            self.parseFieldValue = function (field) {
                if (field.isObject) return undefined;
                var raw = ko.unwrap(field.value);

                if (raw === undefined || raw === null) return undefined;
                var str = String(raw).trim();
                if (field.paramType === "boolean") {
                    return str === "true" ? true : undefined;
                }
                if (str === "") return undefined;

                if (field.isArrayOfObjects) {
                    try {
                        var parsed = JSON.parse(str);
                        return Array.isArray(parsed) ? parsed : undefined;
                    } catch (e) {
                        return undefined;
                    }
                }

                if (field.paramType === "integer") {
                    var n = parseInt(str, 10);
                    return isNaN(n) ? undefined : n;
                }
                if (field.paramType === "number") {
                    var f = parseFloat(str);
                    return isNaN(f) ? undefined : f;
                }
                if (field.paramType === "array") {
                    var parts = str.split(",").map(function (s) {
                        return s.trim();
                    }).filter(Boolean);
                    return parts.length ? parts : undefined;
                }
                return str;
            };

            // Build the JSON request body from all body-typed fields.
            self.buildJsonBody = function (endpoint) {
                var body = {};
                endpoint.inputFields.forEach(function (f) {
                    if (f.location !== "body") return;
                    var v = self.parseFieldValue(f);
                    if (v === undefined) return;
                    self.setNested(body, f.path, v);
                });
                return body;
            };

            // Collect multipart formData fields as [name, value] pairs.
            self.buildFormFields = function (endpoint) {
                var pairs = [];
                endpoint.inputFields.forEach(function (f) {
                    if (f.location !== "formData" || f.paramType === "file") return;
                    var v = self.parseFieldValue(f);
                    if (v === undefined) return;
                    if (Array.isArray(v)) {
                        v.forEach(function (item) {
                            pairs.push([f.name, String(item)]);
                        });
                    } else {
                        pairs.push([f.name, typeof v === "object" ? JSON.stringify(v) : String(v)]);
                    }
                });
                return pairs;
            };

            // Produce a list of file parameters with their chosen filename (or placeholder).
            self.buildFileHints = function (endpoint) {
                var files = [];
                endpoint.inputFields.forEach(function (f) {
                    if (f.paramType !== "file") return;
                    var chosen = endpoint.fileInputs[f.name];
                    files.push({
                        name: f.name,
                        filename: chosen ? chosen.name : "/path/to/your/file",
                        required: f.required
                    });
                });
                return files;
            };

            // Flatten a response schema into a list of display-only fields.
            self.flattenResponseFields = function (schema, depth) {
                var resolved = self.resolveRef(schema);
                if (!resolved) return [];

                if (resolved.type === "object" && resolved.additionalProperties && !resolved.properties) {
                    var valueSchema = self.resolveRef(resolved.additionalProperties);
                    if (valueSchema && valueSchema.properties) {
                        return self.flattenResponseFields(valueSchema, depth);
                    }
                    return [];
                }

                if (!resolved.properties) return [];

                var fields = [];
                self.orderedKeys(schema, resolved).forEach(function (propName) {
                    var prop = resolved.properties[propName];
                    var resolvedProp = self.resolveRef(prop);
                    var isNestedObj = resolvedProp && resolvedProp.type === "object" && resolvedProp.properties;
                    var isMap = resolvedProp && resolvedProp.type === "object" && resolvedProp.additionalProperties && !resolvedProp.properties;

                    var exampleStr = "";
                    if (prop.example !== undefined && prop.example !== null) {
                        exampleStr = typeof prop.example === "object"
                            ? JSON.stringify(prop.example)
                            : String(prop.example);
                    }

                    fields.push({
                        name: propName,
                        typeLabel: isMap ? "map" : self.typeLabel(prop),
                        description: prop.description || "",
                        depth: depth,
                        example: exampleStr
                    });

                    if (isNestedObj) {
                        fields = fields.concat(self.flattenResponseFields(prop, depth + 1));
                    }
                    if (isMap) {
                        var mapValueSchema = self.resolveRef(resolvedProp.additionalProperties);
                        if (mapValueSchema && mapValueSchema.properties) {
                            fields = fields.concat(self.flattenResponseFields(resolvedProp.additionalProperties, depth + 1));
                        }
                    }
                });
                return fields;
            };

            // Build a representative example value from a swagger schema definition.
            self.buildExampleValue = function (schema) {
                var resolved = self.resolveRef(schema);
                if (!resolved) return null;

                if (resolved.type === "object" && resolved.additionalProperties && !resolved.properties) {
                    var obj = {};
                    obj["key"] = self.buildExampleValue(resolved.additionalProperties);
                    return obj;
                }

                if (resolved.type === "object" && resolved.properties) {
                    var obj = {};
                    self.orderedKeys(schema, resolved).forEach(function (key) {
                        obj[key] = self.buildExampleValue(resolved.properties[key]);
                    });
                    return obj;
                }

                if (resolved.type === "array") {
                    var item = resolved.items ? self.buildExampleValue(resolved.items) : "...";
                    return [item];
                }

                if (resolved.example !== undefined && resolved.example !== null) return resolved.example;

                switch (resolved.type) {
                    case "integer":
                        return 0;
                    case "number":
                        return 0.0;
                    case "boolean":
                        return false;
                    case "string":
                        if (resolved.format === "date-time") return "2026-01-01T08:00:00Z";
                        return "";
                    default:
                        return null;
                }
            };

            // Build a full example JSON response wrapped in the BaseResponse envelope.
            self.buildExampleResponse = function (response) {
                var isError = response.code.charAt(0) !== '2';
                var body = null;
                if (response.schema) {
                    body = self.buildExampleValue(response.schema);
                }
                var envelope = {
                    error: isError,
                    message: response.description
                };
                if (body !== null) {
                    envelope.body = body;
                }
                return JSON.stringify(envelope, null, 2);
            };

            // --------- Snippet generators ----------

            self.indentJson = function (obj, indent) {
                var pad = new Array(indent + 1).join(" ");
                return JSON.stringify(obj, null, 2)
                    .split("\n")
                    .map(function (line, i) {
                        return i === 0 ? line : pad + line;
                    })
                    .join("\n");
            };

            self.absoluteUrl = function (endpoint) {
                var path = endpoint.fullPath;
                endpoint.inputFields.forEach(function (f) {
                    if (f.location === "path") {
                        var val = ko.unwrap(f.value);
                        if (val) {
                            path = path.replace("{" + f.name + "}", val);
                        }
                    }
                });
                return window.location.origin + path;
            };

            self.curlSnippet = function (endpoint) {
                var url = self.absoluteUrl(endpoint);
                var lines = [];
                lines.push("curl -X " + endpoint.method + " '" + url + "' \\");
                lines.push("  -H 'Authorization: Bearer <TOKEN>'");

                if (endpoint.isMultipart) {
                    var files = self.buildFileHints(endpoint);
                    var form = self.buildFormFields(endpoint);
                    files.forEach(function (f) {
                        lines[lines.length - 1] += " \\";
                        lines.push("  -F '" + f.name + "=@" + f.filename + "'");
                    });
                    form.forEach(function (pair) {
                        lines[lines.length - 1] += " \\";
                        lines.push("  -F '" + pair[0] + "=" + pair[1].replace(/'/g, "'\\''") + "'");
                    });
                } else if (endpoint.method !== "GET" && endpoint.method !== "DELETE") {
                    var body = self.buildJsonBody(endpoint);
                    if (Object.keys(body).length > 0) {
                        lines[lines.length - 1] += " \\";
                        lines.push("  -H 'Content-Type: application/json' \\");
                        lines.push("  -d '" + self.indentJson(body, 5) + "'");
                    }
                }
                return lines.join("\n");
            };

            self.pythonSnippet = function (endpoint) {
                var url = self.absoluteUrl(endpoint);
                var lines = ["import requests", ""];

                if (endpoint.isMultipart) {
                    var files = self.buildFileHints(endpoint);
                    var form = self.buildFormFields(endpoint);

                    if (files.length > 0) {
                        var openParts = files.map(function (f, idx) {
                            return "open('" + f.filename + "', 'rb') as f" + idx;
                        });
                        lines.push("with " + openParts.join(", ") + ":");

                        lines.push("    files = {");
                        files.forEach(function (f, idx) {
                            lines.push("        '" + f.name + "': f" + idx + ",");
                        });
                        lines.push("    }");
                    } else {
                        lines.push("files = None");
                    }

                    var indent = files.length > 0 ? "    " : "";
                    var dataObj = {};
                    form.forEach(function (pair) {
                        if (dataObj[pair[0]] === undefined) {
                            dataObj[pair[0]] = pair[1];
                        } else if (Array.isArray(dataObj[pair[0]])) {
                            dataObj[pair[0]].push(pair[1]);
                        } else {
                            dataObj[pair[0]] = [dataObj[pair[0]], pair[1]];
                        }
                    });
                    lines.push(indent + "data = " + self.indentJson(dataObj, indent.length + 7));
                    lines.push(indent + "response = requests.request(");
                    lines.push(indent + "    '" + endpoint.method + "',");
                    lines.push(indent + "    '" + url + "',");
                    lines.push(indent + "    headers={'Authorization': 'Bearer <TOKEN>'},");
                    lines.push(indent + "    files=files,");
                    lines.push(indent + "    data=data,");
                    lines.push(indent + ")");
                    lines.push(indent + "print(response.status_code, response.json())");
                } else {
                    var body = self.buildJsonBody(endpoint);
                    lines.push("response = requests.request(");
                    lines.push("    '" + endpoint.method + "',");
                    lines.push("    '" + url + "',");
                    lines.push("    headers={'Authorization': 'Bearer <TOKEN>'},");
                    if (Object.keys(body).length > 0) {
                        lines.push("    json=" + self.indentJson(body, 9) + ",");
                    }
                    lines.push(")");
                    lines.push("print(response.status_code, response.json())");
                }
                return lines.join("\n");
            };

            self.javascriptSnippet = function (endpoint) {
                var url = self.absoluteUrl(endpoint);
                var lines = [];

                if (endpoint.isMultipart) {
                    var files = self.buildFileHints(endpoint);
                    var form = self.buildFormFields(endpoint);

                    lines.push("const formData = new FormData();");
                    files.forEach(function (f) {
                        lines.push("// Attach '" + f.name + "' from an <input type=\"file\"> element:");
                        lines.push("formData.append('" + f.name + "', fileInput.files[0]);");
                    });
                    form.forEach(function (pair) {
                        lines.push("formData.append('" + pair[0] + "', " + JSON.stringify(pair[1]) + ");");
                    });
                    lines.push("");
                    lines.push("const response = await fetch('" + url + "', {");
                    lines.push("    method: '" + endpoint.method + "',");
                    lines.push("    headers: { 'Authorization': 'Bearer <TOKEN>' },");
                    lines.push("    body: formData,");
                    lines.push("});");
                } else {
                    var body = self.buildJsonBody(endpoint);
                    lines.push("const response = await fetch('" + url + "', {");
                    lines.push("    method: '" + endpoint.method + "',");
                    lines.push("    headers: {");
                    lines.push("        'Authorization': 'Bearer <TOKEN>',");
                    if (Object.keys(body).length > 0) {
                        lines.push("        'Content-Type': 'application/json',");
                    }
                    lines.push("    },");
                    if (Object.keys(body).length > 0) {
                        lines.push("    body: JSON.stringify(" + self.indentJson(body, 8) + "),");
                    }
                    lines.push("});");
                }
                lines.push("const data = await response.json();");
                lines.push("console.log(response.status, data);");
                return lines.join("\n");
            };

            self.httpieSnippet = function (endpoint) {
                var url = self.absoluteUrl(endpoint);
                var lines = [];

                if (endpoint.isMultipart) {
                    var files = self.buildFileHints(endpoint);
                    var form = self.buildFormFields(endpoint);

                    lines.push("http --form " + endpoint.method + " '" + url + "' \\");
                    lines.push("  'Authorization: Bearer <TOKEN>'");
                    files.forEach(function (f) {
                        lines[lines.length - 1] += " \\";
                        lines.push("  " + f.name + "@" + f.filename);
                    });
                    form.forEach(function (pair) {
                        lines[lines.length - 1] += " \\";
                        lines.push("  " + pair[0] + "='" + pair[1].replace(/'/g, "'\\''") + "'");
                    });
                } else {
                    var body = self.buildJsonBody(endpoint);
                    lines.push("http " + endpoint.method + " '" + url + "' \\");
                    lines.push("  'Authorization: Bearer <TOKEN>'");

                    Object.keys(body).forEach(function (k) {
                        var v = body[k];
                        var arg;
                        if (typeof v === "string") {
                            arg = k + "='" + v.replace(/'/g, "'\\''") + "'";
                        } else {
                            arg = k + ":=" + JSON.stringify(v);
                        }
                        lines[lines.length - 1] += " \\";
                        lines.push("  " + arg);
                    });
                }
                return lines.join("\n");
            };

            self.renderUsage = function (endpoint) {
                // Touch all field observables so this computed re-runs on edits.
                endpoint.inputFields.forEach(function (f) {
                    if (!f.isObject) ko.unwrap(f.value);
                });

                var tab = endpoint.usageTab();
                switch (tab) {
                    case "python":
                        return self.pythonSnippet(endpoint);
                    case "javascript":
                        return self.javascriptSnippet(endpoint);
                    case "httpie":
                        return self.httpieSnippet(endpoint);
                    default:
                        return self.curlSnippet(endpoint);
                }
            };

            // --------- Parse swagger spec ----------

            var callbackSuccess = function (response) {
                var data = response;
                if (response && response.body) data = response.body;

                self.definitions = data.definitions || {};
                self.title(data.info ? data.info.title : "API Documentation");
                self.version(data.info ? data.info.version : "");
                self.description(data.info ? data.info.description : "");
                var basePath = (data.basePath || "").replace(/\/$/, "");
                self.basePath(basePath || "/");

                var grouped = {};

                $.each(data.paths || {}, function (path, methods) {
                    $.each(methods, function (method, op) {
                        if (["get", "post", "put", "delete", "patch"].indexOf(method) < 0) return;

                        var tag = (op.tags && op.tags[0]) || "general";
                        if (!grouped[tag]) {
                            grouped[tag] = {tag: tag, endpoints: [], collapsed: ko.observable(false)};
                        }

                        var consumes = op.consumes || [];
                        var isMultipart = consumes.indexOf("multipart/form-data") >= 0;

                        var responses = [];
                        $.each(op.responses || {}, function (code, resp) {
                            var responseFields = [];
                            if (resp.schema) {
                                responseFields = self.flattenResponseFields(resp.schema, 0);
                            }
                            responses.push({
                                code: String(code),
                                description: resp.description || "",
                                schema: resp.schema || null,
                                fields: responseFields
                            });
                        });

                        var endpoint = {
                            method: method.toUpperCase(),
                            path: path,
                            fullPath: basePath + path,
                            summary: op.summary || "",
                            description: op.description || "",
                            isMultipart: isMultipart,
                            inputFields: self.buildInputFields(op),
                            responses: responses,
                            fileInputs: {},
                            expanded: ko.observable(false),
                            usageTab: ko.observable("curl")
                        };
                        grouped[tag].endpoints.push(endpoint);
                    });
                });

                var endpointIndex = function (path) {
                    for (var i = 0; i < apiEndpointOrder.length; i++) {
                        if (path === apiEndpointOrder[i] || path.indexOf(apiEndpointOrder[i] + "/") === 0) {
                            return i;
                        }
                    }
                    return apiEndpointOrder.length;
                };

                var arr = [];
                Object.keys(grouped).forEach(function (tag) {
                    grouped[tag].endpoints.sort(function (a, b) {
                        var ia = endpointIndex(a.path);
                        var ib = endpointIndex(b.path);
                        if (ia !== ib) return ia - ib;
                        var methodOrder = {GET: 0, POST: 1};
                        var ma = methodOrder[a.method] !== undefined ? methodOrder[a.method] : 2;
                        var mb = methodOrder[b.method] !== undefined ? methodOrder[b.method] : 2;
                        return ma - mb;
                    });
                    arr.push(grouped[tag]);
                });
                arr.sort(function (a, b) {
                    return a.tag.localeCompare(b.tag);
                });

                self.groups(arr);
                self.loading(false);
            };

            var callbackError = function (response) {
                console.error("[API Docs] failed to load:", response);
                self.error("Could not load API documentation.");
                self.loading(false);
            };

            apiCall(
                "GET",
                "/api/v1/swagger/doc.json",
                {},
                null,
                callbackSuccess,
                callbackError,
                true,
                true
            );
        }

        ViewModel.prototype.toggleGroup = function (group) {
            group.collapsed(!group.collapsed());
        };

        ViewModel.prototype.toggleEndpoint = function (endpoint) {
            var opening = !endpoint.expanded();
            if (opening) {
                this.groups().forEach(function (group) {
                    group.endpoints.forEach(function (ep) {
                        if (ep !== endpoint && ep.expanded()) {
                            ep.expanded(false);
                        }
                    });
                });
            }
            endpoint.expanded(opening);
        };

        ViewModel.prototype.fileSelected = function (event, endpoint, paramName) {
            var files = event.target.files;
            if (files && files.length > 0) {
                endpoint.fileInputs[paramName] = files[0];
            } else {
                delete endpoint.fileInputs[paramName];
            }
            if (endpoint.inputFields.length > 0 && !endpoint.inputFields[0].isObject) {
                var v = endpoint.inputFields[0].value;
                v.valueHasMutated && v.valueHasMutated();
            }
        };

        ViewModel.prototype.copyUsage = function (endpoint) {
            var text = this.renderUsage(endpoint);
            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard.writeText(text);
            } else {
                var ta = document.createElement("textarea");
                ta.value = text;
                document.body.appendChild(ta);
                ta.select();
                document.execCommand("copy");
                document.body.removeChild(ta);
            }
        };

        ViewModel.prototype.generateApiToken = function () {

            confirmOverlay(
                "key",
                "Generate API Access Token",
                "A new long-lived API access token will be generated, valid for 1 year.<br />Any previously issued API token for your account will be invalidated.",
                function () {

                    var callbackSuccess = function (response) {

                        var token = response.body["token"];
                        if (!token) {
                            toast(response.message, "success");
                            return;
                        }

                        infoOverlay(
                            "key",
                            "Generated API Access Token",
                            'Please copy the token below, it will disappear shortly and cannot be retrieved again.<br/><br/>' +
                            '<div class="ui sixteen column centered grid">' +
                            '  <div class="fourteen wide column">' +
                            '    <div class="ui inverted black segment" style="word-break: break-all; font-family: monospace; text-align: left;">' + token + '</div>' +
                            '  </div>' +
                            '</div>',
                            function () {
                                token = "";
                                $('body').css("margin-right", "0px");
                            },
                            30000
                        );
                    };

                    apiCall(
                        "POST",
                        "/api/v1/user/api-token",
                        {},
                        null,
                        callbackSuccess,
                        null
                    );
                }
            );
        };

        ViewModel.prototype.dispose = function () {
        };

        return {
            viewModel: ViewModel,
            template: template
        };
    });
