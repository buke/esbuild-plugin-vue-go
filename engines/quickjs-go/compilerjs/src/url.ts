// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

import { parse, format } from 'url';

export class URL {
  protocol: string;
  slashes: boolean;
  auth: string;
  host: string;
  port: string;
  hostname: string;
  hash: string;
  search: string;
  query: string;
  pathname: string;
  path: string;
  href: string;
  url: string;

  origin: string;
  password: string;
  searchParams: any;
  username: string;

  constructor(url: string) {
    this.parse(url);
  }

  parse(url: string) {
    const parsed = parse(url);
    this.url = url;
    this.protocol = parsed.protocol || '';
    this.slashes = parsed.slashes || false;
    this.auth = parsed.auth || '';
    this.host = parsed.host || '';
    this.port = parsed.port || '';
    this.hostname = parsed.hostname || '';
    this.hash = parsed.hash || '';
    this.search = parsed.search || '';
    this.query = parsed.query || '';
    this.pathname = parsed.pathname || '';
    this.path = parsed.path || '';
    this.href = parsed.href || '';
  }

  toString() {
    return format(this);
  }

  toJSON() {
    return JSON.stringify(this);
  }
}

(globalThis as any).URL = URL;
