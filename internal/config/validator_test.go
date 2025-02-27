package config

import (
	"testing"

	"github.com/cuongtl1992/grp-cli/internal/models"
)

func TestValidatePlan(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		plan    *models.Plan
		wantErr bool
	}{
		{
			name: "valid plan",
			plan: &models.Plan{
				APIVersion: "v1",
				Kind:       "ReleasePlan",
				Metadata: models.Metadata{
					Name: "test-plan",
				},
				Stages: []models.Stage{
					{
						Name: "test-stage",
						Jobs: []models.Job{
							{
								Name: "test-job",
								Type: "test-type",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			plan: &models.Plan{
				Kind: "ReleasePlan",
				Metadata: models.Metadata{
					Name: "test-plan",
				},
				Stages: []models.Stage{
					{
						Name: "test-stage",
						Jobs: []models.Job{
							{
								Name: "test-job",
								Type: "test-type",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing kind",
			plan: &models.Plan{
				APIVersion: "v1",
				Metadata: models.Metadata{
					Name: "test-plan",
				},
				Stages: []models.Stage{
					{
						Name: "test-stage",
						Jobs: []models.Job{
							{
								Name: "test-job",
								Type: "test-type",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing metadata name",
			plan: &models.Plan{
				APIVersion: "v1",
				Kind:       "ReleasePlan",
				Stages: []models.Stage{
					{
						Name: "test-stage",
						Jobs: []models.Job{
							{
								Name: "test-job",
								Type: "test-type",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty stages",
			plan: &models.Plan{
				APIVersion: "v1",
				Kind:       "ReleasePlan",
				Metadata: models.Metadata{
					Name: "test-plan",
				},
				Stages: []models.Stage{},
			},
			wantErr: true,
		},
		{
			name: "duplicate stage names",
			plan: &models.Plan{
				APIVersion: "v1",
				Kind:       "ReleasePlan",
				Metadata: models.Metadata{
					Name: "test-plan",
				},
				Stages: []models.Stage{
					{
						Name: "test-stage",
						Jobs: []models.Job{{Name: "job1", Type: "type1"}},
					},
					{
						Name: "test-stage",
						Jobs: []models.Job{{Name: "job2", Type: "type2"}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid dependency",
			plan: &models.Plan{
				APIVersion: "v1",
				Kind:       "ReleasePlan",
				Metadata: models.Metadata{
					Name: "test-plan",
				},
				Stages: []models.Stage{
					{
						Name: "test-stage",
						Jobs: []models.Job{
							{
								Name:      "test-job",
								Type:      "test-type",
								DependsOn: []string{"non-existent-job"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePlan(tt.plan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
