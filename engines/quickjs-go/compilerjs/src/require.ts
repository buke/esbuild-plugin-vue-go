// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

import * as sass from './sass';

// attention: beaware of cannot import sass module in other place
//  cause vue will using sassRequire to import sass module,
export function sassRequire(module: string) {
  if (module === 'sass') {
    return sass;
  } else {
    return undefined;
  }
}
