package jobrunner

import (
	"testing"

	"github.com/evilmint/haargos-agent-golang/client"
	"github.com/evilmint/haargos-agent-golang/statistics"
	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

func TestJobRunner_updateCore(t *testing.T) {
	type fields struct {
		haargosClient    *client.HaargosClient
		supervisorClient *client.HaargosClient
		logger           *logrus.Logger
		statistics       *statistics.Statistics
	}
	type args struct {
		job              types.GenericJob
		client           *client.HaargosClient
		supervisorClient *client.HaargosClient
		supervisorToken  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JobRunner{
				haargosClient:    tt.fields.haargosClient,
				supervisorClient: tt.fields.supervisorClient,
				logger:           tt.fields.logger,
				statistics:       tt.fields.statistics,
			}
			j.updateCore(tt.args.job, tt.args.client, tt.args.supervisorClient, tt.args.supervisorToken)

		})
	}
}
