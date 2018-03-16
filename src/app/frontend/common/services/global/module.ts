// Copyright 2017 The Kubernetes Authors.
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

import {APP_INITIALIZER, Injector, NgModule} from '@angular/core';

import {AssetsService} from './assets';
import {AuthService} from './authentication';
import {AuthorizerService} from './authorizer';
import {BreadcrumbsService} from './breadcrumbs';
import {ConfigService} from './config';
import {CsrfTokenService} from './csrftoken';
import {GlobalSettingsService} from './globalsettings';
import {LocalSettingsService} from './localsettings';
import {NamespaceService} from './namespace';
import {NotificationsService} from './notifications';
import {KdStateService} from './state';
import {ThemeService} from './theme';
import {TitleService} from './title';

@NgModule({
  providers: [
    AuthorizerService, AssetsService, BreadcrumbsService, LocalSettingsService,
    GlobalSettingsService, ConfigService, TitleService, AuthService, CsrfTokenService,
    NotificationsService, ThemeService, KdStateService, NamespaceService, {
      provide: APP_INITIALIZER,
      useFactory: init,
      deps: [GlobalSettingsService, LocalSettingsService, ConfigService],
      multi: true,
    }
  ],
})
export class GlobalServicesModule {
  static injector: Injector;
  constructor(injector: Injector) {
    GlobalServicesModule.injector = injector;
  }
}

export function init(
    globalSettings: GlobalSettingsService, localSettings: LocalSettingsService,
    config: ConfigService): Function {
  return () => {
    globalSettings.init();
    localSettings.init();
    config.init();
  };
}
