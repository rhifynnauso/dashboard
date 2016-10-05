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

/**
 * @final
 */
export default class ReplicationControllerInfoController {
  /**
   * Constructs replication controller info object.
   */
  constructor() {
    /**
     * Replication controller details. Initialized from the scope.
     * @export {!backendApi.ReplicationControllerDetail}
     */
    this.replicationController;

    /** @export */
    this.i18n = i18n(this.replicationController);
  }

  /**
   * @return {boolean}
   * @export
   */
  areDesiredPodsRunning() {
    return this.replicationController.podInfo.running ===
        this.replicationController.podInfo.desired;
  }
}

/**
 * Definition object for the component that displays replication controller info.
 *
 * @return {!angular.Directive}
 */
export const replicationControllerInfoComponent = {
  controller: ReplicationControllerInfoController,
  templateUrl: 'replicationcontrollerdetail/replicationcontrollerinfo.html',
  bindings: {
    /** {!backendApi.ReplicationControllerDetail} */
    'replicationController': '<',
  },
};

/**
 * @param  {!backendApi.ReplicationControllerDetail} rcDetail
 * @return {!Object} a dictionary of translatable messages
 */
function i18n(rcDetail) {
  return {
    /** @export {string} @desc Subtitle 'Details' for the left section with general information
        about a replication controller on the replication controller details page.*/
    MSG_RC_DETAIL_DETAILS_SUBTITLE: goog.getMsg('Details'),
    /** @export {string} @desc Label 'Label selector' for the replication controller's selector on
        the replication controller details page.*/
    MSG_RC_DETAIL_LABEL_SELECTOR_LABEL: goog.getMsg('Label selector'),
    /** @export {string} @desc Label 'Images' for the list of images used in a replication
        controller, on its details page. */
    MSG_RC_DETAIL_IMAGES_LABEL: goog.getMsg('Images'),
    /** @export {string} @desc Subtitle 'Status' for the right section with pod status information
        on the replication controller details page.*/
    MSG_RC_DETAIL_STATUS_SUBTITLE: goog.getMsg('Status'),
    /** @export {string} @desc Label 'Pods' for the pods in a replication controller on its details
        page.*/
    MSG_RC_DETAIL_PODS_LABEL: goog.getMsg('Pods'),
    /** @export {string} @desc Label 'Pods status' for the status of the pods in a replication
       controller, on the replication controller details page.*/
    MSG_RC_DETAIL_PODS_STATUS_LABEL: goog.getMsg('Pods status'),
    /** @export {string} @desc The message says that that many pods were created
        (replication controller details page). */
    MSG_RC_DETAIL_PODS_CREATED_LABEL:
        goog.getMsg('{$podsCount} created', {'podsCount': rcDetail.podInfo.current}),
    /** @export {string} @desc The message says that that many pods are running
        (replication controller details page). */
    MSG_RC_DETAIL_PODS_RUNNING_LABEL:
        goog.getMsg('{$podsCount} running', {'podsCount': rcDetail.podInfo.running}),
    /** @export {string} @desc The message says that that many pods are pending
        (replication controller details page). */
    MSG_RC_DETAIL_PODS_PENDING_LABEL:
        goog.getMsg('{$podsCount} pending', {'podsCount': rcDetail.podInfo.pending}),
    /** @export {string} @desc The message says that that many pods have failed
        (replication controller details page). */
    MSG_RC_DETAIL_PODS_FAILED_LABEL:
        goog.getMsg('{$podsCount} failed', {'podsCount': rcDetail.podInfo.failed}),
    /** @export {string} @desc The message says that that many pods are desired to run
        (replication controller details page). */
    MSG_RC_DETAIL_PODS_DESIRED_LABEL:
        goog.getMsg('{$podsCount} desired', {'podsCount': rcDetail.podInfo.desired}),
  };
}
