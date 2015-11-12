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
 * @fileoverview Externs for Angular $resource service. They are missing from the original externs.
 * TODO(bryk): Contribute this file to Angular externs.
 *
 * @externs
 */


/**
 * @typedef {function(string):!angular.Resource}
 */
angular.$resource;



/**
 * @constructor
 * @template T
 */
angular.Resource = function() {};


/**
 * @param {!T} data
 * @param {function(!T)=} opt_callback
 * @param {function(!angular.$http.Response)=} opt_errback
 */
angular.Resource.prototype.save = function(data, opt_callback, opt_errback) {};
