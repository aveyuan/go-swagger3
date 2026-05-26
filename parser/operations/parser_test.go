package operations

import (
	"errors"
	"testing"

	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/schema"
	"github.com/stretchr/testify/assert"
)

func Test_ParseHeader(t *testing.T) {
	tests := []struct {
		name               string
		schemaParser       schema.Parser
		wantErr            bool
		errMsg             string
		expectedParameters []oas.ParameterObject
	}{
		{
			name:         "Should add parameters with ref",
			schemaParser: schema.SetupUpSchemaParseMocks(schema.GetSchemaObject(), nil),
			wantErr:      false,
			expectedParameters: []oas.ParameterObject{
				{Ref: "#/components/parameters/ContentType"},
				{Ref: "#/components/parameters/Version"},
				{Ref: "#/components/parameters/Authorization"},
			},
		},
		{
			name:         "Should return error if fails parsing the schema",
			schemaParser: schema.SetupUpSchemaParseMocks(schema.GetSchemaObject(), errors.New("someErr")),
			wantErr:      true,
			errMsg:       "someErr",
		},
		{
			name:         "Should return error schema properties are nil",
			schemaParser: schema.SetupUpSchemaParseMocks(&oas.SchemaObject{}, nil),
			wantErr:      true,
			errMsg:       "NilSchemaProperties : parseHeaders can not parse Header schema comment",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operationParser := parser{Parser: test.schemaParser}
			operationObject := &oas.OperationObject{}
			err := operationParser.parseHeaders("/test/path", "pkgName", operationObject, "comment")
			if test.wantErr {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.errMsg)
			}
			assert.Equal(t, test.expectedParameters, operationObject.Parameters)
		})
	}
}

func TestParseParamComment_FormParam(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{}

	comment := "@Param file form ignored true \"Upload file\" \"/path/to/file\""

	err := p.parseParamComment("example/pkg", "pkg", op, comment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if op.RequestBody == nil {
		t.Error("Expected RequestBody to be set for form param")
	}
}

func TestParseParamComment_BodyParam(t *testing.T) {
	mockSchema := new(MockSchemaParser)

	p := &parser{
		Parser: mockSchema,
	}

	op := &oas.OperationObject{}

	comment := "@Param user body User true \"User info\""

	mockSchema.On("RegisterType", "example/pkg", "pkg", "User").Return("UserSchemaRef", nil)

	err := p.parseParamComment("example/pkg", "pkg", op, comment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if op.RequestBody == nil {
		t.Error("Expected RequestBody to be set for body param")
	}

	mockSchema.AssertExpectations(t)
}

func TestParseParamComment_QueryParam(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{}

	comment := "@Param id query int true \"User ID\""

	err := p.parseParamComment("example/pkg", "pkg", op, comment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(op.Parameters) == 0 {
		t.Error("Expected query param to be added to operation.Parameters")
	}
}

func TestParseParamComment_InvalidComment(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{}

	comment := "@Param invalid format only"

	err := p.parseParamComment("pkg", "pkg", op, comment)
	if err == nil {
		t.Error("Expected error for invalid comment format, but got none")
	}
}

func Test_ParseResponseHeader(t *testing.T) {
	tests := []struct {
		name             string
		comment          string
		existingResponse *oas.ResponseObject

		wantErr bool
		errMsg  string

		expectedStatus string
		expectedHeader *oas.HeaderObject
	}{
		{
			name:           "Should parse response header with example",
			comment:        `200 Set-Cookie string "JWT cookie" "accessToken=eyJhbGciOi...; Path=/; HttpOnly; Secure; SameSite=Lax"`,
			wantErr:        false,
			expectedStatus: "200",
			expectedHeader: &oas.HeaderObject{
				Description: "JWT cookie",
				Schema: &oas.SchemaObject{
					Type:    "string",
					Example: "accessToken=eyJhbGciOi...; Path=/; HttpOnly; Secure; SameSite=Lax",
				},
			},
		},
		{
			name:           "Should parse response header without example",
			comment:        `200 Set-Cookie string "JWT cookie"`,
			wantErr:        false,
			expectedStatus: "200",
			expectedHeader: &oas.HeaderObject{
				Description: "JWT cookie",
				Schema: &oas.SchemaObject{
					Type: "string",
				},
			},
		},
		{
			name:           "Should create ResponseObject if it does not exist yet",
			comment:        `201 X-Request-Id string "Request identifier"`,
			wantErr:        false,
			expectedStatus: "201",
			expectedHeader: &oas.HeaderObject{
				Description: "Request identifier",
				Schema: &oas.SchemaObject{
					Type: "string",
				},
			},
		},
		{
			name:    "Should attach header to existing ResponseObject without overwriting content",
			comment: `200 Set-Cookie string "JWT cookie"`,
			existingResponse: &oas.ResponseObject{
				Description: "Success",
				Content: map[string]*oas.MediaTypeObject{
					oas.ContentTypeJson: {},
				},
			},
			wantErr:        false,
			expectedStatus: "200",
			expectedHeader: &oas.HeaderObject{
				Description: "JWT cookie",
				Schema: &oas.SchemaObject{
					Type: "string",
				},
			},
		},
		{
			name:    "Should return error when status is not an integer",
			comment: `abc Set-Cookie string "JWT cookie"`,
			wantErr: true,
			errMsg:  `parseResponseHeaderComment can not parse comment "abc Set-Cookie string "JWT cookie""`,
		},
		{
			name:    "Should return error when status code is invalid",
			comment: `999 Set-Cookie string "JWT cookie"`,
			wantErr: true,
			errMsg:  "parseResponseHeaderComment: invalid http status code 999",
		},
		{
			name:    "Should return error when comment format is invalid",
			comment: `200 Set-Cookie string`,
			wantErr: true,
			errMsg:  `parseResponseHeaderComment can not parse comment "200 Set-Cookie string"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{}

			operation := &oas.OperationObject{
				Responses: map[string]*oas.ResponseObject{},
			}

			if test.existingResponse != nil {
				operation.Responses[test.expectedStatus] = test.existingResponse
			}

			err := p.parseResponseHeaderComment(operation, test.comment)

			if test.wantErr {
				assert.EqualError(t, err, test.errMsg)
				return
			}

			assert.NoError(t, err)

			responseObj, exists := operation.Responses[test.expectedStatus]
			assert.True(t, exists, "ResponseObject for status %s should exist", test.expectedStatus)

			assert.NotNil(t, responseObj.Headers)

			if test.existingResponse != nil {
				assert.Equal(t, test.existingResponse.Content, responseObj.Content,
					"existing Content should not be overwritten")
			}

			for headerName, headerObj := range responseObj.Headers {
				_ = headerName
				assert.Equal(t, test.expectedHeader, headerObj)
			}
		})
	}
}
