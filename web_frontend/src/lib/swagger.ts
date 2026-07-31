/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2025-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

import { stringifyUnknown } from '@/lib/collections'

/**
 * Swagger / OpenAPI 2.0 (swaggo) parsing + request-snippet generation for the
 * API documentation page.
 *
 * This is a faithful, framework-free TypeScript port of the legacy
 * `web_frontend/src/pages/api/api.js` view-model helpers and the ordering
 * tables in `web_frontend/src/js/utils-api.js`. It is deliberately pure (no
 * Vue, no Knockout, no DOM beyond an injected `origin` string) so it can be
 * unit-tested and so the page's reactivity is owned by the SFC, not by this
 * module.
 *
 * The shapes produced here drive `ApiPage.vue`: groups → endpoints → editable
 * input fields + response schema + per-language usage snippets.
 */

// ── Ordering / hiding tables (ported from utils-api.js) ───────────────────────

/** SwaggerConfig controls application-specific endpoint and schema presentation. */
export interface SwaggerConfig {
  endpointOrder: readonly string[]
  fieldHidden: Readonly<Record<string, readonly string[]>>
  fieldOrder: Readonly<Record<string, readonly string[]>>
  defaultTitle: string
}

// Use empty parser customizations when a consumer supplies no Swagger configuration
const DEFAULT_SWAGGER_CONFIG: SwaggerConfig = {
  endpointOrder: [],
  fieldHidden: {},
  fieldOrder: {},
  defaultTitle: '',
}
/** A swagger schema/definition node. Dynamic JSON; intentionally permissive. */
type SwaggerSchema = Record<string, unknown>

/** SwaggerDoc describes the swagger doc data used by Swagger parsing. */
export interface SwaggerDoc {
  info?: { title?: string; version?: string; description?: string }
  basePath?: string
  paths?: Record<string, Record<string, SwaggerOperation>>
  definitions?: Record<string, SwaggerSchema>
}

// SwaggerOperation describes the swagger operation data used by Swagger parsing
interface SwaggerOperation {
  tags?: string[]
  summary?: string
  description?: string
  consumes?: string[]
  parameters?: SwaggerParameter[]
  responses?: Record<string, { description?: string; schema?: SwaggerSchema }>
}

// SwaggerParameter describes the swagger parameter data used by Swagger parsing
interface SwaggerParameter {
  name: string
  in: string
  type?: string
  required?: boolean
  description?: string
  example?: unknown
  schema?: SwaggerSchema
}

// UsageTab describes the usage tab data used by Swagger parsing
type UsageTab = 'curl' | 'python' | 'javascript' | 'httpie'

/** An editable request parameter (body field, query/path/header/formData param). */
interface InputField {
  name: string
  path: string
  label: string
  location: string
  paramType: string
  typeLabel: string
  required: boolean
  description: string
  example?: unknown
  placeholder: string
  depth: number
  isObject: boolean
  isArrayOfObjects: boolean
  arrayItemFields: ResponseField[]
  /** Current editable value (string for inputs, `null` for object group headers). */
  value: string | null
}

/** A read-only response/parameter schema row. */
interface ResponseField {
  name: string
  typeLabel: string
  description: string
  depth: number
  example: string
}

// ResponseInfo describes the response info data used by Swagger parsing
interface ResponseInfo {
  code: string
  description: string
  schema: SwaggerSchema | null
  fields: ResponseField[]
  /** Pre-rendered example JSON (envelope-wrapped); '' when there is no schema. */
  example: string
}

/** Endpoint describes the endpoint data used by Swagger parsing. */
export interface Endpoint {
  method: string
  path: string
  fullPath: string
  summary: string
  description: string
  isMultipart: boolean
  inputFields: InputField[]
  responses: ResponseInfo[]
  /** Chosen <input type=file> selections, keyed by parameter name. */
  fileInputs: Record<string, File>
  /** UI state — page mutates these directly (reactive via the owning ref). */
  expanded: boolean
  usageTab: UsageTab
}

/** EndpointGroup describes the endpoint group data used by Swagger parsing. */
export interface EndpointGroup {
  tag: string
  endpoints: Endpoint[]
  collapsed: boolean
}

/** ParsedSpec describes the parsed spec data used by Swagger parsing. */
export interface ParsedSpec {
  title: string
  version: string
  description: string
  basePath: string
  groups: EndpointGroup[]
}

// Parse only HTTP methods supported by the API explorer
const HTTP_METHODS = ['get', 'post', 'put', 'delete', 'patch']

// ── Schema helpers (closure over the definitions map) ─────────────────────────

// Resolver describes the schema operations used by Swagger parsing
interface Resolver {
  extractRef(schema: SwaggerSchema | undefined | null): string | null
  resolveRef(schema: SwaggerSchema | undefined | null): SwaggerSchema
  stripPkg(name: string): string
  orderedKeys(schema: SwaggerSchema | null, resolved: SwaggerSchema): string[]
  typeLabel(prop: SwaggerSchema | undefined): string
}

function createResolver(
  definitions: Record<string, SwaggerSchema>,
  config: SwaggerConfig,
): Resolver {
  /** Extract a `$ref`, handling swaggo's `allOf:[{$ref}]` pointer wrapper. */
  function extractRef(schema: SwaggerSchema | undefined | null): string | null {
    if (!schema) return null
    if (typeof schema.$ref === 'string') return schema.$ref

    // Read the schemas combined through allOf
    const allOf = schema.allOf as SwaggerSchema[] | undefined

    // Use the first combined schema as the concrete shape
    const first = allOf?.[0]
    if (first && typeof first.$ref === 'string') {
      return first.$ref
    }
    return null
  }

  /** Resolve a `$ref` (or `allOf` wrapper) to its definition, else the schema. */
  function resolveRef(schema: SwaggerSchema | undefined | null): SwaggerSchema {
    // Extract the referenced schema name when one exists
    const ref = extractRef(schema)
    if (ref) {
      // Remove the Swagger definition prefix from the reference
      const name = ref.replace('#/definitions/', '')
      return definitions[name] ?? {}
    }
    return schema ?? {}
  }

  /** Strip the Go package prefix from a definition name. */
  function stripPkg(name: string): string {
    // Find the package separator before the short type name
    const dot = name.lastIndexOf('.')
    return dot >= 0 ? name.substring(dot + 1) : name
  }

  /** Property keys in report-matching order, honouring the hide/order tables. */
  function orderedKeys(schema: SwaggerSchema | null, resolved: SwaggerSchema): string[] {
    // Read the object properties declared by this schema
    const props = resolved.properties as Record<string, SwaggerSchema> | undefined
    if (!props) return []

    // Start with every property name declared by the schema
    let allKeys = Object.keys(props)

    // Extract the referenced schema name when one exists
    const ref = extractRef(schema)
    if (!ref) return allKeys

    // Reduce the reference to the backend type name used by configuration
    const typeName = stripPkg(ref.split('/').pop() as string)

    // Find fields configured to stay hidden for this backend type
    const hidden = config.fieldHidden[typeName]
    if (hidden) {
      allKeys = allKeys.filter((k) => hidden.indexOf(k) < 0)
    }

    // Find the preferred field order for this backend type
    const order = config.fieldOrder[typeName]
    if (!order) return allKeys

    // Build the final property order without duplicates
    const ordered: string[] = []
    order.forEach((k) => {
      if (allKeys.indexOf(k) >= 0) ordered.push(k)
    })
    allKeys.forEach((k) => {
      if (ordered.indexOf(k) < 0) ordered.push(k)
    })
    return ordered
  }

  /** Human-readable type label for a property schema. */
  function typeLabel(prop: SwaggerSchema | undefined): string {
    if (!prop) return 'any'

    // Extract the referenced schema name when one exists
    const ref = extractRef(prop)
    if (ref) return stripPkg(ref.split('/').pop() as string)

    // Read the schema used for additional map values
    const additional = prop.additionalProperties as SwaggerSchema | undefined
    if (prop.type === 'object' && additional) {
      // Extract the referenced map-value type
      const valRef = extractRef(additional)

      // Convert the map-value reference into a readable type name
      const valType = valRef
        ? stripPkg(valRef.split('/').pop() as string)
        : (additional.type as string) || 'any'
      return `map[string]${valType}`
    }
    if (prop.type === 'array') {
      // Read the schema used by array items
      const items = prop.items as SwaggerSchema | undefined

      // Extract the referenced array-item type
      const itemRef = items ? extractRef(items) : null

      // Convert the array-item reference into a readable type name
      const itemType = itemRef
        ? stripPkg(itemRef.split('/').pop() as string)
        : (items && (items.type as string)) || 'string'
      return `${itemType}[]`
    }

    // Start the displayed type label with the Swagger primitive type
    let label = (prop.type as string) || 'any'
    if (prop.format) label += ` (${prop.format as string})`
    return label
  }

  return { extractRef, resolveRef, stripPkg, orderedKeys, typeLabel }
}

/** Default editable value extracted from a property's example. */
function initialValueFor(paramType: string, example: unknown): string {
  if (example === undefined || example === null) {
    return paramType === 'boolean' ? 'false' : ''
  }
  if (paramType === 'boolean') return example ? 'true' : 'false'
  if (paramType === 'array') {
    return Array.isArray(example)
      ? example.map((item) => stringifyUnknown(item)).join(', ')
      : stringifyUnknown(example)
  }
  return stringifyUnknown(example)
}

// ── Response field flattening + example building ──────────────────────────────

function flattenResponseFields(
  r: Resolver,
  schema: SwaggerSchema | null,
  depth: number,
): ResponseField[] {
  // Resolve references before inspecting the concrete schema
  const resolved = r.resolveRef(schema)
  if (!resolved) return []

  // Read the schema used for additional map values
  const additional = resolved.additionalProperties as SwaggerSchema | undefined
  if (resolved.type === 'object' && additional && !resolved.properties) {
    // Resolve the concrete schema used for map values
    const valueSchema = r.resolveRef(additional)
    if (valueSchema && valueSchema.properties) {
      return flattenResponseFields(r, valueSchema, depth)
    }
    return []
  }

  if (!resolved.properties) return []

  // Collect the fields produced from this schema
  let fields: ResponseField[] = []

  // Read the object properties declared by this schema
  const props = resolved.properties as Record<string, SwaggerSchema>
  r.orderedKeys(schema, resolved).forEach((propName) => {
    // orderedKeys derives from Object.keys(props), so every key indexes an
    // existing property; guard anyway to satisfy the compiler and skip gaps.
    const prop = props[propName]
    if (!prop) return

    // Resolve this property before checking its shape
    const resolvedProp = r.resolveRef(prop)

    // Detect nested objects that should be expanded into child fields
    const isNestedObj = !!(
      resolvedProp &&
      resolvedProp.type === 'object' &&
      resolvedProp.properties
    )

    // Read the map-value schema declared by this property
    const propAdditional = resolvedProp.additionalProperties as SwaggerSchema | undefined

    // Detect object maps that need a representative value
    const isMap = !!(
      resolvedProp &&
      resolvedProp.type === 'object' &&
      propAdditional &&
      !resolvedProp.properties
    )

    // Build a readable example for this response field
    let exampleStr = ''
    if (prop.example !== undefined && prop.example !== null) {
      exampleStr = stringifyUnknown(prop.example)
    }

    fields.push({
      name: propName,
      typeLabel: isMap ? 'map' : r.typeLabel(prop),
      description: (prop.description as string) || '',
      depth,
      example: exampleStr,
    })

    if (isNestedObj) {
      fields = fields.concat(flattenResponseFields(r, prop, depth + 1))
    }
    if (isMap && propAdditional) {
      // Resolve the schema used by values in this map
      const mapValueSchema = r.resolveRef(propAdditional)
      if (mapValueSchema && mapValueSchema.properties) {
        fields = fields.concat(flattenResponseFields(r, propAdditional, depth + 1))
      }
    }
  })
  return fields
}

// Build a representative JSON value from a Swagger schema
function buildExampleValue(r: Resolver, schema: SwaggerSchema | null): unknown {
  // Resolve references before inspecting the concrete schema
  const resolved = r.resolveRef(schema)
  if (!resolved) return null

  // Read the schema used for additional map values
  const additional = resolved.additionalProperties as SwaggerSchema | undefined
  if (resolved.type === 'object' && additional && !resolved.properties) {
    return { key: buildExampleValue(r, additional) }
  }

  if (resolved.type === 'object' && resolved.properties) {
    // Build the representative object one property at a time
    const obj: Record<string, unknown> = {}

    // Read the object properties declared by this schema
    const props = resolved.properties as Record<string, SwaggerSchema>
    r.orderedKeys(schema, resolved).forEach((key) => {
      // orderedKeys derives from Object.keys(props), so key always indexes an
      // existing entry; ?? null bridges to the SwaggerSchema | null parameter.
      obj[key] = buildExampleValue(r, props[key] ?? null)
    })
    return obj
  }

  if (resolved.type === 'array') {
    // Read the schema used by array items
    const items = resolved.items as SwaggerSchema | undefined
    return [items ? buildExampleValue(r, items) : '...']
  }

  if (resolved.example !== undefined && resolved.example !== null) return resolved.example

  switch (resolved.type) {
    case 'integer':
      return 0
    case 'number':
      return 0.0
    case 'boolean':
      return false
    case 'string':
      if (resolved.format === 'date-time') return '2026-01-01T08:00:00Z'
      return ''
    default:
      return null
  }
}

/** Full example response wrapped in the legacy BaseResponse envelope. */
function exampleResponse(
  r: Resolver,
  code: string,
  description: string,
  schema: SwaggerSchema | null,
): string {
  // Treat every non-2xx response as an error example
  const isError = code.charAt(0) !== '2'

  // Build the request or response body separately from its envelope
  let body: unknown = null
  if (schema) body = buildExampleValue(r, schema)

  // Wrap the example in the backend response envelope
  const envelope: Record<string, unknown> = {
    error: isError,
    message: description,
  }
  if (body !== null) envelope.body = body
  return JSON.stringify(envelope, null, 2)
}

// ── Input field flattening ────────────────────────────────────────────────────

function flattenSchema(
  r: Resolver,
  schema: SwaggerSchema | null,
  parentPath: string,
  depth: number,
  location: string,
): InputField[] {
  // Resolve references before inspecting the concrete schema
  const resolved = r.resolveRef(schema)
  if (!resolved || resolved.type !== 'object' || !resolved.properties) return []

  // Index required property names for quick field lookup
  const requiredSet: Record<string, boolean> = {}
  ;((resolved.required as string[]) || []).forEach((req) => {
    requiredSet[req] = true
  })

  // Collect the fields produced from this schema
  let fields: InputField[] = []

  // Read the object properties declared by this schema
  const props = resolved.properties as Record<string, SwaggerSchema>
  r.orderedKeys(schema, resolved).forEach((propName) => {
    // orderedKeys derives from Object.keys(props), so every key indexes an
    // existing property; guard anyway to satisfy the compiler and skip gaps.
    const prop = props[propName]
    if (!prop) return

    // Build the dotted path used to write nested field values
    const path = parentPath ? `${parentPath}.${propName}` : propName

    // Resolve this property before checking its shape
    const resolvedProp = r.resolveRef(prop)

    // Detect object fields that contain nested inputs
    const isObj = !!(resolvedProp && resolvedProp.type === 'object')

    // Track whether this field contains repeated object values
    let isArrayOfObjects = false

    // Collect child fields used by one array item
    let arrayItemFields: ResponseField[] = []

    // Read the schema used by array items
    const items = prop.items as SwaggerSchema | undefined
    if (prop.type === 'array' && items) {
      // Resolve the array-item schema before expanding it
      const itemResolved = r.resolveRef(items)
      if (itemResolved && itemResolved.type === 'object' && itemResolved.properties) {
        isArrayOfObjects = true
        arrayItemFields = flattenResponseFields(r, items, 0)
      }
    }

    // Normalize the Swagger parameter type used by the input control
    let paramType: string
    if (isObj) paramType = 'object'
    else if (prop.type === 'array') paramType = 'array'
    else paramType = (prop.type as string) || 'string'

    // Choose the initial input value from the schema and example
    let initial: string | null = isObj ? null : initialValueFor(paramType, prop.example)

    // Build a concise placeholder that explains the expected value
    let placeholder: string
    if (prop.example !== undefined && prop.example !== null) {
      placeholder = Array.isArray(prop.example)
        ? prop.example.map((item) => stringifyUnknown(item)).join(', ')
        : stringifyUnknown(prop.example)
    } else {
      placeholder = paramType
    }

    if (isArrayOfObjects && items) {
      initial = JSON.stringify([buildExampleValue(r, items)], null, 2)
      placeholder = 'JSON array'
    }

    fields.push({
      name: propName,
      path,
      label: propName,
      location,
      paramType,
      typeLabel: r.typeLabel(prop),
      required: !!requiredSet[propName],
      description: (prop.description as string) || '',
      example: prop.example,
      placeholder,
      depth,
      isObject: isObj,
      isArrayOfObjects,
      arrayItemFields,
      value: initial,
    })

    if (isObj) {
      fields = fields.concat(flattenSchema(r, prop, path, depth + 1, location))
    }
  })

  return fields
}

// Convert Swagger parameters into editable form fields
function buildInputFields(r: Resolver, operation: SwaggerOperation): InputField[] {
  // Collect the fields produced from this schema
  let fields: InputField[] = []
  ;(operation.parameters || []).forEach((p) => {
    if (p.in === 'body' && p.schema) {
      fields = fields.concat(flattenSchema(r, p.schema, '', 0, 'body'))
    } else {
      // Normalize the Swagger parameter type used by the input control
      const paramType = p.type || 'string'

      // Choose the initial input value from the schema and example
      const initial = initialValueFor(paramType, p.example)

      // Build a concise placeholder that explains the expected value
      let placeholder: string
      if (p.example !== undefined && p.example !== null) {
        placeholder = Array.isArray(p.example)
          ? p.example.map((item) => stringifyUnknown(item)).join(', ')
          : stringifyUnknown(p.example)
      } else {
        placeholder = paramType
      }
      fields.push({
        name: p.name,
        path: p.name,
        label: p.name,
        location: p.in,
        paramType,
        typeLabel: r.typeLabel(p as unknown as SwaggerSchema),
        required: !!p.required,
        description: p.description || '',
        example: p.example,
        placeholder,
        depth: 0,
        isObject: false,
        isArrayOfObjects: false,
        arrayItemFields: [],
        value: initial,
      })
    }
  })
  return fields
}

// ── Request body / form value building ────────────────────────────────────────

function setNested(obj: Record<string, unknown>, path: string, value: unknown): void {
  // Split the dotted field path into object keys
  const parts = path.split('.')

  // Walk through the object while creating missing parent objects
  let cur = obj
  for (let i = 0; i < parts.length - 1; i++) {
    // split() yields defined strings for every in-range index; bind to a local
    // so the key is provably a string when indexing below.
    const key = parts[i]
    if (key === undefined) continue

    // Inspect the existing child before replacing or descending into it
    const next = cur[key]
    if (typeof next !== 'object' || next === null) {
      cur[key] = {}
    }
    cur = cur[key] as Record<string, unknown>
  }

  // A non-empty split() always has a last element; guard to keep the index typed.
  const lastKey = parts[parts.length - 1]
  if (lastKey !== undefined) cur[lastKey] = value
}

// ParseFieldValue turns field value into documentation data
function parseFieldValue(field: InputField): unknown {
  if (field.isObject) return undefined

  // Read the current form value before converting its type
  const raw = field.value
  if (raw === undefined || raw === null) return undefined

  // Convert the current form value into trimmed text
  const str = String(raw).trim()

  if (field.paramType === 'boolean') return str === 'true' ? true : undefined
  if (str === '') return undefined

  if (field.isArrayOfObjects) {
    try {
      // Parse structured JSON entered by the user
      const parsed = JSON.parse(str)
      return Array.isArray(parsed) ? parsed : undefined
    } catch {
      return undefined
    }
  }

  if (field.paramType === 'integer') {
    // Parse integer fields as base-10 numbers
    const n = parseInt(str, 10)
    return Number.isNaN(n) ? undefined : n
  }
  if (field.paramType === 'number') {
    // Read the current field or parse its decimal value
    const f = parseFloat(str)
    return Number.isNaN(f) ? undefined : f
  }
  if (field.paramType === 'array') {
    // Split the dotted field path into object keys
    const parts = str
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    return parts.length ? parts : undefined
  }
  return str
}

// Build a nested JSON request body from edited fields
function buildJsonBody(endpoint: Endpoint): Record<string, unknown> {
  // Build the request or response body separately from its envelope
  const body: Record<string, unknown> = {}
  endpoint.inputFields.forEach((f) => {
    if (f.location !== 'body') return

    // Convert the current field value to its request type
    const v = parseFieldValue(f)
    if (v === undefined) return
    setNested(body, f.path, v)
  })
  return body
}

// Build name-value pairs for form submissions
function buildFormFields(endpoint: Endpoint): Array<[string, string]> {
  // Collect non-file values for a form submission
  const pairs: Array<[string, string]> = []
  endpoint.inputFields.forEach((f) => {
    if (f.location !== 'formData' || f.paramType === 'file') return

    // Convert the current field value to its request type
    const v = parseFieldValue(f)
    if (v === undefined) return
    if (Array.isArray(v)) {
      v.forEach((item) => pairs.push([f.name, stringifyUnknown(item)]))
    } else {
      pairs.push([f.name, stringifyUnknown(v)])
    }
  })
  return pairs
}

// FileHint describes the file hint data used by Swagger parsing
interface FileHint {
  name: string
  filename: string
  required: boolean
}

// Describe the files selected for a multipart request
function buildFileHints(endpoint: Endpoint): FileHint[] {
  // Collect files selected for a multipart request
  const files: FileHint[] = []
  endpoint.inputFields.forEach((f) => {
    if (f.paramType !== 'file') return

    // Read the file currently selected for this input
    const chosen = endpoint.fileInputs[f.name]
    files.push({
      name: f.name,
      filename: chosen ? chosen.name : '/path/to/your/file',
      required: f.required,
    })
  })
  return files
}

// ── Snippet generators ────────────────────────────────────────────────────────

function indentJson(obj: unknown, indent: number): string {
  // Indent nested code by the requested amount
  const pad = ' '.repeat(indent)
  return JSON.stringify(obj, null, 2)
    .split('\n')
    .map((line, i) => (i === 0 ? line : pad + line))
    .join('\n')
}

// Fill path parameters and build the complete request URL
function absoluteUrl(endpoint: Endpoint, origin: string): string {
  // Build the dotted path used to write nested field values
  let path = endpoint.fullPath
  endpoint.inputFields.forEach((f) => {
    if (f.location === 'path') {
      // Read the current value before encoding it into the URL path
      const val = f.value
      if (val) path = path.replace(`{${f.name}}`, val)
    }
  })
  return origin + path
}

// Generate the cURL example for one endpoint
function curlSnippet(endpoint: Endpoint, origin: string): string {
  // Build the complete endpoint URL used by this snippet
  const url = absoluteUrl(endpoint, origin)

  // Collect snippet lines before joining them
  const lines: string[] = []
  lines.push(`curl -X ${endpoint.method} '${url}' \\`)
  lines.push("  -H 'Authorization: Bearer <TOKEN>'")

  if (endpoint.isMultipart) {
    // Collect files selected for a multipart request
    const files = buildFileHints(endpoint)

    // Build the normal form values used by this snippet
    const form = buildFormFields(endpoint)
    files.forEach((f) => {
      lines[lines.length - 1] += ' \\'
      lines.push(`  -F '${f.name}=@${f.filename}'`)
    })
    form.forEach((pair) => {
      lines[lines.length - 1] += ' \\'
      lines.push(`  -F '${pair[0]}=${pair[1].replace(/'/g, "'\\''")}'`)
    })
  } else if (endpoint.method !== 'GET' && endpoint.method !== 'DELETE') {
    // Build the request or response body separately from its envelope
    const body = buildJsonBody(endpoint)
    if (Object.keys(body).length > 0) {
      lines[lines.length - 1] += ' \\'
      lines.push("  -H 'Content-Type: application/json' \\")
      lines.push(`  -d '${indentJson(body, 5)}'`)
    }
  }
  return lines.join('\n')
}

// Generate the Python requests example for one endpoint
function pythonSnippet(endpoint: Endpoint, origin: string): string {
  // Build the complete endpoint URL used by this snippet
  const url = absoluteUrl(endpoint, origin)

  // Collect snippet lines before joining them
  const lines: string[] = ['import requests', '']

  if (endpoint.isMultipart) {
    // Collect files selected for a multipart request
    const files = buildFileHints(endpoint)

    // Build the normal form values used by this snippet
    const form = buildFormFields(endpoint)

    if (files.length > 0) {
      // Build one Python context manager for each selected file
      const openParts = files.map((f, idx) => `open('${f.filename}', 'rb') as f${idx}`)
      lines.push(`with ${openParts.join(', ')}:`)
      lines.push('  files = {')
      files.forEach((f, idx) => {
        lines.push(`    '${f.name}': f${idx},`)
      })
      lines.push('  }')
    } else {
      lines.push('files = None')
    }

    // Indent request code when file context managers are present
    const indent = files.length > 0 ? '  ' : ''

    // Convert form pairs into the object used by the Python example
    const dataObj: Record<string, unknown> = {}
    form.forEach((pair) => {
      if (dataObj[pair[0]] === undefined) {
        dataObj[pair[0]] = pair[1]
      } else if (Array.isArray(dataObj[pair[0]])) {
        ;(dataObj[pair[0]] as string[]).push(pair[1])
      } else {
        dataObj[pair[0]] = [dataObj[pair[0]] as string, pair[1]]
      }
    })
    lines.push(indent + 'data = ' + indentJson(dataObj, indent.length + 7))
    lines.push(indent + 'response = requests.request(')
    lines.push(indent + `  '${endpoint.method}',`)
    lines.push(indent + `  '${url}',`)
    lines.push(indent + "  headers={'Authorization': 'Bearer <TOKEN>'},")
    lines.push(indent + '  files=files,')
    lines.push(indent + '  data=data,')
    lines.push(indent + ')')
    lines.push(indent + 'print(response.status_code, response.json())')
  } else {
    // Build the request or response body separately from its envelope
    const body = buildJsonBody(endpoint)
    lines.push('response = requests.request(')
    lines.push(`  '${endpoint.method}',`)
    lines.push(`  '${url}',`)
    lines.push("  headers={'Authorization': 'Bearer <TOKEN>'},")
    if (Object.keys(body).length > 0) {
      lines.push('  json=' + indentJson(body, 7) + ',')
    }
    lines.push(')')
    lines.push('print(response.status_code, response.json())')
  }
  return lines.join('\n')
}

// Generate the JavaScript fetch example for one endpoint
function javascriptSnippet(endpoint: Endpoint, origin: string): string {
  // Build the complete endpoint URL used by this snippet
  const url = absoluteUrl(endpoint, origin)

  // Collect snippet lines before joining them
  const lines: string[] = []

  if (endpoint.isMultipart) {
    // Collect files selected for a multipart request
    const files = buildFileHints(endpoint)

    // Build the normal form values used by this snippet
    const form = buildFormFields(endpoint)

    lines.push('const formData = new FormData();')
    files.forEach((f) => {
      lines.push(`// Attach '${f.name}' from an <input type="file"> element:`)
      lines.push(`formData.append('${f.name}', fileInput.files[0]);`)
    })
    form.forEach((pair) => {
      lines.push(`formData.append('${pair[0]}', ${JSON.stringify(pair[1])});`)
    })
    lines.push('')
    lines.push(`const response = await fetch('${url}', {`)
    lines.push(`  method: '${endpoint.method}',`)
    lines.push("  headers: { 'Authorization': 'Bearer <TOKEN>' },")
    lines.push('  body: formData,')
    lines.push('});')
  } else {
    // Build the request or response body separately from its envelope
    const body = buildJsonBody(endpoint)
    lines.push(`const response = await fetch('${url}', {`)
    lines.push(`  method: '${endpoint.method}',`)
    lines.push('  headers: {')
    lines.push("    'Authorization': 'Bearer <TOKEN>',")
    if (Object.keys(body).length > 0) {
      lines.push("    'Content-Type': 'application/json',")
    }
    lines.push('  },')
    if (Object.keys(body).length > 0) {
      lines.push('  body: JSON.stringify(' + indentJson(body, 6) + '),')
    }
    lines.push('});')
  }
  lines.push('const data = await response.json();')
  return lines.join('\n')
}

// Generate the HTTPie example for one endpoint
function httpieSnippet(endpoint: Endpoint, origin: string): string {
  // Build the complete endpoint URL used by this snippet
  const url = absoluteUrl(endpoint, origin)

  // Collect snippet lines before joining them
  const lines: string[] = []

  if (endpoint.isMultipart) {
    // Collect files selected for a multipart request
    const files = buildFileHints(endpoint)

    // Build the normal form values used by this snippet
    const form = buildFormFields(endpoint)

    lines.push(`http --form ${endpoint.method} '${url}' \\`)
    lines.push("  'Authorization: Bearer <TOKEN>'")
    files.forEach((f) => {
      lines[lines.length - 1] += ' \\'
      lines.push(`  ${f.name}@${f.filename}`)
    })
    form.forEach((pair) => {
      lines[lines.length - 1] += ' \\'
      lines.push(`  ${pair[0]}='${pair[1].replace(/'/g, "'\\''")}'`)
    })
  } else {
    // Build the request or response body separately from its envelope
    const body = buildJsonBody(endpoint)
    lines.push(`http ${endpoint.method} '${url}' \\`)
    lines.push("  'Authorization: Bearer <TOKEN>'")
    Object.keys(body).forEach((k) => {
      // Convert the current field value to its request type
      const v = body[k]

      // Format this value as one HTTPie command argument
      const arg =
        typeof v === 'string' ? `${k}='${v.replace(/'/g, "'\\''")}'` : `${k}:=${JSON.stringify(v)}`
      lines[lines.length - 1] += ' \\'
      lines.push(`  ${arg}`)
    })
  }
  return lines.join('\n')
}

/** Renders the usage snippet for the endpoint's currently selected language tab. */
export function renderUsage(endpoint: Endpoint, origin: string): string {
  // Dispatch to the generator for the selected language, defaulting to curl
  switch (endpoint.usageTab) {
    case 'python':
      return pythonSnippet(endpoint, origin)
    case 'javascript':
      return javascriptSnippet(endpoint, origin)
    case 'httpie':
      return httpieSnippet(endpoint, origin)
    default:
      return curlSnippet(endpoint, origin)
  }
}

// ── Top-level parse ───────────────────────────────────────────────────────────

/** Parse a swagger doc into the grouped, ordered view-model the page renders. */
export function parseSwagger(
  doc: SwaggerDoc,
  config: SwaggerConfig = DEFAULT_SWAGGER_CONFIG,
): ParsedSpec {
  // Read reusable schemas declared by the Swagger document
  const definitions = doc.definitions ?? {}

  // Create the reference resolver shared by the parsing pass
  const r = createResolver(definitions, config)

  // Use the documented API title or its configured fallback
  const title = doc.info?.title || config.defaultTitle

  // Read the documented API version
  const version = doc.info?.version || ''

  // Read the documented API description
  const description = doc.info?.description || ''

  // Remove a trailing slash from the documented base path
  const basePathRaw = (doc.basePath || '').replace(/\/$/, '')

  // Fall back to the root path when no base path is documented
  const basePath = basePathRaw || '/'

  // Group parsed endpoints under their first Swagger tag
  const grouped: Record<string, EndpointGroup> = {}

  Object.entries(doc.paths || {}).forEach(([path, methods]) => {
    Object.entries(methods).forEach(([method, op]) => {
      if (HTTP_METHODS.indexOf(method) < 0) return

      // Use the first endpoint tag or the general fallback group
      const tag = (op.tags && op.tags[0]) || 'general'
      if (!grouped[tag]) {
        grouped[tag] = { tag, endpoints: [], collapsed: false }
      }

      // Read the content types accepted by this endpoint
      const consumes = op.consumes || []

      // Detect endpoints that upload multipart form data
      const isMultipart = consumes.indexOf('multipart/form-data') >= 0

      // Collect the documented responses for this endpoint
      const responses: ResponseInfo[] = []
      Object.entries(op.responses || {}).forEach(([code, resp]) => {
        // Flatten the current response schema into display fields
        let responseFields: ResponseField[] = []
        if (resp.schema) responseFields = flattenResponseFields(r, resp.schema, 0)

        // Normalize the response status code to text
        const codeStr = String(code)

        // Read the documented response description
        const desc = resp.description || ''

        // Keep the schema being parsed
        const schema = resp.schema || null
        responses.push({
          code: codeStr,
          description: desc,
          schema,
          fields: responseFields,
          example: schema ? exampleResponse(r, codeStr, desc, schema) : '',
        })
      })

      grouped[tag].endpoints.push({
        method: method.toUpperCase(),
        path,
        fullPath: basePathRaw + path,
        summary: op.summary || '',
        description: op.description || '',
        isMultipart,
        inputFields: buildInputFields(r, op),
        responses,
        fileInputs: {},
        expanded: false,
        usageTab: 'curl',
      })
    })
  })

  // Rank endpoint paths according to configured navigation order
  const endpointIndex = (path: string): number => {
    for (let i = 0; i < config.endpointOrder.length; i++) {
      if (path === config.endpointOrder[i] || path.indexOf(config.endpointOrder[i] + '/') === 0) {
        return i
      }
    }
    return config.endpointOrder.length
  }

  // Convert the tag lookup into the final ordered group list
  const groups: EndpointGroup[] = []
  Object.keys(grouped).forEach((tag) => {
    // Keys come straight from `grouped`, so the lookup is always present; bind
    // to a local to make that provable and avoid repeated indexing.
    const group = grouped[tag]
    if (!group) return
    group.endpoints.sort((a, b) => {
      // Rank the first endpoint path
      const ia = endpointIndex(a.path)

      // Rank the second endpoint path
      const ib = endpointIndex(b.path)
      if (ia !== ib) return ia - ib

      // Known methods sort first (GET, POST); anything else falls back to 2.
      const methodOrder: Record<string, number> = { GET: 0, POST: 1 }

      // Rank the first endpoint HTTP method
      const ma = methodOrder[a.method] ?? 2

      // Rank the second endpoint HTTP method
      const mb = methodOrder[b.method] ?? 2
      return ma - mb
    })
    groups.push(group)
  })
  groups.sort((a, b) => a.tag.localeCompare(b.tag))

  return { title, version, description, basePath, groups }
}
