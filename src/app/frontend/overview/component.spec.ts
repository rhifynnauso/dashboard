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

import {
  HttpClientTestingModule,
  HttpTestingController,
} from '@angular/common/http/testing';
import { Component, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FlexLayoutModule } from '@angular/flex-layout';
import {
  MatCardModule,
  MatDividerModule,
  MatIconModule,
  MatTooltipModule,
} from '@angular/material';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import {
  AppConfig,
  CronJobList,
  DaemonSetList,
  PodList,
} from '@api/backendapi';

import { AllocationChartComponent } from '../common/components/allocationchart/component';
import { CardComponent } from '../common/components/card/component';
import {
  ListGroupIdentifiers,
  ListIdentifiers,
} from '../common/components/resourcelist/groupids';
import { ConfigService } from '../common/services/global/config';
import { NotificationsService } from '../common/services/global/notifications';

import { OverviewComponent } from './component';
import { Helper, ResourceRatioModes } from './helper';
import {
  emptyResourcesRatio,
  WorkloadStatusComponent,
} from './workloadstatus/component';

const mockDaemonSetData: DaemonSetList = {
  listMeta: { totalItems: 1 },
  daemonSets: [],
  status: { running: 1, pending: 0, succeeded: 0, failed: 0 },
  errors: [],
};

const mockPodsData: PodList = {
  listMeta: { totalItems: 12 },
  pods: [],
  cumulativeMetrics: null,
  status: { running: 9, pending: 1, succeeded: 0, failed: 2 },
  errors: [],
};

const mockCronJobsData: CronJobList = {
  listMeta: { totalItems: 18 },
  items: [],
  status: { running: 8, pending: 1, succeeded: 4, failed: 5 },
  errors: [],
};

@Component({ selector: 'kd-daemon-set-list', template: '' })
class MockDaemonSetListComponent {}

describe('OverviewComponent', () => {
  let httpMock: HttpTestingController;
  let configService: ConfigService;
  let testHostFixture: ComponentFixture<OverviewComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [
        CardComponent,
        OverviewComponent,
        MockDaemonSetListComponent,
        AllocationChartComponent,
        WorkloadStatusComponent,
      ],
      imports: [
        MatIconModule,
        MatCardModule,
        MatDividerModule,
        MatTooltipModule,
        NoopAnimationsModule,
        HttpClientTestingModule,
        FlexLayoutModule,
      ],
      providers: [ConfigService, NotificationsService],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();
    httpMock = TestBed.get(HttpTestingController);
    configService = TestBed.get(ConfigService);
  }));

  beforeEach(() => {
    configService.init();
    const configRequest = httpMock.expectOne('config');
    const config: AppConfig = { serverTime: new Date().getTime() };
    configRequest.flush(config);

    testHostFixture = TestBed.createComponent(OverviewComponent);
  });

  it('should mount with empty resourcesRatio', () => {
    const instance = testHostFixture.componentInstance;
    expect(instance.resourcesRatio).toEqual(emptyResourcesRatio);
  });

  it('should update resourcesRatio', () => {
    const instance = testHostFixture.componentInstance;

    instance.updateResourcesRatio({
      id: ListIdentifiers.daemonSet,
      groupId: ListGroupIdentifiers.workloads,
      items: mockDaemonSetData.listMeta.totalItems,
      filtered: false,
      resourceList: mockDaemonSetData,
    });

    expect(instance.resourcesRatio).toEqual({
      ...emptyResourcesRatio,
      daemonSetRatio: Helper.getResourceRatio(
        mockDaemonSetData.status,
        mockDaemonSetData.listMeta.totalItems
      ),
    });

    expect(instance.showWorkloadStatuses()).toEqual(true);
  });

  it('should update resourcesRatio on pod identifier', () => {
    // This checks the ResourceRatioModes.Completable path
    const instance = testHostFixture.componentInstance;

    instance.updateResourcesRatio({
      id: ListIdentifiers.pod,
      groupId: ListGroupIdentifiers.workloads,
      items: mockPodsData.listMeta.totalItems,
      filtered: false,
      resourceList: mockPodsData,
    });

    expect(instance.resourcesRatio).toEqual({
      ...emptyResourcesRatio,
      podRatio: Helper.getResourceRatio(
        mockPodsData.status,
        mockPodsData.listMeta.totalItems,
        ResourceRatioModes.Completable
      ),
    });

    expect(instance.showWorkloadStatuses()).toEqual(true);
  });

  it('should update resourcesRatio on cron job identifier', () => {
    // This checks the ResourceRatioModes.Suspendable path
    const instance = testHostFixture.componentInstance;

    instance.updateResourcesRatio({
      id: ListIdentifiers.cronJob,
      groupId: ListGroupIdentifiers.workloads,
      items: mockCronJobsData.listMeta.totalItems,
      filtered: false,
      resourceList: mockCronJobsData,
    });

    expect(instance.resourcesRatio).toEqual({
      ...emptyResourcesRatio,
      cronJobRatio: Helper.getResourceRatio(
        mockCronJobsData.status,
        mockCronJobsData.listMeta.totalItems,
        ResourceRatioModes.Suspendable
      ),
    });

    expect(instance.showWorkloadStatuses()).toEqual(true);
  });
});
