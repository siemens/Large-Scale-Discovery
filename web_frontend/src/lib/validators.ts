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

import { createI18n } from 'vue-i18n'
import { z } from 'zod'

import { stringifyUnknown } from '@/lib/collections'

// Local English defaults keep this reusable module independent from the global language catalog.
const localMessages = {
  'en-US': {
    email: "'{value}' is not a valid e-mail address.",
    passwordLength: 'Password must be at least {count} characters.',
    targets: {
      empty: 'Target cannot be empty.',
      port: 'Port must be a number between 0 and 65535.',
      host: "'{value}' is not a valid hostname or IP address.",
      hostWithPort: "'{value}' is not a valid hostname or IP address.",
      hostOrNetwork: "'{value}' is not a valid hostname, IP address or CIDR network range.",
      url: "'{value}' is not a valid HTTP(s) URL.",
      urlOrHostWithPort: "'{value}' is not a valid hostname, IP address or HTTP(s) URL.",
    },
  },
}

// Translate module-owned messages without importing the global catalog
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

interface ParsedIpAddress {
  bits: 32 | 128
  value: bigint
}

/** Validates an e-mail address and returns a user-facing error when it is invalid. */
export function validateEmail(value: string): string | undefined {
  // Trim surrounding whitespace before parsing
  const clean = value.trim()
  return _isEmail(clean) ? undefined : localI18n.global.t('email', { value: clean })
}

/** Validate the minimum password length enforced by the consuming backend field. */
export function validatePasswordLength(value: string, minimumLength: number): string | undefined {
  return value.length >= minimumLength
    ? undefined
    : localI18n.global.t('passwordLength', { count: minimumLength })
}

/** Validates a TCP port and returns a user-facing error when it is invalid. */
export function validatePort(value: string): string | undefined {
  // Trim surrounding whitespace before parsing
  const clean = value.trim()
  return _isPort(clean) ? undefined : localI18n.global.t('targets.port')
}

/**
 * Validates a host and returns a user-facing error when it is invalid.
 * Pass `allowPort: true` to also accept an optional trailing port (e.g. `host:443`).
 */
export function validateHost(value: string, allowPort = false): string | undefined {
  // Trim surrounding whitespace before parsing
  const clean = value.trim()
  if (!clean) return localI18n.global.t('targets.host', { value: clean })
  if (allowPort) {
    return _isHost(clean, true)
      ? undefined
      : localI18n.global.t('targets.hostWithPort', { value: clean })
  }
  return _isIPv4(clean) || _isIPv6(clean) || _isFqdn(clean)
    ? undefined
    : localI18n.global.t('targets.host', { value: clean })
}

/** Validates an IPv4 address, IPv6 address, FQDN, or CIDR network range and returns a user-facing error when it is invalid. */
export function validateHostOrNetwork(value: string): string | undefined {
  // Trim surrounding whitespace before parsing
  const clean = value.trim()
  if (!clean) return localI18n.global.t('targets.empty')
  return _isHostOrNetwork(clean)
    ? undefined
    : localI18n.global.t('targets.hostOrNetwork', { value: clean })
}

/** Validates an HTTP(S) URL and returns a user-facing error when it is invalid. */
export function validateUrl(value: string): string | undefined {
  // Trim surrounding whitespace before parsing
  const clean = value.trim()
  if (!clean) return localI18n.global.t('targets.empty')
  return _isHttpUrl(clean) ? undefined : localI18n.global.t('targets.url', { value: clean })
}

/**
 * Validates an HTTP(S) URL or a bare host and returns a user-facing error when it is invalid.
 * Pass `allowPort: true` to also accept an optional trailing port on the host (e.g. `host:443`).
 */
export function validateUrlOrHost(value: string, allowPort = false): string | undefined {
  // Trim surrounding whitespace before parsing
  const clean = value.trim()
  if (!clean) return localI18n.global.t('targets.empty')
  if (_isHttpUrl(clean)) return undefined
  return (allowPort ? _isHost(clean, true) : _isIPv4(clean) || _isIPv6(clean) || _isFqdn(clean))
    ? undefined
    : localI18n.global.t('targets.urlOrHostWithPort', { value: clean })
}

/** Returns whether an IPv4 or IPv6 address belongs to the supplied CIDR network. */
export function isIpInCidr(address: string, cidr: string): boolean {
  // Parse both values and reject malformed or mixed-family comparisons
  const target = _parseIpAddress(address)
  const slashIndex = cidr.indexOf('/')
  if (!target || slashIndex < 1 || slashIndex !== cidr.lastIndexOf('/')) return false

  // Require a valid network address and a prefix supported by that address family
  const network = _parseIpAddress(cidr.slice(0, slashIndex))
  const prefixText = cidr.slice(slashIndex + 1)
  if (!network || network.bits !== target.bits || !/^\d+$/u.test(prefixText)) return false

  // Compare only the network portion of both addresses
  const prefix = Number(prefixText)
  if (prefix > network.bits) return false
  const hostBits = BigInt(network.bits - prefix)
  return target.value >> hostBits === network.value >> hostBits
}

//
//
// Please use the public validators above.
// Do not use the private functions below.
// Validation errors for shared components should
// be managed centrally here within validators.ts
//
//

/** Returns whether a value is a syntactically valid e-mail address. */
function _isEmail(value: unknown): boolean {
  // Zod's built-in email rule validates RFC 5321 syntax
  const emailSchema = z.email()
  return emailSchema.safeParse(stringifyUnknown(value)).success
}

/** Returns whether a value is a valid TCP port in the inclusive range 0 through 65535. */
function _isPort(value: unknown): boolean {
  // Require digit-only strings so values like "80.0" or " 80" are rejected
  const text = stringifyUnknown(value).trim()
  if (!/^\d+$/.test(text)) return false
  return Number(text) <= 65_535
}

/** Returns whether a value is a valid dotted-decimal IPv4 address. */
function _isIPv4(value: unknown): boolean {
  // Dotted-decimal IPv4 address: four decimal octets.
  const ipv4Pattern = /^(\d{1,3})(\.\d{1,3}){3}$/

  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()
  if (!text) return false
  return ipv4Pattern.test(text) && text.split('.').every((octet) => Number(octet) <= 255)
}

/** Returns whether a value is a syntactically valid IPv6 address. */
function _isIPv6(value: unknown): boolean {
  // IPv6 address: compressed (::) and full 8-group forms, but not mixed IPv4-mapped addresses.
  const ipv6Pattern =
    /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|(([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{1,4})?::(([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{1,4})?)$/

  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()
  return text.length > 0 && ipv6Pattern.test(text)
}

/** Parses an IPv4 or IPv6 address into a comparable integer and family size. */
function _parseIpAddress(value: unknown): ParsedIpAddress | undefined {
  // Convert IPv4 octets directly into one unsigned 32-bit value
  const text = stringifyUnknown(value).trim()
  if (_isIPv4(text)) {
    const parsed = text
      .split('.')
      .reduce((result, octet) => (result << 8n) | BigInt(Number(octet)), 0n)
    return { bits: 32, value: parsed }
  }

  // Expand the one optional IPv6 compression marker to eight complete groups
  if (!_isIPv6(text)) return undefined
  const [leftText = '', rightText = ''] = text.split('::')
  const left = leftText ? leftText.split(':') : []
  const right = rightText ? rightText.split(':') : []
  const missingGroups = 8 - left.length - right.length
  const groups = [...left, ...Array<string>(missingGroups).fill('0'), ...right]

  // Combine the eight 16-bit groups into one unsigned 128-bit value
  const parsed = groups.reduce(
    (result, group) => (result << 16n) | BigInt(Number.parseInt(group, 16)),
    0n,
  )
  return { bits: 128, value: parsed }
}

/**
 * Returns whether a value is a valid fully qualified domain name or `localhost`.
 * IP addresses are intentionally excluded; use {@link _isHost} when they are also allowed.
 */
function _isFqdn(value: unknown): boolean {
  // DNS hostname: labels of up to 63 alphanumeric-or-hyphen characters, total length up to 253.
  const hostnamePattern =
    /^(?=.{1,253}$)([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/

  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()
  if (!text) return false
  if (text === 'localhost') return true
  if (!text.includes('.') || /^[\d.]+$/.test(text)) return false
  return hostnamePattern.test(text)
}

/** Returns whether a value is an IP CIDR range with a decimal prefix length. */
function _isCidr(value: unknown): boolean {
  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()

  // Require exactly one slash with at least one character before it
  const slashIndex = text.indexOf('/')
  if (slashIndex < 1 || slashIndex !== text.lastIndexOf('/')) return false

  // Prefix length must be a plain decimal integer
  const prefixText = text.slice(slashIndex + 1)
  if (!/^\d+$/.test(prefixText)) return false

  // A colon in the host part identifies IPv6; everything else is treated as IPv4.
  // Each family then enforces both address syntax and its own maximum prefix length.
  const prefix = Number(prefixText)
  const host = text.slice(0, slashIndex)
  return host.includes(':') ? prefix <= 128 && _isIPv6(host) : prefix <= 32 && _isIPv4(host)
}

/**
 * Returns whether a value is an IPv4 address, IPv6 address, FQDN, or `localhost`.
 * Pass `allowPort = true` to also accept an optional trailing port (e.g. `host:443`).
 * IPv6 with a port must use the bracketed `[address]:port` form; `[address]` alone is also valid.
 */
function _isHost(value: unknown, allowPort = false): boolean {
  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()
  if (!text) return false

  // A bare host without a port is always valid
  if (_isIPv4(text) || _isIPv6(text) || _isFqdn(text)) return true

  // Without the port flag, bare-host checks are all we do
  if (!allowPort) return false

  // IPv6 with a port must use the bracketed [address]:port form;
  // [address] without a port is also valid — brackets disambiguate the colons
  if (text.startsWith('[')) {
    const match = /^\[([^\]]+)](?::(\d+))?$/.exec(text)
    if (!match || !_isIPv6(match[1] ?? '')) return false
    return !(match[2] !== undefined && !_isPort(match[2]))
  }

  // IPv4 and FQDN use host:port with exactly one colon
  const colonIndex = text.indexOf(':')
  if (colonIndex < 1 || colonIndex !== text.lastIndexOf(':')) return false
  const host = text.slice(0, colonIndex)
  return (_isIPv4(host) || _isFqdn(host)) && _isPort(text.slice(colonIndex + 1))
}

/**
 * Returns whether a value is an IP address, host name, or IP CIDR network.
 * CIDR notation on host names is not accepted.
 */
function _isHostOrNetwork(value: unknown): boolean {
  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()
  if (!text) return false
  return _isIPv4(text) || _isIPv6(text) || _isFqdn(text) || _isCidr(text)
}

/** Returns whether a value is a complete HTTP(S) URL with a supported host and port. */
function _isHttpUrl(value: unknown): boolean {
  // Normalize the supplied text before validating it
  const text = stringifyUnknown(value).trim()

  // Extract the authority (host + optional port) from the URL.
  // Validating it directly avoids `new URL()` silently repairing malformed inputs
  // like `http:host.test` that the backend would reject.
  const authorityMatch = /^https?:\/\/([^/?#]+)/i.exec(text)
  if (!authorityMatch) return false
  const authority = authorityMatch[1] ?? ''

  // Strip optional userinfo (user:pass@) before the host
  const hostAndPort = authority.slice(authority.lastIndexOf('@') + 1)
  if (!hostAndPort || /[\s\\]/.test(hostAndPort)) return false

  if (hostAndPort.startsWith('[')) {
    // IPv6 host — must be in [address] or [address]:port form
    const match = /^\[([^\]]+)](?::(\d+))?$/.exec(hostAndPort)
    if (!match || !_isIPv6(match[1] ?? '')) return false
    if (match[2] !== undefined && !_isPort(match[2])) return false
  } else {
    // IPv4 or FQDN — split on the last colon to separate an optional port
    const colonIndex = hostAndPort.indexOf(':')
    if (colonIndex !== hostAndPort.lastIndexOf(':')) return false
    const host = colonIndex < 0 ? hostAndPort : hostAndPort.slice(0, colonIndex)
    if (!_isIPv4(host) && !_isFqdn(host)) return false
    if (colonIndex >= 0 && !_isPort(hostAndPort.slice(colonIndex + 1))) return false
  }

  try {
    // Final structural check via the browser's URL parser
    const url = new URL(text)
    return (
      (url.protocol === 'http:' || url.protocol === 'https:') &&
      url.hostname.length > 0 &&
      !url.hostname.endsWith('.')
    )
  } catch {
    return false
  }
}
