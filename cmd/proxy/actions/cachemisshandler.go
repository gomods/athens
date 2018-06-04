package actions

import (
	"log"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
)

func cachemissHandler(next buffalo.Handler, w worker.Worker) buffalo.Handler {
	return func(c buffalo.Context) error {
		nextErr := next(c)

		if isModuleNotFoundErr(nextErr) {
			params, err := getAllPathParams(c)
			if err != nil {
				log.Println(err)
				return nextErr
			}

			// TODO: add separate worker instead of inline reports
			if err := queueCacheMissReportJob(params.module, params.version, app.Worker); err != nil {
				log.Println(err)
				return nextErr
			}

		}

		return nextErr
	}
}

func queueCacheMissReportJob(module, version string, w worker.Worker) error {
	return w.Perform(worker.Job{
		Queue:   workerQueue,
		Handler: ReporterWorkerName,
		Args: worker.Args{
			workerModuleKey:  module,
			workerVersionKey: version,
		},
	})
}

func isModuleNotFoundErr(err error) bool {
	s := err.Error()
	return strings.HasPrefix(s, "module ") && strings.HasSuffix(s, "not found")
}
