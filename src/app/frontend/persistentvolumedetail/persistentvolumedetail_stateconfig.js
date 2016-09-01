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

import {actionbarViewName, stateName as chromeStateName} from 'chrome/chrome_state';
import {breadcrumbsConfig} from 'common/components/breadcrumbs/breadcrumbs_service';
import {stateName as persistentVolumeList, stateUrl} from 'persistentvolumelist/persistentvolumelist_state';

import {ActionBarController} from './actionbar_controller';
import {PersistentVolumeDetailController} from './persistentvolumedetail_controller';
import {stateName} from './persistentvolumedetail_state';

/**
 * Configures states for the persistent volume details view.
 *
 * @param {!ui.router.$stateProvider} $stateProvider
 * @ngInject
 */
export default function stateConfig($stateProvider) {
  $stateProvider.state(stateName, {
    url: `${stateUrl}/:objectName`,
    parent: chromeStateName,
    resolve: {
      'persistentVolumeDetailResource': getPersistentVolumeDetailResource,
      'persistentVolumeDetail': getPersistentVolumeDetail,
    },
    data: {
      [breadcrumbsConfig]: {
        'label': '{{$stateParams.objectName}}',
        'parent': persistentVolumeList,
      },
    },
    views: {
      '': {
        controller: PersistentVolumeDetailController,
        controllerAs: '$ctrl',
        templateUrl: 'persistentvolumedetail/persistentvolumedetail.html',
      },
      [actionbarViewName]: {
        controller: ActionBarController,
        controllerAs: '$ctrl',
        templateUrl: 'persistentvolumedetail/actionbar.html',
      },
    },
  });
}

/**
 * @param {!./../common/resource/resourcedetail.StateParams} $stateParams
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource<!backendApi.PersistentVolumeDetail>}
 * @ngInject
 */
export function getPersistentVolumeDetailResource($resource, $stateParams) {
  return $resource(`api/v1/persistentvolume/${$stateParams.objectName}`);
}

/**
 * @param {!angular.Resource<!backendApi.PersistentVolumeDetail>} persistentVolumeDetailResource
 * @return {!angular.$q.Promise}
 * @ngInject
 */
export function getPersistentVolumeDetail(persistentVolumeDetailResource) {
  return persistentVolumeDetailResource.get().$promise;
}
