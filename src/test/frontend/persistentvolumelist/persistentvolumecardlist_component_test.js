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
import persistentVolumeListModule from 'persistentvolumelist/persistentvolumelist_module';

describe('Persistent Volume card list', () => {
  /** @type
   * {!persistentvolumelist/persistentvolumecard_component.PersistentVolumeCardListController} */
  let ctrl;

  beforeEach(() => {
    angular.mock.module(persistentVolumeListModule.name);

    angular.mock.inject(($componentController, $rootScope) => {
      /** @type {!PersistentVolumeCardListController} */
      ctrl = $componentController('kdPersistentVolumeCardList', {$scope: $rootScope}, {});
    });
  });

  it('should instantiate the controller properly', () => { expect(ctrl).not.toBeUndefined(); });

  it('should init i18n', () => { expect(ctrl.i18n).not.toBeUndefined(); });
});
