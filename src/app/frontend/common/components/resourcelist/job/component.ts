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

import { HttpParams } from '@angular/common/http';
import { Component, ComponentFactoryResolver, Input } from '@angular/core';
import { Event, Job, JobList } from '@api/backendapi';
import { Observable } from 'rxjs/Observable';

import { ResourceListWithStatuses } from '../../../resources/list';
import { NotificationsService } from '../../../services/global/notifications';
import { EndpointManager, Resource } from '../../../services/resource/endpoint';
import { NamespacedResourceService } from '../../../services/resource/resource';
import { MenuComponent } from '../../list/column/menu/component';
import { ListGroupIdentifiers, ListIdentifiers } from '../groupids';

@Component({
  selector: 'kd-job-list',
  templateUrl: './template.html',
})
export class JobListComponent extends ResourceListWithStatuses<JobList, Job> {
  @Input() title: string;
  @Input() endpoint = EndpointManager.resource(Resource.job, true).list();

  constructor(
    private readonly job_: NamespacedResourceService<JobList>,
    notifications: NotificationsService,
    resolver: ComponentFactoryResolver
  ) {
    super('job', notifications, resolver);
    this.id = ListIdentifiers.job;
    this.groupId = ListGroupIdentifiers.workloads;

    // Register status icon handlers
    this.registerBinding(
      this.icon.checkCircle,
      'kd-success',
      this.isInSuccessState
    );
    this.registerBinding(
      this.icon.timelapse,
      'kd-muted',
      this.isInPendingState
    );
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn(
      'namespace',
      'name',
      this.shouldShowNamespaceColumn_.bind(this)
    );
  }

  getResourceObservable(params?: HttpParams): Observable<JobList> {
    return this.job_.get(this.endpoint, undefined, undefined, params);
  }

  map(jobList: JobList): Job[] {
    return jobList.jobs;
  }

  isInErrorState(resource: Job): boolean {
    return resource.podInfo.warnings.length > 0;
  }

  isInPendingState(resource: Job): boolean {
    return (
      resource.podInfo.warnings.length === 0 && resource.podInfo.pending > 0
    );
  }

  isInSuccessState(resource: Job): boolean {
    return (
      resource.podInfo.warnings.length === 0 && resource.podInfo.pending === 0
    );
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'pods', 'age', 'images'];
  }

  hasErrors(job: Job): boolean {
    return job.podInfo.warnings.length > 0;
  }

  getEvents(job: Job): Event[] {
    return job.podInfo.warnings;
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
