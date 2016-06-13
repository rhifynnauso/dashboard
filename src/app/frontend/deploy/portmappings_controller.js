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
 * Controller for the port mappings directive.
 *
 * @final
 */
export default class PortMappingsController {
  /** @ngInject */
  constructor() {
    /**
     * Two way data binding from the scope.
     * @export {!Array<!backendApi.PortMapping>}
     */
    this.portMappings = [];

    /**
     * Initialized from the scope.
     * @export {!Array<string>}
     */
    this.protocols;

    /**
     * Binding to outer scope.
     * @export {boolean}
     */
    this.isExternal;

    /**
     * Available service types
     * @export {!Array<ServiceType>}
     */
    this.serviceTypes = [NO_SERVICE, INT_SERVICE, EXT_SERVICE];

    /**
     * Selected service type. Binding to outer scope.
     * @export {ServiceType}
     */
    this.serviceType = NO_SERVICE;

    /** @export */
    this.i18n = i18n;
  }

  /**
   * Call checks on port mapping:
   *  - adds new port mapping when last empty port mapping has been filled
   *  - validates port mapping
   * @param {!angular.FormController|undefined} portMappingForm
   * @param {number} portMappingIndex
   * @export
   */
  checkPortMapping(portMappingForm, portMappingIndex) {
    this.addProtocolIfNeeed_();
    this.validatePortMapping_(portMappingForm, portMappingIndex);
  }

  /**
   * @param {string} defaultProtocol
   * @return {!backendApi.PortMapping}
   * @private
   */
  newEmptyPortMapping_(defaultProtocol) {
    return {port: null, targetPort: null, protocol: defaultProtocol};
  }

  /**
   * @export
   */
  addProtocolIfNeeed_() {
    let lastPortMapping = this.portMappings[this.portMappings.length - 1];
    if (this.isPortMappingFilled_(lastPortMapping)) {
      this.portMappings.push(this.newEmptyPortMapping_(this.protocols[0]));
    }
  }

  /**
   * Validates port mapping. In case when only one port is specified it is considered as invalid.
   * @param {!angular.FormController|undefined} portMappingForm
   * @param {number} portIndex
   * @private
   */
  validatePortMapping_(portMappingForm, portIndex) {
    if (angular.isDefined(portMappingForm)) {
      /** @type {!backendApi.PortMapping} */
      let portMapping = this.portMappings[portIndex];

      /** @type {!angular.NgModelController} */
      let portElem = portMappingForm['port'];
      /** @type {!angular.NgModelController} */
      let targetPortElem = portMappingForm['targetPort'];

      /** @type {boolean} */
      let isValidPort = this.isPortMappingFilledOrEmpty_(portMapping) || !!portMapping.port;
      /** @type {boolean} */
      let isValidTargetPort =
          this.isPortMappingFilledOrEmpty_(portMapping) || !!portMapping.targetPort;

      portElem.$setValidity('empty', isValidPort);
      targetPortElem.$setValidity('empty', isValidTargetPort);
    }
  }

  /**
   * @param {number} index
   * @return {boolean}
   * @export
   */
  isRemovable(index) { return index !== (this.portMappings.length - 1); }

  /**
   * @param {number} index
   * @export
   */
  remove(index) { this.portMappings.splice(index, 1); }

  /**
   * Returns true when the given port mapping is filled by the user, i.e., is not empty.
   * @param {!backendApi.PortMapping} portMapping
   * @return {boolean}
   * @private
   */
  isPortMappingFilled_(portMapping) { return !!portMapping.port && !!portMapping.targetPort; }

  /**
   * Returns true when the given port mapping is filled or empty (both ports), false otherwise.
   * @param {!backendApi.PortMapping} portMapping
   * @return {boolean}
   * @private
   */
  isPortMappingFilledOrEmpty_(portMapping) { return !portMapping.port === !portMapping.targetPort; }

  /**
   * Change the service type. Port mappings are adjusted and external flag.
   * @export
   */
  changeServiceType() {
    // add or remove port mappings
    if (this.serviceType === NO_SERVICE) {
      this.portMappings = [];
    } else if (this.portMappings.length === 0) {
      this.portMappings = [this.newEmptyPortMapping_(this.protocols[0])];
    }

    // set flag
    this.isExternal = this.serviceType.external;
  }

  /**
   * Returns true if the given port mapping is the first in the list.
   * @param {number} index
   * @export
   */
  isFirst(index) { return (index === 0); }
}

const i18n = {
  /** @export {string} @desc Label 'Service' above the service type selection box on the deploy
     page.*/
  MSG_PORT_MAPPINGS_SERVICE_LABEL: goog.getMsg('Service'),
  /** @export {string} @desc Label 'None', which appears as an option in the service type
     selection box on the deploy page.*/
  MSG_PORT_MAPPINGS_SERVICE_TYPE_NONE_LABEL: goog.getMsg('None'),
  /** @export {string} @desc Label 'Internal', which appears as an option in the service type
     selection box on the deploy page.*/
  MSG_PORT_MAPPINGS_SERVICE_TYPE_INTERNAL_LABEL: goog.getMsg('Internal'),
  /** @export {string} @desc Label 'External', which appears as an option in the service type
     selection box on the deploy page.*/
  MSG_PORT_MAPPINGS_SERVICE_TYPE_EXTERNAL_LABEL: goog.getMsg('External'),
  /** @export {string} @desc Label 'Port', which serves as a placeholder for the port input
     field in the port mappings section on the deploy page.*/
  MSG_PORT_MAPPINGS_PORT_LABEL: goog.getMsg('Port'),
  /** @export {string} @desc This warning appears when the user specifies a non-integer port
     number in the port mappings section on the deploy page.*/
  MSG_PORT_MAPPINGS_PORT_INTEGER_WARNING: goog.getMsg('Port must be an integer'),
  /** @export {string} @desc This warning appears when the user specifies a negative port number
     in the port mappings section on the deploy page.*/
  MSG_PORT_MAPPINGS_PORT_NEGATIVE_WARNING: goog.getMsg('Port must be greater than 0'),
  /** @export {string} @desc This warning appears when a typed in port number exceeds the
     maximum allowed value (in the port mappings section on the deploy page).*/
  MSG_PORT_MAPPINGS_PORT_MAX_VALUE_WARNING: goog.getMsg('Port must be less than 65536'),
  /** @export {string} @desc This warning appears when the user specifies an empty port number
     and a target port number is already there (in the port mappings section on the deploy
     page).*/
  MSG_PORT_MAPPINGS_PORT_EMPTY_WARNING:
      goog.getMsg(`Port can't be empty when target port is specified.`),
  /** @export {string} @desc Label 'Target port', which serves as a placeholder for the target
     port input field in the port mappings section on the deploy page.*/
  MSG_PORT_MAPPINGS_TARGET_PORT_LABEL: goog.getMsg('Target port'),
  /** @export {string} @desc This warning appears when the user specifies a non-integer target
     port number in the port mappings section on the deploy page.*/
  MSG_PORT_MAPPINGS_TARGET_PORT_INTEGER_WARNING: goog.getMsg('Target port must be an integer'),
  /** @export {string} @desc This warning appears when the user specifies a negative target port
     number in the port mappings section on the deploy page.*/
  MSG_PORT_MAPPINGS_TARGET_PORT_NEGATIVE_WARNING: goog.getMsg('Target port must be greater than 0'),
  /** @export {string} @desc This warning appears when a typed in target port number exceeds the
     maximum allowed value (in the port mappings section on the deploy page).*/
  MSG_PORT_MAPPINGS_TARGET_PORT_MAX_VALUE_WARNING: goog.getMsg('Target port must be less than 65536'),
  /** @export {string} @desc This warning appears when the user specifies an empty target port
     number and a port number is already there (in the port mappings section on the deploy
     page).*/
  MSG_PORT_MAPPINGS_TARGET_PORT_EMPTY_WARNING:
      goog.getMsg(`Target port can't be empty when port is specified.`),
  /** @export {string} @desc Label 'Protocol' above the protocol selection box in the port
     mappings section, on the deploy page.*/
  MSG_PORT_MAPPINGS_PROTOCOL_LABEL: goog.getMsg('Protocol'),
  /** @export {string} @desc This warning appears when the user does not specify a protocol for
     an existing port mapping (in the port mappings section, on the deploy page).*/
  MSG_PORT_MAPPINGS_PROTOCOL_REQUIRED_WARNING: goog.getMsg('Protocol is required'),
  /** @export {string} @desc This warning appears when the user specifies an invalid protocol
     for a port mapping (in the port mappings section, on the deploy page).*/
  MSG_PORT_MAPPINGS_PROTOCOL_INVALID_WARNING: goog.getMsg('Invalid protocol'),
};

/** @final */
class ServiceType {
  constructor(label, external) {
    /** @export {string} */
    this.label = label;
    /** @export {boolean} */
    this.external = external;
  }
}

/** @type {ServiceType} */
export const NO_SERVICE = new ServiceType(i18n.MSG_PORT_MAPPINGS_SERVICE_TYPE_NONE_LABEL, false);

/** @type {ServiceType} */
export const INT_SERVICE =
    new ServiceType(i18n.MSG_PORT_MAPPINGS_SERVICE_TYPE_INTERNAL_LABEL, false);

/** @type {ServiceType} */
export const EXT_SERVICE =
    new ServiceType(i18n.MSG_PORT_MAPPINGS_SERVICE_TYPE_EXTERNAL_LABEL, true);
