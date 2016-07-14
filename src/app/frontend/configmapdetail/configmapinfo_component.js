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
export default class ConfigMapInfoController {
  /**
   * Constructs pettion controller info object.
   */
  constructor() {
    /**
     * Config map details. Initialized from the scope.
     * @export {!backendApi.ConfigMapDetail}
     */
    this.configMap;

    /** @export */
    this.i18n = i18n;
  }
}

/**
 * Definition object for the component that displays config map info.
 *
 * @return {!angular.Directive}
 */
export const configMapInfoComponent = {
  controller: ConfigMapInfoController,
  templateUrl: 'configmapdetail/configmapinfo.html',
  bindings: {
    /** {!backendApi.ConfigMapDetail} */
    'configMap': '=',
  },
};

const i18n = {
  /** @export {string} @desc Config map info details section name. */
  MSG_CONFIG_MAP_INFO_DETAILS_SECTION: goog.getMsg('Details'),
  /** @export {string} @desc Config map info details section name entry. */
  MSG_CONFIG_MAP_INFO_NAME_ENTRY: goog.getMsg('Name'),
  /** @export {string} @desc Config map info details section namespace entry. */
  MSG_CONFIG_MAP_INFO_NAMESPACE_ENTRY: goog.getMsg('Namespace'),
  /** @export {string} @desc Config map info details section labels entry. */
  MSG_CONFIG_MAP_INFO_LABELS_ENTRY: goog.getMsg('Labels'),
};
