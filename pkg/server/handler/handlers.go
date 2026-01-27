package handler

import (
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/attributes"
	"github.com/rancher/steve/pkg/schema"
)

func k8sAPI(sf schema.Factory, apiOp *types.APIRequest) {
	apiOp.Name = apiOp.Request.PathValue("name")
	apiOp.Type = apiOp.Request.PathValue("type")
	nOrN := apiOp.Request.PathValue("nameorns")

	if nOrN != "" {
		schema := apiOp.Schemas.LookupSchema(apiOp.Type)
		if attributes.Namespaced(schema) {
			apiOp.Namespace = nOrN
		} else {
			apiOp.Name = nOrN
		}
	}

	if namespace := apiOp.Request.PathValue("namespace"); namespace != "" {
		apiOp.Namespace = namespace
	}
}

func apiRoot(sf schema.Factory, apiOp *types.APIRequest) {
	apiOp.Type = "apiRoot"
}
