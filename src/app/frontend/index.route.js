// Copyright 2015 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


/** @ngInject */
export function routerConfig($componentLoaderProvider) {
  $componentLoaderProvider.setTemplateMapping(function(name) {
    return `${ name }/${ name }.html`;
  });
}

export class RouterController {
  /** @ngInject */
  constructor($router) {
    var router = $router;
    router['config']([
      { path: '/', component: 'main' }
    ]);
  }
}

