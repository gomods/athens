'use strict';

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

var scsslint = require('gulp-scss-lint');
require('es6-promise').polyfill();

var runSequence = require('run-sequence');

var browserSync = require('browser-sync').create();
var reload = browserSync.reload;

var src_paths = {
  sass: ['src/scss/*.scss'],
  script: [
    'static/js/*.js',
    '!static/js/*.min.js'
  ],
};

var dest_paths = {
  style: 'static/css/',
  script: 'static/js/',
  browserSync: ''
};


gulp.task('lint:sass', function() {
  return gulp.src(src_paths.sass)
    .pipe(scsslint({
        'config': 'scss-lint.yml'
    }));
});

gulp.task('sass:style', function() {
  return gulp.src(src_paths.sass)
    .pipe($.plumber({
      errorHandler: function(err) {
        console.log(err.messageFormatted);
        this.emit('end');
      }
    }))
    .pipe($.sass( { outputStyle: 'expanded' } ).on( 'error', $.sass.logError ) )
    .pipe($.autoprefixer({
        browsers: ['last 2 versions'],
        cascade: false
    }))
    .pipe(gulp.dest(dest_paths.style))
    .pipe(browserSync.stream())
    .pipe($.cssnano())
    .pipe($.rename({ suffix: '.min' }))
    .pipe(gulp.dest(dest_paths.style));
});

gulp.task('javascript', function() {
  return gulp.src(src_paths.script)
    .pipe($.uglify().on('error', $.util.log))
    .pipe($.rename({ suffix: '.min' }))
    .pipe(gulp.dest(dest_paths.script));
});

gulp.task('lint:javascript', function() {
  return gulp.src(src_paths.script)
    .pipe($.jshint())
    .pipe($.jshint.reporter('jshint-stylish'));
});

gulp.task('browser-sync', function() {
  browserSync.init({
    server: {
      baseDir: dest_paths.browserSync
    }
  });

  gulp.watch(src_paths.sass, ['default']).on('change', reload);
});

gulp.task('lint', ['lint:sass', 'lint:javascript']);
gulp.task('sass', ['sass:style']);
gulp.task('script', ['javascript']);
gulp.task('serve', ['browser-sync']);

gulp.task('default', function(callback) {
  runSequence(
    'lint',
    'sass',
    'script',
    callback
  );
});

gulp.task('watch', function() {
  gulp.watch([src_paths.sass, src_paths.script], ['default']);
});
