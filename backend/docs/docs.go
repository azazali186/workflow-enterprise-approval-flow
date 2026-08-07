package docs

import "github.com/swaggo/swag"

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:     "1.0",
	Title:       "ApprovalFlow API",
	Description: "AeroXe ApprovalFlow - Digital approval and workflow automation platform",
	Host:        "localhost:8080",
	BasePath:    "/api/v1",
	Schemes:     []string{"http", "https"},
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
