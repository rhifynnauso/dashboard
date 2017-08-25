// Copyright 2017 The Kubernetes Dashboard Authors.
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


const logsPerView = 100;
const maxLogSize = 2e9;
// Load logs from the beginning of the log file. This matters only if the log file is too large to
// be loaded completely.
const beginningOfLogFile = 'beginning';
// Load logs from the end of the log file. This matters only if the log file is too large to be
// loaded completely.
const endOfLogFile = 'end';
const oldestTimestamp = 'oldest';
const newestTimestamp = 'newest';

/**
 * Controller for the logs view.
 *
 * @final
 */
export class LogsController {
  /**
   * @param {!./service.LogsService} logsService
   * @param {!angular.$sce} $sce
   * @param {!angular.$document} $document
   * @param {!angular.$resource} $resource
   * @param {!../common/errorhandling/service.ErrorDialog} errorDialog
   * @ngInject
   */
  constructor(logsService, $sce, $document, $resource, errorDialog) {
    /** @private {!angular.$sce} */
    this.sce_ = $sce;

    /** @private {!HTMLDocument} */
    this.document_ = $document[0];

    /** @private {!angular.$resource} */
    this.resource_ = $resource;

    /** @export {!./service.LogsService} */
    this.logsService = logsService;

    /** @export */
    this.i18n = i18n;

    /** @export {!Array<string>} Log set. */
    this.logsSet;

    /** @export {!backendApi.LogDetails} */
    this.podLogs;

    /**
     * Current pod selection
     * @export {string}
     */
    this.pod;

    /**
     * Current container selection
     * @export {string}
     */
    this.container;

    /**
     * Pods and containers available for selection
     * @export {!backendApi.LogSources}
     */
    this.logSources;

    /**
     * Current page selection
     * @private {!backendApi.LogSelection}
     */
    this.currentSelection;

    /** @export {number} */
    this.topIndex = 0;

    /** @private {!../common/errorhandling/service.ErrorDialog} */
    this.errorDialog_ = errorDialog;

    /** @private {!ui.router.$stateParams} */
    this.stateParams_;

    /** @export {!kdUiRouter.$transition$} - initialized from resolve */
    this.$transition$;
  }


  $onInit() {
    this.container = this.podLogs.info.containerName;
    this.pod = this.podLogs.info.podName;
    this.stateParams_ = this.$transition$.params();
    this.updateUiModel(this.podLogs);
    this.topIndex = this.podLogs.logs.length;
  }


  /**
   * Loads maxLogSize oldest lines of logs.
   * @export
   */
  loadOldest() {
    this.loadView(beginningOfLogFile, oldestTimestamp, 0, -maxLogSize - logsPerView, -maxLogSize);
  }

  /**
   * Loads maxLogSize newest lines of logs.
   * @export
   */
  loadNewest() {
    this.loadView(endOfLogFile, newestTimestamp, 0, maxLogSize, maxLogSize + logsPerView);
  }

  /**
   * Shifts view by maxLogSize lines to the past.
   * @export
   */
  loadOlder() {
    this.loadView(
        this.currentSelection.logFilePosition, this.currentSelection.referencePoint.timestamp,
        this.currentSelection.referencePoint.lineNum,
        this.currentSelection.offsetFrom - logsPerView, this.currentSelection.offsetFrom);
  }

  /**
   * Shifts view by maxLogSize lines to the future.
   * @export
   */
  loadNewer() {
    this.loadView(
        this.currentSelection.logFilePosition, this.currentSelection.referencePoint.timestamp,
        this.currentSelection.referencePoint.lineNum, this.currentSelection.offsetTo,
        this.currentSelection.offsetTo + logsPerView);
  }

  /**
   * Downloads and loads slice of logs as specified by offsetFrom and offsetTo.
   * It works just like normal slicing, but indices are referenced relatively to certain reference
   * line.
   * So for example if reference line has index n and we want to download first 10 elements in array
   * we have to use
   * from -n to -n+10.
   * @param {string} logFilePosition
   * @param {string} referenceTimestamp
   * @param {number} referenceLinenum
   * @param {number} offsetFrom
   * @param {number} offsetTo
   * @private
   */
  loadView(logFilePosition, referenceTimestamp, referenceLinenum, offsetFrom, offsetTo) {
    let namespace = this.stateParams_.objectNamespace;

    this.resource_(`api/v1/log/${namespace}/${this.pod}/${this.container}`)
        .get(
            {
              'logFilePosition': logFilePosition,
              'referenceTimestamp': referenceTimestamp,
              'referenceLineNum': referenceLinenum,
              'offsetFrom': offsetFrom,
              'offsetTo': offsetTo,
            },
            (podLogs) => {
              this.updateUiModel(podLogs);
            });
  }

  /**
   * Updates all state parameters and sets the current log view with the data returned from the
   * backend If logs are not available sets logs to no logs available message.
   * @param {!backendApi.LogDetails} podLogs
   * @private
   */
  updateUiModel(podLogs) {
    this.podLogs = podLogs;
    this.currentSelection = podLogs.selection;
    this.logsSet = this.formatAllLogs_(podLogs.logs);
    if (podLogs.info.truncated) {
      this.errorDialog_.open(this.i18n.MSG_LOGS_TRUNCATED_WARNING, '');
    }
  }

  /**
   * Formats logs as HTML.
   *
   * @param {!Array<backendApi.LogLine>} logs
   * @return {!Array<string>}
   * @private
   */
  formatAllLogs_(logs) {
    if (logs.length === 0) {
      logs = [{timestamp: '0', content: this.i18n.MSG_LOGS_ZEROSTATE_TEXT}];
    }
    return logs.map((line) => this.formatLine_(line));
  }


  /**
   * Formats the given log line as raw HTML to display to the user.
   * @param {!backendApi.LogLine} line
   * @return {*}
   * @private
   */
  formatLine_(line) {
    // remove html and add colors
    // We know that trustAsHtml is safe here because escapedLine is escaped to
    // not contain any HTML markup, and formattedLine is the result of passing
    // ecapedLine to ansi_to_html, which is known to only add span tags.
    let escapedContent =
        this.sce_.trustAsHtml(ansi_up.ansi_to_html(this.escapeHtml_(line.content)));

    // add timestamp if needed
    let showTimestamp = this.logsService.getShowTimestamp();
    let logLine = showTimestamp ? `${line.timestamp} ${escapedContent}` : escapedContent;

    return logLine;
  }

  /**
   * Escapes an HTML string (e.g. converts "<foo>bar&baz</foo>" to
   * "&lt;foo&gt;bar&amp;baz&lt;/foo&gt;") by bouncing it through a text node.
   * @param {string} html
   * @return {string}
   * @private
   */
  escapeHtml_(html) {
    let div = this.document_.createElement('div');
    div.appendChild(this.document_.createTextNode(html));
    return div.innerHTML;
  }


  /**
   * Indicates log area font size.
   * @export
   * @return {string}
   */
  getLogsClass() {
    const logsTextSize = 'kd-logs-element';
    if (this.logsService.getCompact()) {
      return `${logsTextSize}-compact`;
    }
    return logsTextSize;
  }

  /**
   * Return proper style class for logs content.
   * @export
   * @returns {string}
   */
  getStyleClass() {
    const logsTextColor = 'kd-logs-text-color';
    if (this.logsService.getInverted()) {
      return `${logsTextColor}-invert`;
    }
    return logsTextColor;
  }

  /**
   * Return proper style class for text color icon.
   * @export
   * @returns {string}
   */
  getColorIconClass() {
    const logsTextColor = 'kd-logs-color-icon';
    if (this.logsService.getInverted()) {
      return `${logsTextColor}-invert`;
    }
    return logsTextColor;
  }

  /**
   * Return proper style class for font size icon.
   * @export
   * @returns {string}
   */
  getSizeIconClass() {
    const logsTextColor = 'kd-logs-size-icon';
    if (this.logsService.getCompact()) {
      return `${logsTextColor}-compact`;
    }
    return logsTextColor;
  }

  /**
   * Return the proper icon depending on the selection state
   * @export
   * @returns {string}
   */
  getTimestampIcon() {
    if (this.logsService.getShowTimestamp()) {
      return 'timer';
    }
    return 'timer_off';
  }


  /**
   * Execute when a user changes the container from which logs should be loaded.

   * @export
   */
  onContainerChange() {
    this.loadNewest();
  }

  /**
   * Execute when a user changes the pod from which logs should be loaded.
   *
   * @export
   */
  onPodChange() {
    this.loadNewest();
  }

  /**
   * Execute when a user changes the selected option for console font size.
   * @export
   */
  onFontSizeChange() {
    this.logsService.setCompact();
  }

  /**
   * Execute when a user changes the selected option for console color.
   * @export
   */
  onTextColorChange() {
    this.logsService.setInverted();
  }

  /**
   * Execute when a user changes the selected option for show timestamp.
   * @export
   */
  onShowTimestamp() {
    this.logsService.setShowTimestamp();
    this.logsSet = this.formatAllLogs_(this.podLogs.logs);
  }
}

/**
 * Returns component definition for logs component.
 *
 * @return {!angular.Component}
 */
export const logsComponent = {
  controller: LogsController,
  controllerAs: 'ctrl',
  templateUrl: 'logs/logs.html',
  bindings: {
    'logSources': '<',
    'podLogs': '<',
    '$transition$': '<',
  },
};

const i18n = {
  /** @export {string} @desc Text for logs card zerostate in logs page. */
  MSG_LOGS_ZEROSTATE_TEXT: goog.getMsg('The selected container has not logged any messages yet.'),
  /** @export {string} @desc Error dialog indicating that parts of the log file is missing due to memory constraints. */
  MSG_LOGS_TRUNCATED_WARNING:
      goog.getMsg('The middle part of the log file cannot be loaded, because it is too big.'),
};
