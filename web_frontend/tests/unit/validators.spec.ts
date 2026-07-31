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

import { describe, expect, it } from 'vitest'

import {
  isIpInCidr,
  validateEmail,
  validateHost,
  validateHostOrNetwork,
  validatePort,
  validateUrl,
  validateUrlOrHost,
} from '@/lib/validators'

// Verify exact CIDR membership independently from target syntax validation
describe('isIpInCidr', () => {
  it('accepts an IPv4 address inside the network', () =>
    expect(isIpInCidr('198.51.100.42', '198.51.100.0/24')).toBe(true))
  it('rejects an IPv4 address outside the network', () =>
    expect(isIpInCidr('203.0.113.42', '198.51.100.0/24')).toBe(false))
  it('accepts an IPv6 address inside the network', () =>
    expect(isIpInCidr('2001:db8:1::42', '2001:db8::/32')).toBe(true))
  it('rejects an IPv6 address outside the network', () =>
    expect(isIpInCidr('2001:db9::42', '2001:db8::/32')).toBe(false))
  it('rejects malformed and mixed-family comparisons', () => {
    expect(isIpInCidr('not-an-ip', '198.51.100.0/24')).toBe(false)
    expect(isIpInCidr('198.51.100.42', '2001:db8::/32')).toBe(false)
  })
})

// ─── validateEmail ────────────────────────────────────────────────────────────

describe('validateEmail', () => {
  it('returns undefined for a valid e-mail', () =>
    expect(validateEmail('user@example.com')).toBeUndefined())
  it('returns undefined for a subdomain e-mail', () =>
    expect(validateEmail('user@mail.example.com')).toBeUndefined())
  it('returns undefined with surrounding whitespace (trimmed)', () =>
    expect(validateEmail('  user@example.com  ')).toBeUndefined())
  it('returns an error string for missing @', () =>
    expect(validateEmail('userexample.com')).toBeTypeOf('string'))
  it('returns an error string for missing domain', () =>
    expect(validateEmail('user@')).toBeTypeOf('string'))
  it('returns an error string for missing local part', () =>
    expect(validateEmail('@example.com')).toBeTypeOf('string'))
  it('returns an error string for empty input', () =>
    expect(validateEmail('')).toBeTypeOf('string'))
})

// ─── validatePort ─────────────────────────────────────────────────────────────

describe('validatePort', () => {
  it('returns undefined for port 0', () => expect(validatePort('0')).toBeUndefined())
  it('returns undefined for port 80', () => expect(validatePort('80')).toBeUndefined())
  it('returns undefined for port 65535', () => expect(validatePort('65535')).toBeUndefined())
  it('returns an error string for port 65536', () =>
    expect(validatePort('65536')).toBeTypeOf('string'))
  it('returns an error string for negative port', () =>
    expect(validatePort('-1')).toBeTypeOf('string'))
  it('returns an error string for fractional port', () =>
    expect(validatePort('80.5')).toBeTypeOf('string'))
  it('returns an error string for non-numeric input', () =>
    expect(validatePort('http')).toBeTypeOf('string'))
  it('returns an error string for empty input', () => expect(validatePort('')).toBeTypeOf('string'))
  it('accepts " 80 " with surrounding whitespace (trimmed internally)', () =>
    expect(validatePort(' 80 ')).toBeUndefined())
})

// ─── validateHost ─────────────────────────────────────────────────────────────

describe('validateHost', () => {
  // ── no port (default allowPort = false) ──────────────────────────────
  it('returns undefined for valid IPv4', () => expect(validateHost('192.168.1.1')).toBeUndefined())
  it('returns undefined for valid IPv6', () => expect(validateHost('2001:db8::1')).toBeUndefined())
  it('returns undefined for example.com', () => expect(validateHost('example.com')).toBeUndefined())
  it('returns undefined for localhost', () => expect(validateHost('localhost')).toBeUndefined())
  it('returns an error for empty input', () => expect(validateHost('')).toBeTypeOf('string'))
  it('returns an error for whitespace-only input', () =>
    expect(validateHost('   ')).toBeTypeOf('string'))
  it('returns an error for IPv4:port when allowPort is false', () =>
    expect(validateHost('10.0.0.1:443')).toBeTypeOf('string'))

  // ── with port (allowPort = true) ─────────────────────────────────────
  it('returns undefined for IPv4:port with allowPort', () =>
    expect(validateHost('10.0.0.1:443', true)).toBeUndefined())
  it('returns undefined for FQDN:port with allowPort', () =>
    expect(validateHost('example.com:8080', true)).toBeUndefined())
  it('returns undefined for [IPv6]:port with allowPort', () =>
    expect(validateHost('[::1]:443', true)).toBeUndefined())
  it('returns undefined for [IPv6] without port with allowPort', () =>
    expect(validateHost('[::1]', true)).toBeUndefined())
  it('returns undefined for bare host with allowPort', () =>
    expect(validateHost('example.com', true)).toBeUndefined())
  it('returns an error for invalid port with allowPort', () =>
    expect(validateHost('example.com:abc', true)).toBeTypeOf('string'))
  it('returns an error for empty with allowPort', () =>
    expect(validateHost('', true)).toBeTypeOf('string'))
})

// ─── validateHostOrNetwork ────────────────────────────────────────────────────

describe('validateHostOrNetwork', () => {
  // ── valid inputs ──────────────────────────────────────────────────────
  it('returns undefined for valid IPv4', () =>
    expect(validateHostOrNetwork('10.0.0.1')).toBeUndefined())
  it('returns undefined for valid IPv6', () =>
    expect(validateHostOrNetwork('2001:db8::1')).toBeUndefined())
  it('returns undefined for FQDN', () =>
    expect(validateHostOrNetwork('example.com')).toBeUndefined())
  it('returns undefined for localhost', () =>
    expect(validateHostOrNetwork('localhost')).toBeUndefined())
  it('returns undefined for IPv4 CIDR', () =>
    expect(validateHostOrNetwork('192.168.1.0/24')).toBeUndefined())
  it('returns undefined for IPv6 CIDR', () =>
    expect(validateHostOrNetwork('2001:db8::/32')).toBeUndefined())

  // ── invalid inputs ────────────────────────────────────────────────────
  it('returns an error for empty input', () =>
    expect(validateHostOrNetwork('')).toBeTypeOf('string'))
  it('returns an error for whitespace-only', () =>
    expect(validateHostOrNetwork('   ')).toBeTypeOf('string'))
  it('returns an error for host:port (ports not accepted here)', () =>
    expect(validateHostOrNetwork('10.0.0.1:443')).toBeTypeOf('string'))
  it('returns an error for URL', () =>
    expect(validateHostOrNetwork('https://example.com')).toBeTypeOf('string'))
  it('returns an error for FQDN with CIDR', () =>
    expect(validateHostOrNetwork('example.com/24')).toBeTypeOf('string'))
  it('returns an error for out-of-range CIDR prefix', () =>
    expect(validateHostOrNetwork('10.0.0.0/33')).toBeTypeOf('string'))
})

// ─── validateUrl ──────────────────────────────────────────────────────────────

describe('validateUrl', () => {
  // ── valid URLs ────────────────────────────────────────────────────────
  it('returns undefined for https://example.com', () =>
    expect(validateUrl('https://example.com')).toBeUndefined())
  it('returns undefined for http://example.com', () =>
    expect(validateUrl('http://example.com')).toBeUndefined())
  it('returns undefined for https://example.com:8443/path', () =>
    expect(validateUrl('https://example.com:8443/path')).toBeUndefined())
  it('returns undefined for https://10.0.0.1', () =>
    expect(validateUrl('https://10.0.0.1')).toBeUndefined())
  it('returns undefined with surrounding whitespace (trimmed)', () =>
    expect(validateUrl('  https://example.com  ')).toBeUndefined())

  // ── invalid inputs ────────────────────────────────────────────────────
  it('returns an error for bare hostname', () =>
    expect(validateUrl('example.com')).toBeTypeOf('string'))
  it('returns an error for hostname:port', () =>
    expect(validateUrl('example.com:8080')).toBeTypeOf('string'))
  it('returns an error for ftp URL', () =>
    expect(validateUrl('ftp://example.com')).toBeTypeOf('string'))
  it('returns an error for empty input', () => expect(validateUrl('')).toBeTypeOf('string'))
  it('returns an error for whitespace-only', () => expect(validateUrl('   ')).toBeTypeOf('string'))
})

// ─── validateUrlOrHost ────────────────────────────────────────────────────────

describe('validateUrlOrHost', () => {
  // ── valid as URL (allowPort = false, default) ─────────────────────────
  it('returns undefined for https://example.com', () =>
    expect(validateUrlOrHost('https://example.com')).toBeUndefined())
  it('returns undefined for http://10.0.0.1', () =>
    expect(validateUrlOrHost('http://10.0.0.1')).toBeUndefined())
  it('returns undefined for https://example.com:8443', () =>
    expect(validateUrlOrHost('https://example.com:8443')).toBeUndefined())

  // ── valid as bare host (allowPort = false) ────────────────────────────
  it('returns undefined for bare IPv4', () => expect(validateUrlOrHost('10.0.0.1')).toBeUndefined())
  it('returns undefined for bare IPv6', () =>
    expect(validateUrlOrHost('2001:db8::1')).toBeUndefined())
  it('returns undefined for FQDN', () => expect(validateUrlOrHost('example.com')).toBeUndefined())
  it('returns undefined for localhost', () =>
    expect(validateUrlOrHost('localhost')).toBeUndefined())

  // ── allowPort = false rejects host:port ──────────────────────────────
  it('returns an error for IPv4:port when allowPort is false (default)', () =>
    expect(validateUrlOrHost('10.0.0.1:443')).toBeTypeOf('string'))
  it('returns an error for FQDN:port when allowPort is false (default)', () =>
    expect(validateUrlOrHost('example.com:8080')).toBeTypeOf('string'))

  // ── allowPort = true accepts host:port ───────────────────────────────
  it('returns undefined for IPv4:port with allowPort', () =>
    expect(validateUrlOrHost('10.0.0.1:443', true)).toBeUndefined())
  it('returns undefined for FQDN:port with allowPort', () =>
    expect(validateUrlOrHost('example.com:8080', true)).toBeUndefined())
  it('returns undefined for [IPv6]:port with allowPort', () =>
    expect(validateUrlOrHost('[2001:db8::1]:443', true)).toBeUndefined())
  it('returns undefined for [IPv6] without port with allowPort', () =>
    expect(validateUrlOrHost('[::1]', true)).toBeUndefined())
  it('returns undefined for https URL with allowPort', () =>
    expect(validateUrlOrHost('https://example.com:8443', true)).toBeUndefined())
  it('returns an error for invalid port with allowPort', () =>
    expect(validateUrlOrHost('example.com:abc', true)).toBeTypeOf('string'))

  // ── empty / whitespace ────────────────────────────────────────────────
  it('returns an error for empty input', () => expect(validateUrlOrHost('')).toBeTypeOf('string'))
  it('returns an error for whitespace-only', () =>
    expect(validateUrlOrHost('   ')).toBeTypeOf('string'))

  // ── ftp / unsupported scheme ──────────────────────────────────────────
  it('returns an error for ftp URL (not http/https)', () =>
    expect(validateUrlOrHost('ftp://example.com')).toBeTypeOf('string'))

  // ── CIDR is not accepted ──────────────────────────────────────────────
  it('returns an error for CIDR notation', () =>
    expect(validateUrlOrHost('192.168.1.0/24')).toBeTypeOf('string'))
})
