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
import {stateName} from '../../persistentvolume/detail/state';

/**
 * Controller for the persistent volume card.
 *
 * @final
 */
class PersistentVolumeCardController {
  /**
   * @param {!ui.router.$state} $state
   * @param {!angular.$interpolate} $interpolate
   * @ngInject
   */
  constructor($state, $interpolate) {
    /**
     * Initialized from the scope.
     * @export {!backendApi.PersistentVolume}
     */
    this.persistentVolume;

    /** @private {!ui.router.$state} */
    this.state_ = $state;

    /** @private */
    this.interpolate_ = $interpolate;
  }

  /**
   * @return {string}
   * @export
   */
  getPersistentVolumeDetailHref() {
    return this.state_.href(stateName, new StateParams('', this.persistentVolume.objectMeta.name));
  }

  /**
   * @export
   * @param  {string} creationDate - creation date of the config map
   * @return {string} localized tooltip with the formated creation date
   */
  getCreatedAtTooltip(creationDate) {
    let filter = this.interpolate_(`{{date | date}}`);
    /** @type {string} @desc Tooltip 'Created at [some date]' showing the exact creation time of
     * persistent volume. */
    let MSG_PERSISTENT_VOLUME_LIST_CREATED_AT_TOOLTIP =
        goog.getMsg('Created at {$creationDate}', {'creationDate': filter({'date': creationDate})});
    return MSG_PERSISTENT_VOLUME_LIST_CREATED_AT_TOOLTIP;
  }
}

/**
 * @return {!angular.Component}
 */
export const persistentVolumeCardComponent = {
  bindings: {
    'persistentVolume': '=',
  },
  controller: PersistentVolumeCardController,
  templateUrl: 'persistentvolume/list/card.html',
};
