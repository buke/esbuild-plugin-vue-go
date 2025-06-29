// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

export function addLeadingUnderscore(path: string) {
  const parts = path.split('/');
  const filename = parts[parts.length - 1];

  parts[parts.length - 1] = '_' + filename;

  return parts.join('/');
}

export function removeLeadingUnderscore(path: string) {
  const parts = path.split('/');
  const filename = parts[parts.length - 1];

  if (filename.startsWith('_')) {
    parts[parts.length - 1] = filename.slice(1);
  }

  return parts.join('/');
}
