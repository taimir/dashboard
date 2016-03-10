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

// logs shown per page
const LOGS_PER_PAGE = 20;

/**
 * Controller for the logs view.
 *
 * @final
 */
export class LogsController {
  /**
   * @param {!backendApi.Logs} podLogs
   * @param {!./logs_service.LogColorInversionService} logsColorInversionService
   * @param {!./logs_state.StateParams} $stateParams
   * @ngInject
   */
  constructor(podLogs, logsColorInversionService, $log, $stateParams, $resource) {
    /** @export {!Array<string>} Log set. */
    this.logsSet = podLogs.logs;

    /** @export {int} Total logs count. */
    this.totalCount = podLogs.total;

    /** @export {int} The index of the first log on the current page. */
    this.currentIndex = podLogs.startIndex;

    /** @export {int} The count of logs currently displayed. */
    this.shownCount = podLogs.logs.length;

    /** @private {!./logs_service.LogColorInversionService} */
    this.logsColorInversionService_ = logsColorInversionService;

    /** @private {string} */
    this.namespace = $stateParams.namespace;

    /** @private {string} */
    this.pod = $stateParams.podId;

    /** @private {string} */
    this.container = $stateParams.container;

    /** @private {!angular.$log} */
    this.log_ = $log;

    /** @private {!angular.$resource} */
    this.resource_ = $resource;
  }

  /**
   * Displays the logs from the previous page.
   * @export
   */
  showPreviousPage() { this.getLogs(Math.max(0, this.currentIndex - LOGS_PER_PAGE)); }

  /**
   * Displays the logs from the first page.
   * @export
   */
  showFirstPage() { this.getLogs(0); }

  /**
   * Displays the logs from the next page.
   * @export
   */
  showNextPage() { this.getLogs(this.currentIndex + LOGS_PER_PAGE); }

  /**
   * Displays the logs from the last page.
   * @export
   */
  showLastPage() { this.getLogs(-1); }

  /**
   * Queries all secrets for the given namespace.
   * @param {int} startIndex
   * @param {int} count
   * @private
   */
  getLogs(startIndex) {
    /** @type {!angular.Resource<!backendApi.SecretsList>} */
    let resource =
        this.resource_(`api/v1/logs/${this.namespace}/${this.pod}/${this.container}/${startIndex}/${LOGS_PER_PAGE}`);
    resource.get(
        (res) => {
          this.logsSet = res.logs;
          this.totalCount = res.total;
          this.currentIndex = res.startIndex;
          this.shownCount = res.logs.length;
        },
        (err) => { this.log_.log(`Error getting logs: ${err}`); });
  }

  /**
   *
   */
  /**
   * Indicates state of log area color.
   * If false: black text is placed on white area. Otherwise colors are inverted.
   * @export
   * @return {boolean}
   */
  isTextColorInverted() { return this.logsColorInversionService_.getInverted(); }

  /**
   * Return proper style class for logs content.
   * @export
   * @returns {string}
   */
  getStyleClass() {
    const logsTextColor = 'kd-logs-text-color';
    if (this.isTextColorInverted()) {
      return `${logsTextColor}-invert`;
    }
    return `${logsTextColor}`;
  }
}
