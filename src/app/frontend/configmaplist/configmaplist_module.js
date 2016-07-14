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

import stateConfig from './configmaplist_stateconfig';
import filtersModule from 'common/filters/filters_module';
import componentsModule from 'common/components/components_module';
import chromeModule from 'chrome/chrome_module';
import configMapDetailModule from 'configmapdetail/configmapdetail_module';
import {configMapCardComponent} from './configmapcard_component';
import {configMapCardListComponent} from './configmapcardlist_component';

/**
 * Angular module for the Config Map list view.
 */
export default angular
    .module(
        'kubernetesDashboard.configMapList',
        [
          'ngMaterial',
          'ngResource',
          'ui.router',
          filtersModule.name,
          componentsModule.name,
          configMapDetailModule.name,
          chromeModule.name,
        ])
    .config(stateConfig)
    .component('kdConfigMapCardList', configMapCardListComponent)
    .component('kdConfigMapCard', configMapCardComponent);
