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
 * Gulp tasks for processing and compiling frontend JavaScript files.
 */
import gulp from 'gulp';
import gulpAngularTemplatecache from 'gulp-angular-templatecache';
import gulpClosureCompiler from 'gulp-closure-compiler';
import gulpHtmlmin from 'gulp-htmlmin';
import path from 'path';
import webpackStream from 'webpack-stream';

import conf from './conf';

/**
 * Compiles frontend JavaScript files into development bundle located in {conf.paths.serve}
 * directory. This has to be done because currently browsers do not handle ES6 syntax and
 * modules correctly.
 *
 * Only dependencies of root application module are included in the bundle.
 */
gulp.task('scripts', function() {
  let webpackOptions = {
    devtool: 'inline-source-map',
    module: {
      // ES6 modules have to be preprocessed with Babel loader to work in browsers.
      loaders: [{test: /\.js$/, exclude: /node_modules/, loaders: ['babel-loader']}],
    },
    output: {filename: 'app-dev.js'},
    resolve: {
      // Set the module resolve root, so that webpack knows how to process non-relative imports.
      // Should be kept in sync with respective Closure Compiler option.
      root: conf.paths.frontendSrc,
    },
    quiet: true,
  };

  return gulp.src(path.join(conf.paths.frontendSrc, 'index_module.js'))
      .pipe(webpackStream(webpackOptions))
      .pipe(gulp.dest(conf.paths.serve));
});

/**
 * Compiles frontend JavaScript files into production bundle located in {conf.paths.prodTmp}
 * directory.
 */
gulp.task('scripts:prod', ['angular-templates'], function() {
  let closureCompilerConfig = {
    fileName: 'app.js',
    // "foo_flag: null" means that a flag is enabled.
    compilerFlags: {
      angular_pass: null,
      closure_entry_point: 'module$index_module',
      compilation_level: 'ADVANCED_OPTIMIZATIONS',
      export_local_property_definitions: null,
      externs: [
        path.join(conf.paths.nodeModules, 'google-closure-compiler/contrib/externs/angular-1.4.js'),
        path.join(
            conf.paths.nodeModules,
            'google-closure-compiler/contrib/externs/angular-1.4-http-promise_templated.js'),
        path.join(
            conf.paths.nodeModules,
            'google-closure-compiler/contrib/externs/angular-1.4-q_templated.js'),
        path.join(
            conf.paths.nodeModules, 'google-closure-compiler/contrib/externs/angular-material.js'),
        path.join(
            conf.paths.nodeModules, 'google-closure-compiler/contrib/externs/angular_ui_router.js'),
        path.join(
            conf.paths.nodeModules,
            'google-closure-compiler/contrib/externs/angular-1.4-resource.js'),
        path.join(conf.paths.externs, '**/*.js'),
      ],
      generate_exports: null,
      js_module_root: path.relative(conf.paths.base, conf.paths.frontendSrc),
      // Enable all compiler checks by default and make them errors.
      jscomp_error: '*',
      // Disable checks that are not applicable to the project.
      jscomp_off: [
        // This check does not work correctly with ES6.
        'inferredConstCheck',
        // Let ESLint handle all lint checks.
        'lintChecks',
      ],
      language_in: 'ECMASCRIPT6_STRICT',
      language_out: 'ECMASCRIPT3',
      manage_closure_dependencies: true,
      use_types_for_optimization: null,
    },
    compilerPath: path.join(conf.paths.nodeModules, 'google-closure-compiler/compiler.jar'),
    // This makes the compiler faster. Requires Java 7+.
    tieredCompilation: true,
  };

  return gulp
      .src([
        // Application source files.
        path.join(conf.paths.frontendSrc, '**/*.js'),
        // Partials generated by other tasks, e.g., Angular templates.
        path.join(conf.paths.partials, '**/*.js'),
        // Include base.js to enable some compiler functions, e.g., @export annotation handling.
        path.join(conf.paths.bowerComponents, 'google-closure-library/closure/goog/base.js'),
      ])
      .pipe(gulpClosureCompiler(closureCompilerConfig))
      .pipe(gulp.dest(conf.paths.prodTmp));
});

/**
 * Compiles Angular HTML template files into one JS file that serves them through $templateCache.
 */
gulp.task('angular-templates', function() {
  return gulp.src(path.join(conf.paths.frontendSrc, '**/!(index).html'))
      .pipe(gulpHtmlmin({
        removeComments: true,
        collapseWhitespace: true,
        conservativeCollapse: true,
      }))
      .pipe(
          gulpAngularTemplatecache(
              'angular-templates.js',
              {
                module: conf.frontend.rootModuleName,
              }))
      .pipe(gulp.dest(conf.paths.partials));
});
