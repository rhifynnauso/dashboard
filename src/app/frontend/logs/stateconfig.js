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

import {stateName as chromeStateName} from 'chrome/state';
import {fillContentConfig} from 'chrome/state';
import {breadcrumbsConfig} from 'common/components/breadcrumbs/service';
import {appendDetailParamsToUrl} from 'common/resource/resourcedetail';

import {stateName} from './state';

/**
 * Configures states for the logs view.
 *
 * @param {!ui.router.$stateProvider} $stateProvider
 * @ngInject
 */
export default function stateConfig($stateProvider) {
  $stateProvider.state(stateName, {
    url: `${appendDetailParamsToUrl('/log')}/:container`,
    params: {
      'container': {
        value: null,
        squash: true,
      },
    },
    parent: chromeStateName,
    resolve: {
      'podContainers': resolvePodContainers,
      'podLogs': resolvePodLogs,
    },
    data: {
      [breadcrumbsConfig]: {
        'label': i18n.MSG_BREADCRUMBS_LOGS_LABEL,
      },
      [fillContentConfig]: true,
    },
    views: {
      '': {
        component: 'kdLogs',
      },
    },
  });
}

/**
 * @param {!./state.StateParams} $stateParams
 * @param {!angular.$resource} $resource
 * @return {!angular.$q.Promise}
 * @ngInject
 */
function resolvePodLogs($stateParams, $resource) {
  let namespace = $stateParams.objectNamespace;
  let podId = $stateParams.objectName;
  let container = $stateParams.container || '';

  /** @type {!angular.Resource<!backendApi.Logs>} */
  let resource = $resource(`api/v1/pod/${namespace}/${podId}/log/${container}`);

  return resource.get().$promise;
}

/**
 * @param {!./state.StateParams} $stateParams
 * @param {!angular.$resource} $resource
 * @return {!angular.$q.Promise}
 * @ngInject
 */
function resolvePodContainers($stateParams, $resource) {
  let namespace = $stateParams.objectNamespace;
  let podId = $stateParams.objectName;

  /** @type {!angular.Resource<!backendApi.PodContainerList>} */
  let resource = $resource(`api/v1/pod/${namespace}/${podId}/container`);
  return resource.get().$promise;
}

const i18n = {
  /** @type {string} @desc Breadcrum label for the logs view. */
  MSG_BREADCRUMBS_LOGS_LABEL: goog.getMsg('Logs'),
};
