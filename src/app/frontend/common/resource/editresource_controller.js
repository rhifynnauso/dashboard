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
 * Controller for the edit resource dialog.
 *
 * @final
 */
export class EditResourceController {
  /**
   * @param {!md.$dialog} $mdDialog
   * @param {!angular.$resource} $resource
   * @param {!angular.$http} $http
   * @param {string} resourceKindName
   * @param {string} resourceUrl
   * @ngInject
   */
  constructor($mdDialog, $resource, $http, resourceKindName, resourceUrl) {
    /** @export {string} */
    this.resourceKindName = resourceKindName;

    /** @export {Object} JSON representation of the edited resource. */
    this.data = null;

    /** @private {string} */
    this.resourceUrl = resourceUrl;

    /** @private {!md.$dialog} */
    this.mdDialog_ = $mdDialog;

    /** @private {!angular.$resource} */
    this.resource_ = $resource;

    /** @private {!angular.$http} */
    this.http_ = $http;

    this.init_();
  }

  /**
   * @private
   */
  init_() {
    let promise = this.http_.get(this.resourceUrl);
    promise.then((/** !angular.$http.Response<Object>*/ response) => {
      this.data = response.data;
    });
  }

  /**
   * @export
   */
  update() {
    return this.http_.put(this.resourceUrl, this.data)
        .then(this.mdDialog_.hide, this.mdDialog_.cancel);
  }

  /**
   * Cancels and closes the dialog.
   *
   * @export
   */
  cancel() {
    this.mdDialog_.cancel();
  }
}
