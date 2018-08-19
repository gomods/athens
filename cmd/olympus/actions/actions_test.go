package actions

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gobuffalo/gocraft-work-adapter"
	"github.com/gobuffalo/suite"
	"github.com/gocraft/work"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/payloads"
)

const (
	testConfigFile = "../../../config.test.toml"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	conf := config.GetConfLogErr(testConfigFile, t)
	app, err := App(conf)
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{suite.NewAction(app)}
	suite.Run(t, as)
}

func (as *ActionSuite) Test_Cache_Miss_Route() {
	mod := &payloads.Module{}
	mod.Name = "moduleName"
	mod.Version = "1.0.0"

	worker, ok := as.App.Worker.(*gwa.Adapter)
	as.True(ok)

	// stop workers so job stays in the queue
	as.NoError(worker.Stop())

	res := as.JSON("/cachemiss").Post(mod)

	// get redis queue
	conn := worker.Enqueur.Pool.Get()
	redisQ := OlympusWorkerName + ":jobs:" + DownloadHandlerName
	defer conn.Close()

	// Fetch the job from the queue
	resp, err := conn.Do("LPOP", redisQ)
	as.NoError(err)

	var job work.Job
	bResp := resp.([]byte)
	as.NoError(json.Unmarshal(bResp, &job))

	module, ok := job.Args[workerModuleKey].(string)
	as.True(ok)
	version, ok := job.Args[workerVersionKey].(string)
	as.True(ok)

	as.Equal("moduleName", module)
	as.Equal("1.0.0", version)
	as.Equal(200, res.Code)
}

func (as *ActionSuite) Test_Push_Notification_Route() {
	p := &payloads.PushNotification{}
	p.OriginURL = "https://mycdn.com/"
	e := eventlog.Event{ID: "1", Module: "mymod", Version: "1.0.0", Time: time.Now(), Op: eventlog.OpAdd}
	p.Events = []eventlog.Event{e}

	worker, ok := as.App.Worker.(*gwa.Adapter)
	as.True(ok)

	// stop workers so job stays in the queue
	as.NoError(worker.Stop())

	// push event
	res := as.JSON("/push").Post(p)
	as.Equal(200, res.Code)

	// get redis queue
	conn := worker.Enqueur.Pool.Get()
	redisQ := OlympusWorkerName + ":jobs:" + PushNotificationHandlerName
	defer conn.Close()

	// fetch the job from the queue
	resp, err := conn.Do("LPOP", redisQ)
	as.NoError(err)

	var job work.Job
	bResp := resp.([]byte)
	as.NoError(json.Unmarshal(bResp, &job))

	pnJSON, ok := job.Args[workerPushNotificationKey].(string)
	as.True(ok)
	pn := &payloads.PushNotification{}
	b := []byte(pnJSON)
	json.Unmarshal(b, pn)

	as.Equal(p.OriginURL, pn.OriginURL)
	as.Equal(p.Events[0].Module, pn.Events[0].Module)
	as.Equal(p.Events[0].Version, pn.Events[0].Version)
}

// TODO: something like this to test Push_Notification_Job handler after mergeDB is completed

//func (as *ActionSuite) Test_Push_Notification_Job() {
//	storage, err := mem.NewStorage()
//	as.NoError(err)
//	eLog, err := mongo.NewLog("mongodb://127.0.0.1:27017")
//	pushHandler := GetProcessPushNotificationJob(storage, eLog)
//
//	p := &payloads.PushNotification{}
//	p.OriginURL = "https://mycdn.com/"
//	e := eventlog.Event{ID: "1", Module: "mymod", Version: "1.0.0", Time: time.Now(), Op: eventlog.OpAdd}
//	p.Events = []eventlog.Event{e}
//	pj, err := json.Marshal(p)
//	as.NoError(err)
//	args := worker.Args{
//		workerPushNotificationKey: string(pj),
//	}
//	pushHandler(args)
//
//	events, err := eLog.Read()
//	as.NoError(err)
//	as.Equal(1, len(events))
//}
