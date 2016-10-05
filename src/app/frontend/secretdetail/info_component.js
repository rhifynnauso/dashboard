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

/**
 * @final
 */
export default class SecretInfoController {
  constructor() {
    /** @export {!backendApi.SecretDetail} Initialized from the scope. */
    this.secret;

    /** @export */
    this.i18n = i18n;
  }
}

/**
 * Definition object for the component that displays secret info.
 *
 * @return {!angular.Directive}
 */
export const secretInfoComponent = {
  controller: SecretInfoController,
  templateUrl: 'secretdetail/info.html',
  bindings: {
    /** {!backendApi.SecretDetail} */
    'secret': '=',
  },
};

const i18n = {
  /** @export {string} @desc Config map info details section name. */
  MSG_SECRET_INFO_DETAILS_SECTION: goog.getMsg('Details'),
};
