// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the 'License');
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an 'AS IS' BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import stateConfig from './persistentvolumelist_stateconfig';
import filtersModule from 'common/filters/filters_module';
import componentsModule from 'common/components/components_module';
import chromeModule from 'chrome/chrome_module';
import persistentVolumeDetailModule from 'persistentvolumedetail/persistentvolumedetail_module';
import {persistentVolumeCardListComponent} from './persistentvolumecardlist_component';
import {persistentVolumeCardComponent} from './persistentvolumecard_component';

/**
 * Angular module for the Persistent Volume list view.
 */
export default angular
    .module(
        'kubernetesDashboard.persistentVolumeList',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          filtersModule.name,
          componentsModule.name,
          persistentVolumeDetailModule.name,
          chromeModule.name,
        ])
    .config(stateConfig)
    .component('kdPersistentVolumeCardList', persistentVolumeCardListComponent)
    .component('kdPersistentVolumeCard', persistentVolumeCardComponent)
    .factory('kdPersistentVolumeListResource', persistentVolumeListResource);

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
function persistentVolumeListResource($resource) {
  return $resource('api/v1/persistentvolume');
}
