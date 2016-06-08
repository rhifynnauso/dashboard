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

import podsListModule from 'podlist/podlist_module';
import podDetailModule from 'poddetail/poddetail_module';

describe('Pod card list controller', () => {
  /**
   * @type {!podlist/podcardlist_component.PodCardListController}
   */
  let ctrl;

  beforeEach(() => {
    angular.mock.module(podsListModule.name);
    angular.mock.module(podDetailModule.name);

    angular.mock.inject(($componentController, $rootScope) => {
      ctrl = $componentController('kdPodCardList', {$scope: $rootScope}, {
        logsHrefFn: function() {},
      });
    });
  });

  it('should execute logs href callback function', () => {
    // given
    let pod = {name: 'test-pod'};
    spyOn(ctrl, 'logsHrefFn');

    // when
    ctrl.getPodLogsHref(pod);

    // then
    expect(ctrl.logsHrefFn).toHaveBeenCalledWith({pod: pod});
  });

  it('should execute logs href callback function', () => {
    expect(ctrl.getPodDetailHref({
      objectMeta: {
        name: 'foo-pod',
        namespace: 'foo-namespace',
      },
    })).toBe('#/pod/foo-namespace/foo-pod');
  });

  it('should check pod status correctly (succeeded is successful)', () => {
    expect(ctrl.isStatusSuccessful({
      name: 'test-pod',
      podPhase: 'Succeeded',
    })).toBeTruthy();
  });

  it('should check pod status correctly (running is successful)', () => {
    expect(ctrl.isStatusSuccessful({
      name: 'test-pod',
      podPhase: 'Running',
    })).toBeTruthy();
  });

  it('should check pod status correctly (failed isn\'t successful)', () => {
    expect(ctrl.isStatusSuccessful({
      name: 'test-pod',
      podPhase: 'Failed',
    })).toBeFalsy();
  });

  it('should check pod status correctly (pending is pending)', () => {
    expect(ctrl.isStatusPending({
      name: 'test-pod',
      podPhase: 'Pending',
    })).toBeTruthy();
  });

  it('should check pod status correctly (failed isn\'t pending)', () => {
    expect(ctrl.isStatusPending({
      name: 'test-pod',
      podPhase: 'Failed',
    })).toBeFalsy();
  });

  it('should check pod status correctly (failed is failed)', () => {
    expect(ctrl.isStatusFailed({
      name: 'test-pod',
      podPhase: 'Failed',
    })).toBeTruthy();
  });

  it('should check pod status correctly (running isn\'t failed)', () => {
    expect(ctrl.isStatusFailed({
      name: 'test-pod',
      podPhase: 'Running',
    })).toBeFalsy();
  });

  it('should format the "pod start date" tooltip correctly', () => {
    expect(ctrl.getStartedAtTooltip('2016-06-06T09:13:12Z')).toBe('Started at 6/6/16 09:13 UTC');
  });
});
