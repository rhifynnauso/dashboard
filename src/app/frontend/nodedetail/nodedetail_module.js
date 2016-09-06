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

import chromeModule from 'chrome/chrome_module';
import componentsModule from 'common/components/components_module';
import filtersModule from 'common/filters/filters_module';
import eventsModule from 'events/events_module';

import {nodeAllocatedResourcesComponent} from './nodeallocatedresources_component';
import {nodeConditionsComponent} from './nodeconditions_component';
import stateConfig from './nodedetail_stateconfig';
import {nodeInfoComponent} from './nodeinfo_component';


/**
 * Angular module for the Node details view.
 *
 * The view shows detailed view of a Node.
 */
export default angular
    .module(
        'kubernetesDashboard.nodeDetail',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          componentsModule.name,
          chromeModule.name,
          filtersModule.name,
          eventsModule.name,
        ])
    .config(stateConfig)
    .component('kdNodeAllocatedResources', nodeAllocatedResourcesComponent)
    .component('kdNodeConditions', nodeConditionsComponent)
    .component('kdNodeInfo', nodeInfoComponent)
    .factory('kdNodeEventsResource', nodeEventsResource)
    .factory('kdNodePodsResource', nodePodsResource);

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
function nodeEventsResource($resource) {
  return $resource('api/v1/node/:name/event');
}

/**
 * @param {!angular.$resource} $resource
 * @return {!angular.Resource}
 * @ngInject
 */
function nodePodsResource($resource) {
  return $resource('api/v1/node/:name/pod');
}
