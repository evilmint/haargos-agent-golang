package jobrunner

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evilmint/haargos-agent-golang/client"
	"github.com/evilmint/haargos-agent-golang/statistics"
	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type JobRunner struct {
	haargosClient    *client.HaargosClient
	supervisorClient *client.HaargosClient
	logger           *logrus.Logger
	statistics       *statistics.Statistics
}

func NewJobRunner(logger *logrus.Logger, haargosClient *client.HaargosClient, supervisorClient *client.HaargosClient, statistics *statistics.Statistics) *JobRunner {
	return &JobRunner{
		haargosClient:    haargosClient,
		supervisorClient: supervisorClient,
		logger:           logger,
		statistics:       statistics,
	}
}

func (j *JobRunner) HandleJobs(haConfigPath string, supervisorToken string) {
	jobs, err := j.haargosClient.FetchJobs()

	if err != nil || jobs == nil {
		j.logger.Errorf("Failed collecting jobs %s", err)
	} else {
		var jobNames = ""
		j.logger.Infof("Collected %d jobs. %s", len(*jobs), jobNames)

		for _, job := range *jobs {
			if job.Type == "update_core" {
				j.updateCore(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "update_addon" {
				j.updateAddon(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "update_os" {
				j.updateOS(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "addon_stop" {
				j.stopAddon(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "addon_start" {
				j.startAddon(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "addon_uninstall" {
				j.uninstallAddon(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "addon_restart" {
				j.restartAddon(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else if job.Type == "addon_update" {
				j.updateAddon(job, j.haargosClient, j.supervisorClient, supervisorToken)
			} else {
				j.logger.Warningf("Unsupported job encountered [type=%s]", job.Type)
			}
		}
	}

	if err != nil {
		j.statistics.IncrementFailedRequestCount()
	}
}

type AddonContext struct {
	Slug string `json:"addon_id"`
}

func (j *JobRunner) stopAddon(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	var addonContext AddonContext
	if err := UnmarshalContext(job.Context, &addonContext); err != nil {
		j.logger.Errorf("Wrong context in job %s", job.Type)
		return
	}

	j.logger.Infof("Job scheduled [type=%s, slug=%s]", job.Type, addonContext.Slug)

	res, err := supervisorClient.GenericPOST(
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)},
		fmt.Sprintf("/addons/%s/stop", addonContext.Slug),
	)

	j.finalizeUpdate(res, err, addonContext, job, client)
}

func (j *JobRunner) restartAddon(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	var addonContext AddonContext
	if err := UnmarshalContext(job.Context, &addonContext); err != nil {
		j.logger.Errorf("Wrong context in job %s", job.Type)
		return
	}

	j.logger.Infof("Job scheduled [type=%s, slug=%s]", job.Type, addonContext.Slug)

	res, err := supervisorClient.GenericPOST(
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)},
		fmt.Sprintf("/addons/%s/restart", addonContext.Slug),
	)

	j.finalizeUpdate(res, err, addonContext, job, client)
}

func (j *JobRunner) startAddon(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	var addonContext AddonContext
	if err := UnmarshalContext(job.Context, &addonContext); err != nil {
		j.logger.Errorf("Wrong context in job %s", job.Type)
		return
	}

	j.logger.Infof("Job scheduled [type=%s, slug=%s]", job.Type, addonContext.Slug)

	res, err := supervisorClient.GenericPOST(
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)},
		fmt.Sprintf("/addons/%s/start", addonContext.Slug),
	)

	j.finalizeUpdate(res, err, addonContext, job, client)
}

func (j *JobRunner) uninstallAddon(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	var addonContext AddonContext
	if err := UnmarshalContext(job.Context, &addonContext); err != nil {
		j.logger.Errorf("Wrong context in job %s", job.Type)
		return
	}

	j.logger.Infof("Job scheduled [type=%s, slug=%s]", job.Type, addonContext.Slug)

	res, err := supervisorClient.GenericPOST(
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)},
		fmt.Sprintf("/addons/%s/uninstall", addonContext.Slug),
	)

	j.finalizeUpdate(res, err, addonContext, job, client)
}

func (j *JobRunner) updateAddon(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	var addonContext AddonContext
	if err := UnmarshalContext(job.Context, &addonContext); err != nil {
		j.logger.Errorf("Wrong context in job %s", job.Type)
		return
	}

	j.logger.Infof("Job scheduled [type=%s, slug=%s]", job.Type, addonContext.Slug)

	res, err := supervisorClient.GenericPOST(
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)},
		fmt.Sprintf("/addons/%s/update", addonContext.Slug),
	)

	j.finalizeUpdate(res, err, addonContext, job, client)
}

func (j *JobRunner) updateAddon(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	var addonContext AddonContext
	if err := UnmarshalContext(job.Context, &addonContext); err != nil {
		j.logger.Errorf("Wrong context in job %s", job.Type)
		return
	}

	j.logger.Infof("Job scheduled [type=%s, slug=%s]", job.Type, addonContext.Slug)

	res, err := supervisorClient.UpdateAddon(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)}, addonContext.Slug)

	j.finalizeUpdate(res, err, addonContext, job, client)
}

func (j *JobRunner) updateOS(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	j.logger.Infof("Job scheduled [type=%s]", job.Type)

	res, err := supervisorClient.UpdateOS(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})

	j.finalizeUpdate(res, err, "", job, client)
}

func (j *JobRunner) postAction(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string, path string) {
	j.logger.Infof("Job scheduled [type=%s]", job.Type)

	res, err := supervisorClient.GenericPOST(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)}, path)

	j.finalizeUpdate(res, err, "", job, client)
}

func (j *JobRunner) finalizeUpdate(res *http.Response, err error, context interface{}, job types.GenericJob, client *client.HaargosClient) {
	if err != nil {
		resString := ""

		if res != nil && res.StatusCode >= 200 && res.StatusCode < 300 {
			resString += fmt.Sprintf(", status=%s", res.Status)
		}
		if res != nil {
			j.logger.Infof("Res is not nil [status=%s]", res.Status)
		} else {
			j.logger.Infof("Res is nil")
		}

		j.logger.Errorf("Job failure [type=%s, context=%s, err=%s%s]", job.Type, context, err, resString)
	}

	if res != nil && (res.StatusCode < 500 && res.StatusCode >= 200) {
		_, err = client.CompleteJob(job)

		if err != nil {
			j.logger.Errorf("Job dequeue failed [type=%s, context=%s, err=%s]", job.Type, context, err)
		} else {
			j.logger.Infof("Job dequeue successful.")
		}
	}
}

func UnmarshalContext(context interface{}, target interface{}) error {
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return fmt.Errorf("error marshaling context: %w", err)
	}

	if err := json.Unmarshal(contextJSON, target); err != nil {
		return fmt.Errorf("error unmarshaling context into target struct: %w", err)
	}

	return nil
}

func (j *JobRunner) updateCore(job types.GenericJob, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	j.logger.Infof("Updating core")
	res, err := supervisorClient.UpdateCore(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})
	j.logger.Infof("Updating core scheduled")

	j.finalizeUpdate(res, err, "", job, client)
}
