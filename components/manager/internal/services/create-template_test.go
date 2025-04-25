package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/mongodb/template"
	"testing"
	"time"
)

func Test_createTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	tempId := uuid.New()
	orgId := uuid.New()

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	timestamp := time.Now().Unix()
	templateEntity := &template.Template{
		ID:           tempId,
		OutputFormat: "xml",
		Description:  "Template Financeiro",
		FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
		CreatedAt:    time.Time{},
	}

	templateTest := `
		<?xml version="1.0" encoding="UTF-8"?>
		{%- if not transaction_id -%}
		{% set transaction_id = "01965f04-7087-735f-a284-3d3e4edc6a48" %}
		{%- endif -%}
		{%- for t in .transaction -%}
		{%- if transaction_id == "" or t.id == transaction_id -%}
		<Transacao>
			<Identificador>{{ t.id }}</Identificador>
			<Descricao>{{ t.description }}</Descricao>
			<Template>{{ t.template }}</Template>
			<DataCriacao>{{ t.created_at }}</DataCriacao>
			<Status>{{ t.status }}</Status>
			<Valor scale="{{ t.amount_scale }}">
				{{ t.amount }}
			</Valor>
			<Moeda>{{ t.asset_code }}</Moeda>
			<PlanoContas>{{ t.chart_of_accounts_group_name }}</PlanoContas>
		
			{% for org in onboarding.organization %}
			<Organizacao>
				<CNPJ>{{ org.legal_document }}</CNPJ>
				<NomeLegal>{{ org.legal_name }}</NomeLegal>
				<NomeFantasia>{{ org.doing_business_as }}</NomeFantasia>
				<Endereco>{{ org.address.line1 }}, {{ org.address.city }} - {{ org.address.state }}</Endereco>
			</Organizacao>
			{% endfor %}
		
			{% for l in onboarding.ledger %}
			<Ledger>
				<Nome>{{ l.name }}</Nome>
				<Status>{{ l.status }}</Status>
			</Ledger>
			{% endfor %}
		
			{% for a in onboarding.asset %}
			<Ativo>
				<Nome>{{ a.name }}</Nome>
				<Tipo>{{ a.type }}</Tipo>
				<Codigo>{{ a.code }}</Codigo>
			</Ativo>
			{% endfor %}
		
		</Transacao>
		{% endif %}
		{%- endfor %}
	`

	tests := []struct {
		name           string
		templateFile   string
		outFormat      string
		description    string
		orgId          uuid.UUID
		mockSetup      func()
		expectErr      bool
		expectedResult *template.Template
	}{
		{
			name:         "Success - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			orgId:        orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(templateEntity, nil)
			},
			expectErr: false,
			expectedResult: &template.Template{
				ID:           tempId,
				OutputFormat: "xml",
				Description:  "Template Financeiro",
				FileName:     fmt.Sprintf("%s_%d.tpl", tempId.String(), timestamp),
				CreatedAt:    time.Time{},
			},
		},
		{
			name:         "Error - Create a template",
			templateFile: templateTest,
			outFormat:    "xml",
			description:  "Template Financeiro",
			orgId:        orgId,
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr:      true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := tempSvc.CreateTemplate(ctx, tt.templateFile, tt.outFormat, tt.description, tt.orgId)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
