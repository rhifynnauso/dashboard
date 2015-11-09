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
 * @fileoverview Gulp tasks that serve the application.
 */
import browserSync from 'browser-sync';
import browserSyncSpa from 'browser-sync-spa';
import child from 'child_process';
import gulp from 'gulp';
import path from 'path';

import conf from './conf';


/**
 * Browser sync instance that serves the application.
 */
export const browserSyncInstance = browserSync.create();


/**
 * Currently running backend process object. Null if the backend is not running.
 *
 * @type {?child.ChildProcess}
 */
let runningBackendProcess = null;


/**
 * Initializes BrowserSync tool. Files are server from baseDir directory list. When
 * includeBowerComponents is true, requests for paths starting with '/bower_components' are
 * routed to bower components directory.
 *
 * @param {!Array<string>|string} baseDir
 * @param {boolean} includeBowerComponents
 */
function browserSyncInit(baseDir, includeBowerComponents) {
  // Enable custom support for Angular apps, e.g., history management.
  browserSyncInstance.use(browserSyncSpa({
    selector: '[ng-app]',
  }));

  let config = {
    browser: [], // Needed so that the browser does not auto-launch.
    directory: false, // Disable directory listings.
    // TODO(bryk): Add proxy to the backend here.
    startPath: '/',
    server: {
      baseDir: baseDir,
    },
  };

  if (includeBowerComponents) {
    config.server.routes = {
      '/bower_components': conf.paths.bowerComponents,
    };
  }

  browserSyncInstance.init(config);
}


/**
 * Serves the application in development mode.
 */
function serveDevelopmentMode() {
  browserSyncInit([
    conf.paths.serve,
    conf.paths.frontendSrc, // For angular templates to work.
    conf.paths.app, // For assets to work.
  ], true);
}


/**
 * Serves the application in development mode. Watches for changes in the source files to rebuild
 * development artifacts.
 */
gulp.task('serve:watch', ['watch'], serveDevelopmentMode);


/**
 * Serves the application in development mode.
 */
gulp.task('serve', ['index'], serveDevelopmentMode);


/**
 * Serves the application in production mode.
 */
gulp.task('serve:prod', ['build-frontend'], function () {
  browserSyncInit(conf.paths.distPublic, false);
});


/**
 * Spawns new backend application process and finishes the task immediately. Previously spawned
 * backend process is killed beforehand, if any.
 */
gulp.task('spawn-backend', ['backend', 'kill-backend'], function () {
  runningBackendProcess = child.spawn(path.join(conf.paths.serve, conf.backend.binaryName));

  runningBackendProcess.on('exit', function() {
    // Mark that there is no backend process running anymore.
    runningBackendProcess = null;
  });

  runningBackendProcess.stdout.on('data', function (data) {
    console.log('' + data);
  });

  runningBackendProcess.stderr.on('data', function (data) {
    console.error('' + data);
  });
});


/**
 * Kills running backend process (if any).
 */
gulp.task('kill-backend', function (doneFn) {
  if (runningBackendProcess) {
    runningBackendProcess.on('exit', function() {
      // Mark that there is no backend process running anymore.
      runningBackendProcess = null;
      // Finish the task only when the backend is actually killed.
      doneFn();
    });
    runningBackendProcess.kill();
  } else {
    doneFn();
  }
});


/**
 * Watches for changes in source files and runs Gulp tasks to rebuild them.
 */
gulp.task('watch', ['index'], function () {
  gulp.watch([path.join(conf.paths.frontendSrc, 'index.html'), 'bower.json'], ['index']);

  gulp.watch([
    path.join(conf.paths.frontendSrc, '**/*.scss'),
  ], function(event) {
    if(event.type === 'changed') {
      // If this is a file change, rebuild only styles - nothing more is needed.
      gulp.start('styles');
    } else {
      // If this is new/deleted file, everything has to be rebuilt.
      gulp.start('index');
    }
  });

  gulp.watch(path.join(conf.paths.frontendSrc, '**/*.js'), ['scripts']);
  gulp.watch(path.join(conf.paths.backendSrc, '**/*.go'), ['spawn-backend']);
});
