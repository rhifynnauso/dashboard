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

import {resourceCardComponent} from './resourcecard_component';
import {resourceCardColumnComponent} from './resourcecardcolumn_component';
import {resourceCardColumnsComponent} from './resourcecardcolumns_component';
import {resourceCardFooterComponent} from './resourcecardfooter_component';

/**
 * Module containing common components for resource cards. A resource card should be used
 * for displaying every kind of Kubernetes resource (e.g., Services, Secrets or Replica Sets).
 */
export default angular
    .module(
        'kubernetesDashboard.common.components.resourcecard',
        [
          'ngMaterial',
        ])
    .component('kdResourceCard', resourceCardComponent)
    .component('kdResourceCardColumn', resourceCardColumnComponent)
    .component('kdResourceCardColumns', resourceCardColumnsComponent)
    .component('kdResourceCardFooter', resourceCardFooterComponent);
