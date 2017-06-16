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

export class AllocatedResourcesChartController {
  /**
   * @ngInject
   * @param {!angular.Scope} $scope
   * @param {!angular.JQLite} $element
   */
  constructor($scope, $element) {
    /** @private {!angular.Scope} */
    this.scope_ = $scope;

    /** @private {!angular.JQLite} */
    this.element_ = $element;


    /**
     * Outer graph percent. Initialized from the scope.
     * @export {number}
     */
    this.outer;

    /**
     * Inner graph percent. Initialized from the scope.
     * @export {number}
     */
    this.inner;
  }

  $onInit() {
    this.generateGraph_();
  }

  /**
   * Initializes pie chart graph. Check documentation at:
   * https://nvd3-community.github.io/nvd3/examples/documentation.html#pieChart
   *
   * @private
   */
  initPieChart_(svg, data, color, margin, ratio) {
    let size = 280;
    let chart = nv.models.pieChart()
                    .showLegend(false)
                    .showLabels(true)
                    .x((d) => {
                      return d.value;
                    })
                    .y((d) => {
                      return d.value;
                    })
                    .donut(true)
                    .donutRatio(ratio)
                    .color([color, '#ddd'])
                    .margin({top: margin, right: margin, bottom: margin, left: margin})
                    .width(size)
                    .height(size)
                    .growOnHover(false)
                    .labelType((d, i) => {
                      // Displays label only for allocated resources.
                      if (i === 0) {
                        return `${d.data.value.toFixed(2)}%`;
                      }
                      return '';
                    });

    chart.tooltip.enabled(false);

    svg.attr('height', size)
        .attr('width', size)
        .append('g')
        .datum(data)
        .transition()
        .duration(350)
        .call(chart);
  }


  /**
   * Generates graph using provided requests and limits bindings.
   * @private
   */
  generateGraph_() {
    nv.addGraph(() => {
      let svg = d3.select(this.element_[0]).append('svg');

      if (this.outer !== undefined) {
        this.initPieChart_(
            svg,
            [
              {value: this.outer},
              {value: 100 - this.outer},
            ],
            '#00c752', 0, 0.61);
      }

      if (this.inner !== undefined) {
        this.initPieChart_(
            svg,
            [
              {value: this.inner},
              {value: 100 - this.inner},
            ],
            '#326de6', 36, 0.55);
      }
    });
  }
}

/**
 * Definition object for the component that displays chart with allocated resources.
 *
 * @type {!angular.Component}
 */
export const allocatedResourcesChartComponent = {
  bindings: {
    'outer': '<',
    'inner': '<',
  },
  controller: AllocatedResourcesChartController,
  templateUrl: 'node/detail/allocatedresourceschart.html',
};
