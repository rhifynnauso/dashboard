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
import {appendDetailParamsToUrl} from 'common/resource/resourcedetail';
import {stateName as configMapList, stateUrl} from 'configmaplist/configmaplist_state';

import {ActionBarController} from './actionbar_controller';
import {ConfigMapDetailController} from './configmapdetail_controller';
import {stateName} from './configmapdetail_state';

/**
 * Configures states for the config map details view.
 *
 * @param {!ui.router.$stateProvider} $stateProvider
 * @ngInject
 */
export default function stateConfig($stateProvider) {
  $stateProvider.state(stateName, {
    url: appendDetailParamsToUrl(stateUrl),
    parent: chromeStateName,
    resolve: {
      'configMapDetailResource': getConfigMapDetailResource,
      'configMapDetail': getConfigMapDetail,
    },
    data: {
      [breadcrumbsConfig]: {
        'label': '{{$stateParams.objectName}}',
        'parent': configMapList,
      },
    },
    views: {
      '': {
        controller: ConfigMapDetailController,
        controllerAs: '$ctrl',
        templateUrl: 'configmapdetail/configmapdetail.html',
      },
      [actionbarViewName]: {
        controller: ActionBarController,
        controllerAs: '$ctrl',
        templateUrl: 'configmapdetail/actionbar.html',
      },
    },
  });
}

/**
 * @param {!./../common/resource/resourcedetail.StateParams} $stateParams
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource<!backendApi.ConfigMapDetail>}
 * @ngInject
 */
export function getConfigMapDetailResource($resource, $stateParams) {
  return $resource(`api/v1/configmap/${$stateParams.objectNamespace}/${$stateParams.objectName}`);
}

/**
 * @param {!angular.Resource<!backendApi.ConfigMapDetail>} configMapDetailResource
 * @return {!angular.$q.Promise}
 * @ngInject
 */
export function getConfigMapDetail(configMapDetailResource) {
  return configMapDetailResource.get().$promise;
}
