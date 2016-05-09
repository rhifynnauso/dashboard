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

import {stateName as logsStateName} from 'logs/logs_state';
import {StateParams as LogsStateParams} from 'logs/logs_state';

/**
 * Controller for the replication controller details view.
 *
 * @final
 */
export default class ReplicationControllerDetailController {
  /**
   * @param {function(string):boolean} $mdMedia Angular Material $mdMedia service
   * @param {!backendApi.ReplicationControllerDetail} replicationControllerDetail
   * @param {!backendApi.Events} replicationControllerEvents
   * @param {!ui.router.$state} $state
   * @param {!../logs/logs_state.StateParams} $stateParams
   * @ngInject
   */
  constructor(
      $mdMedia, replicationControllerDetail, replicationControllerEvents, $state, $stateParams) {
    /** @export {function(string):boolean} */
    this.mdMedia = $mdMedia;

    /** @export {!backendApi.ReplicationControllerDetail} */
    this.replicationControllerDetail = replicationControllerDetail;

    /** @export !Array<!backendApi.Event> */
    this.events = replicationControllerEvents.events;

    /** @private {!ui.router.$state} */
    this.state_ = $state;

    /** @private {!../logs/logs_state.StateParams} */
    this.stateParams_ = $stateParams;
  }

  /**
   * @param {!backendApi.Pod} pod
   * @return {boolean}
   * @export
   */
  hasCpuUsage(pod) {
    return !!pod.metrics && !!pod.metrics.cpuUsageHistory && pod.metrics.cpuUsageHistory.length > 0;
  }

  /**
   * @param {!backendApi.Pod} pod
   * @return {boolean}
   * @export
   */
  hasMemoryUsage(pod) {
    return !!pod.metrics && !!pod.metrics.memoryUsageHistory &&
        pod.metrics.memoryUsageHistory.length > 0;
  }

  /**
   * @param {!backendApi.Pod} pod
   * @return {string}
   * @export
   */
  getPodLogsHref(pod) {
    return this.state_.href(
        logsStateName,
        new LogsStateParams(
            this.stateParams_.namespace, this.stateParams_.replicationController, pod.name));
  }
}
