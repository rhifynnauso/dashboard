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

import resourceCardModule from 'common/components/resourcecard/resourcecard_module';

describe('Delete resource menu item', () => {
  /** @type
   * {!common/components/resourcecard/resourcecarddeletemenuitem_component.ResourceCardDeleteMenuItemController}
   */
  let ctrl;
  /** @type {!angular.$q} */
  let q;
  /** @type {!angular.Scope} */
  let scope;
  /** @type {!ui.router.$state} */
  let state;
  /** @type {!common/resource/verber_service.VerberService} */
  let kdResourceVerberService;
  /** @type {!md.$dialog}*/
  let mdDialog;

  beforeEach(() => {
    angular.mock.module(resourceCardModule.name);

    angular.mock.inject(
        ($rootScope, $componentController, _kdResourceVerberService_, $q, $state, $mdDialog) => {
          ctrl = $componentController('kdResourceCardDeleteMenuItem');
          ctrl.resourceCardCtrl = {
            objectMeta: {name: 'foo-name', namespace: 'foo-namespace'},
            typeMeta: {kind: 'foo'},
          };
          state = $state;
          kdResourceVerberService = _kdResourceVerberService_;
          scope = $rootScope;
          q = $q;
          mdDialog = $mdDialog;
        });
  });

  it('should delete the resource', () => {
    let deferred = q.defer();
    spyOn(kdResourceVerberService, 'showDeleteDialog').and.returnValue(deferred.promise);
    spyOn(state, 'reload');
    ctrl.remove();

    expect(state.reload).not.toHaveBeenCalled();
    deferred.resolve();
    scope.$digest();
    expect(state.reload).toHaveBeenCalled();
  });

  it('should ignore cancels', () => {
    let deferred = q.defer();
    spyOn(kdResourceVerberService, 'showDeleteDialog').and.returnValue(deferred.promise);
    spyOn(state, 'reload');
    spyOn(mdDialog, 'alert').and.callThrough();
    ctrl.remove();

    deferred.reject();
    scope.$digest();
    expect(state.reload).not.toHaveBeenCalled();
    expect(mdDialog.alert).not.toHaveBeenCalled();
  });

  it('should show alert window on error', () => {
    let deferred = q.defer();
    spyOn(kdResourceVerberService, 'showDeleteDialog').and.returnValue(deferred.promise);
    spyOn(state, 'reload');
    spyOn(mdDialog, 'alert').and.callThrough();
    ctrl.remove();

    deferred.reject({data: 'foo-data', statusText: 'foo-text'});
    scope.$digest();
    expect(state.reload).not.toHaveBeenCalled();
    expect(mdDialog.alert).toHaveBeenCalled();
  });
});
