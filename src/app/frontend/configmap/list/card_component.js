// Copyright 2017 The Kubernetes Dashboard Authors.
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

import {StateParams} from '../../common/resource/resourcedetail';
import {stateName} from '../../configmap/detail/state';

/**
 * Controller for the config map card.
 *
 * @final
 */
class ConfigMapCardController {
  /**
   * @param {!ui.router.$state} $state
   * @param {!angular.$interpolate} $interpolate
   * @param {!../../common/namespace/service.NamespaceService} kdNamespaceService
   * @ngInject
   */
  constructor($state, $interpolate, kdNamespaceService) {
    /**
     * Initialized from the scope.
     * @export {!backendApi.ConfigMap}
     */
    this.configMap;

    /** @private {!ui.router.$state} */
    this.state_ = $state;

    /** @private */
    this.interpolate_ = $interpolate;

    /** @private {!../../common/namespace/service.NamespaceService} */
    this.kdNamespaceService_ = kdNamespaceService;
  }

  /**
   * @return {boolean}
   * @export
   */
  areMultipleNamespacesSelected() {
    return this.kdNamespaceService_.areMultipleNamespacesSelected();
  }

  /**
   * @return {string}
   * @export
   */
  getConfigMapDetailHref() {
    return this.state_.href(
        stateName,
        new StateParams(this.configMap.objectMeta.namespace, this.configMap.objectMeta.name));
  }

  /**
   * @export
   * @param  {string} creationDate - creation date of the config map
   * @return {string} localized tooltip with the formated creation date
   */
  getCreatedAtTooltip(creationDate) {
    let filter = this.interpolate_(`{{date | date}}`);
    /** @type {string} @desc Tooltip 'Created at [some date]' showing the exact creation time of
     * config map. */
    let MSG_CONFIG_MAP_LIST_CREATED_AT_TOOLTIP =
        goog.getMsg('Created at {$creationDate}', {'creationDate': filter({'date': creationDate})});
    return MSG_CONFIG_MAP_LIST_CREATED_AT_TOOLTIP;
  }
}

/**
 * @return {!angular.Component}
 */
export const configMapCardComponent = {
  bindings: {
    'configMap': '=',
  },
  controller: ConfigMapCardController,
  templateUrl: 'configmap/list/card.html',
};
