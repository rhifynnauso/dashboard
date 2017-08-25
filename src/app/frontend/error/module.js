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

import chromeModule from '../chrome/module';

import {stateName, StateParams} from './state';
import stateConfig from './stateconfig';

/**
 * Angular module for error views.
 */
export default angular
    .module(
        'kubernetesDashboard.error',
        [
          'ui.router',
          chromeModule.name,
        ])
    .config(stateConfig)
    .run(errorConfig);

/**
 * Configures event catchers for the error views.
 *
 * @param {!kdUiRouter.$state} $state
 * @ngInject
 */
function errorConfig($state) {
  $state.defaultErrorHandler((err) => {
    $state.go(stateName, new StateParams(err.detail, $state.params.namespace));
  });
}
