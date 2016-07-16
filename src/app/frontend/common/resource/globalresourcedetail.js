// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {StateParams} from 'chrome/chrome_state';

/**
 * Parameters for this state.
 *
 * All properties are @exported and in sync with URL param names.
 */
export class GlobalStateParams extends StateParams {
  /**
   * @param {string} objectName
  */
  constructor(objectName) {
    // Base StateParams are inherited from chrome parent state. GlobalStateParams are used on
    // detail pages for non-namespaced objects, which do not require namespace to be set.
    super(undefined);

    /** @export {string} Name of this object. */
    this.objectName = objectName;
  }
}

export function appendDetailParamsToUrl(baseUrl) {
  return `${baseUrl}/:objectName`;
}
