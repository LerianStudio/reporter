package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"mime/multipart"
	"net/textproto"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/mongodb/template"
	"testing"
)

func createFileHeaderFromString(content, filename string) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", "application/tpl")

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, bytes.NewReader([]byte(content)))
	if err != nil {
		return nil, err
	}

	writer.Close()

	// Parse multipart body to get FileHeader
	r := multipart.NewReader(body, writer.Boundary())
	form, err := r.ReadForm(int64(body.Len()))
	if err != nil {
		return nil, err
	}

	files := form.File["file"]
	if len(files) == 0 {
		return nil, errors.New("no file found in form")
	}

	return files[0], nil
}

func Test_updateTemplateById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTempRepo := template.NewMockRepository(ctrl)
	orgId := uuid.New()
	htmlType := "html"

	tempSvc := &UseCase{
		TemplateRepo: mockTempRepo,
	}

	templateTestXML := `
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
	templateTestXMLFileHeader, _ := createFileHeaderFromString(templateTestXML, "teste_template_XML.tpl")

	tests := []struct {
		name         string
		templateFile *multipart.FileHeader
		outFormat    string
		description  string
		orgId        uuid.UUID
		tempId       uuid.UUID
		mockSetup    func()
		expectErr    bool
	}{
		{
			name:         "Success - Update outputFormat template",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			orgId:        uuid.New(),
			tempId:       uuid.New(),
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:         "Error - Update all template fail to find ouputFormat",
			templateFile: templateTestXMLFileHeader,
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
			mockSetup: func() {
				mockTempRepo.EXPECT().
					FindOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, constant.ErrInternalServer)
			},
			expectErr: true,
		},
		{
			name:         "Error - Update all template fail to outputFormat is not equal update file content",
			templateFile: templateTestXMLFileHeader,
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
			mockSetup: func() {
				htmlTypeP := &htmlType
				mockTempRepo.EXPECT().
					FindOutputFormatByID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(htmlTypeP, nil)
			},
			expectErr: true,
		},
		{
			name:         "Error - Update outputFormat template invalid",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "json",
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
			mockSetup:    func() {},
			expectErr:    true,
		},
		{
			name:         "Error - Update outputFormat template where template file content invalid",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "html",
			description:  "Template Financeiro",
			orgId:        orgId,
			tempId:       uuid.New(),
			mockSetup:    func() {},
			expectErr:    true,
		},
		{
			name:         "Error - Update template error",
			templateFile: templateTestXMLFileHeader,
			outFormat:    "xml",
			description:  "Template Atualizado",
			orgId:        orgId,
			tempId:       uuid.New(),
			mockSetup: func() {
				mockTempRepo.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(constant.ErrInternalServer)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			err := tempSvc.UpdateTemplateByID(ctx, tt.outFormat, tt.description, tt.orgId, tt.tempId, tt.templateFile)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
