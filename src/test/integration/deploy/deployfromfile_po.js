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

export default class DeployFromFilePageObject {
  constructor() {
    this.deployFromFileRadioButtonQuery = by.xpath('//md-radio-button[@value="File"]');
    this.deployFromFileRadioButton = element(this.deployFromFileRadioButtonQuery);

    this.deployButtonQuery = by.xpath('//button[@type="submit"]');
    this.deployButton = element(this.deployButtonQuery);

    this.cancelButtonQuery = by.xpath('//button[@ng-click="ctrl.cancel()"]');
    this.cancelButton = element(this.cancelButtonQuery);

    this.inputContainerQuery = by.tagName('md-input-container');
    this.inputContainer = element(this.inputContainerQuery);

    this.filePickerQuery = by.css('.kd-upload-file-picker');
    this.filePicker = element(this.filePickerQuery);

    this.mdDialogQuery = by.tagName('md-dialog');
    this.mdDialog = element(this.mdDialogQuery);
  }
}
